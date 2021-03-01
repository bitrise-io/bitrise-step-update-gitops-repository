package gitops

import (
	"context"
	"fmt"
	"io/ioutil"
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

	tmpRepoPath string
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
	tmpRepoPath, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp dir for repo: %w", err)
	}
	repo := &gitRepo{
		pr:          p.PullRequestOpener,
		githubRepo:  p.GithubRepo,
		branch:      p.Branch,
		tmpRepoPath: tmpRepoPath,
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
	// Stage all changes, commit them to the current branch
	// and push the commit to the remote repository.
	gitArgs := [][]string{
		{"add", "--all"},
		{"commit", "-m", message},
		{"push", "--all", "-u"},
	}
	for _, a := range gitArgs {
		if _, err := r.git(a...); err != nil {
			return err
		}
	}
	return nil
}

func (r gitRepo) currentBranch() (string, error) {
	branch, err := r.git("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(branch), nil
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
