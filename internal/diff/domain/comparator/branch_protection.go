package comparator

import (
	"context"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/service"
	"github.com/myzkey/gh-repo-settings/internal/diff/presentation"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// BranchProtectionGateway provides access to branch protection data
type BranchProtectionGateway interface {
	GetBranchProtection(ctx context.Context, branch string) (model.BranchProtectionCurrent, error)
}

// BranchProtectionComparator compares branch protection rules
// This is an application service that orchestrates domain logic
type BranchProtectionComparator struct {
	gateway BranchProtectionGateway
	rules   map[string]*config.BranchRule
}

// NewBranchProtectionComparator creates a new BranchProtectionComparator
func NewBranchProtectionComparator(gateway BranchProtectionGateway, rules map[string]*config.BranchRule) *BranchProtectionComparator {
	return &BranchProtectionComparator{
		gateway: gateway,
		rules:   rules,
	}
}

// NewBranchProtectionComparatorWithClient creates a comparator with a GitHub client
// This is a convenience constructor that creates the gateway internally
func NewBranchProtectionComparatorWithClient(client github.GitHubClient, rules map[string]*config.BranchRule) *BranchProtectionComparator {
	return &BranchProtectionComparator{
		gateway: &githubBranchProtectionGateway{client: client},
		rules:   rules,
	}
}

// Compare compares the current branch protection with the desired configuration
func (c *BranchProtectionComparator) Compare(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	for branchName, rule := range c.rules {
		current, err := c.gateway.GetBranchProtection(ctx, branchName)
		if err != nil {
			if apperrors.Is(err, apperrors.ErrBranchNotProtected) {
				// Branch protection doesn't exist, will be added
				plan.Add(model.NewAddChange(
					model.CategoryBranchProtection,
					branchName,
					presentation.FormatBranchRule(rule),
				))
				continue
			}
			return nil, err
		}

		// Map config to domain model
		desired := mapBranchRuleToDomain(rule)

		// Use pure domain service for comparison
		branchChanges := service.CompareBranchRule(branchName, current, desired)
		plan.AddAll(branchChanges)
	}

	return plan, nil
}

// mapBranchRuleToDomain converts config.BranchRule to domain model
func mapBranchRuleToDomain(rule *config.BranchRule) model.BranchProtectionDesired {
	return model.BranchProtectionDesired{
		RequiredReviews:      rule.RequiredReviews,
		DismissStaleReviews:  rule.DismissStaleReviews,
		RequireCodeOwner:     rule.RequireCodeOwner,
		StrictStatusChecks:   rule.StrictStatusChecks,
		StatusChecks:         rule.StatusChecks,
		EnforceAdmins:        rule.EnforceAdmins,
		RequireLinearHistory: rule.RequireLinearHistory,
		AllowForcePushes:     rule.AllowForcePushes,
		AllowDeletions:       rule.AllowDeletions,
		RequireSignedCommits: rule.RequireSignedCommits,
	}
}

// githubBranchProtectionGateway is an internal gateway implementation
type githubBranchProtectionGateway struct {
	client github.GitHubClient
}

func (g *githubBranchProtectionGateway) GetBranchProtection(
	ctx context.Context,
	branch string,
) (model.BranchProtectionCurrent, error) {
	data, err := g.client.GetBranchProtection(ctx, branch)
	if err != nil {
		return model.BranchProtectionCurrent{}, err
	}

	return model.BranchProtectionCurrent{
		RequiredReviews:      extractRequiredReviews(data),
		DismissStaleReviews:  extractDismissStaleReviews(data),
		RequireCodeOwner:     extractRequireCodeOwner(data),
		StrictStatusChecks:   extractStrictStatusChecks(data),
		StatusChecks:         extractStatusChecks(data),
		EnforceAdmins:        extractEnforceAdmins(data),
		RequireLinearHistory: extractRequireLinearHistory(data),
		AllowForcePushes:     extractAllowForcePushes(data),
		AllowDeletions:       extractAllowDeletions(data),
		RequireSignedCommits: extractRequireSignedCommits(data),
	}, nil
}

func extractRequiredReviews(data *github.BranchProtectionData) int {
	if data.RequiredPullRequestReviews != nil && data.RequiredPullRequestReviews.RequiredApprovingReviewCount != nil {
		return *data.RequiredPullRequestReviews.RequiredApprovingReviewCount
	}
	return 0
}

func extractDismissStaleReviews(data *github.BranchProtectionData) bool {
	if data.RequiredPullRequestReviews != nil {
		return data.RequiredPullRequestReviews.DismissStaleReviews
	}
	return false
}

func extractRequireCodeOwner(data *github.BranchProtectionData) bool {
	if data.RequiredPullRequestReviews != nil {
		return data.RequiredPullRequestReviews.RequireCodeOwnerReviews
	}
	return false
}

func extractStrictStatusChecks(data *github.BranchProtectionData) bool {
	if data.RequiredStatusChecks != nil && data.RequiredStatusChecks.Strict != nil {
		return *data.RequiredStatusChecks.Strict
	}
	return false
}

func extractStatusChecks(data *github.BranchProtectionData) []string {
	if data.RequiredStatusChecks != nil {
		return data.RequiredStatusChecks.Contexts
	}
	return nil
}

func extractEnforceAdmins(data *github.BranchProtectionData) bool {
	if data.EnforceAdmins != nil {
		return data.EnforceAdmins.Enabled
	}
	return false
}

func extractRequireLinearHistory(data *github.BranchProtectionData) bool {
	if data.RequiredLinearHistory != nil && data.RequiredLinearHistory.Enabled != nil {
		return *data.RequiredLinearHistory.Enabled
	}
	return false
}

func extractAllowForcePushes(data *github.BranchProtectionData) bool {
	if data.AllowForcePushes != nil && data.AllowForcePushes.Enabled != nil {
		return *data.AllowForcePushes.Enabled
	}
	return false
}

func extractAllowDeletions(data *github.BranchProtectionData) bool {
	if data.AllowDeletions != nil && data.AllowDeletions.Enabled != nil {
		return *data.AllowDeletions.Enabled
	}
	return false
}

func extractRequireSignedCommits(data *github.BranchProtectionData) bool {
	if data.RequiredSignatures != nil {
		return data.RequiredSignatures.Enabled
	}
	return false
}
