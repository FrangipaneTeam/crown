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

	if !event.GetIssue().IsPullRequest() {
		zerolog.Ctx(ctx).Debug().Msg("Issue comment event is not for a pull request")
		return nil
	}

	switch event.GetAction() {
	case "opened":

		// * Check if the issue have a conventional issue title

		// ParseIssue
		issueTitle, err := conventionalissue.Parse(ghc.GetIssue().GetTitle())
		if err != nil {
			ghc.Logger.Debug().Msg("Issue title is not conventional issue format")
			comments.CreateIssueComment(ghc, comments.IDIssuesTitleInvalid, ghc.GetIssue().GetTitle())
			return nil
		}

		_, err = ghc.GetLabel(labeler.FormatedLabelScope(issueTitle.GetScope()))
		if err != nil {
			comments.CreateIssueComment(ghc, comments.IDIssuesLabelNotExists, issueTitle.GetScope(), issueTitle.GetScope())
			return nil
		} else {
			labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(issueTitle.GetScope()))
		}

	case "edited":

		// * Check if the issue have a conventional issue title

		oldTitle := event.GetChanges().GetTitle().GetFrom()
		if oldTitle != ghc.GetIssue().GetTitle() {
			issueTitleOld, err := conventionalissue.Parse(oldTitle)
			if err == nil {
				ghc.RemoveLabelForIssue(labeler.FormatedLabelScope(issueTitleOld.GetScope()))
			}

			issueTitleNew, err := conventionalissue.Parse(ghc.GetIssue().GetTitle())
			if err != nil {
				comments.EditIssueComment(ghc, comments.IDIssuesTitleInvalid, ghc.GetIssue().GetTitle())
			} else {
				if err := comments.RemoveIssueComment(ghc, comments.IDIssuesTitleInvalid); err != nil {
					ghc.Logger.Error().Err(err).Msg("Failed to remove issue title invalid comment")
				}

				_, err = ghc.GetLabel(labeler.FormatedLabelScope(issueTitleNew.GetScope()))
				if err != nil {
					comments.EditIssueComment(ghc, comments.IDIssuesLabelNotExists, issueTitleNew.GetScope(), issueTitleNew.GetScope())
				} else {
					comments.RemoveIssueComment(ghc, comments.IDIssuesLabelNotExists)
					labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(issueTitleNew.GetScope()))
				}
			}
		}

	case "labeled":

		l := event.GetLabel()
		if strings.HasPrefix(l.GetName(), "category/") {
			labelsCategory = append(labelsCategory, l.GetName())
		}

		if commentID, ok := comments.IsIssueCommentExist(ghc, comments.IDIssuesLabelNotExists); ok {
			ghc.Logger.Debug().Msg("Issue comment IDIssuesLabelNotExists already exist")
			body := comments.GetIssueBody(ghc, commentID)
			if ok, value := comments.ExtraIssueComment(body, comments.IDIssuesLabelNotExists, comments.ExtraBotLabel); ok {
				ghc.Logger.Debug().Msgf("Issue comment IDIssuesLabelNotExists already exist for label %s", value)
				if labeler.FormatedLabelScope(value) == l.GetName() {
					comments.RemoveIssueComment(ghc, comments.IDIssuesLabelNotExists)
				}
			}
		}

	}

	lbls, err := ghc.SearchLabelInIssue("^category/")
	if err != nil {
		ghc.Logger.Error().Err(err).Msg("Failed to search label in issue")
	}

	o := make([]string, 0)

	for _, lbl := range lbls {
		if _, ok := common.Find(labelsCategory, lbl.GetName()); !ok {
			ghc.RemoveLabelForIssue(lbl.GetName())
		}
		o = append(o, lbl.GetName())
	}

	for _, lbl := range labelsCategory {
		if _, ok := common.Find(o, lbl); !ok {
			ghc.AddLabelToIssue(lbl)
		}
	}

	return nil

}
