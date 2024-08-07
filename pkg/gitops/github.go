package gitops

import (
	"context"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	gogh "github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

//go:generate moq -out github_moq_test.go . pullRequestOpener
type pullRequestOpener interface {
	OpenPullRequest(context.Context, openPullRequestParams) (string, error)
}

// githubClient implements the pullRequestOpener interface.
var _ pullRequestOpener = (*githubClient)(nil)

type githubClient struct {
	client *gogh.Client
	repo   *githubRepo
}

// NewGithubClient returns a new Github client to interact with a given repository.
func NewGithubClient(ctx context.Context, repo *githubRepo) (*githubClient, error) {
	// Initialize client for Github API.
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(repo.token)},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSource)
	ghClient := gogh.NewClient(tokenClient)
	return &githubClient{
		client: ghClient,
		repo:   repo,
	}, nil
}

type openPullRequestParams struct {
	title string
	body  string
	head  string
	base  string
}

func (gh githubClient) OpenPullRequest(ctx context.Context, p openPullRequestParams) (string, error) {
	// Title is required for PRs. Generate one if it's omitted.
	if p.title == "" {
		p.title = "Merge " + p.head
	}

	if len(p.title) > 255 {
		p.title = p.title[:255]
	}

	req := &gogh.NewPullRequest{
		Title: gogh.String(p.title),
		Body:  gogh.String(p.body),
		Head:  gogh.String(p.head),
		Base:  gogh.String(p.base),
	}
	pr, _, err := gh.client.PullRequests.Create(ctx, gh.repo.owner, gh.repo.name, req)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	return *pr.HTMLURL, nil
}

// githubRepo represents a Github repository.
type githubRepo struct {
	url   stepconf.Secret
	token stepconf.Secret
	owner string
	name  string
}

// NewGithubRepo return a new Github repository.
func NewGithubRepo(url, user string, token stepconf.Secret) (*githubRepo, error) {
	// Trim prefix.
	prefix := "https://github.com/"
	if !strings.HasPrefix(url, prefix) {
		return nil, fmt.Errorf("must start with %q", prefix)
	}
	url = strings.TrimPrefix(url, prefix)

	// Trim suffix.
	suffix := ".git"
	if !strings.HasSuffix(url, suffix) {
		return nil, fmt.Errorf("must end with %q", suffix)
	}
	url = strings.TrimSuffix(url, suffix)

	// Split remaining URL for owner and repository name.
	a := strings.Split(url, "/")
	if len(a) != 2 {
		return nil, fmt.Errorf("must separate owner from repo with one /")
	}

	// Construct an authenticated HTTPS URL from plain URL and credentials.
	authURL := fmt.Sprintf(
		"https://%s:%s@github.com/%s/%s.git", user, string(token), a[0], a[1])

	return &githubRepo{
		url:   stepconf.Secret(authURL),
		token: token,
		owner: a[0],
		name:  a[1],
	}, nil
}
