package config

// mergeConfigs merges src into dst (src values override dst values)
func mergeConfigs(dst, src *Config) {
	if src.Repo != nil {
		if dst.Repo == nil {
			dst.Repo = &RepoConfig{}
		}
		mergeRepoConfig(dst.Repo, src.Repo)
	}

	if len(src.Topics) > 0 {
		dst.Topics = src.Topics
	}

	if src.Labels != nil {
		if dst.Labels == nil {
			dst.Labels = &LabelsConfig{}
		}
		mergeLabelsConfig(dst.Labels, src.Labels)
	}

	if src.BranchProtection != nil {
		if dst.BranchProtection == nil {
			dst.BranchProtection = make(map[string]*BranchRule)
		}
		for k, v := range src.BranchProtection {
			if dst.BranchProtection[k] == nil {
				dst.BranchProtection[k] = &BranchRule{}
			}
			mergeBranchRule(dst.BranchProtection[k], v)
		}
	}

	if src.Env != nil {
		if dst.Env == nil {
			dst.Env = &EnvConfig{}
		}
		mergeEnvConfig(dst.Env, src.Env)
	}

	if src.Actions != nil {
		if dst.Actions == nil {
			dst.Actions = &ActionsConfig{}
		}
		mergeActionsConfig(dst.Actions, src.Actions)
	}
}

// mergeRepoConfig merges repo configurations
func mergeRepoConfig(dst, src *RepoConfig) {
	if src.Description != nil {
		dst.Description = src.Description
	}
	if src.Homepage != nil {
		dst.Homepage = src.Homepage
	}
	if src.Visibility != nil {
		dst.Visibility = src.Visibility
	}
	if src.AllowMergeCommit != nil {
		dst.AllowMergeCommit = src.AllowMergeCommit
	}
	if src.AllowRebaseMerge != nil {
		dst.AllowRebaseMerge = src.AllowRebaseMerge
	}
	if src.AllowSquashMerge != nil {
		dst.AllowSquashMerge = src.AllowSquashMerge
	}
	if src.DeleteBranchOnMerge != nil {
		dst.DeleteBranchOnMerge = src.DeleteBranchOnMerge
	}
	if src.AllowUpdateBranch != nil {
		dst.AllowUpdateBranch = src.AllowUpdateBranch
	}
}

// mergeLabelsConfig merges labels configurations
func mergeLabelsConfig(dst, src *LabelsConfig) {
	if src.ReplaceDefault {
		dst.ReplaceDefault = src.ReplaceDefault
	}
	if len(src.Items) > 0 {
		dst.Items = src.Items
	}
}

// mergeBranchRule merges branch protection rules
func mergeBranchRule(dst, src *BranchRule) {
	if src.RequiredReviews != nil {
		dst.RequiredReviews = src.RequiredReviews
	}
	if src.DismissStaleReviews != nil {
		dst.DismissStaleReviews = src.DismissStaleReviews
	}
	if src.RequireCodeOwner != nil {
		dst.RequireCodeOwner = src.RequireCodeOwner
	}
	if src.RequireStatusChecks != nil {
		dst.RequireStatusChecks = src.RequireStatusChecks
	}
	if len(src.StatusChecks) > 0 {
		dst.StatusChecks = src.StatusChecks
	}
	if src.StrictStatusChecks != nil {
		dst.StrictStatusChecks = src.StrictStatusChecks
	}
	if len(src.RequiredDeployments) > 0 {
		dst.RequiredDeployments = src.RequiredDeployments
	}
	if src.RequireSignedCommits != nil {
		dst.RequireSignedCommits = src.RequireSignedCommits
	}
	if src.RequireLinearHistory != nil {
		dst.RequireLinearHistory = src.RequireLinearHistory
	}
	if src.EnforceAdmins != nil {
		dst.EnforceAdmins = src.EnforceAdmins
	}
	if src.RestrictCreations != nil {
		dst.RestrictCreations = src.RestrictCreations
	}
	if src.RestrictPushes != nil {
		dst.RestrictPushes = src.RestrictPushes
	}
	if src.AllowForcePushes != nil {
		dst.AllowForcePushes = src.AllowForcePushes
	}
	if src.AllowDeletions != nil {
		dst.AllowDeletions = src.AllowDeletions
	}
}

// mergeEnvConfig merges environment configurations
func mergeEnvConfig(dst, src *EnvConfig) {
	if len(src.Variables) > 0 {
		if dst.Variables == nil {
			dst.Variables = make(map[string]string)
		}
		for k, v := range src.Variables {
			dst.Variables[k] = v
		}
	}
	if len(src.Secrets) > 0 {
		dst.Secrets = src.Secrets
	}
}

// mergeActionsConfig merges actions configurations
func mergeActionsConfig(dst, src *ActionsConfig) {
	if src.Enabled != nil {
		dst.Enabled = src.Enabled
	}
	if src.AllowedActions != nil {
		dst.AllowedActions = src.AllowedActions
	}
	if src.SelectedActions != nil {
		if dst.SelectedActions == nil {
			dst.SelectedActions = &SelectedActionsConfig{}
		}
		if src.SelectedActions.GithubOwnedAllowed != nil {
			dst.SelectedActions.GithubOwnedAllowed = src.SelectedActions.GithubOwnedAllowed
		}
		if src.SelectedActions.VerifiedAllowed != nil {
			dst.SelectedActions.VerifiedAllowed = src.SelectedActions.VerifiedAllowed
		}
		if len(src.SelectedActions.PatternsAllowed) > 0 {
			dst.SelectedActions.PatternsAllowed = src.SelectedActions.PatternsAllowed
		}
	}
	if src.DefaultWorkflowPermissions != nil {
		dst.DefaultWorkflowPermissions = src.DefaultWorkflowPermissions
	}
	if src.CanApprovePullRequestReviews != nil {
		dst.CanApprovePullRequestReviews = src.CanApprovePullRequestReviews
	}
}
