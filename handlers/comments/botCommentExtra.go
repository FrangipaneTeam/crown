package comments

type BotCommentExtra int

const (
	// ! Always add new IDs at the END of the list
	ExtraBotID BotCommentExtra = 48753691 << iota
	ExtraBotLabel
	// ! Always add new IDs at the END of the list
)

// GetKey returns the key of the extra
func (c BotCommentExtra) GetKey() string {
	return issuesCommentsExtra[c].key
}
