package ghclient

import "github.com/google/go-github/v47/github"

// GetCommits returns the commits of the pull request.
func (g *GHClient) GetCommits() ([]*github.RepositoryCommit, error) {
	commits, _, err := g.client.PullRequests.ListCommits(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), nil)
	if err != nil {
		return nil, err
	}

	return commits, nil
}
