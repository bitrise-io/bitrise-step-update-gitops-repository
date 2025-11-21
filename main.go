package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bitrise-io/bitrise-step-update-gitops-repository/pkg/gitops"
	"github.com/bitrise-io/go-steputils/stepconf"
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
	stepconf.Print(cfg)

	// Create reference to Github repository.
	ghRepo, err := gitops.NewGithubRepo(
		cfg.DeployRepositoryURL, cfg.DeployUser, cfg.DeployToken)
	if err != nil {
		return fmt.Errorf("new github repository reference: %w", err)
	}

	// Create Github client.
	gh, err := gitops.NewGithubClient(ctx, ghRepo)
	if err != nil {
		return fmt.Errorf("new github client: %w", err)
	}

	// Create local clone of the remote repository.
	localRepo, err := gitops.NewGitRepo(ctx, gitops.NewGitRepoParams{
		PullRequestOpener: gh,
		GithubRepo:        ghRepo,
		Branch:            cfg.DeployBranch,
	})
	if err != nil {
		return fmt.Errorf("new repository: %w", err)
	}
	defer localRepo.Close(ctx)

	var renderer gitops.AllFilesRenderer

	if len(cfg.Deployments) == 0 {
		return fmt.Errorf("no deployments configured")
	}

	if cfg.ReplacerMode {
		renderer = gitops.MultiReplacer{
			Delimiter:       cfg.Delimiter,
			DestinationRepo: localRepo,
			Deployments:     cfg.Deployments,
		}
	} else {
		renderer = gitops.MultiTemplates{
			SourceFolder:    cfg.TemplatesFolder,
			Deployments:     cfg.Deployments,
			DestinationRepo: localRepo,
		}
	}

	// Update files of gitops repository.
	i := gitops.Integration{
		Repo:      localRepo,
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
