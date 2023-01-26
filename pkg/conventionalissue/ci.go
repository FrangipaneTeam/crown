package conventionalissue

import (
	"errors"
	"regexp"
	"strings"
)

type IssueTitle struct {
	scope, message string
}

// Parse parse the message of an issue and returns a conventional issue.
// Pattern is: [<scope>] <message>
func Parse(msg string) (*IssueTitle, error) {

	x := regexp.MustCompile(`\[([a-zA-Z]+)\/?([a-zA-Z]+)?\](.*)`)
	matches := x.FindStringSubmatch(msg)

	if len(matches) == 0 || len(matches) < 3 || matches[0] == "" {
		return nil, errors.New("invalid issue format. Format is \"[<scope>] <message>\"")
	}

	return &IssueTitle{
		scope:   matches[1],
		message: matches[2],
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
