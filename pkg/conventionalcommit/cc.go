package conventionalcommit

import (
	parser "github.com/conventionalcommit/parser"
)

type Cc struct {
	*parser.Commit

	CommitType  CommitType
	CommitScope CommitScope
}

type (
	CommitType  string
	CommitScope string
)

// String returns the string representation of a commit type.
func (c CommitType) String() string {
	return string(c)
}

// String returns the string representation of a commit scope.
func (c CommitScope) String() string {
	return string(c)
}

// ParseCommit parses a commit message and returns a conventional commit.
func ParseCommit(msg string) (*Cc, error) {
	c, err := parser.New().Parse(msg)
	if err != nil {
		return nil, err
	}

	x := &Cc{
		Commit: c,
	}

	x.commitType()
	if err := x.commitScope(); err != nil {
		return nil, err
	}

	return x, nil
}

// ParseCommits parses a list of commit messages and returns a list of conventional commits.
func ParseCommits(msgs []string) ([]*Cc, error) {
	var commits []*Cc

	for _, msg := range msgs {
		c, err := ParseCommit(msg)
		if err != nil {
			return nil, err
		}

		commits = append(commits, c)
	}

	return commits, nil
}
