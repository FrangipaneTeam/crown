package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/azrod/common-go"
	"github.com/google/go-github/v47/github"
	"github.com/kr/pretty"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"

	"github.com/FrangipaneTeam/crown/handlers/comments"
	"github.com/FrangipaneTeam/crown/handlers/status"
	"github.com/FrangipaneTeam/crown/pkg/conventionalcommit"
	"github.com/FrangipaneTeam/crown/pkg/conventionalsizepr"
	"github.com/FrangipaneTeam/crown/pkg/db"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/labeler"
	"github.com/FrangipaneTeam/crown/pkg/statustype"
)

// Handler for pull request events
// More details : https://docs.github.com/fr/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request

type PullRequestHandler struct {
	githubapp.ClientCreator
}

// Handles returns the list of events this handler handles.
func (h *PullRequestHandler) Handles() []string {
	return []string{"pull_request"}
}

// Handle processes the event.
func (h *PullRequestHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse issue comment event payload")
	}

	ghc, err := ghclient.NewGHClient(ctx, h, event)
	if err != nil {
		return errors.Wrap(err, "failed to create github client for issue comment event")
	}

	if strings.HasSuffix(ghc.GetAuthor(), "[bot]") {
		ghc.Logger.Debug().Msg("Issue was created by a bot")
		return nil
	}

	// if err = ghc.BranchRequiredStatusChecks([]*github.RequiredStatusCheck{
	// 	{
	// 		Context: status.PR_Check_Title.String(),
	// 		AppID:   github.Int64(config.AppID),
	// 	},
	// 	{
	// 		Context: status.PR_Check_SizeChanges.String(),
	// 		AppID:   github.Int64(config.AppID),
	// 	},
	// 	{
	// 		Context: status.PR_Labeler.String(),
	// 		AppID:   github.Int64(config.AppID),
	// 	},
	// 	{
	// 		Context: status.PR_Check_Commits.String(),
	// 		AppID:   github.Int64(config.AppID),
	// 	},
	// }); err != nil {
	// 	ghc.Logger.Error().Err(err).Msg("Failed to set branch protection")
	// }

	ghc.Logger.Debug().Msgf("Event action is %s in Handle PullRequest", event.GetAction())

	core := &corePR{
		ghc:            ghc,
		eDB:            db.EventDBNew(db.DBEvent),
		event:          event,
		labelsCategory: &[]string{},
		labelsType:     &[]string{},
	}

	core.commitSHA = event.GetPullRequest().GetHead().GetSHA()
	if core.commitSHA == "" {
		ghc.Logger.Error().Msg("Failed to get commit SHA")
		return nil
	}

	// Generic Action
	switch event.GetAction() {
	case "opened", "edited", "synchronize":

		// ? Init status
		// At this instant, All status are pending
		core.PR_Check_Title = status.NewStatus(ghc, status.PR_Check_Title, core.commitSHA)
		core.PR_Check_commits = status.NewStatus(ghc, status.PR_Check_Commits, core.commitSHA)
		core.PR_Labeler = status.NewStatus(ghc, status.PR_Labeler, core.commitSHA)
		core.PR_Check_SizeChanges = status.NewStatus(ghc, status.PR_Check_SizeChanges, core.commitSHA)

		// Check if title respect conventional commit
		core.CheckTitle()
		// Check if commits respect conventional commit
		core.CheckCommits()
		// Check if PR respect size
		core.CheckSizePR()
		// Check if author is COMMUNITY
		// core.Community()

		core.WriteDB()

		core.ComputeLabels()

	default:
		return nil
	}

	return nil
}

type corePR struct {
	ghc            *ghclient.GHClient
	eDB            *db.EventDB
	event          github.PullRequestEvent
	commitSHA      string
	labelsCategory *[]string
	labelsType     *[]string

	PR_Check_Title       *status.Status
	PR_Check_commits     *status.Status
	PR_Check_SizeChanges *status.Status
	PR_Labeler           *status.Status
}

// WriteDB Record data in DB.
func (core *corePR) WriteDB() {
	x, err := json.Marshal(db.Event{
		InstallationID: core.ghc.GetInstallationID(),
		RepoOwner:      core.ghc.GetRepoOwner(),
		RepoName:       core.ghc.GetRepoName(),
		LabelsCategory: *core.labelsCategory,
		LabelsType:     *core.labelsType,
	})
	if err != nil {
		core.ghc.Logger.Error().Err(err).Msg("Failed to marshal event")
	} else {
		err = core.eDB.Set([]byte(core.PathDB()), x)
		if err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to set event in DB")
		}
	}
}

// ReadDB Read data from DB.
func (core *corePR) ReadDB() {
	x, err := core.eDB.Get([]byte(core.PathDB()))
	if err != nil {
		core.ghc.Logger.Error().Err(err).Msg("Failed to get event in DB")
	} else {
		dbEvent := db.Event{}
		if err = json.Unmarshal(x, &dbEvent); err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to unmarshal event")
			pretty.Println(string(x))
		} else {
			core.LoadLabels(dbEvent)
		}
	}
}

// GetLabels return the labels.
func (core *corePR) GetLabels() []string {
	x := make([]string, 0)
	x = append(x, *core.labelsCategory...)
	x = append(x, *core.labelsType...)
	return x
}

// Load labels from DB.
func (core *corePR) LoadLabels(d db.Event) {
	core.labelsCategory = &d.LabelsCategory
	core.labelsType = &d.LabelsType
}

// PathDB return the path of the DB.
func (core *corePR) PathDB() string {
	return fmt.Sprintf("%d/%s/%s/%d", core.ghc.GetInstallationID(), core.ghc.GetRepoOwner(), core.ghc.GetRepoName(), core.event.PullRequest.GetNumber())
}

// CheckTitle check if the title is valid
// Check if the PullRequest have a conventional commit title format.
func (core *corePR) CheckTitle() {
	// ? ParseTitle
	MsgPRTitleInvalid := comments.NewCommentMsg(core.ghc, comments.IDPRTitleInvalid, comments.PRTitleInvalidValues{
		Title: core.ghc.GetPullRequest().GetTitle(),
	})
	if MsgPRTitleInvalid == nil {
		core.ghc.Logger.Error().Msg("Failed to create comment")
	}
	PrTitle, err := conventionalcommit.ParseCommit(core.ghc.GetPullRequest().GetTitle())
	if err != nil {
		core.ghc.Logger.Debug().Msg("Error while parsing PR title")
		if err := core.PR_Check_Title.SetState(statustype.Failure); err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to set status")
		}
		MsgPRTitleInvalid.EditIssueComment()
	} else {
		// Scope
		if PrTitle.Scope() != "" {
			MsgPRIssuesLabelNotExists := comments.NewCommentMsg(core.ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
				Label: labeler.LabelScope(PrTitle.Scope()).GetLongName(),
			})
			if MsgPRIssuesLabelNotExists == nil {
				core.ghc.Logger.Error().Msg("Failed to create comment")
			}
			_, err = core.ghc.GetLabel(labeler.LabelScope(PrTitle.Scope()).GetLongName())
			if err != nil {
				core.PR_Labeler.SetState(statustype.Failure)
				if err := MsgPRIssuesLabelNotExists.EditIssueComment(); err != nil {
					core.ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
				}
			} else {
				if _, ok := common.Find(*core.labelsCategory, labeler.LabelScope(PrTitle.Scope()).GetLongName()); !ok {
					*core.labelsCategory = append(*core.labelsCategory, labeler.LabelScope(PrTitle.Scope()).GetLongName())
				}
			}
		}

		// Type
		if v, ok := labeler.FindLabelerType(PrTitle); !ok {
			core.PR_Check_Title.SetState(statustype.Failure)
			MsgPRTitleInvalid.EditIssueComment()
		} else {
			_, err = core.ghc.GetLabel(v.GetLongName())
			if err != nil {
				err = core.ghc.CreateLabel(v.GitHubLabel())
				if err != nil {
					core.PR_Check_Title.SetState(statustype.Failure)
					core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
				} else {
					if _, ok := common.Find(*core.labelsType, labeler.LabelScope(v.GetLongName()).GetLongName()); !ok {
						*core.labelsType = append(*core.labelsType, v.GetLongName())
					}
				}
			} else {
				if _, ok := common.Find(*core.labelsType, labeler.LabelScope(v.GetLongName()).GetLongName()); !ok {
					*core.labelsType = append(*core.labelsType, v.GetLongName())
				}
			}
		}

		// Breaking change
		if PrTitle.IsBreakingChange() {
			_, err := core.ghc.GetLabel(labeler.BreakingChange.GetLongName())
			if err != nil {
				err = core.ghc.CreateLabel(labeler.BreakingChange.GithubLabel())
				if err != nil {
					core.PR_Check_Title.SetState(statustype.Failure)
					core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
				} else {
					if _, ok := common.Find(*core.labelsType, labeler.BreakingChange.GetLongName()); !ok {
						*core.labelsType = append(*core.labelsType, labeler.BreakingChange.GetLongName())
					}
				}
			} else {
				if _, ok := common.Find(*core.labelsType, labeler.BreakingChange.GetLongName()); !ok {
					*core.labelsType = append(*core.labelsType, labeler.BreakingChange.GetLongName())
				}
			}
		}

		if err := core.PR_Check_Title.IsSuccess(); err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to set status")
		}

		// If PR title is valid, remove issue comment
		if core.PR_Check_Title.GetState() == statustype.Success {
			// Remove issue comment if commit message is valid
			if err := MsgPRTitleInvalid.RemoveIssueComment(); err != nil {
				core.ghc.Logger.Error().Err(err).Msg("Failed to remove issue comment")
			}
		}
	}
}

// CheckCommits check if the commits are valid
// Check if the PullRequest have a conventional commit format.
func (core *corePR) CheckCommits() {
	// ? ParseCommits
	commits, err := core.ghc.GetCommits()
	if err != nil {
		core.PR_Check_commits.SetState(statustype.Failure)
		core.ghc.Logger.Error().Err(err).Msg("Failed to get commits")
	} else {
		allCommitsSHA := make([]string, 0)

		for _, commit := range commits {
			MsgPRCommitInvalid := comments.NewCommentMsg(core.ghc, comments.IDPRCommitInvalid, comments.PRCommitInvalidValues{
				CommitMsg: commit.GetCommit().GetMessage(),
				CommitSHA: commit.GetSHA(),
			})
			if MsgPRCommitInvalid == nil {
				core.ghc.Logger.Error().Msg("Failed to create comment")
				continue
			}

			allCommitsSHA = append(allCommitsSHA, commit.GetSHA())

			core.ghc.Logger.Debug().Msgf("Commit message is %s", commit.GetCommit().GetMessage())

			cm, err := conventionalcommit.ParseCommit(commit.GetCommit().GetMessage())
			if err != nil {
				core.ghc.Logger.Error().Str("message", commit.GetCommit().GetMessage()).Str("commitID", commit.GetSHA()).Msg("Commit message is not conventional commit format")
				core.PR_Check_commits.SetState(statustype.Failure)
				if err := MsgPRCommitInvalid.EditIssueComment(); err != nil {
					core.ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
					continue
				}
			} else {
				// Remove issue comment if commit message is valid
				if err := MsgPRCommitInvalid.RemoveIssueComment(); err != nil {
					core.ghc.Logger.Error().Err(err).Msg("Failed to remove issue comment")
					continue
				}

				// Scope
				if cm.Scope() != "" {
					MsgPRIssuesLabelNotExists := comments.NewCommentMsg(core.ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
						Label: labeler.LabelScope(cm.Scope()).GetLongName(),
					})
					if MsgPRIssuesLabelNotExists == nil {
						core.ghc.Logger.Error().Msg("Failed to create comment")
						continue
					}
					_, err = core.ghc.GetLabel(labeler.LabelScope(cm.Scope()).GetLongName())
					if err != nil {
						core.PR_Labeler.SetState(statustype.Failure)
						if eer := MsgPRIssuesLabelNotExists.EditIssueComment(); err != nil {
							core.ghc.Logger.Error().Err(eer).Msg("Failed to edit issue comment")
							continue
						}
					} else {
						if _, ok := common.Find(*core.labelsCategory, labeler.LabelScope(cm.Scope()).GetLongName()); !ok {
							*core.labelsCategory = append(*core.labelsCategory, labeler.LabelScope(cm.Scope()).GetLongName())
						}
					}
				}
				// Type
				if v, ok := labeler.FindLabelerType(cm); !ok {
					core.ghc.Logger.Error().Str("message", commit.GetCommit().GetMessage()).Str("commitID", commit.GetSHA()).Msg("Commit message is not conventional commit format")
					core.PR_Check_commits.SetState(statustype.Failure)
					if err := MsgPRCommitInvalid.EditIssueComment(); err != nil {
						core.ghc.Logger.Error().Err(err).Msg("Failed to edit issue comment")
						continue
					}
				} else {
					_, err = core.ghc.GetLabel(v.GetLongName())
					if err != nil {
						err = core.ghc.CreateLabel(v.GitHubLabel())
						if err != nil {
							core.PR_Labeler.SetState(statustype.Failure)
							core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
						} else {
							if _, ok := common.Find(*core.labelsType, v.GetLongName()); !ok {
								*core.labelsType = append(*core.labelsType, v.GetLongName())
							}
						}
					} else {
						if _, ok := common.Find(*core.labelsType, v.GetLongName()); !ok {
							*core.labelsType = append(*core.labelsType, v.GetLongName())
						}
					}
				}

				// Breaking change
				if cm.IsBreakingChange() {
					_, err := core.ghc.GetLabel(labeler.BreakingChange.GetLongName())
					if err != nil {
						err = core.ghc.CreateLabel(labeler.BreakingChange.GithubLabel())
						if err != nil {
							core.PR_Labeler.SetState(statustype.Failure)
							core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
						} else {
							if _, ok := common.Find(*core.labelsType, labeler.BreakingChange.GetLongName()); !ok {
								*core.labelsType = append(*core.labelsType, labeler.BreakingChange.GetLongName())
							}
						}
					} else {
						if _, ok := common.Find(*core.labelsType, labeler.BreakingChange.GetLongName()); !ok {
							*core.labelsType = append(*core.labelsType, labeler.BreakingChange.GetLongName())
						}
					}
				}

				if err := core.PR_Check_commits.IsSuccess(); err != nil {
					core.ghc.Logger.Error().Err(err).Msg("Failed to set status")
				}
			}
		}

		cts, err := core.ghc.ListComments()
		if err != nil {
			core.ghc.Logger.Error().Err(err).Msg("Failed to get comments")
		} else {
			if len(cts) > 0 {
				for _, comment := range cts {
					if ok, value := comments.ExtraIssueComment(comment.GetBody(), comments.IDPRCommitInvalid, comments.ExtraBotID); ok && comments.IDPRCommitInvalid.IsValid(value) {
						if ok, value := comments.ExtraIssueComment(comment.GetBody(), comments.IDPRCommitInvalid, comments.ExtraCommitID); ok {
							if _, ok := common.Find(allCommitsSHA, value); !ok {
								// Delete commit for invalid commit message with commitID has been deleted
								core.ghc.DeleteComment(comment.GetID())
							}
						}
					}
				}
			}
		}
	}
}

// CheckSizePR check size of PR
// Check if the PR is too big.
func (core *corePR) CheckSizePR() {
	// * Calcul Additions and Deletions
	size := conventionalsizepr.NewPRSize(core.event.PullRequest.GetAdditions(), core.event.PullRequest.GetDeletions())

	MsgPRSizeTooBig := comments.NewCommentMsg(core.ghc, comments.IDPRSizeTooBig, nil)
	if MsgPRSizeTooBig == nil {
		core.ghc.Logger.Error().Msg("Failed to create comment message")
	}

	if size.IsTooBig() {
		MsgPRSizeTooBig.EditIssueComment()
	} else {
		MsgPRSizeTooBig.RemoveIssueComment()
	}

	l := labeler.FindLabelerSize(size.GetSize())
	_, err := core.ghc.GetLabel(l.GetLongName())
	if err != nil {
		err = core.ghc.CreateLabel(l.GithubLabel())
		if err != nil {
			core.PR_Check_SizeChanges.SetState(statustype.Failure)
			core.ghc.Logger.Error().Err(err).Msg("Failed to create label size")
		} else {
			*core.labelsType = append(*core.labelsType, l.GetLongName())
		}
	} else {
		*core.labelsType = append(*core.labelsType, l.GetLongName())
	}

	if err := core.PR_Check_SizeChanges.IsSuccess(); err != nil {
		core.ghc.Logger.Error().Err(err).Msg("Failed to set status")
	}
}

// ComputeLabels compute labels.
func (core *corePR) ComputeLabels() {
	o := make([]string, 0)
	allLabels := make([]string, 0)
	allLabels = append(allLabels, *core.labelsType...)
	allLabels = append(allLabels, *core.labelsCategory...)

	for _, lbl := range core.event.PullRequest.Labels {
		core.ghc.Logger.Debug().Msgf("Label is %s", lbl.GetName())
		if _, ok := common.Find(allLabels, lbl.GetName()); !ok {
			if err := core.ghc.RemoveLabelForIssue(lbl.GetName()); err != nil {
				core.PR_Labeler.SetState(statustype.Failure)
				core.ghc.Logger.Error().Err(err).Msg("Failed to remove label")
			}
		}
		o = append(o, lbl.GetName())
	}

	for _, lbl := range allLabels {
		if _, ok := common.Find(o, lbl); !ok {
			if err := core.ghc.AddLabelToIssue(lbl); err != nil {
				core.PR_Labeler.SetState(statustype.Failure)
				core.ghc.Logger.Error().Err(err).Msg("Failed to add label")
			}
		}
	}

	if err := core.PR_Labeler.IsSuccess(); err != nil {
		core.ghc.Logger.Error().Err(err).Msg("Failed to set status")
	}
}

// Community labels.
func (core *corePR) Community() {
	if _, ok := common.Find([]string{"NONE", "CONTRIBUTOR"}, core.event.GetPullRequest().GetAuthorAssociation()); ok {
		_, err := core.ghc.GetLabel(labeler.LabelerCommunity().GetName())
		if err != nil {
			err = core.ghc.CreateLabel(labeler.LabelerCommunity().GithubLabel())
			if err != nil {
				core.ghc.Logger.Error().Err(err).Msg("Failed to create label")
			} else {
				if _, ok := common.Find(*core.labelsType, labeler.LabelerCommunity().GetName()); !ok {
					*core.labelsType = append(*core.labelsType, labeler.LabelerCommunity().GetName())
				}
			}
		} else {
			if _, ok := common.Find(*core.labelsType, labeler.LabelerCommunity().GetName()); !ok {
				*core.labelsType = append(*core.labelsType, labeler.LabelerCommunity().GetName())
			}
		}
	}
}
