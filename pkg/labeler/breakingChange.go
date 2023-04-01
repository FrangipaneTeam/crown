package labeler

import "github.com/google/go-github/v47/github"

const (
	breakingChangeLongName                       = "BreakingChange"
	BreakingChange         LabelerBreakingChange = 666
)

type LabelerBreakingChange int //nolint:revive

// GetGithubLabel returns the github label of the breaking change.
func (c LabelerBreakingChange) GithubLabel() github.Label {
	return github.Label{
		Name:  github.String(breakingChangeLongName),
		Color: github.String("ff0000"),
	}
}

// GetLongName returns the long name of the breaking change.
func (c LabelerBreakingChange) GetLongName() string {
	return breakingChangeLongName
}
