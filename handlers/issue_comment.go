package handlers

import (
	"context"
	"encoding/json"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	"github.com/FrangipaneTeam/crown/pkg/slashcommand"
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

	ghc.Logger.Debug().Msgf("Event action is %s", event.GetAction())

	commentBody := event.Comment.GetBody()
	commentID := event.Comment.GetID()
	user := event.Comment.GetUser()
	if ok, _ := ghc.IsInFrangipaneOrg(user.GetLogin()); ok {
		if foundSlashCommand, cmd := slashcommand.FindSlashCommand(commentBody, commentID, int64(event.Issue.GetNumber())); foundSlashCommand {
			ghc.Logger.Debug().Msgf("Found slash command %s with verb %s and description %s from %s", cmd.Command, cmd.Verb, cmd.Desc, user.GetName())
			if err := slashcommand.ExecuteSlashCommand(ghc, cmd); err != nil {
				return err
			}
		}
	} else {
		ghc.Logger.Debug().Msgf("User %s is not in %s org", user.GetLogin(), ghc.GetOrg())
	}

	return nil
}
