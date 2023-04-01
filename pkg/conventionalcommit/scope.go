package conventionalcommit

import (
	"errors"
	"regexp"
)

// commitScope returns the scope of the commit.
func (l *Cc) commitScope() error {
	if l.Scope() == "" {
		l.CommitScope = ""
		return nil
	}

	x := regexp.MustCompile(`([a-zA-Z]+)\/?([a-zA-Z]+)?`)
	matches := x.FindStringSubmatch(l.Scope())

	if len(matches) == 0 || len(matches) < 3 || matches[0] == "" {
		return errors.New("invalid scope commit format. Format is \"feat(scope): message or feat(scope/subscope): message\"")
	}

	l.CommitScope = CommitScope(matches[0])

	return nil
}

// GetScope returns the scope of the commit.
func (l *Cc) GetScope() CommitScope {
	return l.CommitScope
}
