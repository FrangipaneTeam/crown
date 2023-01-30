package comments

import (
	"fmt"
	"net/http"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/google/go-github/v47/github"
)

var (
	issuesComments = map[BotCommentID]string{
		// Issues comments
		IDIssuesTitleInvalid:   "The issue title `%s` is not conventional issue format.\nPlease follow this format : `[SCOPE] title`",
		IDIssuesLabelNotExists: "The label `%s` not existing in this repository. \nif you are an administrator you can write a comment with the command `/label:add %s` to automatically create the label",
		// PR comments
		IDPRTitleInvalid:  "The pull request title `%s` is not conventional commit format.\nPlease follow this format : `type(scope): subject` or `type: subject`\n\nFor more information about conventional commit, please visit [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/)",
		IDPRCommitInvalid: "The commit message `%s` is not conventional commit format.\nPlease follow this formats :\n* `type(scope): subject`\n* `type: subject`\n\nFor more information about conventional commit, please visit [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/)",
		IDPRSizeTooBig:    "Thank you for your contribution, but this PR exceeds the recommended size of 1000 lines. Please make sure you are NOT addressing multiple issues with one PR.\nNote this PR might be rejected due to its size.",
	}

	issuesCommentsExtra = map[BotCommentExtra]commentExtra{
		ExtraBotID:    {key: "botid", value: nil},
		ExtraBotLabel: {key: "bot_label", value: nil},
		ExtraCommitID: {key: "commit_id", value: nil},
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
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("<!-- %s: %d -->\n", c.key, c.value)
	case BotCommentID:
		return fmt.Sprintf("<!-- %s: %d -->\n", c.key, c.value.(BotCommentID))
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

	args   []interface{}
	values interface{}

	// IsIssueCommentExist checks if the comment ID exists
	IsIssueCommentExist func() (commentID int64, exist bool)

	// CreateIssueComment creates a comment on the issue if it does not exist.
	CreateIssueComment func() error

	// EditIssueComment edits the comment ID if it exists.
	// If the comment ID does not exist, it creates it.
	EditIssueComment func() error

	// RemoveIssueComment removes the comment ID.
	// If the comment ID does not exist, it does nothing.
	// If the comment ID exists, it removes it.
	RemoveIssueComment func() error
}

// setExtra sets the extra value
func (c *commentMsg) setExtra(extra BotCommentExtra, v interface{}) {
	w := issuesCommentsExtra[extra]
	w.SetValue(v)
	c.extra[extra] = w
}

// NewCommentMsg creates a new comment message
func NewCommentMsg(ghc *ghclient.GHClient, id BotCommentID, values interface{}) *commentMsg {
	x := &commentMsg{
		ghc:        ghc,
		id:         id,
		msgPattern: issuesComments[id],
		extra:      make(map[BotCommentExtra]commentExtra),
		values:     values,
	}

	x.setExtra(ExtraBotID, id.Id())
	x.IsIssueCommentExist = func() (commentID int64, exist bool) {
		cts, err := x.ghc.ListComments()
		if err != nil {
			x.ghc.Logger.Error().Err(err).Msg("Failed to get comments")
		} else {
			if len(cts) > 0 {
				for _, comment := range cts {
					if ok, value := ExtraIssueComment(comment.GetBody(), id, ExtraBotID); ok && id.IsValid(value) {
						return comment.GetID(), true
					}
				}
			}
		}

		return 0, false
	}
	x.CreateIssueComment = func() error {
		_, ok := x.IsIssueCommentExist()
		if !ok {
			prComment := github.IssueComment{
				Body: x.createIssueMessage(),
			}
			if err := x.ghc.CreateComment(prComment); err != nil {
				x.ghc.Logger.Error().Err(err).Msg("Failed to comment on issue")
				return err
			}
		}
		return nil
	}
	x.EditIssueComment = func() error {
		commentID, ok := x.IsIssueCommentExist()
		if ok {
			prComment := github.IssueComment{
				Body: x.createIssueMessage(),
			}
			if err := x.ghc.EditComment(commentID, prComment); err != nil {
				if err.(*github.ErrorResponse).Response.StatusCode == http.StatusNotFound {
					return x.CreateIssueComment()
				}
				x.ghc.Logger.Error().Err(err).Int64("commentID", commentID).Msg("Failed to edit comment on issue")
				return err
			}
		} else {
			return x.CreateIssueComment()
		}
		return nil
	}
	x.RemoveIssueComment = func() error {
		commentID, ok := x.IsIssueCommentExist()
		if ok {
			if err := x.ghc.DeleteComment(commentID); err != nil {
				x.ghc.Logger.Error().Err(err).Msg("Failed to delete comment")
				return err
			}
		}

		return nil
	}
	switch id {
	case IDIssuesTitleInvalid:
		if x.values == nil {
			x.ghc.Logger.Error().Msg("values is nil")
			return nil
		}
		if vals, ok := x.values.(IssuesTitleInvalidValues); ok {
			x.msgComputed = fmt.Sprintf(issuesComments[id], vals.Title)
		}

	case IDPRTitleInvalid:
		if x.values == nil {
			x.ghc.Logger.Error().Msg("values is nil")
			return nil
		}
		if vals, ok := x.values.(PRTitleInvalidValues); ok {
			x.msgComputed = fmt.Sprintf(issuesComments[id], vals.Title)
		}

	case IDIssuesLabelNotExists:
		if x.values == nil {
			x.ghc.Logger.Error().Msg("values is nil")
			return nil
		}
		if vals, ok := x.values.(IssuesLabelNotExistsValues); ok {
			ebl := issuesCommentsExtra[ExtraBotLabel]
			ebl.SetValue(vals.Label)
			x.extra[ExtraBotLabel] = ebl

			x.msgComputed = fmt.Sprintf(issuesComments[id], vals.Label, vals.Label)
			x.IsIssueCommentExist = func() (commentID int64, exist bool) {
				cts, err := x.ghc.ListComments()
				if err != nil {
					x.ghc.Logger.Error().Err(err).Msg("Failed to get comments")
				} else {

					if len(cts) > 0 {
						for _, comment := range cts {
							if ok, value := ExtraIssueComment(comment.GetBody(), id, ExtraBotID); ok && id.IsValid(value) {
								ghc.Logger.Debug().Msg("Found ExtraBotID")
								if ok, label := ExtraIssueComment(comment.GetBody(), id, ExtraBotLabel); ok && label == vals.Label {
									return comment.GetID(), true
								}
							}
						}
					}
				}

				return 0, false
			}
		} else {
			x.ghc.Logger.Error().Msg("values is not IssuesLabelNotExistsValues")
			return nil
		}
	case IDPRCommitInvalid:
		if x.values == nil {
			x.ghc.Logger.Error().Msg("values is nil")
			return nil
		}

		if vals, ok := x.values.(PRCommitInvalidValues); ok {
			if vals.CommitSHA == "" || vals.CommitMsg == "" {
				x.ghc.Logger.Error().Msg("commit sha or commit msg is empty")
				return nil
			}

			eci := issuesCommentsExtra[ExtraCommitID]
			eci.SetValue(vals.CommitSHA)
			x.extra[ExtraCommitID] = eci

			x.msgComputed = fmt.Sprintf(issuesComments[id], vals.CommitMsg)
			x.IsIssueCommentExist = func() (commentID int64, exist bool) {
				cts, err := x.ghc.ListComments()
				if err != nil {
					x.ghc.Logger.Error().Err(err).Msg("Failed to get comments")
				} else {
					if len(cts) > 0 {
						for _, comment := range cts {
							if ok, value := ExtraIssueComment(comment.GetBody(), id, ExtraBotID); ok && id.IsValid(value) {
								if ok, commitID := ExtraIssueComment(comment.GetBody(), id, ExtraCommitID); ok && commitID == vals.CommitSHA {
									return comment.GetID(), true
								}
							}
						}
					}
				}

				return 0, false
			}
		} else {
			x.ghc.Logger.Error().Msg("values is not PRCommitInvalidValues")
			return nil
		}

	default:
		x.msgComputed = fmt.Sprintf(issuesComments[id])
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
