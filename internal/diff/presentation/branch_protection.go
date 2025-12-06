package presentation

import (
	"fmt"
	"strings"

	"github.com/myzkey/gh-repo-settings/internal/config"
)

// FormatBranchRule formats a branch rule for display
// This is presentation logic for human-readable output
func FormatBranchRule(rule *config.BranchRule) string {
	var parts []string
	if rule.RequiredReviews != nil {
		parts = append(parts, fmt.Sprintf("required_reviews=%d", *rule.RequiredReviews))
	}
	if rule.DismissStaleReviews != nil && *rule.DismissStaleReviews {
		parts = append(parts, "dismiss_stale_reviews=true")
	}
	if rule.RequireCodeOwner != nil && *rule.RequireCodeOwner {
		parts = append(parts, "require_code_owner=true")
	}
	if rule.StrictStatusChecks != nil && *rule.StrictStatusChecks {
		parts = append(parts, "strict_status_checks=true")
	}
	if len(rule.StatusChecks) > 0 {
		parts = append(parts, fmt.Sprintf("status_checks=%v", rule.StatusChecks))
	}
	if rule.EnforceAdmins != nil && *rule.EnforceAdmins {
		parts = append(parts, "enforce_admins=true")
	}
	if rule.RequireLinearHistory != nil && *rule.RequireLinearHistory {
		parts = append(parts, "require_linear_history=true")
	}
	if rule.RequireSignedCommits != nil && *rule.RequireSignedCommits {
		parts = append(parts, "require_signed_commits=true")
	}
	if rule.AllowForcePushes != nil && *rule.AllowForcePushes {
		parts = append(parts, "allow_force_pushes=true")
	}
	if rule.AllowDeletions != nil && *rule.AllowDeletions {
		parts = append(parts, "allow_deletions=true")
	}
	if len(parts) == 0 {
		return "new protection"
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}
