package gitops

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/stepconf"
	"gopkg.in/yaml.v3"
)

type config struct {
	// DeployRepositoryURL is the URL of the deployment (GitOps) repository.
	DeployRepositoryURL string `env:"deploy_repository_url,required"`
	// DeployFolder is the folder to render templates to in the deploy repository.
	DeployFolder string `env:"deploy_path,required"`
	// DeployBranch is the branch to render templates to in the deploy repository.
	DeployBranch string `env:"deploy_branch,required"`
	// PullRequest won't push to the branch. It will open a PR only instead.
	PullRequest bool `env:"pull_request"`
	// PullRequestTitle is the title of the opened pull request.
	PullRequestTitle string `env:"pull_request_title"`
	// PullRequestBody is the body of the opened pull request.
	PullRequestBody string `env:"pull_request_body"`
	// RawValues are unparsed version of `Values` field (to-be-parsed manually).
	RawValues string `env:"values"`
	// Values are values applied to the template files.
	Values map[string]string
	// TemplatesFolder is the path to the deployment templates folder.
	TemplatesFolder string `env:"templates_folder_path,dir"`
	// DeployToken is the Personal Access Token to interact with Github API.
	DeployToken stepconf.Secret `env:"deploy_token,required"`
	// DeployUser is the username associated with the Personal Access Token.
	DeployUser string `env:"deploy_user"`
	// CommitMessage is the created commit's message.
	CommitMessage string `env:"commit_message,required"`
}

// NewConfig returns a new configuration initialized from environment variables.
func NewConfig() (config, error) {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		return config{}, fmt.Errorf("parse step config: %w", err)
	}

	return cfg, yaml.Unmarshal([]byte(cfg.RawValues), &cfg.Values)
}
