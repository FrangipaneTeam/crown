package ghclient

import "github.com/google/go-github/v47/github"

// CreateComment creates a comment on the issue.
func (g *GHClient) CreateComment(comment github.IssueComment) error {
	_, _, err := g.client.Issues.CreateComment(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), &comment)
	if err != nil {
		return err
	}

	return nil
}

// ListComments lists all comments on the issue.
func (g *GHClient) ListComments() ([]*github.IssueComment, error) {
	comments, _, err := g.client.Issues.ListComments(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), nil)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// DeleteComment deletes a comment on the issue.
func (g *GHClient) DeleteComment(commentID int64) error {
	_, err := g.client.Issues.DeleteComment(g.context, g.repoOwner, g.repoName, commentID)
	if err != nil {
		return err
	}

	return nil
}

// EditComment edits a comment on the issue.
func (g *GHClient) EditComment(commentID int64, comment github.IssueComment) error {

	_, _, err := g.client.Issues.EditComment(g.context, g.repoOwner, g.repoName, commentID, &comment)
	if err != nil {
		return err
	}

	return nil
}
