package conventionalissue

import (
	"errors"
	"regexp"
	"strings"

	"github.com/FrangipaneTeam/crown/pkg/common"
)

type IssueTitle struct {
	scope, message string
}

// Parse parse the message of an issue and returns a conventional issue.
// Pattern is: [<SCOPE>] <message>
func Parse(titleBody string) (*IssueTitle, error) {
	var re = regexp.MustCompile(`(?m)\[(?P<scope>[A-Z]+)\/?([A-Z]+)?\]\s+(?P<value>[-a-zA-Z_():\s]+)$`)

	match := re.FindString(titleBody)
	m := common.ReSubMatchMap(re, match)
	if len(m) == 0 {
		return nil, errors.New("invalid issue format. Format is \"[<SCOPE>] <message>\"")
	}

	if m["scope"] == "" || m["value"] == "" {
		return nil, errors.New("invalid issue format. Format is \"[<SCOPE>] <message>\"")
	}

	return &IssueTitle{
		scope:   m["scope"],
		message: m["value"],
	}, nil

}

// GetScope returns the scope of the issue.
func (i *IssueTitle) GetScope() string {
	return strings.ToLower(i.scope)
}

// GetMessage returns the message of the issue.
func (i *IssueTitle) GetMessage() string {
	return i.message
}
