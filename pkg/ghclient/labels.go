package ghclient

import (
	"errors"
	"regexp"

	"github.com/google/go-github/v47/github"
)

/*

	> PULL REQUEST

*/

// SearchLabelsInIssue searches for labels in the issue.
// pattern is a regular expression.
func (g *GHClient) SearchLabelInIssue(pattern string) ([]*github.Label, error) {
	labels, _, err := g.client.Issues.ListLabelsByIssue(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	g.Logger.Debug().Interface("labels", labels).Msgf("Searching for labels matching %s", pattern)

	var foundLabels []*github.Label
	for _, label := range labels {
		if regexp.MustCompile(pattern).MatchString(label.GetName()) {
			foundLabels = append(foundLabels, label)
		}
	}

	return foundLabels, nil

}

// GetLabelsInIssue returns the labels of the issue.
func (g *GHClient) GetLabelsInIssue() ([]*github.Label, error) {
	labels, _, err := g.client.Issues.ListLabelsByIssue(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return labels, nil
}

// ExistsLabel returns true if the label exists.
func (g *GHClient) ExistsLabelInIssue(label string) (bool, error) {
	labels, err := g.GetLabelsInIssue()
	if err != nil {
		return false, err
	}

	for _, l := range labels {
		if l.GetName() == label {
			return true, nil
		}
	}

	return false, nil
}

// GetLabelInIssue returns the label with the given name.
func (g *GHClient) GetLabelInIssue(label string) (x *github.Label, err error) {
	labels, err := g.GetLabelsInIssue()
	if err != nil {
		return nil, err
	}

	for _, l := range labels {
		if l.GetName() == label {
			return l, nil
		}
	}

	return nil, errors.New("label not found")
}

// AddLabelToIssue adds a label to the issue.
func (g *GHClient) AddLabelToIssue(label string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), []string{label})
	if err != nil {
		return err
	}
	return nil
}

// AddLabelsToIssue adds labels to the issue.
func (g *GHClient) AddLabelsToIssue(labels []string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), labels)
	if err != nil {
		return err
	}
	return nil
}

// RemoveLabelForIssue removes a label from the issue.
func (g *GHClient) RemoveLabelForIssue(label string) error {
	_, err := g.client.Issues.RemoveLabelForIssue(g.context, g.repoOwner, g.repoName, g.GetIssueNumber(), label)
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

// GetLabels returns the labels of the repository.
func (g *GHClient) GetLabels() ([]*github.Label, error) {
	labels, _, err := g.client.Issues.ListLabels(g.context, g.repoOwner, g.repoName, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return labels, nil
}

// GetLabel returns the label with the given name of the repository.
func (g *GHClient) GetLabel(label string) (x *github.Label, err error) {
	x, _, err = g.client.Issues.GetLabel(g.context, g.repoOwner, g.repoName, label)
	if err != nil {
		return nil, err
	}
	return x, nil
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
