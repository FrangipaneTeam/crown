package slashcommand

import (
	"fmt"
	"regexp"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/google/go-github/v47/github"
)

const (
	labelCategoryColor = "BFD4F2"

	startChar = "/"
	labelCmd  = "label"

	addVerb    = "add"
	removeVerb = "remove"
)

var (
	slashcommandRe = regexp.MustCompile(`/(\w+):(\w+)\s+(\S+)`)
)

type SlashCommand struct {
	Command string
	Verb    string
	Desc    string
	IssueID int64
}

// FindSlashCommand finds slash command in body string.
func FindSlashCommand(body string, issueID int64) (bool, *SlashCommand) {
	fmt.Println(body)
	cmd := slashcommandRe.FindStringSubmatch(body)
	fmt.Println(len(cmd))
	if len(cmd) != 4 {
		return false, nil
	}

	switch cmd[1] {
	case labelCmd:
		return true, &SlashCommand{
			Command: labelCmd,
			Verb:    cmd[2],
			Desc:    cmd[3],
			IssueID: issueID,
		}
	default:
		return false, nil
	}
}

// ExecuteSlashCommand executes slash command.
func ExecuteSlashCommand(ghc *ghclient.GHClient, cmd *SlashCommand) error {
	switch cmd.Command {
	case labelCmd:
		switch cmd.Verb {
		case addVerb:
			labelColor := labelCategoryColor

			newLabel := github.Label{
				Name:  &cmd.Desc,
				Color: &labelColor,
			}
			if err := ghc.CreateLabel(newLabel); err != nil {
				if err := ghc.AddCommentReaction(cmd.IssueID, "-1"); err != nil {
					ghc.Logger.Err(err).Msg("failed to add reaction")
				}
				return err
			}

			if err := ghc.AddCommentReaction(cmd.IssueID, "+1"); err != nil {
				ghc.Logger.Err(err).Msg("failed to add reaction")
				return err
			}

			ghc.AddLabelToIssue(cmd.Desc)

		case removeVerb:
			if err := ghc.RemoveLabelsForIssue([]string{cmd.Desc}); err != nil {
				if err := ghc.AddCommentReaction(cmd.IssueID, "-1"); err != nil {
					ghc.Logger.Err(err).Msg("failed to add reaction")
				}
				return err
			}

			if err := ghc.AddCommentReaction(cmd.IssueID, "+1"); err != nil {
				ghc.Logger.Err(err).Msg("failed to add reaction")
				return err
			}

		}
	}
	return nil
}
