package labeler

import "github.com/google/go-github/v47/github"

// LabelCommunity is the label for the community category
const labelCommunity = "Community"

type LabelCommunity string

// GithubLabel returns the github label of the label community
func (c *LabelCommunity) GithubLabel() github.Label {
	return github.Label{
		Name:  github.String(string(*c)),
		Color: github.String("fbca04"),
	}
}

// GetName returns the name of the label
func (c *LabelCommunity) GetName() string {
	return string(*c)
}

func LabelerCommunity() *LabelCommunity {
	x := LabelCommunity(labelCommunity)
	return &x
}
