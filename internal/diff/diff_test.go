package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/github"
)

func ptr[T any](v T) *T {
	return &v
}

func TestToStringSet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]bool
	}{
		{
			name:  "empty slice",
			input: []string{},
			want:  map[string]bool{},
		},
		{
			name:  "single item",
			input: []string{"a"},
			want:  map[string]bool{"a": true},
		},
		{
			name:  "multiple items",
			input: []string{"a", "b", "c"},
			want:  map[string]bool{"a": true, "b": true, "c": true},
		},
		{
			name:  "duplicates",
			input: []string{"a", "b", "a"},
			want:  map[string]bool{"a": true, "b": true},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSet(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("toStringSet() length = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("toStringSet()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestCalculatorCompareRepo(t *testing.T) {
	tests := []struct {
		name     string
		current  *github.RepoData
		config   *config.RepoConfig
		expected []Change
	}{
		{
			name: "no changes",
			current: &github.RepoData{
				Description: ptr("Test"),
				Visibility:  "public",
			},
			config: &config.RepoConfig{
				Description: ptr("Test"),
				Visibility:  ptr("public"),
			},
			expected: []Change{},
		},
		{
			name: "description change",
			current: &github.RepoData{
				Description: ptr("Old"),
			},
			config: &config.RepoConfig{
				Description: ptr("New"),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: "repo",
					Key:      "description",
					Old:      "Old",
					New:      "New",
				},
			},
		},
		{
			name: "visibility change",
			current: &github.RepoData{
				Visibility: "private",
			},
			config: &config.RepoConfig{
				Visibility: ptr("public"),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: "repo",
					Key:      "visibility",
					Old:      "private",
					New:      "public",
				},
			},
		},
		{
			name: "merge options change",
			current: &github.RepoData{
				AllowMergeCommit: true,
				AllowSquashMerge: false,
			},
			config: &config.RepoConfig{
				AllowMergeCommit: ptr(false),
				AllowSquashMerge: ptr(true),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: "repo",
					Key:      "allow_merge_commit",
					Old:      true,
					New:      false,
				},
				{
					Type:     ChangeUpdate,
					Category: "repo",
					Key:      "allow_squash_merge",
					Old:      false,
					New:      true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.RepoData = tt.current

			cfg := &config.Config{Repo: tt.config}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(plan.Changes) != len(tt.expected) {
				t.Errorf("expected %d changes, got %d", len(tt.expected), len(plan.Changes))
				return
			}

			for i, exp := range tt.expected {
				got := plan.Changes[i]
				if got.Type != exp.Type || got.Category != exp.Category || got.Key != exp.Key {
					t.Errorf("change %d mismatch: expected %+v, got %+v", i, exp, got)
				}
			}
		})
	}
}

func TestCalculatorCompareLabels(t *testing.T) {
	tests := []struct {
		name     string
		current  []github.LabelData
		config   *config.LabelsConfig
		expected struct {
			adds    int
			updates int
			deletes int
		}
	}{
		{
			name: "add new label",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a"},
			},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a"},
					{Name: "feature", Color: "a2eeef"},
				},
			},
			expected: struct {
				adds    int
				updates int
				deletes int
			}{adds: 1, updates: 0, deletes: 0},
		},
		{
			name: "update existing label",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a", Description: "Old description"},
			},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "ff0000", Description: "New description"},
				},
			},
			expected: struct {
				adds    int
				updates int
				deletes int
			}{adds: 0, updates: 1, deletes: 0},
		},
		{
			name: "delete label with replace_default",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a"},
				{Name: "old-label", Color: "000000"},
			},
			config: &config.LabelsConfig{
				ReplaceDefault: true,
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a"},
				},
			},
			expected: struct {
				adds    int
				updates int
				deletes int
			}{adds: 0, updates: 0, deletes: 1},
		},
		{
			name: "no delete without replace_default",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a"},
				{Name: "old-label", Color: "000000"},
			},
			config: &config.LabelsConfig{
				ReplaceDefault: false,
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a"},
				},
			},
			expected: struct {
				adds    int
				updates int
				deletes int
			}{adds: 0, updates: 0, deletes: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Labels = tt.current

			cfg := &config.Config{Labels: tt.config}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var adds, updates, deletes int
			for _, c := range plan.Changes {
				switch c.Type {
				case ChangeAdd:
					adds++
				case ChangeUpdate:
					updates++
				case ChangeDelete:
					deletes++
				}
			}

			if adds != tt.expected.adds {
				t.Errorf("expected %d adds, got %d", tt.expected.adds, adds)
			}
			if updates != tt.expected.updates {
				t.Errorf("expected %d updates, got %d", tt.expected.updates, updates)
			}
			if deletes != tt.expected.deletes {
				t.Errorf("expected %d deletes, got %d", tt.expected.deletes, deletes)
			}
		})
	}
}

func TestCalculatorCompareBranchProtection(t *testing.T) {
	tests := []struct {
		name          string
		current       map[string]*github.BranchProtectionData
		config        map[string]*config.BranchRule
		expectedCount int
		isAdd         bool
	}{
		{
			name:    "add new protection",
			current: map[string]*github.BranchProtectionData{},
			config: map[string]*config.BranchRule{
				"main": {
					RequiredReviews: ptr(2),
				},
			},
			expectedCount: 1,
			isAdd:         true,
		},
		{
			name: "update existing protection",
			current: map[string]*github.BranchProtectionData{
				"main": {
					RequiredPullRequestReviews: &struct {
						RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
						DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
						RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
					}{
						RequiredApprovingReviewCount: 1,
					},
				},
			},
			config: map[string]*config.BranchRule{
				"main": {
					RequiredReviews: ptr(2),
				},
			},
			expectedCount: 1,
			isAdd:         false,
		},
		{
			name: "no changes",
			current: map[string]*github.BranchProtectionData{
				"main": {
					RequiredPullRequestReviews: &struct {
						RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
						DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
						RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
					}{
						RequiredApprovingReviewCount: 2,
					},
				},
			},
			config: map[string]*config.BranchRule{
				"main": {
					RequiredReviews: ptr(2),
				},
			},
			expectedCount: 0,
			isAdd:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.BranchProtections = tt.current
			if len(tt.current) == 0 {
				mock.GetBranchProtectionError = apperrors.ErrBranchNotProtected
			}

			cfg := &config.Config{BranchProtection: tt.config}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			branchChanges := 0
			for _, c := range plan.Changes {
				if c.Category == "branch_protection" {
					branchChanges++
					if tt.isAdd && c.Type != ChangeAdd {
						t.Errorf("expected add change, got %v", c.Type)
					}
				}
			}

			if branchChanges != tt.expectedCount {
				t.Errorf("expected %d branch protection changes, got %d", tt.expectedCount, branchChanges)
			}
		})
	}
}

func TestCalculatorCheckSecrets(t *testing.T) {
	tests := []struct {
		name           string
		currentSecrets []string
		configSecrets  []string
		expectMissing  int
	}{
		{
			name:           "all secrets present",
			currentSecrets: []string{"API_KEY", "DEPLOY_TOKEN"},
			configSecrets:  []string{"API_KEY", "DEPLOY_TOKEN"},
			expectMissing:  0,
		},
		{
			name:           "some secrets missing",
			currentSecrets: []string{"API_KEY"},
			configSecrets:  []string{"API_KEY", "DEPLOY_TOKEN", "SECRET_KEY"},
			expectMissing:  2,
		},
		{
			name:           "all secrets missing",
			currentSecrets: []string{},
			configSecrets:  []string{"API_KEY"},
			expectMissing:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Secrets = tt.currentSecrets

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Secrets: tt.configSecrets,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckSecrets: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			missingCount := 0
			for _, c := range plan.Changes {
				if c.Category == "secrets" && c.Type == ChangeMissing {
					missingCount++
				}
			}

			if missingCount != tt.expectMissing {
				t.Errorf("expected %d missing secrets, got %d", tt.expectMissing, missingCount)
			}
		})
	}
}

func TestCalculatorCheckVariables(t *testing.T) {
	tests := []struct {
		name        string
		currentVars []github.VariableData
		configVars  map[string]string
		expectAdds  int
	}{
		{
			name:        "all variables present with same values",
			currentVars: []github.VariableData{{Name: "NODE_ENV", Value: "prod"}, {Name: "DEBUG", Value: "true"}},
			configVars:  map[string]string{"NODE_ENV": "prod"},
			expectAdds:  0,
		},
		{
			name:        "some variables to add",
			currentVars: []github.VariableData{{Name: "NODE_ENV", Value: "prod"}},
			configVars:  map[string]string{"NODE_ENV": "prod", "LOG_LEVEL": "info"},
			expectAdds:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Variables = tt.currentVars

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Variables: tt.configVars,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckEnv: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			addCount := 0
			for _, c := range plan.Changes {
				if c.Category == "variables" && c.Type == ChangeAdd {
					addCount++
				}
			}

			if addCount != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, addCount)
			}
		})
	}
}

func TestPlanHasChanges(t *testing.T) {
	tests := []struct {
		name     string
		changes  []Change
		expected bool
	}{
		{
			name:     "no changes",
			changes:  []Change{},
			expected: false,
		},
		{
			name:     "has changes",
			changes:  []Change{{Type: ChangeUpdate}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Changes: tt.changes}
			if plan.HasChanges() != tt.expected {
				t.Errorf("expected HasChanges() = %v", tt.expected)
			}
		})
	}
}

func TestPlanHasMissingSecrets(t *testing.T) {
	tests := []struct {
		name     string
		changes  []Change
		expected bool
	}{
		{
			name:     "no missing secrets",
			changes:  []Change{{Category: "repo", Type: ChangeUpdate}},
			expected: false,
		},
		{
			name:     "has missing secrets",
			changes:  []Change{{Category: "secrets", Type: ChangeMissing}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Changes: tt.changes}
			if plan.HasMissingSecrets() != tt.expected {
				t.Errorf("expected HasMissingSecrets() = %v", tt.expected)
			}
		})
	}
}

func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeAdd, "add"},
		{ChangeUpdate, "update"},
		{ChangeDelete, "delete"},
		{ChangeMissing, "missing"},
		{ChangeType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.changeType.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.changeType.String())
			}
		})
	}
}

func TestPlanHasMissingVariables(t *testing.T) {
	tests := []struct {
		name     string
		changes  []Change
		expected bool
	}{
		{
			name:     "no missing variables",
			changes:  []Change{{Category: "repo", Type: ChangeUpdate}},
			expected: false,
		},
		{
			name:     "has missing variables",
			changes:  []Change{{Category: "env", Type: ChangeMissing}},
			expected: true,
		},
		{
			name:     "has env but not missing type",
			changes:  []Change{{Category: "env", Type: ChangeUpdate}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Changes: tt.changes}
			if plan.HasMissingVariables() != tt.expected {
				t.Errorf("expected HasMissingVariables() = %v", tt.expected)
			}
		})
	}
}

func TestCalculatorCompareTopics(t *testing.T) {
	tests := []struct {
		name          string
		currentTopics []string
		configTopics  []string
		expectChange  bool
	}{
		{
			name:          "no changes",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go", "cli"},
			expectChange:  false,
		},
		{
			name:          "topics changed",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go", "github"},
			expectChange:  true,
		},
		{
			name:          "topics added",
			currentTopics: []string{"go"},
			configTopics:  []string{"go", "cli"},
			expectChange:  true,
		},
		{
			name:          "topics removed",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go"},
			expectChange:  true,
		},
		{
			name:          "empty to some",
			currentTopics: []string{},
			configTopics:  []string{"go"},
			expectChange:  true,
		},
		{
			name:          "order changed - no change",
			currentTopics: []string{"cli", "go"},
			configTopics:  []string{"go", "cli"},
			expectChange:  false, // order is ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.RepoData = &github.RepoData{
				Topics: tt.currentTopics,
			}

			cfg := &config.Config{Topics: tt.configTopics}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasTopicsChange := false
			for _, c := range plan.Changes {
				if c.Category == "topics" {
					hasTopicsChange = true
					break
				}
			}

			if hasTopicsChange != tt.expectChange {
				t.Errorf("expected topics change = %v, got %v", tt.expectChange, hasTopicsChange)
			}
		})
	}
}

func TestCompareBranchRuleAllFields(t *testing.T) {
	// Test all branch rule fields for coverage
	current := &github.BranchProtectionData{
		RequiredPullRequestReviews: &struct {
			RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
			DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
			RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
		}{
			RequiredApprovingReviewCount: 1,
			DismissStaleReviews:          false,
			RequireCodeOwnerReviews:      false,
		},
		RequiredStatusChecks: &struct {
			Strict   bool     `json:"strict"`
			Contexts []string `json:"contexts"`
		}{
			Strict:   false,
			Contexts: []string{"test"},
		},
		EnforceAdmins: &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: false},
		RequiredLinearHistory: &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: false},
		AllowForcePushes: &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: false},
		AllowDeletions: &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: false},
		RequiredSignatures: &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: false},
	}

	desired := &config.BranchRule{
		RequiredReviews:      ptr(2),
		DismissStaleReviews:  ptr(true),
		RequireCodeOwner:     ptr(true),
		StrictStatusChecks:   ptr(true),
		StatusChecks:         []string{"test", "build"},
		EnforceAdmins:        ptr(true),
		RequireLinearHistory: ptr(true),
		AllowForcePushes:     ptr(true),
		AllowDeletions:       ptr(true),
		RequireSignedCommits: ptr(true),
	}

	changes := compareBranchRule("main", current, desired)

	expectedKeys := map[string]bool{
		"main.required_reviews":       true,
		"main.dismiss_stale_reviews":  true,
		"main.require_code_owner":     true,
		"main.strict_status_checks":   true,
		"main.status_checks":          true,
		"main.enforce_admins":         true,
		"main.require_linear_history": true,
		"main.allow_force_pushes":     true,
		"main.allow_deletions":        true,
		"main.require_signed_commits": true,
	}

	if len(changes) != len(expectedKeys) {
		t.Errorf("expected %d changes, got %d", len(expectedKeys), len(changes))
	}

	for _, c := range changes {
		if !expectedKeys[c.Key] {
			t.Errorf("unexpected change key: %s", c.Key)
		}
	}
}

func TestCompareBranchRuleNilCurrent(t *testing.T) {
	// Test with nil structs in current (defaults to false/0)
	current := &github.BranchProtectionData{}

	desired := &config.BranchRule{
		RequiredReviews:      ptr(2),
		DismissStaleReviews:  ptr(true),
		RequireCodeOwner:     ptr(true),
		StrictStatusChecks:   ptr(true),
		StatusChecks:         []string{"test"},
		EnforceAdmins:        ptr(true),
		RequireLinearHistory: ptr(true),
		AllowForcePushes:     ptr(true),
		AllowDeletions:       ptr(true),
		RequireSignedCommits: ptr(true),
	}

	changes := compareBranchRule("main", current, desired)

	// All should have changes since current defaults are 0/false
	if len(changes) != 10 {
		t.Errorf("expected 10 changes, got %d", len(changes))
	}
}

func TestFormatBranchRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     *config.BranchRule
		contains []string
	}{
		{
			name:     "empty rule",
			rule:     &config.BranchRule{},
			contains: []string{"new protection"},
		},
		{
			name: "full rule",
			rule: &config.BranchRule{
				RequiredReviews:      ptr(2),
				DismissStaleReviews:  ptr(true),
				RequireCodeOwner:     ptr(true),
				StrictStatusChecks:   ptr(true),
				StatusChecks:         []string{"test", "build"},
				EnforceAdmins:        ptr(true),
				RequireLinearHistory: ptr(true),
				RequireSignedCommits: ptr(true),
				AllowForcePushes:     ptr(true),
				AllowDeletions:       ptr(true),
			},
			contains: []string{
				"required_reviews=2",
				"dismiss_stale_reviews=true",
				"require_code_owner=true",
				"strict_status_checks=true",
				"status_checks=",
				"enforce_admins=true",
				"require_linear_history=true",
				"require_signed_commits=true",
				"allow_force_pushes=true",
				"allow_deletions=true",
			},
		},
		{
			name: "partial rule - false values not shown",
			rule: &config.BranchRule{
				RequiredReviews:     ptr(1),
				DismissStaleReviews: ptr(false),
			},
			contains: []string{"required_reviews=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBranchRule(tt.rule)
			for _, s := range tt.contains {
				if !contains(result, s) {
					t.Errorf("expected result to contain %q, got %q", s, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCalculatorCompareActions(t *testing.T) {
	tests := []struct {
		name                string
		currentPerms        *github.ActionsPermissionsData
		currentWorkflow     *github.ActionsWorkflowPermissionsData
		currentSelected     *github.ActionsSelectedData
		config              *config.ActionsConfig
		expectedChangeCount int
		expectedKeys        []string
	}{
		{
			name: "no changes",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "all",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				Enabled:                      ptr(true),
				AllowedActions:               ptr("all"),
				DefaultWorkflowPermissions:   ptr("read"),
				CanApprovePullRequestReviews: ptr(false),
			},
			expectedChangeCount: 0,
			expectedKeys:        []string{},
		},
		{
			name: "enabled change",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "all",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				Enabled: ptr(false),
			},
			expectedChangeCount: 1,
			expectedKeys:        []string{"enabled"},
		},
		{
			name: "allowed_actions change",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "all",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				AllowedActions: ptr("local_only"),
			},
			expectedChangeCount: 1,
			expectedKeys:        []string{"allowed_actions"},
		},
		{
			name: "workflow permissions change",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "all",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				DefaultWorkflowPermissions:   ptr("write"),
				CanApprovePullRequestReviews: ptr(true),
			},
			expectedChangeCount: 2,
			expectedKeys:        []string{"default_workflow_permissions", "can_approve_pull_request_reviews"},
		},
		{
			name: "multiple changes",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "all",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				Enabled:                      ptr(true),
				AllowedActions:               ptr("selected"),
				DefaultWorkflowPermissions:   ptr("write"),
				CanApprovePullRequestReviews: ptr(true),
			},
			expectedChangeCount: 3,
			expectedKeys:        []string{"allowed_actions", "default_workflow_permissions", "can_approve_pull_request_reviews"},
		},
		{
			name: "selected actions change",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: "selected",
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			currentSelected: &github.ActionsSelectedData{
				GithubOwnedAllowed: true,
				VerifiedAllowed:    false,
				PatternsAllowed:    []string{},
			},
			config: &config.ActionsConfig{
				SelectedActions: &config.SelectedActionsConfig{
					GithubOwnedAllowed: ptr(false),
					VerifiedAllowed:    ptr(true),
				},
			},
			expectedChangeCount: 2,
			expectedKeys:        []string{"github_owned_allowed", "verified_allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.ActionsPermissions = tt.currentPerms
			mock.ActionsWorkflowPerms = tt.currentWorkflow
			if tt.currentSelected != nil {
				mock.ActionsSelected = tt.currentSelected
			}

			cfg := &config.Config{Actions: tt.config}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			actionsChanges := 0
			foundKeys := make(map[string]bool)
			for _, c := range plan.Changes {
				if c.Category == "actions" {
					actionsChanges++
					foundKeys[c.Key] = true
				}
			}

			if actionsChanges != tt.expectedChangeCount {
				t.Errorf("expected %d actions changes, got %d", tt.expectedChangeCount, actionsChanges)
			}

			for _, key := range tt.expectedKeys {
				if !foundKeys[key] {
					t.Errorf("expected change for key %q not found", key)
				}
			}
		})
	}
}

func TestCompareActionsWithPatternsAllowed(t *testing.T) {
	mock := github.NewMockClient()
	mock.ActionsPermissions = &github.ActionsPermissionsData{
		Enabled:        true,
		AllowedActions: "selected",
	}
	mock.ActionsWorkflowPerms = &github.ActionsWorkflowPermissionsData{
		DefaultWorkflowPermissions:   "read",
		CanApprovePullRequestReviews: false,
	}
	mock.ActionsSelected = &github.ActionsSelectedData{
		GithubOwnedAllowed: true,
		VerifiedAllowed:    false,
		PatternsAllowed:    []string{"actions/*"},
	}

	cfg := &config.Config{
		Actions: &config.ActionsConfig{
			SelectedActions: &config.SelectedActionsConfig{
				PatternsAllowed: []string{"actions/*", "github/*"},
			},
		},
	}
	calc := NewCalculator(mock, cfg)

	plan, err := calc.Calculate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, c := range plan.Changes {
		if c.Key == "patterns_allowed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected patterns_allowed change")
	}
}

func TestCalculatorCompareRepoAllFields(t *testing.T) {
	// Test all repo fields for coverage
	mock := github.NewMockClient()
	mock.RepoData = &github.RepoData{
		Description:         ptr("old"),
		Homepage:            ptr("https://old.com"),
		Visibility:          "private",
		AllowMergeCommit:    true,
		AllowRebaseMerge:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
		AllowUpdateBranch:   false,
	}

	cfg := &config.Config{
		Repo: &config.RepoConfig{
			Description:         ptr("new"),
			Homepage:            ptr("https://new.com"),
			Visibility:          ptr("public"),
			AllowMergeCommit:    ptr(false),
			AllowRebaseMerge:    ptr(false),
			AllowSquashMerge:    ptr(false),
			DeleteBranchOnMerge: ptr(true),
			AllowUpdateBranch:   ptr(true),
		},
	}
	calc := NewCalculator(mock, cfg)

	plan, err := calc.Calculate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKeys := map[string]bool{
		"description":            true,
		"homepage":               true,
		"visibility":             true,
		"allow_merge_commit":     true,
		"allow_rebase_merge":     true,
		"allow_squash_merge":     true,
		"delete_branch_on_merge": true,
		"allow_update_branch":    true,
	}

	for _, c := range plan.Changes {
		if c.Category == "repo" {
			if !expectedKeys[c.Key] {
				t.Errorf("unexpected repo change key: %s", c.Key)
			}
			delete(expectedKeys, c.Key)
		}
	}

	if len(expectedKeys) > 0 {
		t.Errorf("missing repo changes: %v", expectedKeys)
	}
}

func TestCalculatorErrors(t *testing.T) {
	t.Run("GetRepo error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetRepoError = apperrors.ErrRepoNotFound

		cfg := &config.Config{Repo: &config.RepoConfig{Description: ptr("test")}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetLabels error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetLabelsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Labels: &config.LabelsConfig{Items: []config.Label{{Name: "bug", Color: "d73a4a"}}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetBranchProtection error (not ErrBranchNotProtected)", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetBranchProtectionError = apperrors.ErrPermissionDenied

		cfg := &config.Config{BranchProtection: map[string]*config.BranchRule{"main": {RequiredReviews: ptr(1)}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetSecrets error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetSecretsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Env: &config.EnvConfig{Secrets: []string{"KEY"}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{CheckSecrets: true})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetVariables error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetVariablesError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Env: &config.EnvConfig{Variables: map[string]string{"VAR": "value"}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{CheckEnv: true})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetActionsPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetActionsPermissionsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Actions: &config.ActionsConfig{Enabled: ptr(true)}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetActionsWorkflowPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.ActionsPermissions = &github.ActionsPermissionsData{Enabled: true}
		mock.GetActionsWorkflowPermissionsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Actions: &config.ActionsConfig{Enabled: ptr(true)}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestJoinParts(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "empty",
			parts:    []string{},
			expected: "",
		},
		{
			name:     "single",
			parts:    []string{"a=1"},
			expected: "a=1",
		},
		{
			name:     "multiple",
			parts:    []string{"a=1", "b=2", "c=3"},
			expected: "a=1, b=2, c=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinParts(tt.parts)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPtrEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        *string
		b        *string
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "a nil b not nil",
			a:        nil,
			b:        ptr("test"),
			expected: false,
		},
		{
			name:     "a not nil b nil",
			a:        ptr("test"),
			b:        nil,
			expected: false,
		},
		{
			name:     "both equal",
			a:        ptr("test"),
			b:        ptr("test"),
			expected: true,
		},
		{
			name:     "both not equal",
			a:        ptr("test1"),
			b:        ptr("test2"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ptrEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPtrVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "",
		},
		{
			name:     "not nil",
			input:    ptr("test"),
			expected: "test",
		},
		{
			name:     "empty string",
			input:    ptr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ptrVal(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringSliceEqualIgnoreOrder(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "same order",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different order",
			a:        []string{"c", "a", "b"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different length",
			a:        []string{"a", "b"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "different content",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "d"},
			expected: false,
		},
		{
			name:     "duplicates same",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "b", "a"},
			expected: true,
		},
		{
			name:     "duplicates different count",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "b", "b"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSliceEqualIgnoreOrder(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("stringSliceEqualIgnoreOrder(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
