package comments

import (
	"fmt"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
)

var (
	issuesComments = map[BotCommentID]string{
		IDIssuesTitleInvalid:   "The issue title `%s` is not conventional issue format.\nPlease follow this format : `[SCOPE] title`",
		IDIssuesLabelNotExists: "The label `%s` not existing in this repository. \nif you are an administrator you can write a comment with the command `/label:add category/%s` to automatically create the label",
		IDPRTitleInvalid:       "The pull request title `%s` is not conventional commit format.\nPlease follow this format : `type(scope): subject` or `type: subject`\n\nFor more information about conventional commit, please visit [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/)",
		IDPRCommitInvalid:      "The commit message `%s` is not conventional commit format.\nPlease follow this format : `type(scope): subject` or `type: subject`\n\nFor more information about conventional commit, please visit [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/)",
	}

	issuesCommentsExtra = map[BotCommentExtra]commentExtra{
		ExtraBotID:    {key: "botid", value: nil},
		ExtraBotLabel: {key: "bot_label", value: nil},
	}
)

type commentExtra struct {
	key   string
	value interface{}
}

func (c *commentExtra) SetValue(v interface{}) {
	c.value = v
}

// injectBotID injects the bot id in the message
func (c commentExtra) injectBotExtras() string {
	switch c.value.(type) {
	case int, int8, int16, int32, int64, BotCommentID:
		return fmt.Sprintf("<!-- %s: %v -->\n", c.key, c.value.(BotCommentID))
	default:
		return fmt.Sprintf("<!-- %s: %s -->\n", c.key, c.value)
	}
}

type commentMsg struct {
	ghc *ghclient.GHClient

	id BotCommentID

	msgPattern  string
	msgComputed string
	extra       map[BotCommentExtra]commentExtra

	args []interface{}
}

// setExtra sets the extra value
func (c *commentMsg) setExtra(extra BotCommentExtra, v interface{}) {
	w := issuesCommentsExtra[extra]
	w.SetValue(v)

	c.extra[extra] = w
}

// NewCommentMsg creates a new comment message
func newCommentMsg(ghc *ghclient.GHClient, id BotCommentID, args ...interface{}) *commentMsg {
	x := &commentMsg{
		ghc:        ghc,
		id:         id,
		msgPattern: issuesComments[id],
		extra:      make(map[BotCommentExtra]commentExtra),
		args:       args,
	}

	x.setExtra(ExtraBotID, id)

	switch id {
	case IDIssuesLabelNotExists:
		if len(x.args) == 0 {
			x.ghc.Logger.Error().Msg("args is empty")
			return nil
		}

		ebl := issuesCommentsExtra[ExtraBotLabel]
		ebl.SetValue(x.args[0])
		x.extra[ExtraBotLabel] = ebl

		x.msgComputed = fmt.Sprintf(issuesComments[id], x.args[0], x.args[0])
	default:
		x.msgComputed = fmt.Sprintf(issuesComments[id], x.args)
	}

	return x
}

// createIssueMessage creates a comment on the issue if the title is not conventional issue format
func (n *commentMsg) createIssueMessage() *string {

	var msg string
	for _, extra := range n.extra {
		msg += extra.injectBotExtras()
	}

	if n.msgComputed == "" {
		n.ghc.Logger.Error().Msg("msgComputed is empty")
		return nil
	}

	msg += n.msgComputed

	return &msg
}
