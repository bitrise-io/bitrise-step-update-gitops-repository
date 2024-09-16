package gitops

import (
	"fmt"
	"strings"

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
	TemplatesFolder string `env:"templates_folder_path"`
	// DeployToken is the Personal Access Token to interact with Github API.
	DeployToken stepconf.Secret `env:"deploy_token,required"`
	// DeployUser is the username associated with the Personal Access Token.
	DeployUser string `env:"deploy_user"`
	// CommitMessage is the created commit's message.
	CommitMessage string `env:"commit_message,required"`
	// ReplacerMode matches & replaces unknown values by key+delimiter instead of templating
	ReplacerMode bool `env:"replacer_mode,required"`
	// Delimiter indicates the delimiter between key and value in replacer mode
	Delimiter string `env:"delimiter"`
	// RawFiles are unparsed version of `Files` field (to-be-parsed manually).
	RawFiles string `env:"files"`
	// Files are required in replacer mode. List of files to find values in for replacement.
	Files []string
}

func (c config) validate() error {
	if !c.ReplacerMode {
		if len(c.TemplatesFolder) == 0 {
			return requiredError("TemplatesFolder")
		}
		return nil
	}

	if len(c.Delimiter) == 0 {
		return requiredError("Delimiter")
	}
	if len(c.Files) == 0 {
		return requiredError("Files")
	}

	return nil
}

// NewConfig returns a new configuration initialized from environment variables.
func NewConfig() (config, error) {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		return config{}, fmt.Errorf("parse step config: %w", err)
	}

	if err := yaml.Unmarshal([]byte(cfg.RawValues), &cfg.Values); err != nil {
		return config{}, fmt.Errorf("parse values: %w", err)
	}

	if cfg.ReplacerMode {
		files, err := parseStringSlice([]byte(cfg.RawFiles), cfg.Files)
		if err != nil {
			return config{}, fmt.Errorf("parsing files to string slice: %w", err)
		}
		cfg.Files = files
	}

	if err := cfg.validate(); err != nil {
		return config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func requiredError(field string) error {
	return fmt.Errorf("\n- %s: required variable is not present", field)
}

func parseStringSlice(raw []byte, field []string) ([]string, error) {
	if err := yaml.Unmarshal(raw, &field); err != nil {
		return nil, fmt.Errorf("parse string to string slice: %w", err)
	}
	elements := make([]string, 0)
	for _, s := range field {
		elements = append(elements, strings.Fields(s)...)
	}

	return elements, nil
}
