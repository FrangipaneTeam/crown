package handlers

import (
	"context"
	"encoding/json"

	"github.com/FrangipaneTeam/crown/handlers/comments"
	"github.com/FrangipaneTeam/crown/pkg/conventionalcommit"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/labeler"
	"github.com/azrod/common-go"
	"github.com/google/go-github/v47/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

// Handler for pull request events
// More details : https://docs.github.com/fr/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request

type PullRequestHandler struct {
	githubapp.ClientCreator
}

// Handles returns the list of events this handler handles
func (h *PullRequestHandler) Handles() []string {
	return []string{"pull_request"}
}

// Handle processes the event
func (h *PullRequestHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse issue comment event payload")
	}

	labelsCategory := make([]string, 0)
	labelsType := make([]string, 0)

	ghc, err := ghclient.NewGHClient(ctx, h, event)
	if err != nil {
		return errors.Wrap(err, "failed to create github client for issue comment event")
	}

	ghc.Logger.Debug().Msgf("Event action is %s in Handle PullRequest", event.GetAction())

	switch event.GetAction() {
	case "opened", "edited", "synchronize":

		// * Check if the issue have a conventional issue title

		// ParseIssue
		PrTitle, err := conventionalcommit.ParseCommit(ghc.GetPullRequest().GetTitle())
		if err != nil {
			ghc.Logger.Debug().Msg("PullRequest title is not conventional commit format")
			comments.EditIssueComment(ghc, comments.IDPRTitleInvalid, ghc.GetPullRequest().GetTitle())
		} else {

			// Scope
			_, err = ghc.GetLabel(labeler.FormatedLabelScope(PrTitle.Scope()))
			if err != nil {
				comments.EditIssueComment(ghc, comments.IDPRTitleInvalid, PrTitle.GetScope())
			} else {
				labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(PrTitle.Scope()))
			}

			// Type
			if v, ok := labeler.FindLabelerType(PrTitle); !ok {
				comments.EditIssueComment(ghc, comments.IDPRTypeInvalid, v.GetShortName())
			} else {
				_, err = ghc.GetLabel(v.GetLongName())
				if err != nil {
					err = ghc.CreateLabel(v.GitHubLabel())
					if err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to create label")
					} else {
						labelsType = append(labelsType, v.GetLongName())
					}
				} else {
					labelsType = append(labelsType, v.GetLongName())
				}
			}

			// Breaking change
			if PrTitle.IsBreakingChange() {
				_, err := ghc.GetLabel(labeler.BreakingChange.GetLongName())
				if err != nil {
					err = ghc.CreateLabel(labeler.BreakingChange.GithubLabel())
					if err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to create label")
					} else {
						labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
					}
				} else {
					labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
				}
			}
		}
		// End of check PR title

		// * Check if the commits messages are conventional commit format

		commits, err := ghc.GetCommits()
		if err != nil {
			ghc.Logger.Error().Err(err).Msg("Failed to get commits")
		} else {
			for _, commit := range commits {

				ghc.Logger.Debug().Msgf("Commit message is %s", commit.GetCommit().GetMessage())

				cm, err := conventionalcommit.ParseCommit(commit.GetCommit().GetMessage())
				if err != nil {
					ghc.Logger.Debug().Msg("Commit message is not conventional commit format")
					comments.CreateIssueComment(ghc, comments.IDPRCommitInvalid, commit.GetCommit().GetMessage())
				} else {

					// Scope
					_, err = ghc.GetLabel(labeler.FormatedLabelScope(cm.Scope()))
					if err != nil {
						comments.CreateIssueComment(ghc, comments.IDIssuesLabelNotExists, cm.GetScope())
					} else {
						if _, ok := common.Find(labelsCategory, labeler.FormatedLabelScope(cm.Scope())); !ok {
							labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(cm.Scope()))
						}
					}

					// Type
					if v, ok := labeler.FindLabelerType(cm); !ok {
						comments.CreateIssueComment(ghc, comments.IDPRTypeInvalid, v.GetShortName())
					} else {
						_, err = ghc.GetLabel(v.GetLongName())
						if err != nil {
							err = ghc.CreateLabel(v.GitHubLabel())
							if err != nil {
								ghc.Logger.Error().Err(err).Msg("Failed to create label")
							} else {
								labelsType = append(labelsType, v.GetLongName())
							}
						} else {
							labelsType = append(labelsType, v.GetLongName())
						}
					}

					// Breaking change
					if cm.IsBreakingChange() {
						_, err := ghc.GetLabel(labeler.BreakingChange.GetLongName())
						if err != nil {
							err = ghc.CreateLabel(labeler.BreakingChange.GithubLabel())
							if err != nil {
								ghc.Logger.Error().Err(err).Msg("Failed to create label")
							} else {
								labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
							}
						} else {
							labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
						}
					}

				}
			}
		}

		// * Calcul Additions and Deletions
		// * Check if the PR is too big

	default:
		return nil
	}

	o := make([]string, 0)

	allLabels := append(labelsType, labelsCategory...)

	ghc.Logger.Debug().Msgf("Labels are %v", event.PullRequest.Labels)

	for _, lbl := range event.PullRequest.Labels {
		ghc.Logger.Debug().Msgf("Label is %s", lbl.GetName())
		if _, ok := common.Find(allLabels, lbl.GetName()); !ok {
			ghc.RemoveLabelForIssue(lbl.GetName())
		}
		o = append(o, lbl.GetName())
	}

	for _, lbl := range allLabels {
		if _, ok := common.Find(o, lbl); !ok {
			ghc.AddLabelToIssue(lbl)
		}
	}

	return nil

}
