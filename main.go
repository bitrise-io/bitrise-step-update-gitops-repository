package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bitrise-io/bitrise-step-update-gitops-repository/pkg/gitops"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Read gitops related config from environment.
	cfg, err := gitops.NewConfig()
	if err != nil {
		return fmt.Errorf("new gitops config: %w", err)
	}

	// Create Github client.
	gh, err := gitops.NewGithub(ctx, cfg.DeployRepositoryURL, cfg.DeployToken)
	if err != nil {
		return fmt.Errorf("new github client: %w", err)
	}

	// Temporary SSH key (used by git commands).
	sshKey, err := gitops.NewSSHKey(ctx, gh)
	if err != nil {
		return fmt.Errorf("new temporary ssh key: %w", err)
	}

	// Create local clone of the remote repository.
	repo, err := gitops.NewRepository(ctx, gitops.NewRepositoryParams{
		Github: gh,
		SSHKey: sshKey,
		Remote: gitops.RemoteConfig{
			URL:    cfg.DeployRepositoryURL,
			Branch: cfg.DeployBranch,
		},
	})
	if err != nil {
		sshKey.Close(ctx) // repo initialization failed, it cannot close it
		return fmt.Errorf("new repository: %w", err)
	}
	defer repo.Close(ctx)

	// Create templates renderer.
	renderer := gitops.Templates{
		SourceFolder:      cfg.TemplatesFolder,
		Values:            cfg.Values,
		DestinationRepo:   repo,
		DestinationFolder: cfg.DeployFolder,
	}

	// Update files of gitops repository.
	i := gitops.Integration{
		Repo:      repo,
		ExportEnv: gitops.EnvmanExport,
		Renderer:  renderer,
	}
	if err := i.UpdateFiles(ctx, gitops.UpdateFilesParams{
		PullRequest:      cfg.PullRequest,
		PullRequestTitle: cfg.PullRequestTitle,
		PullRequestBody:  cfg.PullRequestBody,
		CommitMessage:    cfg.CommitMessage,
	}); err != nil {
		return fmt.Errorf("update files in gitops repo: %w", err)
	}
	return nil
}
