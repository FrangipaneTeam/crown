package ghclient

import "github.com/google/go-github/v47/github"

// SearchLabelsInIssue searches for labels in the issue.
func (g *GHClient) SearchLabelInIssue(pattern string) ([]*github.LabelResult, error) {
	labels, _, err := g.client.Search.Labels(g.context, *g.repo.ID, pattern, &github.SearchOptions{})
	if err != nil {
		return nil, err
	}

	var found []*github.LabelResult
	for _, label := range labels.Labels {
		if label.GetName() == pattern {
			found = append(found, label)
		}
	}

	return found, nil
}

// ExistsLabel returns true if the label exists.
func (g *GHClient) ExistsLabel(label string) (bool, error) {
	_, _, err := g.client.Issues.GetLabel(g.context, g.repoOwner, g.repoName, label)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetLabel returns the label with the given name.
func (g *GHClient) GetLabel(label string) (x *github.Label, err error) {
	x, _, err = g.client.Issues.GetLabel(g.context, g.repoOwner, g.repoName, label)
	if err != nil {
		return nil, err
	}
	return x, nil
}

// AddLabelToIssue adds a label to the issue.
func (g *GHClient) AddLabelToIssue(label string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(g.context, g.repoOwner, g.repoName, g.issue.GetNumber(), []string{label})
	if err != nil {
		return err
	}
	return nil
}

// AddLabelsToIssue adds labels to the issue.
func (g *GHClient) AddLabelsToIssue(labels []string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(g.context, g.repoOwner, g.repoName, g.issue.GetNumber(), labels)
	if err != nil {
		return err
	}
	return nil
}

// RemoveLabelForIssue removes a label from the issue.
func (g *GHClient) RemoveLabelForIssue(label string) error {
	_, err := g.client.Issues.RemoveLabelForIssue(g.context, g.repoOwner, g.repoName, g.issue.GetNumber(), label)
	if err != nil {
		return err
	}
	g.Logger.Debug().Msgf("Removing %s label", label)
	return nil
}

// RemoveLabelsForIssue removes labels from the issue.
func (g *GHClient) RemoveLabelsForIssue(labels []string) error {
	for _, label := range labels {
		err := g.RemoveLabelForIssue(label)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateLabel creates a label.
func (g *GHClient) CreateLabel(label github.Label) error {
	_, _, err := g.client.Issues.CreateLabel(g.context, g.repoOwner, g.repoName, &label)
	if err != nil {
		return err
	}
	return nil
}

// AddCommentReaction adds a reaction to a comment.
func (g *GHClient) AddCommentReaction(commentID int64, reaction string) error {
	_, _, err := g.client.Reactions.CreateIssueCommentReaction(g.context, g.repoOwner, g.repoName, commentID, reaction)
	if err != nil {
		return err
	}
	return nil
}
