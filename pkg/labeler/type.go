package labeler

import (
	"github.com/google/go-github/v47/github"

	"github.com/FrangipaneTeam/crown/pkg/conventionalcommit"
)

const (
	// ! Always add new IDs at the END of the list.
	Feature LabelerType = 123 << iota
	Fix
	Refactor
	Docs
	Chore
	Style
	Perf
	Test
	CI
	// ! Always add new IDs at the END of the list.
)

var LabelsType = map[LabelerType]labelType{
	Feature: {
		shortName: "feat",
		longName:  "Feature",
	},
	Fix: {
		shortName: "fix",
		longName:  "Fix",
	},
	Refactor: {
		shortName: "refactor",
		longName:  "Refactor",
	},
	Docs: {
		shortName: "docs",
		longName:  "Docs",
	},
	Chore: {
		shortName: "chore",
		longName:  "Chore",
	},
	Style: {
		shortName: "style",
		longName:  "Style",
	},
	Perf: {
		shortName: "perf",
		longName:  "Perf",
	},
	Test: {
		shortName: "test",
		longName:  "Test",
	},
	CI: {
		shortName: "ci",
		longName:  "CI",
	},
}

type labelType struct {
	shortName string
	longName  string
	color     string
}

type LabelerType int //nolint:revive

// FincLabelerType returns the LabelerType of the label.
func FindLabelerType(label *conventionalcommit.Cc) (LabelerType, bool) {
	for k, v := range LabelsType {
		if v.shortName == label.Type() {
			return k, true
		}
	}

	return 0, false
}

// GetShortName returns the short name of the label.
func (c LabelerType) GetShortName() string {
	return LabelsType[c].shortName
}

// GetShortNameP returns the short name of the label as a pointer.
func (c LabelerType) GetShortNameP() *string {
	return github.String(LabelsType[c].shortName)
}

// GetLongName returns the long name of the label.
func (c LabelerType) GetLongName() string {
	return LabelsType[c].longName
}

// GetLongNameP returns the long name of the label as a pointer.
func (c LabelerType) GetLongNameP() *string {
	return github.String(LabelsType[c].longName)
}

// GetColor returns the color of the label.
func (c LabelerType) GetColor() string {
	return LabelsType[c].color
}

// GetColorP returns the color of the label as a pointer.
func (c LabelerType) GetColorP() *string {
	return github.String(LabelsType[c].color)
}

// IsValid returns true if the string passed in is equal to LabelerType.
func (c LabelerType) IsValid(id string) bool {
	return c.GetShortName() == id
}

// GitHubLabel returns the label of the LabelerType.
func (c LabelerType) GitHubLabel() github.Label {
	x := github.Label{
		Name: c.GetLongNameP(),
	}

	if c.GetColor() != "" {
		x.Color = c.GetColorP()
	}

	return x
}
