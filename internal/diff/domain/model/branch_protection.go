package model

// BranchProtectionCurrent represents the current state of branch protection
// This is a domain model independent of infrastructure (GitHub API)
type BranchProtectionCurrent struct {
	RequiredReviews      int
	DismissStaleReviews  bool
	RequireCodeOwner     bool
	StrictStatusChecks   bool
	StatusChecks         []string
	EnforceAdmins        bool
	RequireLinearHistory bool
	AllowForcePushes     bool
	AllowDeletions       bool
	RequireSignedCommits bool
}

// BranchProtectionDesired represents the desired state of branch protection
// This is a domain model independent of configuration format
type BranchProtectionDesired struct {
	RequiredReviews      *int
	DismissStaleReviews  *bool
	RequireCodeOwner     *bool
	StrictStatusChecks   *bool
	StatusChecks         []string
	EnforceAdmins        *bool
	RequireLinearHistory *bool
	AllowForcePushes     *bool
	AllowDeletions       *bool
	RequireSignedCommits *bool
}
