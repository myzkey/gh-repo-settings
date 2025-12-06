package service

import "github.com/myzkey/gh-repo-settings/internal/diff/domain/model"

// CompareBranchRule compares current and desired branch protection rules
// This is a pure domain service with no infrastructure dependencies
func CompareBranchRule(
	branch string,
	current model.BranchProtectionCurrent,
	desired model.BranchProtectionDesired,
) []model.Change {
	var changes []model.Change
	prefix := branch + "."

	// Required reviews (int comparison)
	if desired.RequiredReviews != nil && *desired.RequiredReviews != current.RequiredReviews {
		changes = append(changes, model.NewUpdateChange(
			model.CategoryBranchProtection,
			prefix+"required_reviews",
			current.RequiredReviews,
			*desired.RequiredReviews,
		))
	}

	// Boolean fields
	addBoolChange(&changes, prefix+"dismiss_stale_reviews", desired.DismissStaleReviews, current.DismissStaleReviews)
	addBoolChange(&changes, prefix+"require_code_owner", desired.RequireCodeOwner, current.RequireCodeOwner)
	addBoolChange(&changes, prefix+"strict_status_checks", desired.StrictStatusChecks, current.StrictStatusChecks)
	addBoolChange(&changes, prefix+"enforce_admins", desired.EnforceAdmins, current.EnforceAdmins)
	addBoolChange(&changes, prefix+"require_linear_history", desired.RequireLinearHistory, current.RequireLinearHistory)
	addBoolChange(&changes, prefix+"allow_force_pushes", desired.AllowForcePushes, current.AllowForcePushes)
	addBoolChange(&changes, prefix+"allow_deletions", desired.AllowDeletions, current.AllowDeletions)
	addBoolChange(&changes, prefix+"require_signed_commits", desired.RequireSignedCommits, current.RequireSignedCommits)

	// Status checks (slice comparison)
	if desired.StatusChecks != nil && !stringSliceEqual(desired.StatusChecks, current.StatusChecks) {
		changes = append(changes, model.NewUpdateChange(
			model.CategoryBranchProtection,
			prefix+"status_checks",
			current.StatusChecks,
			desired.StatusChecks,
		))
	}

	return changes
}

// addBoolChange adds a change if the desired value differs from current
func addBoolChange(changes *[]model.Change, key string, desired *bool, current bool) {
	if desired == nil {
		return
	}
	if *desired == current {
		return
	}
	*changes = append(*changes, model.NewUpdateChange(
		model.CategoryBranchProtection,
		key,
		current,
		*desired,
	))
}

// stringSliceEqual compares two string slices for equality
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
