package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FrangipaneTeam/crown/handlers/comments"
	"github.com/FrangipaneTeam/crown/handlers/status"
	"github.com/FrangipaneTeam/crown/pkg/db"
	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/labeler"
	"github.com/FrangipaneTeam/crown/pkg/slashcommand"
	"github.com/FrangipaneTeam/crown/pkg/statustype"
	"github.com/azrod/common-go"
	"github.com/google/go-github/v47/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

type IssueCommentHandler struct {
	githubapp.ClientCreator
}

// Handles returns the list of events this handler handles
func (h *IssueCommentHandler) Handles() []string {
	return []string{"issue_comment"}
}

type coreIssueComment struct {
	ghc            *ghclient.GHClient
	eDB            *db.EventDB
	event          github.IssueCommentEvent
	labelsCategory *[]string
	labelsType     *[]string

	PR_Check_Title       *status.Status
	PR_Check_commits     *status.Status
	PR_Check_SizeChanges *status.Status
	PR_Labeler           *status.Status
}

// Handle processes the event
func (h *IssueCommentHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
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

	core := &coreIssueComment{
		ghc:            ghc,
		eDB:            db.EventDBNew(db.DBEvent),
		event:          event,
		labelsCategory: &[]string{},
		labelsType:     &[]string{},
	}

	ghc.Logger.Debug().Msgf("Event action is %s", event.GetAction())

	dbEvent := db.Event{}
	x, err := core.eDB.Get([]byte(core.PathDB()))
	if err != nil {
		ghc.Logger.Error().Err(err).Msg("Failed to get event in DB")
	} else {
		if err = json.Unmarshal(x, &dbEvent); err != nil {
			ghc.Logger.Error().Err(err).Msg("Failed to unmarshal event")
		} else {
			core.LoadLabels(dbEvent)
		}
	}

	commentBody := event.GetComment().GetBody()
	commentID := event.GetComment().GetID()
	user := event.GetComment().GetUser()

	if ok, _ := ghc.IsInOrganization(user.GetLogin()); ok {
		if foundSlashCommand, cmd, err := slashcommand.FindSlashCommand(commentBody); foundSlashCommand {
			switch cmd := cmd.(type) {
			case slashcommand.SlashCommandLabel:
				ghc.Logger.Debug().Msgf("Found slash command %s with verb %s from %s", cmd.Action, cmd.Verb, user.GetName())
				switch cmd.Action {
				case slashcommand.CommandLabel:
					label := labeler.LabelScope(cmd.Label)
					switch cmd.Verb {
					case slashcommand.VerbAdd:
						if err := ghc.CreateLabel(label.GithubLabel()); err != nil {
							if err := ghc.AddCommentReaction(commentID, "-1"); err != nil {
								ghc.Logger.Err(err).Msg("failed to add reaction")
							}
							return err
						}

						if err := ghc.AddCommentReaction(commentID, "+1"); err != nil {
							ghc.Logger.Err(err).Msg("failed to add reaction")
							return err
						}

						MsgPRIssuesLabelNotExists := comments.NewCommentMsg(ghc, comments.IDIssuesLabelNotExists, comments.IssuesLabelNotExistsValues{
							Label: label.GetLongName(),
						})

						// Remove comment if label added is in comment
						MsgPRIssuesLabelNotExists.RemoveIssueComment()

						*core.labelsCategory = append(*core.labelsCategory, label.GetLongName())

					case slashcommand.VerbRemove:
						for i, l := range *core.labelsCategory {
							if l == label.GetLongName() {
								*core.labelsCategory = append((*core.labelsCategory)[:i], (*core.labelsCategory)[i+1:]...)
								break
							}
						}
					}

				}
			}
			core.ComputeLabels()
		} else if err != nil {
			return err
		}
	} else {
		ghc.Logger.Debug().Msgf("User %s is not in %s org", user.GetLogin(), ghc.GetOrg())
	}

	return nil
}

// GetLabels return the labels
func (core *coreIssueComment) GetLabels() []string {
	x := make([]string, 0)
	x = append(x, *core.labelsCategory...)
	x = append(x, *core.labelsType...)
	return x
}

// Load labels from DB
func (core *coreIssueComment) LoadLabels(d db.Event) {
	core.labelsCategory = &d.LabelsCategory
	core.labelsType = &d.LabelsType
}

// PathDB return the path of the DB
func (core *coreIssueComment) PathDB() string {
	return fmt.Sprintf("%d/%s/%s/%d", core.ghc.GetInstallationID(), core.ghc.GetRepoOwner(), core.ghc.GetRepoName(), core.event.GetIssue().GetNumber())
}

// ComputeLabels compute labels
func (core *coreIssueComment) ComputeLabels() {
	o := make([]string, 0)
	allLabels := make([]string, 0)
	allLabels = append(allLabels, *core.labelsType...)
	allLabels = append(allLabels, *core.labelsCategory...)

	for _, lbl := range core.event.GetIssue().Labels {
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
