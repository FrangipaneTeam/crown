package comments

import (
	"regexp"

	"github.com/FrangipaneTeam/crown/pkg/common"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/google/go-github/v47/github"
)

// CreateIssueComment creates a comment on the issue if it does not exist.
func CreateIssueComment(ghClient *ghclient.GHClient, id BotCommentID, args ...interface{}) {

	commentMsg := newCommentMsg(ghClient, id, args...)

	_, ok := IsIssueCommentExist(ghClient, id)
	if !ok {
		prComment := github.IssueComment{
			Body: commentMsg.createIssueMessage(),
		}
		if err := ghClient.CreateComment(prComment); err != nil {
			ghClient.Logger.Error().Err(err).Msg("Failed to comment on issue")
		}
	}

}

// EditIssueComment edits the comment ID if it exists.
// If the comment ID does not exist, it creates it.
func EditIssueComment(ghClient *ghclient.GHClient, id BotCommentID, args ...interface{}) {

	commentMsg := newCommentMsg(ghClient, id, args...)

	commentID, ok := IsIssueCommentExist(ghClient, id)
	if ok {
		prComment := github.IssueComment{
			Body: commentMsg.createIssueMessage(),
		}
		if err := ghClient.EditComment(commentID, prComment); err != nil {
			ghClient.Logger.Error().Err(err).Msg("Failed to edit comment on issue")
		}
	} else {
		CreateIssueComment(ghClient, id, args...)
	}
}

// IsIssueCommentExist checks if the comment ID exists
func IsIssueCommentExist(ghClient *ghclient.GHClient, id BotCommentID) (commentID int64, exist bool) {

	cts, err := ghClient.ListComments()
	if err != nil {
		ghClient.Logger.Error().Err(err).Msg("Failed to get comments")
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

// RemoveIssueComment removes the comment ID. If the comment ID does not exist, it does nothing.
// If the comment ID exists, it removes it.
func RemoveIssueComment(ghClient *ghclient.GHClient, id BotCommentID) error {

	commentID, ok := IsIssueCommentExist(ghClient, id)
	if ok {
		if err := ghClient.DeleteComment(commentID); err != nil {
			ghClient.Logger.Error().Err(err).Msg("Failed to delete comment")
			return err
		}
	}

	return nil
}

// GetIssueBody returns the body of the issue.
func GetIssueBody(ghClient *ghclient.GHClient, commentID int64) string {

	cts, err := ghClient.ListComments()
	if err != nil {
		ghClient.Logger.Error().Err(err).Msg("Failed to get comments")
	} else {
		if len(cts) > 0 {
			for _, comment := range cts {
				if comment.GetID() == commentID {
					return comment.GetBody()
				}
			}
		}
	}

	return ""
}

// ExtraIssueComment Extract key and value from the comment ID.
// If the comment ID does not exist, it returns an empty string.
func ExtraIssueComment(msgBody string, id BotCommentID, key BotCommentExtra) (found bool, value string) {

	var (
		keyExpected = key.GetKey()
		re          = regexp.MustCompile(`(?m)<!--\s+(?P<key>[a-z_]+):\s+(?P<value>\S+)\s+-->`)
	)

	for _, match := range re.FindAllString(msgBody, -1) {
		m := common.ReSubMatchMap(re, match)
		if m["key"] == keyExpected {
			return true, m["value"]
		}
	}

	return false, ""

}
