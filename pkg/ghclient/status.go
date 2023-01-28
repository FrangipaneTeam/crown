package ghclient

import "github.com/google/go-github/v47/github"

// CreateStatus creates a status on the issue.
func (g *GHClient) CreateStatus(status github.RepoStatus, commitSHA string) error {
	g.Logger.Debug().Msgf("Creating status %s with state %s", *status.Context, *status.State)
	_, _, err := g.client.Repositories.CreateStatus(g.context, g.repoOwner, g.repoName, commitSHA, &status)
	return err
}

// ListStatuses lists all statuses on the issue.
func (g *GHClient) ListStatuses(commitSHA string) ([]*github.RepoStatus, error) {
	statuses, _, err := g.client.Repositories.ListStatuses(g.context, g.repoOwner, g.repoName, commitSHA, nil)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

// EditStatus edits a status on the issue.
func (g *GHClient) EditStatus(status github.RepoStatus, commitSHA string) error {
	return g.CreateStatus(status, commitSHA)
}
