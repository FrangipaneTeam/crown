package handlers

import (
	"context"

	"github.com/google/go-github/v47/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
)

const (
	FrangipaneOrg = "FrangipaneTeam"
)

type GHClient struct {
	githubapp.ClientCreator
	Context context.Context
	Client  *github.Client
	Repo    *github.Repository
	Issue   *github.Issue

	RepoOwner     string
	RepoName      string
	FrangipaneOrg string
	Author        string

	logger         zerolog.Logger
	installationID int64
}

func NewGHClient(ctx context.Context, client *github.Client, event interface{}) *GHClient {

	var ghClient *GHClient

	switch event := event.(type) {
	case *github.IssuesEvent:
		ghClient = &GHClient{
			Repo:           event.GetRepo(),
			Issue:          event.GetIssue(),
			installationID: githubapp.GetInstallationIDFromEvent(event),
		}
	default:
		return nil
	}

	ghClient.init()
	ghClient.newContext(ctx)
	ghClient.newGHClient()

	return ghClient
}

func (g *GHClient) init() {
	g.RepoOwner = g.Repo.GetOwner().GetLogin()
	g.RepoName = g.Repo.GetName()
	g.Author = g.Issue.GetUser().GetLogin()
	g.FrangipaneOrg = FrangipaneOrg
}

func (g *GHClient) newContext(ctx context.Context) {
	g.Context, g.logger = githubapp.PrepareRepoContext(ctx, g.installationID, g.Repo)
}

func (g *GHClient) newGHClient() error {
	var err error
	g.Client, err = g.NewInstallationClient(g.installationID)
	if err != nil {
		return err
	}
	return nil
}

// GetRepo returns the repository.
func (g *GHClient) GetRepo() *github.Repository {
	return g.Repo
}

// GetRepoOwner returns the repository owner.
func (g *GHClient) GetRepoOwner() string {
	return g.RepoOwner
}

// GetRepoName returns the repository name.
func (g *GHClient) GetRepoName() string {
	return g.RepoName
}

// GetAuthor returns the author of the issue.
func (g *GHClient) GetAuthor() string {
	return g.Author
}

// IsInFrangipaneOrg returns true if the user is in the FrangipaneTeam organization.
func (g *GHClient) IsInFrangipaneOrg(user string) (bool, error) {
	inOrg, _, err := g.Client.Organizations.IsMember(context.Background(), g.FrangipaneOrg, user)
	if err != nil {
		return false, err
	}

	return inOrg, err
}

// CreateComment creates a comment on the issue.
func (g *GHClient) CreateComment(comment *github.IssueComment) error {
	_, _, err := g.Client.Issues.CreateComment(g.Context, g.RepoOwner, g.RepoName, g.Issue.GetNumber(), comment)
	if err != nil {
		return err
	}

	return nil
}
