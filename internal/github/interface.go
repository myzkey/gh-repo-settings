package github

import "context"

// Client interface defines all GitHub operations
type GitHubClient interface {
	// Repository operations
	GetRepo(ctx context.Context) (*RepoData, error)
	UpdateRepo(ctx context.Context, settings map[string]interface{}) error

	// Topics operations
	SetTopics(ctx context.Context, topics []string) error

	// Label operations
	GetLabels(ctx context.Context) ([]LabelData, error)
	CreateLabel(ctx context.Context, name, color, description string) error
	UpdateLabel(ctx context.Context, oldName, newName, color, description string) error
	DeleteLabel(ctx context.Context, name string) error

	// Branch protection operations
	GetBranchProtection(ctx context.Context, branch string) (*BranchProtectionData, error)
	UpdateBranchProtection(ctx context.Context, branch string, settings *BranchProtectionSettings) error

	// Secrets and variables
	GetSecrets(ctx context.Context) ([]string, error)
	GetVariables(ctx context.Context) ([]string, error)

	// Actions permissions
	GetActionsPermissions(ctx context.Context) (*ActionsPermissionsData, error)
	UpdateActionsPermissions(ctx context.Context, enabled bool, allowedActions string) error
	GetActionsSelectedActions(ctx context.Context) (*ActionsSelectedData, error)
	UpdateActionsSelectedActions(ctx context.Context, settings *ActionsSelectedData) error
	GetActionsWorkflowPermissions(ctx context.Context) (*ActionsWorkflowPermissionsData, error)
	UpdateActionsWorkflowPermissions(ctx context.Context, permissions string, canApprove bool) error

	// Repository info
	RepoOwner() string
	RepoName() string
}

// BranchProtectionSettings represents settings to update branch protection
type BranchProtectionSettings struct {
	RequiredReviews         *int     `json:"required_approving_review_count,omitempty"`
	DismissStaleReviews     *bool    `json:"dismiss_stale_reviews,omitempty"`
	RequireCodeOwnerReviews *bool    `json:"require_code_owner_reviews,omitempty"`
	RequireStatusChecks     *bool    `json:"-"`
	StatusChecks            []string `json:"contexts,omitempty"`
	StrictStatusChecks      *bool    `json:"strict,omitempty"`
	EnforceAdmins           *bool    `json:"enforce_admins,omitempty"`
	RequireLinearHistory    *bool    `json:"required_linear_history,omitempty"`
	AllowForcePushes        *bool    `json:"allow_force_pushes,omitempty"`
	AllowDeletions          *bool    `json:"allow_deletions,omitempty"`
	RequireSignedCommits    *bool    `json:"required_signatures,omitempty"`
}

// Ensure Client implements GitHubClient
var _ GitHubClient = (*Client)(nil)
