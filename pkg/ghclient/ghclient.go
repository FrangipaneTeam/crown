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
	issue     *github.Issue

	repoOwner     string
	repoName      string
	frangipaneOrg string
	author        string

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
		}
	case github.IssueCommentEvent:
		ghClient = &GHClient{
			repo:           event.GetRepo(),
			issue:          event.GetIssue(),
			installationID: githubapp.GetInstallationIDFromEvent(&event),
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
	g.frangipaneOrg = FrangipaneOrg
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

// GetOrg returns the organization.
func (g *GHClient) GetOrg() string {
	return g.frangipaneOrg
}

// IsInFrangipaneOrg returns true if the user is in the FrangipaneTeam organization.
func (g *GHClient) IsInFrangipaneOrg(user string) (bool, error) {
	inOrg, _, err := g.client.Organizations.IsMember(g.context, g.GetOrg(), user)
	if err != nil {
		return false, err
	}

	return inOrg, err
}
