//nolint:deadcode,varcheck
package tracker

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/rs/zerolog"

	"github.com/FrangipaneTeam/crown/pkg/common"
)

var logger zerolog.Logger

type TypeResource string

const (
	intervalScanIssue = 10 * time.Minute
	intervalScanPR    = 10 * time.Minute

	TypeIssue   TypeResource = "issues"
	TypePR      TypeResource = "pullrequests"
	TypeRelease TypeResource = "releases"
)

type GithubRepository struct {
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
	ID        int64  `json:"id"`
}

// GetRepoOwner returns the owner of the repository.
func (g *GithubRepository) GetRepoOwner() string {
	return g.RepoOwner
}

// GetRepoName returns the name of the repository.
func (g *GithubRepository) GetRepoName() string {
	return g.RepoName
}

// GetID returns the ID of the repository.
func (g *GithubRepository) GetID() int64 {
	return g.ID
}

type GithubRename struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type GithubTimeline struct {
	// ID is the unique identifier of the timeline event.
	ID int64 `json:"id,omitempty"`

	// The commit message.
	Message string `json:"message,omitempty"`

	// Event identifies the actual type of Event that occurred. Possible values
	// are:
	//
	//     assigned
	//       The issue was assigned to the assignee.
	//
	//     closed
	//       The issue was closed by the actor. When the commit_id is present, it
	//       identifies the commit that closed the issue using "closes / fixes #NN"
	//       syntax.
	//
	//     commented
	//       A comment was added to the issue.
	//
	//     committed
	//       A commit was added to the pull request's 'HEAD' branch. Only provided
	//       for pull requests.
	//
	//     cross-referenced
	//       The issue was referenced from another issue. The 'source' attribute
	//       contains the 'id', 'actor', and 'url' of the reference's source.
	//
	//     demilestoned
	//       The issue was removed from a milestone.
	//
	//     head_ref_deleted
	//       The pull request's branch was deleted.
	//
	//     head_ref_restored
	//       The pull request's branch was restored.
	//
	//     labeled
	//       A label was added to the issue.
	//
	//     locked
	//       The issue was locked by the actor.
	//
	//     mentioned
	//       The actor was @mentioned in an issue body.
	//
	//     merged
	//       The issue was merged by the actor. The 'commit_id' attribute is the
	//       SHA1 of the HEAD commit that was merged.
	//
	//     milestoned
	//       The issue was added to a milestone.
	//
	//     referenced
	//       The issue was referenced from a commit message. The 'commit_id'
	//       attribute is the commit SHA1 of where that happened.
	//
	//     renamed
	//       The issue title was changed.
	//
	//     reopened
	//       The issue was reopened by the actor.
	//
	//     reviewed
	//       The pull request was reviewed.
	//
	//     subscribed
	//       The actor subscribed to receive notifications for an issue.
	//
	//     unassigned
	//       The assignee was unassigned from the issue.
	//
	//     unlabeled
	//       A label was removed from the issue.
	//
	//     unlocked
	//       The issue was unlocked by the actor.
	//
	//     unsubscribed
	//       The actor unsubscribed to stop receiving notifications for an issue.
	//
	Event string `json:"event,omitempty"`

	// The string SHA of a commit that referenced this Issue or Pull Request.
	CommitID string `json:"commit_id,omitempty"`
	// The timestamp indicating when the event occurred.
	CreatedAt Timestamp `json:"created_at,omitempty"`

	// An object containing rename details including 'from' and 'to' attributes.
	// Only provided for 'renamed' events.
	Rename GithubRename `json:"rename,omitempty"`

	// The state of a submitted review. Can be one of: 'commented',
	// 'changes_requested' or 'approved'.
	// Only provided for 'reviewed' events.
	State string `json:"state,omitempty"`

	// The review summary text.
	Body        string    `json:"body,omitempty"`
	SubmittedAt Timestamp `json:"submitted_at,omitempty"`
}

// generatePathDB generate the path of the key in the database.
func generatePathDB(x TypeResource, installationID int64, repoOwner, repoName string, repoID int64) string {
	// format of the key in the database is
	// <TypeResource>/<InstallationID>/<RepoOwner>/<RepoName>/<RepoID>
	// example: issue/1234556789/owner/repo/8
	return fmt.Sprintf("%s/%d/%s/%s/%d", x, installationID, repoOwner, repoName, repoID)
}

func (t *trackBase) PathDB(x TypeResource) string {
	// format of the key in the database is
	// <TypeResource>/<InstallationID>/<RepoOwner>/<RepoName>/<RepoID>
	// example: issue/1234556789/owner/repo/8
	return generatePathDB(x, t.GetInstallationID(), t.GetTargetRepository().GetRepoOwner(), t.GetTargetRepository().GetRepoName(), t.GetTargetRepository().GetID())
}

type trackCore struct {
	// Core
	ghc    *github.Client
	ctx    context.Context
	cancel context.CancelFunc
}
type trackBase struct {
	// LastScanAt is the timestamp of the last scan
	LastScanAt Timestamp `json:"last_scan_at"`
	// StatusOfLastScan is the status of the last scan
	StatusOfLastScan bool `json:"status_of_last_scan"`

	// SourcesRepository is the list of the repository where the issue/pr is referenced
	SourcesRepository []GithubRepository `json:"sources_repository"`

	// InstallationID is the installation id of the github app
	InstallationID int64 `json:"installation_id"`

	// TargetRepository is the tracked repository
	TargetRepository GithubRepository `json:"target_repository"`

	// Issue/PR
	Status   string    `json:"status"`
	Title    string    `json:"title"`
	CreateAt Timestamp `json:"create_at"`
	UpdateAt Timestamp `json:"update_at"`
	ClosedAt Timestamp `json:"closed_at"`
	Timeline []GithubTimeline
}

// IsExist check if the ID is already in the list.
func (t *trackBase) IsExist(iD int64) bool {
	for _, x := range t.SourcesRepository {
		if x.GetID() == iD {
			return true
		}
	}
	return false
}

// GetLastScanAt return the last scan timestamp.
func (t *trackBase) GetLastScanAt() Timestamp {
	return t.LastScanAt
}

// GetStatusOfLastScan return the status of the last scan.
func (t *trackBase) GetStatusOfLastScan() bool {
	return t.StatusOfLastScan
}

// GetSourcesRepository return the list of the repository where the issue/pr is referenced.
func (t *trackBase) GetSourcesRepository() *[]GithubRepository {
	return &t.SourcesRepository
}

// GetInstallationID return the installation id.
func (t *trackBase) GetInstallationID() int64 {
	return t.InstallationID
}

// GetTargetRepository return the tracked repository.
func (t *trackBase) GetTargetRepository() *GithubRepository {
	return &t.TargetRepository
}

// GetStatus return the status of the issue/pr.
func (t *trackBase) GetStatus() string {
	return t.Status
}

// GetTitle return the title of the issue/pr.
func (t *trackBase) GetTitle() string {
	return t.Title
}

// GetCreateAt return the creation timestamp of the issue/pr.
func (t *trackBase) GetCreateAt() Timestamp {
	return t.CreateAt
}

// GetUpdateAt return the last update timestamp of the issue/pr.
func (t *trackBase) GetUpdateAt() Timestamp {
	return t.UpdateAt
}

// GetClosedAt return the closed timestamp of the issue/pr.
func (t *trackBase) GetClosedAt() Timestamp {
	return t.ClosedAt
}

// GetTimeline return the timeline of the issue/pr.
func (t *trackBase) GetTimeline() []GithubTimeline {
	return t.Timeline
}

type TrackPR struct {
	trackBase

	// IsMerged return true if the PR is merged
	IsMerged func() bool
}

// Init initialize the tracker.
func Init(x zerolog.Logger) {
	logger = x
}

// parseTrackIssueURL parse the TrackIssueURL and return the repoOwner, repoName, id and error.
func parseTrackIssueURL(trackIssueURL string) (GithubRepository, error) {
	// The format of the TrackIssueURL is
	// repoOwner/repoName#ID (ex: FrangipaneTeam/crown#1)
	// or https://github.com/repoOwner/repoName/issues/ID (ex: https://github.com/FrangipaneTeam/crown/issues/140)

	regexS := []*regexp.Regexp{
		// First format (repoOwner/repoName#ID)
		regexp.MustCompile(`^(?P<repoOwner>\S+)\/(?P<repoName>\S+)#(?P<idIssue>[0-9]+)$`),
		// Second format ()
		regexp.MustCompile(`^.*\/(?P<repoOwner>\S+)\/(?P<repoName>\S+)\/issues\/(?P<idIssue>[0-9]+)$`),
	}

	// Parse TrackIssueURL
	for _, re := range regexS {
		match := re.FindString(trackIssueURL)
		m := common.ReSubMatchMap(re, match)
		if len(m) == 0 {
			continue
		}

		// id is int64
		id, err := strconv.ParseInt(m["idIssue"], 10, 64)
		if err != nil {
			return GithubRepository{}, err
		}

		return GithubRepository{
			RepoOwner: m["repoOwner"],
			RepoName:  m["repoName"],
			ID:        id,
		}, nil
	}

	return GithubRepository{}, fmt.Errorf("unable to parse TrackIssueURL")
}

// sourceRepositoryAlreadyExist check if the source repository is already in the list.
func (t *trackBase) sourceRepositoryAlreadyExist(repoOwner, repoName string, issueID int64) bool {
	for _, s := range t.SourcesRepository {
		if s.RepoOwner == repoOwner && s.RepoName == repoName && s.ID == issueID {
			return true
		}
	}
	return false
}

// AddSourceRepository add a source repository to the track issue.
func (t *trackBase) AddSourceRepository(repoOwner, repoName string, issueID int64) {
	if !t.sourceRepositoryAlreadyExist(repoOwner, repoName, issueID) {
		t.SourcesRepository = append(t.SourcesRepository, GithubRepository{
			RepoOwner: repoOwner,
			RepoName:  repoName,
			ID:        issueID,
		})
	}
}
