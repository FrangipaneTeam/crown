package comments

import "fmt"

type BotCommentID int

const (
	// ! Always add new IDs at the END of the list
	IDIssuesTitleInvalid BotCommentID = 597659851 << iota
	IDIssuesLabelNotExists
	IDPRTitleInvalid
	IDPRCommitInvalid
	// ! Always add new IDs at the END of the list
)

// Int64 returns a pointer to the int64 value passed in.
func Int64(v int64) *int64 {
	return &v
}

// Int64Value returns the value of the int64 pointer passed in or 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// IsValid returns true if the string passed in is equal to BotCommentID
func (c BotCommentID) IsValid(id string) bool {
	return fmt.Sprint(c) == id
}
