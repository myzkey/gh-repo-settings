package github

import (
	"github.com/myzkey/gh-repo-settings/internal/githubopenapi"
)

// Type aliases for generated OpenAPI types.
// These provide backward compatibility while using the generated types.

// RepoData is an alias for the generated FullRepository type.
type RepoData = githubopenapi.FullRepository

// LabelData is an alias for the generated Label type.
type LabelData = githubopenapi.Label

// BranchProtectionData is an alias for the generated BranchProtection type.
type BranchProtectionData = githubopenapi.BranchProtection

// ActionsPermissionsData is an alias for the generated ActionsRepositoryPermissions type.
type ActionsPermissionsData = githubopenapi.ActionsRepositoryPermissions

// ActionsSelectedData is an alias for the generated SelectedActions type.
type ActionsSelectedData = githubopenapi.SelectedActions

// ActionsWorkflowPermissionsData is an alias for the generated ActionsGetDefaultWorkflowPermissions type.
type ActionsWorkflowPermissionsData = githubopenapi.ActionsGetDefaultWorkflowPermissions

// PagesData is an alias for the generated GitHubPage type.
type PagesData = githubopenapi.GitHubPage

// PagesSourceData is an alias for the generated PagesSourceHash type.
type PagesSourceData = githubopenapi.PagesSourceHash

// VariableData is an alias for the generated ActionsVariable type.
type VariableData = githubopenapi.ActionsVariable

// CurrentSettings represents the current GitHub repository settings for JSON output.
// This is a custom type for export functionality, not from OpenAPI.
type CurrentSettings struct {
	Repo             *CurrentRepoSettings          `json:"repo,omitempty"`
	Topics           []string                      `json:"topics,omitempty"`
	Labels           []LabelData                   `json:"labels,omitempty"`
	BranchProtection map[string]*CurrentBranchRule `json:"branch_protection,omitempty"`
	Actions          *CurrentActionsSettings       `json:"actions,omitempty"`
	Pages            *PagesData                    `json:"pages,omitempty"`
	Variables        []VariableData                `json:"variables,omitempty"`
	Secrets          []string                      `json:"secrets,omitempty"`
}

// CurrentRepoSettings represents current repository settings for export.
type CurrentRepoSettings struct {
	Description         string `json:"description,omitempty"`
	Homepage            string `json:"homepage,omitempty"`
	Visibility          string `json:"visibility"`
	AllowMergeCommit    bool   `json:"allow_merge_commit"`
	AllowRebaseMerge    bool   `json:"allow_rebase_merge"`
	AllowSquashMerge    bool   `json:"allow_squash_merge"`
	DeleteBranchOnMerge bool   `json:"delete_branch_on_merge"`
	AllowUpdateBranch   bool   `json:"allow_update_branch"`
}

// CurrentBranchRule represents current branch protection rule for export.
type CurrentBranchRule struct {
	RequiredReviews      *int     `json:"required_reviews,omitempty"`
	DismissStaleReviews  *bool    `json:"dismiss_stale_reviews,omitempty"`
	RequireCodeOwner     *bool    `json:"require_code_owner,omitempty"`
	RequireStatusChecks  *bool    `json:"require_status_checks,omitempty"`
	StrictStatusChecks   *bool    `json:"strict_status_checks,omitempty"`
	StatusChecks         []string `json:"status_checks,omitempty"`
	EnforceAdmins        *bool    `json:"enforce_admins,omitempty"`
	RequireLinearHistory *bool    `json:"require_linear_history,omitempty"`
	AllowForcePushes     *bool    `json:"allow_force_pushes,omitempty"`
	AllowDeletions       *bool    `json:"allow_deletions,omitempty"`
}

// CurrentActionsSettings represents current actions settings for export.
type CurrentActionsSettings struct {
	Enabled                      bool   `json:"enabled"`
	AllowedActions               string `json:"allowed_actions"`
	DefaultWorkflowPermissions   string `json:"default_workflow_permissions,omitempty"`
	CanApprovePullRequestReviews *bool  `json:"can_approve_pull_request_reviews,omitempty"`
}
