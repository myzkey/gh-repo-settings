package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

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
				AllowedActions: allowedActions("all"),
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
				AllowedActions: allowedActions("all"),
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
				AllowedActions: allowedActions("all"),
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
				AllowedActions: allowedActions("all"),
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
				AllowedActions: allowedActions("all"),
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
				AllowedActions: allowedActions("selected"),
			},
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			currentSelected: &github.ActionsSelectedData{
				GithubOwnedAllowed: ptr(true),
				VerifiedAllowed:    ptr(false),
				PatternsAllowed:    &[]string{},
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
			for _, c := range plan.Changes() {
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
		AllowedActions: allowedActions("selected"),
	}
	mock.ActionsWorkflowPerms = &github.ActionsWorkflowPermissionsData{
		DefaultWorkflowPermissions:   "read",
		CanApprovePullRequestReviews: false,
	}
	mock.ActionsSelected = &github.ActionsSelectedData{
		GithubOwnedAllowed: ptr(true),
		VerifiedAllowed:    ptr(false),
		PatternsAllowed:    &[]string{"actions/*"},
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
	for _, c := range plan.Changes() {
		if c.Key == "patterns_allowed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected patterns_allowed change")
	}
}
