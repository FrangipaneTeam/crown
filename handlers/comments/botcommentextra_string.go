// Code generated by "stringer -type=BotCommentExtra"; DO NOT EDIT.

package comments

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ExtraBotID-48753691]
	_ = x[ExtraBotLabel-97507382]
	_ = x[ExtraCommitID-195014764]
}

const (
	_BotCommentExtra_name_0 = "ExtraBotID"
	_BotCommentExtra_name_1 = "ExtraBotLabel"
	_BotCommentExtra_name_2 = "ExtraCommitID"
)

func (i BotCommentExtra) String() string {
	switch {
	case i == 48753691:
		return _BotCommentExtra_name_0
	case i == 97507382:
		return _BotCommentExtra_name_1
	case i == 195014764:
		return _BotCommentExtra_name_2
	default:
		return "BotCommentExtra(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
