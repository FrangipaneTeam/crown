package status

import (
	"github.com/google/go-github/v47/github"

	"github.com/FrangipaneTeam/crown/pkg/ghclient"
	status "github.com/FrangipaneTeam/crown/pkg/statustype"
)

//nolint:revive,stylecheck
const (
	PR_Check_Title StatusCategory = 99 << iota
	PR_Check_Commits
	PR_Check_SizeChanges
	PR_Labeler

	Issue_Check_Title
	Issue_Labeler
)

//go:generate stringer -type=StatusCategory
type StatusCategory int //nolint:revive

type statusMessages struct {
	Success string
	Failure string
	Pending string
}

type Status struct {
	ghc            *ghclient.GHClient
	repoStatus     github.RepoStatus
	statusMessages statusMessages
	commitSHA      string
}

// ListOfStatuses is a list of all statuses.
var listOfStatuses = map[StatusCategory]Status{
	PR_Check_Title: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(PR_Check_Title.String()),
		},
		statusMessages: statusMessages{
			Success: "PR title is valid",
			Failure: "PR title is invalid",
			Pending: "Checking PR title",
		},
	},
	PR_Check_Commits: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(PR_Check_Commits.String()),
		},
		statusMessages: statusMessages{
			Success: "PR has valid commits",
			Failure: "PR has invalid commits",
			Pending: "Checking PR commits",
		},
	},
	PR_Check_SizeChanges: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(PR_Check_SizeChanges.String()),
		},
		statusMessages: statusMessages{
			Success: "Successfully checked size changes",
			Failure: "Failed to check size changes",
			Pending: "Checking size changes",
		},
	},
	PR_Labeler: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(PR_Labeler.String()),
		},
		statusMessages: statusMessages{
			Success: "Successfully labeled PR",
			Failure: "Failed to label PR",
			Pending: "Labeling PR",
		},
	},
	Issue_Check_Title: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(Issue_Check_Title.String()),
		},
		statusMessages: statusMessages{
			Success: "Issue title is valid",
			Failure: "Issue title is invalid",
			Pending: "Checking issue title",
		},
	},
	Issue_Labeler: {
		repoStatus: github.RepoStatus{
			State:   github.String(status.Pending.String()),
			Context: github.String(Issue_Labeler.String()),
		},
		statusMessages: statusMessages{
			Success: "Successfully labeled issue",
			Failure: "Failed to label issue",
			Pending: "Labeling issue",
		},
	},
}

// NewStatus returns a new status.
func NewStatus(ghc *ghclient.GHClient, category StatusCategory, commitSHA string) *Status {
	if x, ok := listOfStatuses[category]; ok {
		s := &x
		s.ghc = ghc
		s.commitSHA = commitSHA
		if err := s.SetState(status.Pending); err != nil {
			return nil
		}
		return s
	}
	return nil
}

// SetState sets the state of the status.
func (s *Status) SetState(state status.Status) error {
	s.repoStatus.State = github.String(state.String())

	switch state {
	case status.Success:
		s.repoStatus.Description = github.String(s.statusMessages.Success)
	case status.Failure, status.Error:
		s.repoStatus.Description = github.String(s.statusMessages.Failure)
	case status.Pending:
		s.repoStatus.Description = github.String(s.statusMessages.Pending)
	}

	return s.ghc.EditStatus(s.repoStatus, s.commitSHA)
}

// IsSuccess sets the state of the status to success if state is not failure or error.
func (s *Status) IsSuccess() error {
	if status.Status(*s.repoStatus.State) == status.Failure || status.Status(*s.repoStatus.State) == status.Error {
		return nil
	}
	return s.SetState(status.Success)
}

// GetState returns the state of the status.
func (s *Status) GetState() status.Status {
	return status.Status(*s.repoStatus.State)
}
