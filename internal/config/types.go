package config

// Config represents the full configuration for repository settings
type Config struct {
	Extends          []string               `yaml:"extends,omitempty" json:"extends,omitempty" jsonschema:"description=List of preset URLs or file paths to inherit from"`
	Repo             *RepoConfig            `yaml:"repo,omitempty" json:"repo,omitempty" jsonschema:"description=Repository settings"`
	Topics           []string               `yaml:"topics,omitempty" json:"topics,omitempty" jsonschema:"description=Repository topics"`
	Labels           *LabelsConfig          `yaml:"labels,omitempty" json:"labels,omitempty" jsonschema:"description=Issue labels configuration"`
	BranchProtection map[string]*BranchRule `yaml:"branch_protection,omitempty" json:"branch_protection,omitempty" jsonschema:"description=Branch protection rules keyed by branch name"`
	Secrets          *SecretsConfig         `yaml:"secrets,omitempty" json:"secrets,omitempty" jsonschema:"description=Required secrets configuration"`
	Env              *EnvConfig             `yaml:"env,omitempty" json:"env,omitempty" jsonschema:"description=Required environment variables configuration"`
	Actions          *ActionsConfig         `yaml:"actions,omitempty" json:"actions,omitempty" jsonschema:"description=GitHub Actions permissions configuration"`
	Pages            *PagesConfig           `yaml:"pages,omitempty" json:"pages,omitempty" jsonschema:"description=GitHub Pages configuration"`
}

// RepoConfig represents repository settings
type RepoConfig struct {
	Description         *string `yaml:"description,omitempty" json:"description,omitempty" jsonschema:"description=Repository description"`
	Homepage            *string `yaml:"homepage,omitempty" json:"homepage,omitempty" jsonschema:"description=Homepage URL"`
	Visibility          *string `yaml:"visibility,omitempty" json:"visibility,omitempty" jsonschema:"description=Repository visibility,enum=public,enum=private,enum=internal"`
	AllowMergeCommit    *bool   `yaml:"allow_merge_commit,omitempty" json:"allow_merge_commit,omitempty" jsonschema:"description=Allow merge commits"`
	AllowRebaseMerge    *bool   `yaml:"allow_rebase_merge,omitempty" json:"allow_rebase_merge,omitempty" jsonschema:"description=Allow rebase merging"`
	AllowSquashMerge    *bool   `yaml:"allow_squash_merge,omitempty" json:"allow_squash_merge,omitempty" jsonschema:"description=Allow squash merging"`
	DeleteBranchOnMerge *bool   `yaml:"delete_branch_on_merge,omitempty" json:"delete_branch_on_merge,omitempty" jsonschema:"description=Auto-delete head branches after merge"`
	AllowUpdateBranch   *bool   `yaml:"allow_update_branch,omitempty" json:"allow_update_branch,omitempty" jsonschema:"description=Allow updating PR branches"`
}

// LabelsConfig represents label configuration
type LabelsConfig struct {
	ReplaceDefault bool    `yaml:"replace_default,omitempty" json:"replace_default,omitempty" jsonschema:"description=Delete labels not in config"`
	Items          []Label `yaml:"items,omitempty" json:"items,omitempty" jsonschema:"description=List of label definitions"`
}

// Label represents a single label
type Label struct {
	Name        string `yaml:"name" json:"name" jsonschema:"description=Label name,required"`
	Color       string `yaml:"color" json:"color" jsonschema:"description=Hex color without #,required,pattern=^[0-9a-fA-F]{6}$"`
	Description string `yaml:"description,omitempty" json:"description,omitempty" jsonschema:"description=Label description"`
}

// BranchRule represents branch protection rules
type BranchRule struct {
	// Pull request reviews
	RequiredReviews     *int  `yaml:"required_reviews,omitempty" json:"required_reviews,omitempty" jsonschema:"description=Number of required approving reviews,minimum=0,maximum=6"`
	DismissStaleReviews *bool `yaml:"dismiss_stale_reviews,omitempty" json:"dismiss_stale_reviews,omitempty" jsonschema:"description=Dismiss approvals when new commits are pushed"`
	RequireCodeOwner    *bool `yaml:"require_code_owner,omitempty" json:"require_code_owner,omitempty" jsonschema:"description=Require review from CODEOWNERS"`

	// Status checks
	RequireStatusChecks *bool    `yaml:"require_status_checks,omitempty" json:"require_status_checks,omitempty" jsonschema:"description=Require status checks to pass"`
	StatusChecks        []string `yaml:"status_checks,omitempty" json:"status_checks,omitempty" jsonschema:"description=List of required status check names"`
	StrictStatusChecks  *bool    `yaml:"strict_status_checks,omitempty" json:"strict_status_checks,omitempty" jsonschema:"description=Require branches to be up to date"`

	// Deployments
	RequiredDeployments []string `yaml:"required_deployments,omitempty" json:"required_deployments,omitempty" jsonschema:"description=Required deployment environments"`

	// Commit requirements
	RequireSignedCommits *bool `yaml:"require_signed_commits,omitempty" json:"require_signed_commits,omitempty" jsonschema:"description=Require signed commits"`
	RequireLinearHistory *bool `yaml:"require_linear_history,omitempty" json:"require_linear_history,omitempty" jsonschema:"description=Require linear history (no merge commits)"`

	// Push/merge restrictions
	EnforceAdmins     *bool `yaml:"enforce_admins,omitempty" json:"enforce_admins,omitempty" jsonschema:"description=Enforce rules for administrators"`
	RestrictCreations *bool `yaml:"restrict_creations,omitempty" json:"restrict_creations,omitempty" jsonschema:"description=Restrict branch creation"`
	RestrictPushes    *bool `yaml:"restrict_pushes,omitempty" json:"restrict_pushes,omitempty" jsonschema:"description=Restrict who can push"`
	AllowForcePushes  *bool `yaml:"allow_force_pushes,omitempty" json:"allow_force_pushes,omitempty" jsonschema:"description=Allow force pushes"`
	AllowDeletions    *bool `yaml:"allow_deletions,omitempty" json:"allow_deletions,omitempty" jsonschema:"description=Allow branch deletion"`
}

// SecretsConfig represents secrets configuration
type SecretsConfig struct {
	Required []string `yaml:"required,omitempty" json:"required,omitempty" jsonschema:"description=List of required secret names"`
}

// EnvConfig represents environment variables configuration
type EnvConfig struct {
	Required []string `yaml:"required,omitempty" json:"required,omitempty" jsonschema:"description=List of required environment variable names"`
}

// ActionsConfig represents GitHub Actions permissions configuration
type ActionsConfig struct {
	Enabled        *bool   `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"description=Enable or disable GitHub Actions"`
	AllowedActions *string `yaml:"allowed_actions,omitempty" json:"allowed_actions,omitempty" jsonschema:"description=Which actions are allowed,enum=all,enum=local_only,enum=selected"`

	SelectedActions *SelectedActionsConfig `yaml:"selected_actions,omitempty" json:"selected_actions,omitempty" jsonschema:"description=Configuration for selected actions (when allowed_actions is 'selected')"`

	DefaultWorkflowPermissions   *string `yaml:"default_workflow_permissions,omitempty" json:"default_workflow_permissions,omitempty" jsonschema:"description=Default GITHUB_TOKEN permissions,enum=read,enum=write"`
	CanApprovePullRequestReviews *bool   `yaml:"can_approve_pull_request_reviews,omitempty" json:"can_approve_pull_request_reviews,omitempty" jsonschema:"description=Allow GitHub Actions to create and approve pull requests"`
}

// SelectedActionsConfig represents the configuration for selected actions
type SelectedActionsConfig struct {
	GithubOwnedAllowed *bool    `yaml:"github_owned_allowed,omitempty" json:"github_owned_allowed,omitempty" jsonschema:"description=Allow actions created by GitHub"`
	VerifiedAllowed    *bool    `yaml:"verified_allowed,omitempty" json:"verified_allowed,omitempty" jsonschema:"description=Allow actions from verified creators"`
	PatternsAllowed    []string `yaml:"patterns_allowed,omitempty" json:"patterns_allowed,omitempty" jsonschema:"description=Patterns for allowed actions (e.g. 'actions/*')"`
}

// PagesConfig represents GitHub Pages configuration
type PagesConfig struct {
	BuildType *string            `yaml:"build_type,omitempty" json:"build_type,omitempty" jsonschema:"description=Build type for GitHub Pages,enum=workflow,enum=legacy"`
	Source    *PagesSourceConfig `yaml:"source,omitempty" json:"source,omitempty" jsonschema:"description=Source configuration (for legacy build type)"`
}

// PagesSourceConfig represents the source configuration for GitHub Pages
type PagesSourceConfig struct {
	Branch *string `yaml:"branch,omitempty" json:"branch,omitempty" jsonschema:"description=Branch name for Pages source"`
	Path   *string `yaml:"path,omitempty" json:"path,omitempty" jsonschema:"description=Path within the branch (/ or /docs),enum=/,enum=/docs"`
}
