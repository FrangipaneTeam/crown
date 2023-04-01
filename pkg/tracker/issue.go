package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v47/github"

	"github.com/FrangipaneTeam/crown/pkg/config"
	"github.com/FrangipaneTeam/crown/pkg/db"
)

const (
	closed = "closed"
)

type TrackIssue struct {
	core trackCore
	base trackBase
}

func TrackNewIssue(repoOwner, repoName string, installationID, issueID int64, trackIssueURL string) error {
	// Parse TrackIssueURL
	ghRepository, err := parseTrackIssueURL(trackIssueURL)
	if err != nil {
		return err
	}

	pathDB := generatePathDB(TypeIssue, installationID, repoOwner, repoName, issueID)

	// Find if the issue is already tracked
	if ok := db.TrackDB().KeyExists(db.Byte(pathDB)); ok {
		if raw, err := db.TrackDB().Get(db.Byte(pathDB)); err == nil {
			var t trackBase
			if err := json.Unmarshal(raw, &t); err == nil {
				logger.Debug().Msgf("Issue already tracked: %s", pathDB)
				t.AddSourceRepository(repoOwner, repoName, issueID)

				tJSON, err := json.Marshal(t)
				if err != nil {
					return err
				}

				if err = db.TrackDB().Set(db.Byte(pathDB), tJSON); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		// Create TrackIssue
		t := &TrackIssue{
			base: trackBase{
				TargetRepository: ghRepository,
				SourcesRepository: []GithubRepository{
					{
						RepoOwner: repoOwner,
						RepoName:  repoName,
						ID:        issueID,
					},
				},
				InstallationID:   installationID,
				StatusOfLastScan: false,
			},
		}

		t.base.LastScanAt.Now()
		t.base.CreateAt.Now()
		t.base.UpdateAt.Now()
		t.base.ClosedAt.Now()

		tJSON, err := json.Marshal(t.base)
		if err != nil {
			return err
		}

		return db.TrackDB().Set(db.Byte(pathDB), tJSON)
	}

	return nil
}

// newGithubClient returns a new github client.
func (c *TrackIssue) newGithubClient() error {
	if config.AppID == 0 || c.base.InstallationID == 0 {
		return fmt.Errorf("app id or installation id is not set")
	}

	if config.PrivateKey == nil {
		return fmt.Errorf("private key is not set")
	}

	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, c.base.InstallationID, config.PrivateKey)
	if err != nil {
		return err
	}

	c.core.ctx, c.core.cancel = context.WithTimeout(context.Background(), 30*time.Second)

	// Use installation transport with client.
	c.core.ghc = github.NewClient(&http.Client{Transport: itr})

	return nil
}

// githubParams returns the github params for the issue
// Returns the owner, the repo and the issue number.
func (g *GithubRepository) githubParams() (string, string, int) {
	return g.RepoOwner, g.RepoName, int(g.ID)
}

// ScanIsNecessary checks if the issue is already tracked.
func (c *TrackIssue) ScanIsNecessary() bool {
	return c.base.LastScanAt.Time.Before(time.Now().Add(-intervalScanIssue)) || !c.base.StatusOfLastScan
}

// Scan scans the issue.
func (c *TrackIssue) Scan() error {
	logger.Debug().Msgf("Start scan issue %s/%s/%d", c.base.TargetRepository.RepoOwner, c.base.TargetRepository.RepoName, c.base.TargetRepository.ID)

	var newUpdate bool

	if c.core.ghc == nil {
		if err := c.newGithubClient(); err != nil {
			c.base.StatusOfLastScan = false
			return err
		}
	}

	timeline, _, err := c.core.ghc.Issues.ListIssueTimeline(c.core.ctx, c.base.TargetRepository.RepoOwner, c.base.TargetRepository.RepoName, int(c.base.TargetRepository.ID), nil)
	if err != nil {
		return err
	}

	if c.base.Timeline == nil {
		c.base.Timeline = make([]GithubTimeline, 0)
	}

	if len(c.base.Timeline) != len(timeline) {
		for _, event := range timeline {
			if c.base.LastScanAt.Time.After(event.GetCreatedAt()) {
				continue
			}
			if !c.base.IsExist(event.GetID()) {
				x := GithubTimeline{
					Event:    event.GetEvent(),
					CommitID: event.GetCommitID(),
					Body:     event.GetBody(),
					Message:  event.GetMessage(),
					State:    event.GetState(),

					Rename: GithubRename{
						From: event.GetRename().GetFrom(),
						To:   event.GetRename().GetTo(),
					},
				}
				x.CreatedAt.Time = event.GetCreatedAt()
				x.SubmittedAt.Time = event.GetSubmittedAt()

				c.base.Timeline = append(c.base.Timeline, x)
			}
		}
	}

	if c.base.Status == "" || c.base.Title == "" {
		issue, _, err := c.core.ghc.Issues.Get(c.core.ctx, c.base.TargetRepository.RepoOwner, c.base.TargetRepository.RepoName, int(c.base.TargetRepository.ID))
		if err != nil {
			return err
		}
		c.base.Status = issue.GetState()
		c.base.Title = issue.GetTitle()
	}

	for _, event := range timeline {
		// if evenc.CreatedAt after last scan
		// ignore event
		if c.base.LastScanAt.Time.After(event.GetCreatedAt()) {
			continue
		}

		switch event.GetEvent() {
		case "commented":
			newUpdate = true
		case closed:
			c.base.Status = closed
			newUpdate = true
		}
	}

	if newUpdate {
		logger.Debug().Msgf("New update for issue %s/%s/%d", c.base.TargetRepository.RepoOwner, c.base.TargetRepository.RepoName, c.base.TargetRepository.ID)
		if c.IsClose() {
			if err = c.PublishMessageClosed(); err != nil {
				logger.Error().Err(err).Msg("error publish message closed")
			}
		} else {
			if err = c.PublishMessageStatus(); err != nil {
				logger.Error().Err(err).Msg("error publish message status")
			}
		}
	}

	c.base.LastScanAt.Time = time.Now()
	c.base.StatusOfLastScan = true
	logger.Debug().Msgf("End scan issue %s/%s/%d", c.base.TargetRepository.RepoOwner, c.base.TargetRepository.RepoName, c.base.TargetRepository.ID)
	return nil
}

// IsClose checks if the issue is closed.
func (c *TrackIssue) IsClose() bool {
	return c.base.Status == closed
}

// PublishMessageClosed publish the message when the issue is closed.
func (c *TrackIssue) PublishMessageClosed() error {
	// TODO add real message
	for _, source := range c.base.SourcesRepository {
		repoO, repoN, issueID := source.githubParams()
		if _, _, err := c.core.ghc.Issues.CreateComment(c.core.ctx, repoO, repoN, issueID, &github.IssueComment{
			Body: github.String("Issue closed"),
		}); err != nil {
			continue
		}
	}
	return nil
}

// PublishMessageStatus publish the status of the scan.
func (c *TrackIssue) PublishMessageStatus() error {
	// TODO add real message
	for _, source := range c.base.SourcesRepository {
		repoO, repoN, issueID := source.githubParams()
		if _, _, err := c.core.ghc.Issues.CreateComment(c.core.ctx, repoO, repoN, issueID, &github.IssueComment{
			Body: github.String("Status: " + c.base.Status),
		}); err != nil {
			continue
		}
	}
	return nil
}
