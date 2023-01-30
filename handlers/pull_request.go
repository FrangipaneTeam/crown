package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/FrangipaneTeam/crown/handlers/comments"
	"github.com/FrangipaneTeam/crown/handlers/status"
	"github.com/FrangipaneTeam/crown/pkg/config"
	"github.com/FrangipaneTeam/crown/pkg/conventionalcommit"
	"github.com/FrangipaneTeam/crown/pkg/conventionalsizepr"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/labeler"
	"github.com/FrangipaneTeam/crown/pkg/statustype"
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

	ghc, err := ghclient.NewGHClient(ctx, h, event)
	if err != nil {
		return errors.Wrap(err, "failed to create github client for issue comment event")
	}

	if err = ghc.BranchRequiredStatusChecks([]*github.RequiredStatusCheck{
		{
			Context: status.PR_Check_Title.String(),
			AppID:   github.Int64(config.AppID),
		},
		{
			Context: status.PR_Check_SizeChanges.String(),
			AppID:   github.Int64(config.AppID),
		},
		{
			Context: status.PR_Labeler.String(),
			AppID:   github.Int64(config.AppID),
		},
		{
			Context: status.PR_Check_Commits.String(),
			AppID:   github.Int64(config.AppID),
		},
	}); err != nil {
		ghc.Logger.Error().Err(err).Msg("Failed to set branch protection")
	}

	ghc.Logger.Debug().Msgf("Event action is %s in Handle PullRequest", event.GetAction())

	// Specific Action
	switch event.GetAction() {
	case "labeled":
		l := event.GetLabel()
		MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
			Label: l.GetName(),
		})

		// Remove comment if label added is in comment
		MsgPRIssuesLabelNotExists.RemoveIssueComment()
	}

	// Generic Action
	switch event.GetAction() {
	case "opened", "edited", "synchronize", "labeled":
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(500+rand.Intn(3000-500+1)) * time.Millisecond)

		labelsCategory := make([]string, 0)
		labelsType := make([]string, 0)
		labelsSize := make([]string, 0)

		commitSHA := event.GetPullRequest().GetHead().GetSHA()
		if commitSHA == "" {
			ghc.Logger.Error().Msg("Failed to get commit SHA")
			return nil
		}

		// * Init status
		// At this instant, All status are pending
		PR_Check_Title := status.NewStatus(ghc, status.PR_Check_Title, commitSHA)
		PR_Check_commits := status.NewStatus(ghc, status.PR_Check_Commits, commitSHA)
		PR_Labeler := status.NewStatus(ghc, status.PR_Labeler, commitSHA)
		PR_Check_SizeChanges := status.NewStatus(ghc, status.PR_Check_SizeChanges, commitSHA)

		// * Check if the PullRequest have a conventional commit title format
		// ? ParseTitle
		MsgPRTitleInvalid := comments.NewCommentMsg(ghc, comments.IDPRTitleInvalid, comments.PRTitleInvalidValues{
			Title: ghc.GetPullRequest().GetTitle(),
		})
		if MsgPRTitleInvalid == nil {
			ghc.Logger.Error().Msg("Failed to create comment")
			return nil
		}
		PrTitle, err := conventionalcommit.ParseCommit(ghc.GetPullRequest().GetTitle())
		if err != nil {
			ghc.Logger.Debug().Msg("Error while parsing PR title")
			if err := PR_Check_Title.SetState(statustype.Failure); err != nil {
				ghc.Logger.Error().Err(err).Msg("Failed to set status")
			}
			MsgPRTitleInvalid.EditIssueComment()
		} else {

			// Scope
			if PrTitle.Scope() != "" {
				MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
					Label: labeler.FormatedLabelScope(PrTitle.Scope()),
				})
				if MsgPRIssuesLabelNotExists == nil {
					ghc.Logger.Error().Msg("Failed to create comment")
				}
				_, err = ghc.GetLabel(labeler.FormatedLabelScope(PrTitle.Scope()))
				if err != nil {
					PR_Labeler.SetState(statustype.Failure)
					if err := MsgPRIssuesLabelNotExists.EditIssueComment(); err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
					}
				} else {
					if _, ok := common.Find(labelsCategory, labeler.FormatedLabelScope(PrTitle.Scope())); !ok {
						labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(PrTitle.Scope()))
					}
				}
			}

			// Type
			if v, ok := labeler.FindLabelerType(PrTitle); !ok {
				PR_Check_Title.SetState(statustype.Failure)
				MsgPRTitleInvalid.EditIssueComment()
			} else {
				_, err = ghc.GetLabel(v.GetLongName())
				if err != nil {
					err = ghc.CreateLabel(v.GitHubLabel())
					if err != nil {
						PR_Check_Title.SetState(statustype.Failure)
						ghc.Logger.Error().Err(err).Msg("Failed to create label")
					} else {
						if _, ok := common.Find(labelsType, labeler.FormatedLabelScope(v.GetLongName())); !ok {
							labelsType = append(labelsType, v.GetLongName())
						}
					}
				} else {
					if _, ok := common.Find(labelsType, labeler.FormatedLabelScope(v.GetLongName())); !ok {
						labelsType = append(labelsType, v.GetLongName())
					}
				}
			}

			// Breaking change
			if PrTitle.IsBreakingChange() {
				_, err := ghc.GetLabel(labeler.BreakingChange.GetLongName())
				if err != nil {
					err = ghc.CreateLabel(labeler.BreakingChange.GithubLabel())
					if err != nil {
						PR_Check_Title.SetState(statustype.Failure)
						ghc.Logger.Error().Err(err).Msg("Failed to create label")
					} else {
						if _, ok := common.Find(labelsType, labeler.FormatedLabelScope(labeler.BreakingChange.GetLongName())); !ok {
							labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
						}
					}
				} else {
					if _, ok := common.Find(labelsType, labeler.FormatedLabelScope(labeler.BreakingChange.GetLongName())); !ok {
						labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
					}
				}
			}

			if err := PR_Check_Title.IsSuccess(); err != nil {
				ghc.Logger.Error().Err(err).Msg("Failed to set status")
			}

			// If PR title is valid, remove issue comment
			if PR_Check_Title.GetState() == statustype.Success {
				// Remove issue comment if commit message is valid
				if err := MsgPRTitleInvalid.RemoveIssueComment(); err != nil {
					ghc.Logger.Error().Err(err).Msg("Failed to remove issue comment")
				}
			}
		}
		// End of check PR title

		// * Check if the commits messages are conventional commit format

		// ? ParseCommits

		commits, err := ghc.GetCommits()
		if err != nil {
			PR_Check_commits.SetState(statustype.Failure)
			ghc.Logger.Error().Err(err).Msg("Failed to get commits")
		} else {

			allCommitsSHA := make([]string, 0)

			for _, commit := range commits {

				MsgPRCommitInvalid := comments.NewCommentMsg(ghc, comments.IDPRCommitInvalid, comments.PRCommitInvalidValues{
					CommitMsg: commit.GetCommit().GetMessage(),
					CommitSHA: commit.GetSHA(),
				})
				if MsgPRCommitInvalid == nil {
					ghc.Logger.Error().Msg("Failed to create comment")
					continue
				}

				allCommitsSHA = append(allCommitsSHA, commit.GetSHA())

				ghc.Logger.Debug().Msgf("Commit message is %s", commit.GetCommit().GetMessage())

				cm, err := conventionalcommit.ParseCommit(commit.GetCommit().GetMessage())
				if err != nil {
					ghc.Logger.Error().Str("message", commit.GetCommit().GetMessage()).Str("commitID", commit.GetSHA()).Msg("Commit message is not conventional commit format")
					PR_Check_commits.SetState(statustype.Failure)
					if err := MsgPRCommitInvalid.EditIssueComment(); err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
						continue
					}
				} else {

					// Remove issue comment if commit message is valid
					if err := MsgPRCommitInvalid.RemoveIssueComment(); err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to remove issue comment")
						continue
					}

					// Scope
					if cm.Scope() != "" {
						MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
							Label: labeler.FormatedLabelScope(cm.Scope()),
						})
						if MsgPRIssuesLabelNotExists == nil {
							ghc.Logger.Error().Msg("Failed to create comment")
							continue
						}
						_, err = ghc.GetLabel(labeler.FormatedLabelScope(cm.Scope()))
						if err != nil {
							PR_Labeler.SetState(statustype.Failure)
							if eer := MsgPRIssuesLabelNotExists.EditIssueComment(); err != nil {
								ghc.Logger.Error().Err(eer).Msg("Failed to edit issue comment")
								continue
							}
						} else {
							if _, ok := common.Find(labelsCategory, labeler.FormatedLabelScope(cm.Scope())); !ok {
								labelsCategory = append(labelsCategory, labeler.FormatedLabelScope(cm.Scope()))
							}
						}
					}
					// Type
					if v, ok := labeler.FindLabelerType(cm); !ok {
						ghc.Logger.Error().Str("message", commit.GetCommit().GetMessage()).Str("commitID", commit.GetSHA()).Msg("Commit message is not conventional commit format")
						PR_Check_commits.SetState(statustype.Failure)
						if err := MsgPRCommitInvalid.EditIssueComment(); err != nil {
							ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
							continue
						}
					} else {
						_, err = ghc.GetLabel(v.GetLongName())
						if err != nil {
							err = ghc.CreateLabel(v.GitHubLabel())
							if err != nil {
								PR_Labeler.SetState(statustype.Failure)
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
								PR_Labeler.SetState(statustype.Failure)
								ghc.Logger.Error().Err(err).Msg("Failed to create label")
							} else {
								labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
							}
						} else {
							labelsType = append(labelsType, labeler.BreakingChange.GetLongName())
						}
					}

					if err := PR_Check_commits.IsSuccess(); err != nil {
						ghc.Logger.Error().Err(err).Msg("Failed to set status")
					}

				}
			}

			cts, err := ghc.ListComments()
			if err != nil {
				ghc.Logger.Error().Err(err).Msg("Failed to get comments")
			} else {
				if len(cts) > 0 {
					for _, comment := range cts {
						if ok, value := comments.ExtraIssueComment(comment.GetBody(), comments.IDPRCommitInvalid, comments.ExtraBotID); ok && comments.IDPRCommitInvalid.IsValid(value) {
							if ok, value := comments.ExtraIssueComment(comment.GetBody(), comments.IDPRCommitInvalid, comments.ExtraCommitID); ok {
								if _, ok := common.Find(allCommitsSHA, value); !ok {
									// Delete commit for invalid commit message with commitID has been deleted
									ghc.DeleteComment(comment.GetID())
								}
							}
						}
					}
				}
			}

		}

		// * Calcul Additions and Deletions
		size := conventionalsizepr.NewPRSize(event.PullRequest.GetAdditions(), event.PullRequest.GetDeletions())

		// * Check if the PR is too big
		MsgPRSizeTooBig := comments.NewCommentMsg(ghc, comments.IDPRSizeTooBig, nil)
		if MsgPRSizeTooBig == nil {
			ghc.Logger.Error().Msg("Failed to create comment message")
			return nil
		}

		if size.IsTooBig() {
			MsgPRSizeTooBig.EditIssueComment()
		} else {
			MsgPRSizeTooBig.RemoveIssueComment()
		}

		l := labeler.FindLabelerSize(size.GetSize())
		_, err = ghc.GetLabel(l.GetLongName())
		if err != nil {
			err = ghc.CreateLabel(l.GithubLabel())
			if err != nil {
				PR_Check_SizeChanges.SetState(statustype.Failure)
				ghc.Logger.Error().Err(err).Msg("Failed to create label size")
			} else {
				labelsSize = append(labelsSize, l.GetLongName())
			}
		} else {
			labelsSize = append(labelsSize, l.GetLongName())
		}

		if err := PR_Check_SizeChanges.IsSuccess(); err != nil {
			ghc.Logger.Error().Err(err).Msg("Failed to set status")
		}

		o := make([]string, 0)
		allLabels := make([]string, 0)
		allLabels = append(allLabels, labelsType...)
		allLabels = append(allLabels, labelsCategory...)
		allLabels = append(allLabels, labelsSize...)

		for _, lbl := range event.PullRequest.Labels {
			ghc.Logger.Debug().Msgf("Label is %s", lbl.GetName())
			if _, ok := common.Find(allLabels, lbl.GetName()); !ok {
				if err := ghc.RemoveLabelForIssue(lbl.GetName()); err != nil {
					PR_Labeler.SetState(statustype.Failure)
					ghc.Logger.Error().Err(err).Msg("Failed to remove label")
				}
			}
			o = append(o, lbl.GetName())
		}

		for _, lbl := range allLabels {
			if _, ok := common.Find(o, lbl); !ok {
				if err := ghc.AddLabelToIssue(lbl); err != nil {
					PR_Labeler.SetState(statustype.Failure)
					ghc.Logger.Error().Err(err).Msg("Failed to add label")
				}
			}
		}

		if err := PR_Labeler.IsSuccess(); err != nil {
			ghc.Logger.Error().Err(err).Msg("Failed to set status")
		}

	default:
		return nil
	}

	return nil

}
