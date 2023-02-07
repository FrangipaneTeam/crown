package labeler

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v47/github"
)

const (
	prefixScope = "category"
)

type LabelerScope scope

type scope struct {
	scope    string
	longName string
	color    string
}

// LabelScope return LabelerScope from scope
func LabelScope(scope string) *LabelerScope {
	// if string scope already contain category/ prefix, remove it
	if strings.HasPrefix(scope, prefixScope) {
		scope = strings.TrimPrefix(scope, prefixScope+"/")
	}

	return &LabelerScope{
		scope:    scope,
		longName: formatedLabelScope(scope),
		color:    "BFD4F2",
	}
}

// FormatedLabelScope returns the label in the form of "category/<scope>".
func formatedLabelScope(scope string) string {
	return fmt.Sprintf("%s/%s", prefixScope, scope)
}

// GetLongName returns the long name of the label
func (c *LabelerScope) GetLongName() string {
	return c.longName
}

// GithubLabel returns the github label of the label
func (c *LabelerScope) GithubLabel() github.Label {
	return github.Label{
		Name:  github.String(c.longName),
		Color: github.String(c.color),
	}
}
