package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/FrangipaneTeam/crown/handlers/comments"
	"github.com/FrangipaneTeam/crown/pkg/conventionalissue"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/labeler"
	"github.com/azrod/common-go"
	"github.com/google/go-github/v47/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Handler for issues events
// More details : https://docs.github.com/fr/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#issues

type IssuesHandler struct {
	githubapp.ClientCreator
}

// Handles returns the list of events this handler handles
func (h *IssuesHandler) Handles() []string {
	return []string{"issues"}
}

// Handle processes the event
func (h *IssuesHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssuesEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse issues event payload")
	}

	ghc, err := ghclient.NewGHClient(ctx, h, event)
	if err != nil {
		return errors.Wrap(err, "failed to create github client")
	}

	ghc.Logger.Debug().Msgf("Event action is %s in Handle issues", event.GetAction())

	labelsCategory := make([]string, 0)

	if strings.HasSuffix(ghc.GetAuthor(), "[bot]") {
		ghc.Logger.Debug().Msg("Issue was created by a bot")
		return nil
	}

	if event.GetIssue().IsPullRequest() {
		zerolog.Ctx(ctx).Debug().Msg("Issue comment event is not for a pull request")
		return nil
	}

	switch event.GetAction() {
	case "opened", "edited":

		// * Check if the issue have a conventional issue title
		MsgIssuesTitleInvalid := comments.NewCommentMsg(ghc, comments.IDIssuesTitleInvalid, comments.IssuesTitleInvalidValues{
			Title: event.GetIssue().GetTitle(),
		})
		if MsgIssuesTitleInvalid == nil {
			ghc.Logger.Error().Msg("Failed to create comment")
			return nil
		}

		// ? ParseIssueTitle
		issueTitle, err := conventionalissue.Parse(event.GetIssue().GetTitle())
		if err != nil {
			ghc.Logger.Debug().Msg("Issue title is not conventional issue format")
			MsgIssuesTitleInvalid.EditIssueComment()
			return nil
		}

		if issueTitle.GetScope() != "" {
			if err := MsgIssuesTitleInvalid.RemoveIssueComment(); err != nil {
				ghc.Logger.Error().Err(err).Msg("Failed to remove comment")
			}

			MsgIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
				Label: labeler.FormatedLabelScope(issueTitle.GetScope()),
			})
			if MsgIssuesLabelNotExists == nil {
				ghc.Logger.Error().Msg("Failed to create comment")
			}

			_, err = ghc.GetLabel(labeler.FormatedLabelScope(issueTitle.GetScope()))
			if err != nil {
				MsgIssuesLabelNotExists.EditIssueComment()
			} else {
				MsgIssuesLabelNotExists.RemoveIssueComment()
				labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(issueTitle.GetScope()))
			}
		} else {
			ghc.Logger.Debug().Msg("Issue title has no scope")
			MsgIssuesTitleInvalid.EditIssueComment()
		}

		o := make([]string, 0)
		allLabels := make([]string, 0)
		allLabels = append(allLabels, labelsCategory...)

		for _, lbl := range event.Issue.Labels {
			ghc.Logger.Debug().Msgf("Label is %s", lbl.GetName())
			if _, ok := common.Find(allLabels, lbl.GetName()); !ok {
				if err := ghc.RemoveLabelForIssue(lbl.GetName()); err != nil {
					ghc.Logger.Error().Err(err).Msg("Failed to remove label")
				}
			}
			o = append(o, lbl.GetName())
		}

		for _, lbl := range allLabels {
			if _, ok := common.Find(o, lbl); !ok {
				if err := ghc.AddLabelToIssue(lbl); err != nil {
					ghc.Logger.Error().Err(err).Msg("Failed to add label")
				}
			}
		}

		// lbls, err := ghc.SearchLabelInIssue("^category/")
		// if err != nil {
		// 	ghc.Logger.Error().Err(err).Msg("Failed to search label in issue")
		// }

		// o := make([]string, 0)

		// for _, lbl := range lbls {
		// 	if _, ok := common.Find(labelsCategory, lbl.GetName()); !ok {
		// 		ghc.RemoveLabelForIssue(lbl.GetName())
		// 	}
		// 	o = append(o, lbl.GetName())
		// }

		// for _, lbl := range labelsCategory {
		// 	if _, ok := common.Find(o, lbl); !ok {
		// 		ghc.AddLabelToIssue(lbl)
		// 	}
		// }

	// case "edited":

	// 	// * Check if the issue have a conventional issue title

	// 	oldTitle := event.GetChanges().GetTitle().GetFrom()
	// 	if oldTitle != ghc.GetIssue().GetTitle() {
	// 		issueTitleOld, err := conventionalissue.Parse(oldTitle)
	// 		if err == nil {
	// 			ghc.RemoveLabelForIssue(labeler.FormatedLabelScope(issueTitleOld.GetScope()))
	// 		}

	// 		issueTitleNew, err := conventionalissue.Parse(ghc.GetIssue().GetTitle())
	// 		if err != nil {
	// 			comments.EditIssueComment(ghc, comments.IDIssuesTitleInvalid, ghc.GetIssue().GetTitle())
	// 		} else {
	// 			if err := comments.RemoveIssueComment(ghc, comments.IDIssuesTitleInvalid); err != nil {
	// 				ghc.Logger.Error().Err(err).Msg("Failed to remove issue title invalid comment")
	// 			}

	// 			_, err = ghc.GetLabel(labeler.FormatedLabelScope(issueTitleNew.GetScope()))
	// 			if err != nil {
	// 				comments.EditIssueComment(ghc, comments.IDIssuesLabelNotExists, issueTitleNew.GetScope(), issueTitleNew.GetScope())
	// 			} else {
	// 				comments.RemoveIssueComment(ghc, comments.IDIssuesLabelNotExists)
	// 				labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(issueTitleNew.GetScope()))
	// 			}
	// 		}
	// 	}

	case "labeled":

		ghc.Logger.Debug().Msgf("Label is %s", event.GetLabel().GetName())

		MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
			Label: event.GetLabel().GetName(),
		})

		// Remove comment if label added is in comment
		MsgPRIssuesLabelNotExists.RemoveIssueComment()

	}

	return nil

}
