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
		name            string
		currentSecrets  []string
		requiredSecrets []string
		expectMissing   int
	}{
		{
			name:            "all secrets present",
			currentSecrets:  []string{"API_KEY", "DEPLOY_TOKEN"},
			requiredSecrets: []string{"API_KEY", "DEPLOY_TOKEN"},
			expectMissing:   0,
		},
		{
			name:            "some secrets missing",
			currentSecrets:  []string{"API_KEY"},
			requiredSecrets: []string{"API_KEY", "DEPLOY_TOKEN", "SECRET_KEY"},
			expectMissing:   2,
		},
		{
			name:            "all secrets missing",
			currentSecrets:  []string{},
			requiredSecrets: []string{"API_KEY"},
			expectMissing:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Secrets = tt.currentSecrets

			cfg := &config.Config{
				Secrets: &config.SecretsConfig{
					Required: tt.requiredSecrets,
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
		name             string
		currentVars      []string
		requiredVars     []string
		expectMissing    int
	}{
		{
			name:          "all variables present",
			currentVars:   []string{"NODE_ENV", "DEBUG"},
			requiredVars:  []string{"NODE_ENV"},
			expectMissing: 0,
		},
		{
			name:          "some variables missing",
			currentVars:   []string{"NODE_ENV"},
			requiredVars:  []string{"NODE_ENV", "LOG_LEVEL"},
			expectMissing: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Variables = tt.currentVars

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Required: tt.requiredVars,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckEnv: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			missingCount := 0
			for _, c := range plan.Changes {
				if c.Category == "env" && c.Type == ChangeMissing {
					missingCount++
				}
			}

			if missingCount != tt.expectMissing {
				t.Errorf("expected %d missing variables, got %d", tt.expectMissing, missingCount)
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
