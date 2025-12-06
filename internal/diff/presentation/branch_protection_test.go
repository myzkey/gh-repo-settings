package presentation

import (
	"strings"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
)

func ptr[T any](v T) *T {
	return &v
}

func TestFormatBranchRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     *config.BranchRule
		contains []string
		equals   string
	}{
		{
			name:   "empty rule returns 'new protection'",
			rule:   &config.BranchRule{},
			equals: "new protection",
		},
		{
			name: "required_reviews only",
			rule: &config.BranchRule{
				RequiredReviews: ptr(2),
			},
			contains: []string{"required_reviews=2"},
		},
		{
			name: "dismiss_stale_reviews true",
			rule: &config.BranchRule{
				DismissStaleReviews: ptr(true),
			},
			contains: []string{"dismiss_stale_reviews=true"},
		},
		{
			name: "dismiss_stale_reviews false (not included)",
			rule: &config.BranchRule{
				DismissStaleReviews: ptr(false),
			},
			equals: "new protection",
		},
		{
			name: "require_code_owner true",
			rule: &config.BranchRule{
				RequireCodeOwner: ptr(true),
			},
			contains: []string{"require_code_owner=true"},
		},
		{
			name: "strict_status_checks true",
			rule: &config.BranchRule{
				StrictStatusChecks: ptr(true),
			},
			contains: []string{"strict_status_checks=true"},
		},
		{
			name: "status_checks list",
			rule: &config.BranchRule{
				StatusChecks: []string{"ci", "lint"},
			},
			contains: []string{"status_checks=", "ci", "lint"},
		},
		{
			name: "enforce_admins true",
			rule: &config.BranchRule{
				EnforceAdmins: ptr(true),
			},
			contains: []string{"enforce_admins=true"},
		},
		{
			name: "require_linear_history true",
			rule: &config.BranchRule{
				RequireLinearHistory: ptr(true),
			},
			contains: []string{"require_linear_history=true"},
		},
		{
			name: "require_signed_commits true",
			rule: &config.BranchRule{
				RequireSignedCommits: ptr(true),
			},
			contains: []string{"require_signed_commits=true"},
		},
		{
			name: "allow_force_pushes true",
			rule: &config.BranchRule{
				AllowForcePushes: ptr(true),
			},
			contains: []string{"allow_force_pushes=true"},
		},
		{
			name: "allow_deletions true",
			rule: &config.BranchRule{
				AllowDeletions: ptr(true),
			},
			contains: []string{"allow_deletions=true"},
		},
		{
			name: "multiple settings",
			rule: &config.BranchRule{
				RequiredReviews:     ptr(2),
				DismissStaleReviews: ptr(true),
				RequireCodeOwner:    ptr(true),
				EnforceAdmins:       ptr(true),
			},
			contains: []string{
				"required_reviews=2",
				"dismiss_stale_reviews=true",
				"require_code_owner=true",
				"enforce_admins=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBranchRule(tt.rule)

			if tt.equals != "" {
				if result != tt.equals {
					t.Errorf("FormatBranchRule() = %q, want %q", result, tt.equals)
				}
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("FormatBranchRule() = %q, want to contain %q", result, substr)
				}
			}
		})
	}
}

func TestFormatBranchRule_OutputFormat(t *testing.T) {
	t.Run("output is wrapped in braces", func(t *testing.T) {
		rule := &config.BranchRule{
			RequiredReviews: ptr(1),
		}
		result := FormatBranchRule(rule)

		if !strings.HasPrefix(result, "{") {
			t.Errorf("expected output to start with '{', got %q", result)
		}
		if !strings.HasSuffix(result, "}") {
			t.Errorf("expected output to end with '}', got %q", result)
		}
	})

	t.Run("multiple parts are comma-separated", func(t *testing.T) {
		rule := &config.BranchRule{
			RequiredReviews:     ptr(2),
			DismissStaleReviews: ptr(true),
		}
		result := FormatBranchRule(rule)

		if !strings.Contains(result, ", ") {
			t.Errorf("expected comma-separated parts, got %q", result)
		}
	})
}

func TestFormatBranchRule_BooleanFalseNotIncluded(t *testing.T) {
	// All boolean fields set to false should result in "new protection"
	rule := &config.BranchRule{
		DismissStaleReviews:  ptr(false),
		RequireCodeOwner:     ptr(false),
		StrictStatusChecks:   ptr(false),
		EnforceAdmins:        ptr(false),
		RequireLinearHistory: ptr(false),
		RequireSignedCommits: ptr(false),
		AllowForcePushes:     ptr(false),
		AllowDeletions:       ptr(false),
	}

	result := FormatBranchRule(rule)

	if result != "new protection" {
		t.Errorf("expected 'new protection' for all false booleans, got %q", result)
	}
}

func TestFormatBranchRule_NilValues(t *testing.T) {
	// nil values should not be included
	rule := &config.BranchRule{
		RequiredReviews:     nil,
		DismissStaleReviews: nil,
		StatusChecks:        nil,
	}

	result := FormatBranchRule(rule)

	if result != "new protection" {
		t.Errorf("expected 'new protection' for all nil values, got %q", result)
	}
}

func TestFormatBranchRule_EmptyStatusChecks(t *testing.T) {
	// Empty status checks slice should not be included
	rule := &config.BranchRule{
		StatusChecks: []string{},
	}

	result := FormatBranchRule(rule)

	if result != "new protection" {
		t.Errorf("expected 'new protection' for empty status checks, got %q", result)
	}
}
