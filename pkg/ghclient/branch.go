package ghclient

import (
	"github.com/google/go-github/v47/github"
)

//go:generate mockgen -source=branch.go -destination=branch_mock.go -package=ghclient

// BranchRequiredStatusChecks represents the required status checks for a branch.
func (g *GHClient) BranchRequiredStatusChecks(requiredStatusCheck []*github.RequiredStatusCheck) error {
	_, _, err := g.client.Repositories.UpdateBranchProtection(
		g.context,
		g.repoOwner,
		g.repoName,
		"main",
		&github.ProtectionRequest{
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Checks: requiredStatusCheck,
			},
		})
	return err
}
