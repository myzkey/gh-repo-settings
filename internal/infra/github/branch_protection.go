package github

import (
	"context"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// GetBranchProtection fetches branch protection rules
func (c *Client) GetBranchProtection(ctx context.Context, branch string) (*BranchProtectionData, error) {
	var data BranchProtectionData
	if err := c.getJSON(ctx, c.branchPath(branch, "protection"), &data); err != nil {
		// Check if branch protection doesn't exist (404)
		var apiErr *apperrors.APIError
		if apperrors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return nil, apperrors.ErrBranchNotProtected
		}
		return nil, err
	}
	return &data, nil
}

// UpdateBranchProtection updates branch protection rules
func (c *Client) UpdateBranchProtection(ctx context.Context, branch string, settings *BranchProtectionSettings) error {
	// Build the protection payload
	payload := map[string]interface{}{
		"enforce_admins":          settings.EnforceAdmins != nil && *settings.EnforceAdmins,
		"required_linear_history": settings.RequireLinearHistory != nil && *settings.RequireLinearHistory,
		"allow_force_pushes":      settings.AllowForcePushes != nil && *settings.AllowForcePushes,
		"allow_deletions":         settings.AllowDeletions != nil && *settings.AllowDeletions,
		"restrictions":            nil,
	}

	// Required pull request reviews
	if settings.RequiredReviews != nil || settings.DismissStaleReviews != nil || settings.RequireCodeOwnerReviews != nil {
		reviews := map[string]interface{}{}
		if settings.RequiredReviews != nil {
			reviews["required_approving_review_count"] = *settings.RequiredReviews
		}
		if settings.DismissStaleReviews != nil {
			reviews["dismiss_stale_reviews"] = *settings.DismissStaleReviews
		}
		if settings.RequireCodeOwnerReviews != nil {
			reviews["require_code_owner_reviews"] = *settings.RequireCodeOwnerReviews
		}
		payload["required_pull_request_reviews"] = reviews
	} else {
		payload["required_pull_request_reviews"] = nil
	}

	// Required status checks
	if settings.RequireStatusChecks != nil && *settings.RequireStatusChecks {
		checks := map[string]interface{}{
			"strict": settings.StrictStatusChecks != nil && *settings.StrictStatusChecks,
		}
		if len(settings.StatusChecks) > 0 {
			checks["contexts"] = settings.StatusChecks
		} else {
			checks["contexts"] = []string{}
		}
		payload["required_status_checks"] = checks
	} else {
		payload["required_status_checks"] = nil
	}

	_, err := c.callJSON(ctx, httpPut, c.branchPath(branch, "protection"), payload)
	return err
}
