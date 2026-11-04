package gitops

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gitRepoCases = map[string]struct {
	upstreamBranch string
	repoURL        string
}{
	"master of testhub.assert/test/foo": {
		upstreamBranch: "master",
		repoURL:        "testhub.assert/test/foo",
	},
	"staging of testhub.assert/test/bar": {
		upstreamBranch: "staging",
		repoURL:        "testhub.assert/test/bar",
	},
}

func TestGitRepo(t *testing.T) {
	ctx := context.Background()

	for name, tc := range gitRepoCases {
		t.Run(name, func(t *testing.T) {
			upstreamPath, close := localUpstreamRepo(t, tc.upstreamBranch)
			defer close()

			// Initialize mock Github client.
			wantPullRequestURL := fmt.Sprintf("https://%s/pr/15", tc.repoURL)
			var gotHead, gotBase string
			pr := &pullRequestOpenerMock{
				OpenPullRequestFunc: func(_ context.Context, p openPullRequestParams) (string, error) {
					gotHead = p.head
					gotBase = p.base
					return wantPullRequestURL, nil
				},
			}

			repo, err := NewGitRepo(ctx, NewGitRepoParams{
				PullRequestOpener: pr,
				GithubRepo: &githubRepo{
					url: stepconf.Secret(upstreamPath),
				},
				Branch: tc.upstreamBranch,
			})
			t.Run("create new local repository clone", func(t *testing.T) {
				require.NoError(t, err, "newRepository")
			})

			t.Run("repository is clean without changes", func(t *testing.T) {
				clean, err := repo.workingDirectoryClean()
				require.NoError(t, err)
				require.True(t, clean)
			})

			t.Run("repository is dirty after changes", func(t *testing.T) {
				changePath := path.Join(repo.localPath(), "empty.go")
				write(t, changePath, "package empty")

				clean, err := repo.workingDirectoryClean()
				require.NoError(t, err)
				require.False(t, clean)
			})

			t.Run("commit and push changes to upstream", func(t *testing.T) {
				err := repo.gitCommitAndPush("test commit")
				require.NoError(t, err, "commit and push")

				clean, err := repo.workingDirectoryClean()
				require.NoError(t, err, "working directory clean")
				require.True(t, clean)
			})

			t.Run("open pull request from a new branch", func(t *testing.T) {
				require.NoError(t, repo.gitCheckoutNewBranch(), "new branch")
				changePath := path.Join(repo.localPath(), "another.go")
				write(t, changePath, "package another")

				err := repo.gitCommitAndPush("another commit")
				require.NoError(t, err, "commit and push another")

				gotPullRequestURL, err := repo.openPullRequest(ctx, "", "")
				require.NoError(t, err, "open pull request")
				assert.Equal(t, wantPullRequestURL, gotPullRequestURL, "pr url")

				assert.Equal(t, tc.upstreamBranch, gotBase, "pr base")

				assert.NotEqual(t, gotBase, gotHead, "pr head != base")
				wantHead, err := repo.currentBranch()
				require.NoError(t, err, "current branch")
				assert.Equal(t, wantHead, gotHead, "pr head = current branch")
			})
		})
	}
}

func TestGitCommitAndPushOnlyPushesCurrentBranch(t *testing.T) {
	ctx := context.Background()
	const baseBranch = "staging"

	newClone := func(t *testing.T, upstreamPath string) *gitRepo {
		repo, err := NewGitRepo(ctx, NewGitRepoParams{
			GithubRepo: &githubRepo{url: stepconf.Secret(upstreamPath)},
			Branch:     baseBranch,
		})
		require.NoError(t, err, "new git repo")
		t.Cleanup(func() { repo.Close(ctx) })
		return repo
	}

	t.Run("other local branches are not pushed", func(t *testing.T) {
		upstreamPath, close := localUpstreamRepo(t, baseBranch)
		defer close()
		repo := newClone(t, upstreamPath)

		git(t, repo.localPath(), "branch", "unrelated")
		write(t, path.Join(repo.localPath(), "change.go"), "package change")
		require.NoError(t, repo.gitCommitAndPush("direct commit"), "commit and push")

		assert.ElementsMatch(t, []string{baseBranch}, branches(t, upstreamPath), "upstream branches")
		assert.Equal(t,
			gitOutput(t, repo.localPath(), "rev-parse", "HEAD"),
			gitOutput(t, upstreamPath, "rev-parse", "refs/heads/"+baseBranch),
			"base branch updated")
	})

	t.Run("stale base branch is not pushed in pull request mode", func(t *testing.T) {
		upstreamPath, close := localUpstreamRepo(t, baseBranch)
		defer close()
		repo := newClone(t, upstreamPath)

		// Another run updates the base branch after this clone was made,
		// leaving the local copy of the base branch behind the remote.
		write(t, path.Join(upstreamPath, "concurrent.go"), "package concurrent")
		git(t, upstreamPath, "add", "--all")
		git(t, upstreamPath, "commit", "-m", "concurrent commit")
		concurrentHead := gitOutput(t, upstreamPath, "rev-parse", "HEAD")

		require.NoError(t, repo.gitCheckoutNewBranch(), "new branch")
		write(t, path.Join(repo.localPath(), "change.go"), "package change")
		require.NoError(t, repo.gitCommitAndPush("pull request commit"), "commit and push")

		head, err := repo.currentBranch()
		require.NoError(t, err, "current branch")
		assert.ElementsMatch(t, []string{baseBranch, head}, branches(t, upstreamPath), "upstream branches")
		assert.Equal(t,
			concurrentHead,
			gitOutput(t, upstreamPath, "rev-parse", "refs/heads/"+baseBranch),
			"base branch untouched")
		assert.Equal(t,
			gitOutput(t, repo.localPath(), "rev-parse", "HEAD"),
			gitOutput(t, upstreamPath, "rev-parse", "refs/heads/"+head),
			"head branch pushed")
	})

	t.Run("branch sharing its name with a tag is pushed", func(t *testing.T) {
		upstreamPath, close := localUpstreamRepo(t, baseBranch)
		defer close()
		git(t, upstreamPath, "tag", baseBranch)
		taggedCommit := gitOutput(t, upstreamPath, "rev-parse", "refs/tags/"+baseBranch)
		repo := newClone(t, upstreamPath)
		require.Equal(t, baseBranch,
			gitOutput(t, repo.localPath(), "tag", "--list", baseBranch),
			"clone has the same-named tag")

		branch, err := repo.currentBranch()
		require.NoError(t, err, "current branch")
		assert.Equal(t, baseBranch, branch, "current branch")

		write(t, path.Join(repo.localPath(), "change.go"), "package change")
		require.NoError(t, repo.gitCommitAndPush("direct commit"), "commit and push")

		assert.ElementsMatch(t, []string{baseBranch}, branches(t, upstreamPath), "upstream branches")
		assert.Equal(t,
			gitOutput(t, repo.localPath(), "rev-parse", "HEAD"),
			gitOutput(t, upstreamPath, "rev-parse", "refs/heads/"+baseBranch),
			"base branch updated")
		assert.Equal(t,
			taggedCommit,
			gitOutput(t, upstreamPath, "rev-parse", "refs/tags/"+baseBranch),
			"tag untouched")
	})

	t.Run("detached HEAD is rejected", func(t *testing.T) {
		upstreamPath, close := localUpstreamRepo(t, baseBranch)
		defer close()
		repo := newClone(t, upstreamPath)

		git(t, repo.localPath(), "checkout", "--detach")
		write(t, path.Join(repo.localPath(), "change.go"), "package change")
		require.Error(t, repo.gitCommitAndPush("detached commit"), "commit and push")
		assert.ElementsMatch(t, []string{baseBranch}, branches(t, upstreamPath), "upstream branches")
	})
}

func TestGitCommitAndPushRetriesTransientRejections(t *testing.T) {
	ctx := context.Background()
	const branch = "staging"
	retryDelays := []time.Duration{time.Second, 2 * time.Second}

	newClone := func(t *testing.T, upstreamPath string) (*gitRepo, *[]time.Duration) {
		repo, err := NewGitRepo(ctx, NewGitRepoParams{
			GithubRepo: &githubRepo{url: stepconf.Secret(upstreamPath)},
			Branch:     branch,
		})
		require.NoError(t, err, "new git repo")
		t.Cleanup(func() { repo.Close(ctx) })
		var slept []time.Duration
		repo.pushRetryDelays = retryDelays
		repo.sleep = func(d time.Duration) { slept = append(slept, d) }
		return repo, &slept
	}

	cases := map[string]struct {
		rejections   int    // pushes the remote rejects before accepting one
		message      string // what the remote reports when rejecting
		wantAttempts int
		wantSlept    []time.Duration
		wantErr      bool
	}{
		"internal server error is retried": {
			rejections:   1,
			message:      "fatal error in commit_refs",
			wantAttempts: 2,
			wantSlept:    retryDelays[:1],
		},
		"protected branch rejection is retried": {
			rejections:   1,
			message:      "error: GH006: Protected branch update failed for refs/heads/staging.",
			wantAttempts: 2,
			wantSlept:    retryDelays[:1],
		},
		"retries are bounded": {
			rejections:   10,
			message:      "fatal error in commit_refs",
			wantAttempts: len(retryDelays) + 1,
			wantSlept:    retryDelays,
			wantErr:      true,
		},
		"other rejections are not retried": {
			rejections:   10,
			message:      "error: push declined by policy",
			wantAttempts: 1,
			wantErr:      true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			upstreamPath, close := localUpstreamRepo(t, branch)
			defer close()
			attempts := rejectPushes(t, upstreamPath, tc.rejections, tc.message)
			repo, slept := newClone(t, upstreamPath)

			write(t, path.Join(repo.localPath(), "change.go"), "package change")
			err := repo.gitCommitAndPush("commit")

			assert.Equal(t, tc.wantAttempts, attempts(), "push attempts")
			assert.Equal(t, tc.wantSlept, *slept, "waits between attempts")
			if tc.wantErr {
				require.Error(t, err, "commit and push")
				assert.Contains(t, err.Error(), tc.message, "error reports the rejection")
				return
			}
			require.NoError(t, err, "commit and push")
			assert.Equal(t,
				gitOutput(t, repo.localPath(), "rev-parse", "HEAD"),
				gitOutput(t, upstreamPath, "rev-parse", "refs/heads/"+branch),
				"branch updated")
		})
	}

	t.Run("non-fast-forward is not retried", func(t *testing.T) {
		upstreamPath, close := localUpstreamRepo(t, branch)
		defer close()
		attempts := rejectPushes(t, upstreamPath, 0, "")
		repo, slept := newClone(t, upstreamPath)

		// The branch moves on the remote after the clone was made.
		write(t, path.Join(upstreamPath, "concurrent.go"), "package concurrent")
		git(t, upstreamPath, "add", "--all")
		git(t, upstreamPath, "commit", "-m", "concurrent commit")

		write(t, path.Join(repo.localPath(), "change.go"), "package change")
		require.Error(t, repo.gitCommitAndPush("commit"), "commit and push")
		assert.Empty(t, *slept, "waits between attempts")
		assert.Zero(t, attempts(), "pushes reaching the remote")
	})
}

// rejectPushes makes the upstream repository reject its first n pushes,
// reporting message to the pushing client. The returned function tells how
// many pushes have reached the upstream repository so far.
func rejectPushes(t *testing.T, upstreamPath string, n int, message string) func() int {
	hooksDir := t.TempDir()
	counterPath := path.Join(t.TempDir(), "attempts")
	hook := fmt.Sprintf(`#!/bin/sh
n=$(cat %[1]q 2>/dev/null || echo 0)
n=$((n + 1))
echo "$n" > %[1]q
if [ "$n" -le %[2]d ]; then
	echo %[3]q >&2
	exit 1
fi
`, counterPath, n, message)
	require.NoError(t, os.WriteFile(path.Join(hooksDir, "pre-receive"), []byte(hook), 0700), "write hook")
	git(t, upstreamPath, "config", "core.hooksPath", hooksDir)

	return func() int {
		out, err := os.ReadFile(counterPath)
		if os.IsNotExist(err) {
			return 0
		}
		require.NoError(t, err, "read push attempts")
		attempts, err := strconv.Atoi(strings.TrimSpace(string(out)))
		require.NoError(t, err, "parse push attempts")
		return attempts
	}
}

func localUpstreamRepo(t *testing.T, branch string) (string, func()) {
	repoPath, err := ioutil.TempDir("", "")
	require.NoError(t, err, "new temp directory for local upstream")
	readmePath := path.Join(repoPath, "README.md")

	git(t, repoPath, "init")
	git(t, repoPath, "checkout", "-b", branch)
	write(t, readmePath, "A local upstream repository for testing.")
	git(t, repoPath, "add", "--all")
	git(t, repoPath, "commit", "-m", "initial commit")
	// allow push from another git repository
	git(t, repoPath, "config", "receive.denyCurrentBranch", "ignore")

	return repoPath, func() {
		os.RemoveAll(repoPath)
	}
}

func write(t *testing.T, path, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err, "ioutil.WriteFile(%s)", path)
}

func git(t *testing.T, repoPath string, args ...string) {
	// Change current directory to the repositorys local clone.
	originalDir, err := os.Getwd()
	require.NoError(t, err, "get current dir")
	require.NoError(t, os.Chdir(repoPath), "change to upstream repo")
	// Defer a revert of the current directory to the original one.
	defer func() {
		require.NoError(t, os.Chdir(originalDir), "change to original dir")
	}()

	cmd := exec.Command("git", args...)
	require.NoError(t, cmd.Run(), "git %+v", args)
}

func gitOutput(t *testing.T, repoPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	require.NoError(t, err, "git %+v", args)
	return strings.TrimSpace(string(out))
}

func branches(t *testing.T, repoPath string) []string {
	return strings.Fields(gitOutput(t, repoPath,
		"for-each-ref", "--format=%(refname:lstrip=2)", "refs/heads/"))
}
