package gitops

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:generate moq -out local_repository_moq_test.go . localRepository
type localRepository interface {
	Close(ctx context.Context)
	localPath() string
	gitClone() error
	workingDirectoryClean() (bool, error)
	gitCheckoutNewBranch() error
	gitCommitAndPush(message string) error
	openPullRequest(ctx context.Context, title, body string) (string, error)
}

// gitRepo implements the localRepository interface.
var _ localRepository = (*gitRepo)(nil)

type gitRepo struct {
	pr         pullRequestOpener
	githubRepo *githubRepo
	branch     string

	// pushRetryDelays are the waits before each retry of a push that failed
	// with one of transientPushErrors; their count bounds the retries.
	pushRetryDelays []time.Duration
	sleep           func(time.Duration)

	tmpRepoPath string
}

// defaultPushRetryDelays allow four push attempts within about 15 seconds.
var defaultPushRetryDelays = []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}

// transientPushErrors are server-side push rejections that GitHub has
// returned while another push to the same repository was in progress, and
// that did not recur when the run was repeated.
//
// A genuine branch protection violation also reports GH006; retrying it only
// delays the same error. A retry can never overwrite someone else's commit:
// the push is not forced, so if the branch really moved on the remote, the
// next attempt is rejected as non-fast-forward, which is not retried.
var transientPushErrors = []string{
	"fatal error in commit_refs",
	"GH006: Protected branch update failed",
}

// NewGitRepoParams are parameters for NewGitRepo function.
type NewGitRepoParams struct {
	PullRequestOpener pullRequestOpener
	GithubRepo        *githubRepo
	Branch            string
}

// NewGitRepo returns a new local clone of a remote repository.
// It should be closed after usage.
func NewGitRepo(ctx context.Context, p NewGitRepoParams) (*gitRepo, error) {
	// Temporary directory for local clone of repository.
	tmpRepoPath, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp dir for repo: %w", err)
	}
	repo := &gitRepo{
		pr:          p.PullRequestOpener,
		githubRepo:  p.GithubRepo,
		branch:      p.Branch,
		tmpRepoPath: tmpRepoPath,

		pushRetryDelays: defaultPushRetryDelays,
		sleep:           time.Sleep,
	}
	if err := repo.gitClone(); err != nil {
		return nil, fmt.Errorf("git clone repo: %w", err)
	}
	return repo, nil
}

// Close closes all related resoruces of the repository.
// This is a best-effort operation, possible errors are logged as warning,
// not returned as an actual error.
func (r gitRepo) Close(ctx context.Context) {
	// Delete temporary repository from the local filesystem.
	if err := os.RemoveAll(r.tmpRepoPath); err != nil {
		log.Printf("warning: remove temporary repository: %s\n", err)
	}
}

func (r gitRepo) localPath() string {
	return r.tmpRepoPath
}

func (r gitRepo) gitClone() error {
	_, err := r.git("clone",
		"--branch", r.branch, "--single-branch",
		string(r.githubRepo.url), ".")
	return err
}

func (r gitRepo) workingDirectoryClean() (bool, error) {
	status, err := r.git("status")
	if err != nil {
		return false, err
	}
	return strings.Contains(status, "nothing to commit"), nil
}

func (r gitRepo) gitCheckoutNewBranch() error {
	// Generate branch name based on the current time.
	t := time.Now()
	branch := fmt.Sprintf("ci-%d-%02d-%02dT%02d-%02d-%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	// Execute git checkout to a new branch with that name.
	if _, err := r.git("checkout", "-b", branch); err != nil {
		return fmt.Errorf("checkout new branch %q: %w", branch, err)
	}
	return nil
}

func (r gitRepo) gitCommitAndPush(message string) error {
	// The commit lands on the checked out branch: the deploy branch itself,
	// or the freshly created head branch in pull request mode.
	branch, err := r.currentBranch()
	if err != nil {
		return fmt.Errorf("current branch: %w", err)
	}

	// Stage all changes, commit them to the current branch
	// and push only that branch to the remote repository.
	// In pull request mode the local deploy branch is left as it was cloned;
	// if it has moved on the remote since, pushing it as well would make the
	// push fail even though the pull request branch itself was pushed.
	gitArgs := [][]string{
		{"add", "--all"},
		{"commit", "-m", message},
	}
	for _, a := range gitArgs {
		if _, err := r.git(a...); err != nil {
			return err
		}
	}
	return r.gitPush(branch)
}

func (r gitRepo) gitPush(branch string) error {
	for attempt := 0; ; attempt++ {
		_, err := r.git("push", "origin", pushRefspec(branch))
		if err == nil || attempt >= len(r.pushRetryDelays) || !isTransientPushError(err) {
			return err
		}
		delay := r.pushRetryDelays[attempt]
		log.Printf("warning: push rejected by the remote, retrying in %s: %s\n", delay, err)
		r.sleep(delay)
	}
}

func isTransientPushError(err error) bool {
	for _, s := range transientPushErrors {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}
	return false
}

// pushRefspec returns a refspec that updates exactly the given branch on the
// remote from the local branch of the same name. Fully qualified refs are used
// so a tag sharing the branch's name cannot make the refspec ambiguous.
func pushRefspec(branch string) string {
	ref := "refs/heads/" + branch
	return ref + ":" + ref
}

// currentBranch returns the name of the checked out branch.
// It fails when HEAD is detached.
func (r gitRepo) currentBranch() (string, error) {
	// The full ref is read and its prefix stripped here, because git's own
	// abbreviation (rev-parse --abbrev-ref, symbolic-ref --short) yields
	// "heads/<branch>" when a tag has the same name as the branch.
	out, err := r.git("symbolic-ref", "HEAD")
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(out)
	branch, ok := strings.CutPrefix(ref, "refs/heads/")
	if !ok {
		return "", fmt.Errorf("HEAD points to %q, not a local branch", ref)
	}
	return branch, nil
}

func (r gitRepo) git(args ...string) (string, error) {
	// Change current directory to the repositorys local clone.
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current dir: %w", err)
	}
	if err := os.Chdir(r.tmpRepoPath); err != nil {
		return "", fmt.Errorf("change dir to %q: %w", r.tmpRepoPath, err)
	}

	cmd := exec.Command("git", args...)
	// Run git command and returns its combined output of stdout and stderr.
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errChdir := os.Chdir(originalDir); errChdir != nil {
			err = fmt.Errorf("%w (revert to original dir: %s)", err, errChdir)
		}
		return "", fmt.Errorf("run command %v: %w (output: %s)", args, err, out)
	}
	if err := os.Chdir(originalDir); err != nil {
		return "", fmt.Errorf("revert to original dir: %w", err)
	}
	return string(out), nil
}

func (r gitRepo) openPullRequest(ctx context.Context, title, body string) (string, error) {
	// PR will be open from the current branch.
	currBranch, err := r.currentBranch()
	if err != nil {
		return "", fmt.Errorf("current branch: %w", err)
	}
	// Open pull request from current branch to the base branch.
	url, err := r.pr.OpenPullRequest(ctx, openPullRequestParams{
		title: title,
		body:  body,
		head:  currBranch,
		base:  r.branch,
	})
	if err != nil {
		return "", fmt.Errorf("call github: %w", err)
	}
	return url, nil
}
