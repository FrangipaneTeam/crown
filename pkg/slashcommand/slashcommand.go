package slashcommand

import (
	"errors"
	"regexp"

	"github.com/google/go-github/v47/github"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/tracker"
)

const (
	labelCategoryColor = "BFD4F2"

	labelCmd = "label"
	trackCmd = "track"

	addVerb    = "add"
	removeVerb = "remove"
)

var slashcommandRe = regexp.MustCompile(`/(\w+):(\w+)\s+(\S+)`)

type SlashCommand struct {
	Command string
	Verb    string
	Desc    string
	IssueID int64
	Issue   *github.Issue
}

// FindSlashCommand finds slash command in body string.
func FindSlashCommand(body string) (bool, interface{}, error) {
	cmd := slashcommandRe.FindStringSubmatch(body)
	if len(cmd) != 4 {
		return false, nil, errors.New("invalid command")
	}

	switch cmd[1] {
	case labelCmd:
		v, err := findVerb(cmd[2])
		if err != nil {
			return false, nil, err
		}
		return true, Label{
			Action: CommandLabel,
			Verb:   v,
			Label:  cmd[3],
		}, nil

	// case trackCmd:
	// 	return true, &SlashCommand{
	// 		Command: trackCmd,
	// 		Verb:    cmd[2],
	// 		Desc:    cmd[3],
	// 		IssueID: issueID,
	// 	}
	default:
		return false, nil, errors.New("invalid command")
	}
}

// ExecuteSlashCommand executes slash command.
// Deprecated.
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

			if err := ghc.AddLabelToIssue(cmd.Desc); err != nil {
				ghc.Logger.Err(err).Msg("failed to add label")
				return err
			}

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
	case trackCmd:
		switch cmd.Verb { //nolint:gocritic
		case addVerb:

			if err := tracker.TrackNewIssue(ghc.GetRepoOwner(), ghc.GetRepoName(), ghc.GetInstallationID(), cmd.IssueID, cmd.Desc); err != nil {
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
