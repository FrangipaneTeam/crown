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

	core := &coreIssues{
		ghc:            ghc,
		event:          event,
		labelsCategory: &[]string{},
	}

	ghc.Logger.Debug().Msgf("Event action is %s in Handle issues", event.GetAction())

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
		core.ParseTitle()

		core.ComputeLabels()

	case "labeled":

		MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
			Label: event.GetLabel().GetName(),
		})

		// Remove comment if label added is in comment
		MsgPRIssuesLabelNotExists.RemoveIssueComment()

		// TODO Add closed action
		// Untrack external issue

	}

	return nil

}

type coreIssues struct {
	ghc            *ghclient.GHClient
	event          github.IssuesEvent
	labelsCategory *[]string
}

// ParseTitle parse the title of the issue
func (core *coreIssues) ParseTitle() error {

	MsgIssuesTitleInvalid := comments.NewCommentMsg(core.ghc, comments.IDIssuesTitleInvalid, comments.IssuesTitleInvalidValues{
		Title: core.event.GetIssue().GetTitle(),
	})
	if MsgIssuesTitleInvalid == nil {
		core.ghc.Logger.Error().Msg("Failed to create comment")
		return nil
	}

	// ? ParseIssueTitle
	issueTitle, err := conventionalissue.Parse(core.event.GetIssue().GetTitle())
	if err != nil {
		core.ghc.Logger.Debug().Msg("Issue title is not conventional issue format")
		MsgIssuesTitleInvalid.EditIssueComment()
		return nil
	}

	if issueTitle.GetScope() != "" {
		if err := MsgIssuesTitleInvalid.RemoveIssueComment(); err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to remove comment")
		}

		MsgIssuesLabelNotExists := comments.NewCommentMsg(core.ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
			Label: labeler.FormatedLabelScope(issueTitle.GetScope()),
		})
		if MsgIssuesLabelNotExists == nil {
			core.ghc.Logger.Error().Msg("Failed to create comment")
		}

		_, err = core.ghc.GetLabel(labeler.FormatedLabelScope(issueTitle.GetScope()))
		if err != nil {
			MsgIssuesLabelNotExists.EditIssueComment()
		} else {
			MsgIssuesLabelNotExists.RemoveIssueComment()
			if _, ok := common.Find(*core.labelsCategory, labeler.FormatedLabelScope(issueTitle.GetScope())); !ok {
				*core.labelsCategory = append(*core.labelsCategory, labeler.FormatedLabelScope(issueTitle.GetScope()))
			}
		}
	} else {
		core.ghc.Logger.Debug().Msg("Issue title has no scope")
		MsgIssuesTitleInvalid.EditIssueComment()
	}

	return nil

}

// ComputeLabels compute labels to add to the issue
func (core *coreIssues) ComputeLabels() error {

	o := make([]string, 0)
	allLabels := make([]string, 0)
	allLabels = append(allLabels, *core.labelsCategory...)

	for _, lbl := range core.event.Issue.Labels {
		core.ghc.Logger.Debug().Msgf("Label is %s", lbl.GetName())
		if _, ok := common.Find(allLabels, lbl.GetName()); !ok {
			if err := core.ghc.RemoveLabelForIssue(lbl.GetName()); err != nil {
				core.ghc.Logger.Error().Err(err).Msg("Failed to remove label")
			}
		}
		o = append(o, lbl.GetName())
	}

	for _, lbl := range allLabels {
		if _, ok := common.Find(o, lbl); !ok {
			if err := core.ghc.AddLabelToIssue(lbl); err != nil {
				core.ghc.Logger.Error().Err(err).Msg("Failed to add label")
			}
		}
	}

	return nil
}
