package handlers

import (
	"context"
	"encoding/json"
	"strings"

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
		labelsType:     &[]string{},
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
	case "opened":

		// Author community
		core.Community()
		core.ComputeLabels()

	}

	return nil

}

type coreIssues struct {
	ghc            *ghclient.GHClient
	event          github.IssuesEvent
	labelsCategory *[]string
	labelsType     *[]string
}

// Community labels
func (core *coreIssues) Community() {
	if *core.event.Issue.AuthorAssociation == "NONE" || *core.event.Issue.AuthorAssociation == "CONTRIBUTOR" {
		_, err := core.ghc.GetLabel(labeler.LabelerCommunity().GetName())
		if err != nil {
			err = core.ghc.CreateLabel(labeler.LabelerCommunity().GithubLabel())
			if err != nil {
				core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
			} else {
				if _, ok := common.Find(*core.labelsType, labeler.FormatedLabelScope(labeler.LabelerCommunity().GetName())); !ok {
					*core.labelsType = append(*core.labelsType, labeler.LabelerCommunity().GetName())
				}
			}
		} else {
			if _, ok := common.Find(*core.labelsType, labeler.FormatedLabelScope(labeler.LabelerCommunity().GetName())); !ok {
				*core.labelsType = append(*core.labelsType, labeler.LabelerCommunity().GetName())
			}
		}
	}
}

// ComputeLabels compute labels to add to the issue
func (core *coreIssues) ComputeLabels() error {

	o := make([]string, 0)
	allLabels := make([]string, 0)
	allLabels = append(allLabels, *core.labelsCategory...)
	allLabels = append(allLabels, *core.labelsType...)

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
