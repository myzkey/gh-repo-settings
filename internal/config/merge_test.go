package config

import "testing"

func ptr(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrInt(i int) *int {
	return &i
}

func TestMergeConfigs(t *testing.T) {
	trueVal := true
	falseVal := false
	desc := "Test"
	reviews := 2

	dst := &Config{
		Repo: &RepoConfig{
			Visibility:       ptr("public"),
			AllowMergeCommit: &trueVal,
		},
	}

	src := &Config{
		Repo: &RepoConfig{
			Description:      &desc,
			AllowMergeCommit: &falseVal,
		},
		BranchProtection: map[string]*BranchRule{
			"main": {
				RequiredReviews: &reviews,
			},
		},
	}

	mergeConfigs(dst, src)

	if dst.Repo.Visibility == nil || *dst.Repo.Visibility != "public" {
		t.Error("expected visibility to be preserved")
	}
	if dst.Repo.Description == nil || *dst.Repo.Description != "Test" {
		t.Error("expected description to be added")
	}
	if dst.Repo.AllowMergeCommit == nil || *dst.Repo.AllowMergeCommit != false {
		t.Error("expected allow_merge_commit to be overridden")
	}
	if dst.BranchProtection == nil || dst.BranchProtection["main"] == nil {
		t.Error("expected branch protection to be added")
	}
}

func TestMergeConfigsNilDst(t *testing.T) {
	dst := &Config{}
	src := &Config{
		Repo: &RepoConfig{
			Description: ptr("Test"),
		},
		Topics: []string{"go", "cli"},
		Labels: &LabelsConfig{
			ReplaceDefault: true,
			Items:          []Label{{Name: "bug", Color: "d73a4a"}},
		},
		Env: &EnvConfig{
			Variables: map[string]string{"NODE_ENV": "production"},
			Secrets:   []string{"SECRET_KEY"},
		},
		Actions: &ActionsConfig{
			Enabled: ptrBool(true),
		},
	}

	mergeConfigs(dst, src)

	if dst.Repo == nil || dst.Repo.Description == nil {
		t.Fatal("expected Repo to be set")
	}
	if len(dst.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(dst.Topics))
	}
	if dst.Labels == nil || !dst.Labels.ReplaceDefault {
		t.Error("expected Labels to be set")
	}
	if dst.Env == nil || len(dst.Env.Secrets) != 1 {
		t.Error("expected Env.Secrets to be set")
	}
	if dst.Env == nil || len(dst.Env.Variables) != 1 {
		t.Error("expected Env.Variables to be set")
	}
	if dst.Actions == nil || !*dst.Actions.Enabled {
		t.Error("expected Actions to be set")
	}
}

func TestMergeConfigsTopics(t *testing.T) {
	dst := &Config{Topics: []string{"go", "cli"}}
	src := &Config{Topics: []string{"github", "automation"}}

	mergeConfigs(dst, src)

	if len(dst.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(dst.Topics))
	}
	if dst.Topics[0] != "github" || dst.Topics[1] != "automation" {
		t.Errorf("expected topics [github, automation], got %v", dst.Topics)
	}
}

func TestMergeConfigsEnv(t *testing.T) {
	dst := &Config{
		Env: &EnvConfig{
			Variables: map[string]string{"OLD_VAR": "old"},
			Secrets:   []string{"OLD_SECRET"},
		},
	}
	src := &Config{
		Env: &EnvConfig{
			Variables: map[string]string{"NEW_VAR": "new", "OLD_VAR": "updated"},
			Secrets:   []string{"NEW_SECRET_1", "NEW_SECRET_2"},
		},
	}

	mergeConfigs(dst, src)

	if len(dst.Env.Variables) != 2 {
		t.Errorf("expected 2 variables, got %d", len(dst.Env.Variables))
	}
	if dst.Env.Variables["OLD_VAR"] != "updated" {
		t.Errorf("expected OLD_VAR to be 'updated', got %s", dst.Env.Variables["OLD_VAR"])
	}
	if len(dst.Env.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(dst.Env.Secrets))
	}
}

func TestMergeRepoConfig(t *testing.T) {
	tests := []struct {
		name     string
		dst      *RepoConfig
		src      *RepoConfig
		checkDst func(*testing.T, *RepoConfig)
	}{
		{
			name: "override all fields",
			dst: &RepoConfig{
				Description:         ptr("old"),
				Homepage:            ptr("https://old.com"),
				Visibility:          ptr("private"),
				AllowMergeCommit:    ptrBool(true),
				AllowRebaseMerge:    ptrBool(true),
				AllowSquashMerge:    ptrBool(true),
				DeleteBranchOnMerge: ptrBool(false),
				AllowUpdateBranch:   ptrBool(false),
			},
			src: &RepoConfig{
				Description:         ptr("new"),
				Homepage:            ptr("https://new.com"),
				Visibility:          ptr("public"),
				AllowMergeCommit:    ptrBool(false),
				AllowRebaseMerge:    ptrBool(false),
				AllowSquashMerge:    ptrBool(false),
				DeleteBranchOnMerge: ptrBool(true),
				AllowUpdateBranch:   ptrBool(true),
			},
			checkDst: func(t *testing.T, dst *RepoConfig) {
				if *dst.Description != "new" {
					t.Error("Description not overridden")
				}
				if *dst.Homepage != "https://new.com" {
					t.Error("Homepage not overridden")
				}
				if *dst.Visibility != "public" {
					t.Error("Visibility not overridden")
				}
				if *dst.AllowMergeCommit != false {
					t.Error("AllowMergeCommit not overridden")
				}
				if *dst.AllowRebaseMerge != false {
					t.Error("AllowRebaseMerge not overridden")
				}
				if *dst.AllowSquashMerge != false {
					t.Error("AllowSquashMerge not overridden")
				}
				if *dst.DeleteBranchOnMerge != true {
					t.Error("DeleteBranchOnMerge not overridden")
				}
				if *dst.AllowUpdateBranch != true {
					t.Error("AllowUpdateBranch not overridden")
				}
			},
		},
		{
			name: "partial override preserves existing",
			dst: &RepoConfig{
				Description: ptr("existing"),
				Visibility:  ptr("private"),
			},
			src: &RepoConfig{
				Homepage: ptr("https://new.com"),
			},
			checkDst: func(t *testing.T, dst *RepoConfig) {
				if *dst.Description != "existing" {
					t.Error("Description should be preserved")
				}
				if *dst.Visibility != "private" {
					t.Error("Visibility should be preserved")
				}
				if *dst.Homepage != "https://new.com" {
					t.Error("Homepage should be added")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeRepoConfig(tt.dst, tt.src)
			tt.checkDst(t, tt.dst)
		})
	}
}

func TestMergeLabelsConfig(t *testing.T) {
	tests := []struct {
		name     string
		dst      *LabelsConfig
		src      *LabelsConfig
		checkDst func(*testing.T, *LabelsConfig)
	}{
		{
			name: "full override",
			dst: &LabelsConfig{
				ReplaceDefault: false,
				Items:          []Label{{Name: "bug", Color: "d73a4a"}},
			},
			src: &LabelsConfig{
				ReplaceDefault: true,
				Items:          []Label{{Name: "enhancement", Color: "a2eeef"}},
			},
			checkDst: func(t *testing.T, dst *LabelsConfig) {
				if !dst.ReplaceDefault {
					t.Error("ReplaceDefault should be true")
				}
				if len(dst.Items) != 1 || dst.Items[0].Name != "enhancement" {
					t.Error("Items should be replaced")
				}
			},
		},
		{
			name: "partial - false does not override true",
			dst: &LabelsConfig{
				ReplaceDefault: true,
				Items:          []Label{{Name: "bug", Color: "d73a4a"}},
			},
			src: &LabelsConfig{
				ReplaceDefault: false,
			},
			checkDst: func(t *testing.T, dst *LabelsConfig) {
				if !dst.ReplaceDefault {
					t.Error("ReplaceDefault should remain true")
				}
				if len(dst.Items) != 1 || dst.Items[0].Name != "bug" {
					t.Error("Items should remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeLabelsConfig(tt.dst, tt.src)
			tt.checkDst(t, tt.dst)
		})
	}
}

func TestMergeBranchRule(t *testing.T) {
	tests := []struct {
		name     string
		dst      *BranchRule
		src      *BranchRule
		checkDst func(*testing.T, *BranchRule)
	}{
		{
			name: "override all fields",
			dst: &BranchRule{
				RequiredReviews:     ptrInt(1),
				DismissStaleReviews: ptrBool(false),
				StatusChecks:        []string{"test"},
			},
			src: &BranchRule{
				RequiredReviews:      ptrInt(2),
				DismissStaleReviews:  ptrBool(true),
				RequireCodeOwner:     ptrBool(true),
				RequireStatusChecks:  ptrBool(true),
				StatusChecks:         []string{"build", "lint"},
				StrictStatusChecks:   ptrBool(true),
				RequiredDeployments:  []string{"staging"},
				RequireSignedCommits: ptrBool(true),
				RequireLinearHistory: ptrBool(true),
				EnforceAdmins:        ptrBool(true),
				RestrictCreations:    ptrBool(true),
				RestrictPushes:       ptrBool(true),
				AllowForcePushes:     ptrBool(false),
				AllowDeletions:       ptrBool(false),
			},
			checkDst: func(t *testing.T, dst *BranchRule) {
				if *dst.RequiredReviews != 2 {
					t.Error("RequiredReviews not overridden")
				}
				if !*dst.DismissStaleReviews {
					t.Error("DismissStaleReviews not overridden")
				}
				if !*dst.RequireCodeOwner {
					t.Error("RequireCodeOwner not set")
				}
				if !*dst.RequireStatusChecks {
					t.Error("RequireStatusChecks not set")
				}
				if len(dst.StatusChecks) != 2 || dst.StatusChecks[0] != "build" {
					t.Error("StatusChecks not overridden")
				}
				if !*dst.StrictStatusChecks {
					t.Error("StrictStatusChecks not set")
				}
				if len(dst.RequiredDeployments) != 1 || dst.RequiredDeployments[0] != "staging" {
					t.Error("RequiredDeployments not set")
				}
				if !*dst.RequireSignedCommits {
					t.Error("RequireSignedCommits not set")
				}
				if !*dst.RequireLinearHistory {
					t.Error("RequireLinearHistory not set")
				}
				if !*dst.EnforceAdmins {
					t.Error("EnforceAdmins not set")
				}
				if !*dst.RestrictCreations {
					t.Error("RestrictCreations not set")
				}
				if !*dst.RestrictPushes {
					t.Error("RestrictPushes not set")
				}
				if *dst.AllowForcePushes {
					t.Error("AllowForcePushes should be false")
				}
				if *dst.AllowDeletions {
					t.Error("AllowDeletions should be false")
				}
			},
		},
		{
			name: "partial override preserves existing",
			dst: &BranchRule{
				RequiredReviews:     ptrInt(2),
				DismissStaleReviews: ptrBool(true),
			},
			src: &BranchRule{
				RequireCodeOwner: ptrBool(true),
			},
			checkDst: func(t *testing.T, dst *BranchRule) {
				if *dst.RequiredReviews != 2 {
					t.Error("RequiredReviews should be preserved")
				}
				if !*dst.DismissStaleReviews {
					t.Error("DismissStaleReviews should be preserved")
				}
				if !*dst.RequireCodeOwner {
					t.Error("RequireCodeOwner should be added")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeBranchRule(tt.dst, tt.src)
			tt.checkDst(t, tt.dst)
		})
	}
}

func TestMergeActionsConfig(t *testing.T) {
	tests := []struct {
		name     string
		dst      *ActionsConfig
		src      *ActionsConfig
		checkDst func(*testing.T, *ActionsConfig)
	}{
		{
			name: "override all fields",
			dst: &ActionsConfig{
				Enabled:        ptrBool(false),
				AllowedActions: ptr("all"),
				SelectedActions: &SelectedActionsConfig{
					GithubOwnedAllowed: ptrBool(true),
					PatternsAllowed:    []string{"actions/*"},
				},
				DefaultWorkflowPermissions:   ptr("read"),
				CanApprovePullRequestReviews: ptrBool(false),
			},
			src: &ActionsConfig{
				Enabled:        ptrBool(true),
				AllowedActions: ptr("selected"),
				SelectedActions: &SelectedActionsConfig{
					VerifiedAllowed: ptrBool(true),
					PatternsAllowed: []string{"github/*"},
				},
				DefaultWorkflowPermissions:   ptr("write"),
				CanApprovePullRequestReviews: ptrBool(true),
			},
			checkDst: func(t *testing.T, dst *ActionsConfig) {
				if !*dst.Enabled {
					t.Error("Enabled not overridden")
				}
				if *dst.AllowedActions != "selected" {
					t.Error("AllowedActions not overridden")
				}
				if !*dst.SelectedActions.GithubOwnedAllowed {
					t.Error("GithubOwnedAllowed should be preserved")
				}
				if !*dst.SelectedActions.VerifiedAllowed {
					t.Error("VerifiedAllowed should be added")
				}
				if len(dst.SelectedActions.PatternsAllowed) != 1 || dst.SelectedActions.PatternsAllowed[0] != "github/*" {
					t.Error("PatternsAllowed should be overridden")
				}
				if *dst.DefaultWorkflowPermissions != "write" {
					t.Error("DefaultWorkflowPermissions not overridden")
				}
				if !*dst.CanApprovePullRequestReviews {
					t.Error("CanApprovePullRequestReviews not overridden")
				}
			},
		},
		{
			name: "nil dst.SelectedActions",
			dst:  &ActionsConfig{},
			src: &ActionsConfig{
				SelectedActions: &SelectedActionsConfig{
					GithubOwnedAllowed: ptrBool(true),
				},
			},
			checkDst: func(t *testing.T, dst *ActionsConfig) {
				if dst.SelectedActions == nil {
					t.Fatal("SelectedActions should be created")
				}
				if !*dst.SelectedActions.GithubOwnedAllowed {
					t.Error("GithubOwnedAllowed should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeActionsConfig(tt.dst, tt.src)
			tt.checkDst(t, tt.dst)
		})
	}
}

func TestMergeConfigsBranchProtectionMultipleBranches(t *testing.T) {
	dst := &Config{
		BranchProtection: map[string]*BranchRule{
			"main": {
				RequiredReviews: ptrInt(1),
			},
		},
	}
	src := &Config{
		BranchProtection: map[string]*BranchRule{
			"main": {
				RequiredReviews: ptrInt(2),
			},
			"develop": {
				RequiredReviews: ptrInt(1),
			},
		},
	}

	mergeConfigs(dst, src)

	if dst.BranchProtection["main"] == nil || *dst.BranchProtection["main"].RequiredReviews != 2 {
		t.Error("main branch should be updated to 2 reviews")
	}
	if dst.BranchProtection["develop"] == nil || *dst.BranchProtection["develop"].RequiredReviews != 1 {
		t.Error("develop branch should be added")
	}
}
