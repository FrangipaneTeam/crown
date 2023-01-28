package comments

import (
	"regexp"

	"github.com/FrangipaneTeam/crown/pkg/common"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
)

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
