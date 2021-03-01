package gitops

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

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
