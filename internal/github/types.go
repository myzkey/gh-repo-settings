package github

// RepoData represents repository data from GitHub API
type RepoData struct {
	Description         *string  `json:"description"`
	Homepage            *string  `json:"homepage"`
	Visibility          string   `json:"visibility"`
	AllowMergeCommit    bool     `json:"allow_merge_commit"`
	AllowRebaseMerge    bool     `json:"allow_rebase_merge"`
	AllowSquashMerge    bool     `json:"allow_squash_merge"`
	DeleteBranchOnMerge bool     `json:"delete_branch_on_merge"`
	AllowUpdateBranch   bool     `json:"allow_update_branch"`
	Topics              []string `json:"topics"`
}

// LabelData represents label data from GitHub API
type LabelData struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// BranchProtectionData represents branch protection data from GitHub API
type BranchProtectionData struct {
	RequiredPullRequestReviews *struct {
		RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
		DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
		RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
	} `json:"required_pull_request_reviews"`

	RequiredStatusChecks *struct {
		Strict   bool     `json:"strict"`
		Contexts []string `json:"contexts"`
	} `json:"required_status_checks"`

	EnforceAdmins *struct {
		Enabled bool `json:"enabled"`
	} `json:"enforce_admins"`

	RequiredLinearHistory *struct {
		Enabled bool `json:"enabled"`
	} `json:"required_linear_history"`

	AllowForcePushes *struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_force_pushes"`

	AllowDeletions *struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_deletions"`

	RequiredSignatures *struct {
		Enabled bool `json:"enabled"`
	} `json:"required_signatures"`
}

// ActionsPermissionsData represents GitHub Actions permissions from API
type ActionsPermissionsData struct {
	Enabled        bool   `json:"enabled"`
	AllowedActions string `json:"allowed_actions"` // "all", "local_only", "selected"
}

// ActionsSelectedData represents selected actions configuration from API
type ActionsSelectedData struct {
	GithubOwnedAllowed bool     `json:"github_owned_allowed"`
	VerifiedAllowed    bool     `json:"verified_allowed"`
	PatternsAllowed    []string `json:"patterns_allowed"`
}

// ActionsWorkflowPermissionsData represents workflow permissions from API
type ActionsWorkflowPermissionsData struct {
	DefaultWorkflowPermissions   string `json:"default_workflow_permissions"` // "read" or "write"
	CanApprovePullRequestReviews bool   `json:"can_approve_pull_request_reviews"`
}

// PagesData represents GitHub Pages data from API
type PagesData struct {
	BuildType string           `json:"build_type"` // "workflow" or "legacy"
	Source    *PagesSourceData `json:"source"`
}

// PagesSourceData represents the source configuration for GitHub Pages
type PagesSourceData struct {
	Branch string `json:"branch"`
	Path   string `json:"path"` // "/" or "/docs"
}

// VariableData represents repository variable data from GitHub API
type VariableData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
