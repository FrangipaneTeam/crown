package ghclient

import (
	"context"
	"errors"

	"github.com/google/go-github/v47/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
)

const (
	FrangipaneOrg = "FrangipaneTeam"
)

type GHClient struct {
	githubapp githubapp.ClientCreator
	context   context.Context
	client    *github.Client
	repo      *github.Repository

	issue       *github.Issue
	pullRequest *github.PullRequest

	repoOwner    string
	repoName     string
	organization string
	author       string
	issueNumber  int

	Logger         zerolog.Logger
	installationID int64
}

func NewGHClient(ctx context.Context, ghapp githubapp.ClientCreator, event interface{}) (*GHClient, error) {

	var ghClient *GHClient

	switch event := event.(type) {
	case github.IssuesEvent:
		ghClient = &GHClient{
			repo:           event.GetRepo(),
			issue:          event.GetIssue(),
			installationID: githubapp.GetInstallationIDFromEvent(&event),
			issueNumber:    event.GetIssue().GetNumber(),
		}
	case github.IssueCommentEvent:
		ghClient = &GHClient{
			repo:           event.GetRepo(),
			issue:          event.GetIssue(),
			installationID: githubapp.GetInstallationIDFromEvent(&event),
			issueNumber:    event.GetIssue().GetNumber(),
		}
	case github.PullRequestEvent:
		ghClient = &GHClient{
			repo:           event.GetRepo(),
			pullRequest:    event.GetPullRequest(),
			installationID: githubapp.GetInstallationIDFromEvent(&event),
			issueNumber:    event.GetPullRequest().GetNumber(),
		}
	default:
		return nil, errors.New("event not supported")
	}

	if ghapp == nil {
		return nil, errors.New("githubapp is nil")
	} else {
		ghClient.githubapp = ghapp
	}

	ghClient.init()
	ghClient.newContext(ctx)
	err := ghClient.newGHClient()
	if err != nil {
		return nil, err
	}

	return ghClient, nil
}

func (g *GHClient) init() {
	g.repoOwner = g.repo.GetOwner().GetLogin()
	g.repoName = g.repo.GetName()
	g.author = g.issue.GetUser().GetLogin()
	g.organization = g.repoOwner
}

func (g *GHClient) newContext(ctx context.Context) {
	g.context, g.Logger = githubapp.PrepareRepoContext(ctx, g.installationID, g.repo)
}

func (g *GHClient) newGHClient() error {
	var err error
	x, err := g.githubapp.NewInstallationClient(g.installationID)
	if err != nil {
		return err
	}
	g.client = x
	return nil
}

// GetRepo returns the repository.
func (g *GHClient) GetRepo() *github.Repository {
	return g.repo
}

// GetRepoOwner returns the repository owner.
func (g *GHClient) GetRepoOwner() string {
	return g.repoOwner
}

// GetRepoName returns the repository name.
func (g *GHClient) GetRepoName() string {
	return g.repoName
}

// GetAuthor returns the author of the issue.
func (g *GHClient) GetAuthor() string {
	return g.author
}

// GetIssue returns the issue.
func (g *GHClient) GetIssue() *github.Issue {
	return g.issue
}

// GetPullRequest returns the pull request.
func (g *GHClient) GetPullRequest() *github.PullRequest {
	return g.pullRequest
}

// GetOrg returns the organization.
func (g *GHClient) GetOrg() string {
	return g.organization
}

// GetIssueNumber returns the issue number.
func (g *GHClient) GetIssueNumber() int {
	return g.issueNumber
}

// GetInstallationID returns the installation ID.
func (g *GHClient) GetInstallationID() int64 {
	return g.installationID
}

// IsInOrganization returns true if the user is in the organization.
func (g *GHClient) IsInOrganization(user string) (bool, error) {
	inOrg, _, err := g.client.Organizations.IsMember(g.context, g.GetOrg(), user)
	if err != nil {
		return false, err
	}

	return inOrg, err
}
