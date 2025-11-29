package config

// Config represents the full configuration
type Config struct {
	Repo             *RepoConfig            `yaml:"repo,omitempty"`
	Topics           []string               `yaml:"topics,omitempty"`
	Labels           *LabelsConfig          `yaml:"labels,omitempty"`
	BranchProtection map[string]*BranchRule `yaml:"branch_protection,omitempty"`
	Secrets          *SecretsConfig         `yaml:"secrets,omitempty"`
	Env              *EnvConfig             `yaml:"env,omitempty"`
}

// RepoConfig represents repository settings
type RepoConfig struct {
	Description         *string `yaml:"description,omitempty"`
	Homepage            *string `yaml:"homepage,omitempty"`
	Visibility          *string `yaml:"visibility,omitempty"`
	AllowMergeCommit    *bool   `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool   `yaml:"allow_rebase_merge,omitempty"`
	AllowSquashMerge    *bool   `yaml:"allow_squash_merge,omitempty"`
	DeleteBranchOnMerge *bool   `yaml:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch   *bool   `yaml:"allow_update_branch,omitempty"`
}

// LabelsConfig represents label configuration
type LabelsConfig struct {
	ReplaceDefault bool    `yaml:"replace_default,omitempty"`
	Items          []Label `yaml:"items,omitempty"`
}

// Label represents a single label
type Label struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Description string `yaml:"description,omitempty"`
}

// BranchRule represents branch protection rules
type BranchRule struct {
	// Pull request reviews
	RequiredReviews     *int  `yaml:"required_reviews,omitempty"`
	DismissStaleReviews *bool `yaml:"dismiss_stale_reviews,omitempty"`
	RequireCodeOwner    *bool `yaml:"require_code_owner,omitempty"`

	// Status checks
	RequireStatusChecks *bool    `yaml:"require_status_checks,omitempty"`
	StatusChecks        []string `yaml:"status_checks,omitempty"`
	StrictStatusChecks  *bool    `yaml:"strict_status_checks,omitempty"`

	// Deployments
	RequiredDeployments []string `yaml:"required_deployments,omitempty"`

	// Commit requirements
	RequireSignedCommits *bool `yaml:"require_signed_commits,omitempty"`
	RequireLinearHistory *bool `yaml:"require_linear_history,omitempty"`

	// Push/merge restrictions
	EnforceAdmins     *bool `yaml:"enforce_admins,omitempty"`
	RestrictCreations *bool `yaml:"restrict_creations,omitempty"`
	RestrictPushes    *bool `yaml:"restrict_pushes,omitempty"`
	AllowForcePushes  *bool `yaml:"allow_force_pushes,omitempty"`
	AllowDeletions    *bool `yaml:"allow_deletions,omitempty"`
}

// SecretsConfig represents secrets configuration
type SecretsConfig struct {
	Required []string `yaml:"required,omitempty"`
}

// EnvConfig represents environment variables configuration
type EnvConfig struct {
	Required []string `yaml:"required,omitempty"`
}
