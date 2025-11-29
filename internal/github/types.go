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
