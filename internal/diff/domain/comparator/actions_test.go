package comparator

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestActionsComparator_ComparePermissions(t *testing.T) {
	tests := []struct {
		name         string
		currentPerms *github.ActionsPermissionsData
		config       *config.ActionsConfig
		expectedKeys []string
	}{
		{
			name: "no changes when config matches",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("all"),
			},
			config: &config.ActionsConfig{
				Enabled:        ptr(true),
				AllowedActions: ptr("all"),
			},
			expectedKeys: []string{},
		},
		{
			name: "enabled change detected",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("all"),
			},
			config: &config.ActionsConfig{
				Enabled: ptr(false),
			},
			expectedKeys: []string{"enabled"},
		},
		{
			name: "allowed_actions change detected",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("all"),
			},
			config: &config.ActionsConfig{
				AllowedActions: ptr("selected"),
			},
			expectedKeys: []string{"allowed_actions"},
		},
		{
			name: "nil config fields produce no changes",
			currentPerms: &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("all"),
			},
			config:       &config.ActionsConfig{},
			expectedKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.ActionsPermissions = tt.currentPerms
			mock.ActionsWorkflowPerms = &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions: "read",
			}

			comparator := NewActionsComparator(mock, tt.config)
			plan, err := comparator.Compare(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			keySet := make(map[string]bool)
			for _, c := range plan.Changes() {
				keySet[c.Key] = true
				if c.Category != model.CategoryActions {
					t.Errorf("expected category %s, got %s", model.CategoryActions, c.Category)
				}
			}

			if len(keySet) != len(tt.expectedKeys) {
				t.Errorf("expected %d changes, got %d", len(tt.expectedKeys), len(keySet))
			}

			for _, key := range tt.expectedKeys {
				if !keySet[key] {
					t.Errorf("expected change for key %q not found", key)
				}
			}
		})
	}
}

func TestActionsComparator_CompareSelectedActions(t *testing.T) {
	tests := []struct {
		name            string
		currentSelected *github.ActionsSelectedData
		config          *config.ActionsConfig
		expectedKeys    []string
	}{
		{
			name: "github_owned_allowed change",
			currentSelected: &github.ActionsSelectedData{
				GithubOwnedAllowed: ptr(true),
				VerifiedAllowed:    ptr(false),
			},
			config: &config.ActionsConfig{
				SelectedActions: &config.SelectedActionsConfig{
					GithubOwnedAllowed: ptr(false),
				},
			},
			expectedKeys: []string{"github_owned_allowed"},
		},
		{
			name: "verified_allowed change",
			currentSelected: &github.ActionsSelectedData{
				GithubOwnedAllowed: ptr(true),
				VerifiedAllowed:    ptr(false),
			},
			config: &config.ActionsConfig{
				SelectedActions: &config.SelectedActionsConfig{
					VerifiedAllowed: ptr(true),
				},
			},
			expectedKeys: []string{"verified_allowed"},
		},
		{
			name: "patterns_allowed change",
			currentSelected: &github.ActionsSelectedData{
				PatternsAllowed: &[]string{"actions/*"},
			},
			config: &config.ActionsConfig{
				SelectedActions: &config.SelectedActionsConfig{
					PatternsAllowed: []string{"actions/*", "github/*"},
				},
			},
			expectedKeys: []string{"patterns_allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.ActionsPermissions = &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("selected"),
			}
			mock.ActionsSelected = tt.currentSelected
			mock.ActionsWorkflowPerms = &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions: "read",
			}

			comparator := NewActionsComparator(mock, tt.config)
			plan, err := comparator.Compare(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			keySet := make(map[string]bool)
			for _, c := range plan.Changes() {
				keySet[c.Key] = true
			}

			for _, key := range tt.expectedKeys {
				if !keySet[key] {
					t.Errorf("expected change for key %q not found", key)
				}
			}
		})
	}
}

func TestActionsComparator_CompareWorkflowPermissions(t *testing.T) {
	tests := []struct {
		name            string
		currentWorkflow *github.ActionsWorkflowPermissionsData
		config          *config.ActionsConfig
		expectedKeys    []string
	}{
		{
			name: "default_workflow_permissions change",
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				DefaultWorkflowPermissions: ptr("write"),
			},
			expectedKeys: []string{"default_workflow_permissions"},
		},
		{
			name: "can_approve_pull_request_reviews change",
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				CanApprovePullRequestReviews: ptr(true),
			},
			expectedKeys: []string{"can_approve_pull_request_reviews"},
		},
		{
			name: "both workflow permissions change",
			currentWorkflow: &github.ActionsWorkflowPermissionsData{
				DefaultWorkflowPermissions:   "read",
				CanApprovePullRequestReviews: false,
			},
			config: &config.ActionsConfig{
				DefaultWorkflowPermissions:   ptr("write"),
				CanApprovePullRequestReviews: ptr(true),
			},
			expectedKeys: []string{"default_workflow_permissions", "can_approve_pull_request_reviews"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.ActionsPermissions = &github.ActionsPermissionsData{
				Enabled:        true,
				AllowedActions: allowedActions("all"),
			}
			mock.ActionsWorkflowPerms = tt.currentWorkflow

			comparator := NewActionsComparator(mock, tt.config)
			plan, err := comparator.Compare(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			keySet := make(map[string]bool)
			for _, c := range plan.Changes() {
				keySet[c.Key] = true
			}

			for _, key := range tt.expectedKeys {
				if !keySet[key] {
					t.Errorf("expected change for key %q not found", key)
				}
			}
		})
	}
}

func TestActionsComparator_Errors(t *testing.T) {
	t.Run("GetActionsPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetActionsPermissionsError = apperrors.ErrPermissionDenied

		comparator := NewActionsComparator(mock, &config.ActionsConfig{
			Enabled: ptr(true),
		})

		_, err := comparator.Compare(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("GetActionsWorkflowPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.ActionsPermissions = &github.ActionsPermissionsData{Enabled: true}
		mock.GetActionsWorkflowPermissionsError = apperrors.ErrPermissionDenied

		comparator := NewActionsComparator(mock, &config.ActionsConfig{
			Enabled: ptr(true),
		})

		_, err := comparator.Compare(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
