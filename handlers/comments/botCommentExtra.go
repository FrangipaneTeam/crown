package comments

//go:generate stringer -type=BotCommentExtra
type BotCommentExtra int

const (
	// ! Always add new IDs at the END of the list
	ExtraBotID BotCommentExtra = 48753691 << iota
	ExtraBotLabel
	ExtraCommitID
	// ! Always add new IDs at the END of the list
)

// GetKey returns the key of the extra
func (c BotCommentExtra) GetKey() string {
	return issuesCommentsExtra[c].key
}

type PRTitleInvalidValues struct {
	Title string
}

type PRCommitInvalidValues struct {
	CommitMsg string
	CommitSHA string
}

type IssuesLabelNotExistsValues struct {
	Label string
}

type IssuesTitleInvalidValues struct {
	Title string
	Scope string
}
