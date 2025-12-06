package comparator

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestRepoComparator_Compare(t *testing.T) {
	tests := []struct {
		name         string
		current      *github.RepoData
		config       *config.RepoConfig
		expectedKeys []string
		expectError  bool
	}{
		{
			name: "no changes when config matches current",
			current: &github.RepoData{
				Description: nullStr("Test repo"),
				Visibility:  ptr("public"),
			},
			config: &config.RepoConfig{
				Description: ptr("Test repo"),
				Visibility:  ptr("public"),
			},
			expectedKeys: []string{},
		},
		{
			name: "nil config fields produce no changes",
			current: &github.RepoData{
				Description: nullStr("Test repo"),
				Visibility:  ptr("public"),
			},
			config:       &config.RepoConfig{},
			expectedKeys: []string{},
		},
		{
			name: "description change detected",
			current: &github.RepoData{
				Description: nullStr("Old description"),
			},
			config: &config.RepoConfig{
				Description: ptr("New description"),
			},
			expectedKeys: []string{"description"},
		},
		{
			name: "visibility change detected",
			current: &github.RepoData{
				Visibility: ptr("private"),
			},
			config: &config.RepoConfig{
				Visibility: ptr("public"),
			},
			expectedKeys: []string{"visibility"},
		},
		{
			name: "boolean fields change detected",
			current: &github.RepoData{
				AllowMergeCommit:    ptr(true),
				AllowRebaseMerge:    ptr(true),
				AllowSquashMerge:    ptr(false),
				DeleteBranchOnMerge: ptr(false),
				AllowUpdateBranch:   ptr(false),
			},
			config: &config.RepoConfig{
				AllowMergeCommit:    ptr(false),
				AllowRebaseMerge:    ptr(false),
				AllowSquashMerge:    ptr(true),
				DeleteBranchOnMerge: ptr(true),
				AllowUpdateBranch:   ptr(true),
			},
			expectedKeys: []string{
				"allow_merge_commit",
				"allow_rebase_merge",
				"allow_squash_merge",
				"delete_branch_on_merge",
				"allow_update_branch",
			},
		},
		{
			name: "homepage change detected",
			current: &github.RepoData{
				Homepage: nullStr("https://old.example.com"),
			},
			config: &config.RepoConfig{
				Homepage: ptr("https://new.example.com"),
			},
			expectedKeys: []string{"homepage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.RepoData = tt.current

			comparator := NewRepoComparator(mock, tt.config)
			plan, err := comparator.Compare(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(plan.Changes()) != len(tt.expectedKeys) {
				t.Errorf("expected %d changes, got %d", len(tt.expectedKeys), len(plan.Changes()))
			}

			keySet := make(map[string]bool)
			for _, c := range plan.Changes() {
				keySet[c.Key] = true
				if c.Category != model.CategoryRepo {
					t.Errorf("expected category %s, got %s", model.CategoryRepo, c.Category)
				}
				if c.Type != model.ChangeUpdate {
					t.Errorf("expected type %s, got %s", model.ChangeUpdate, c.Type)
				}
			}

			for _, key := range tt.expectedKeys {
				if !keySet[key] {
					t.Errorf("expected change for key %q not found", key)
				}
			}
		})
	}
}

func TestRepoComparator_GetRepoError(t *testing.T) {
	mock := github.NewMockClient()
	mock.GetRepoError = apperrors.ErrRepoNotFound

	comparator := NewRepoComparator(mock, &config.RepoConfig{
		Description: ptr("test"),
	})

	_, err := comparator.Compare(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestTopicsComparator_Compare(t *testing.T) {
	tests := []struct {
		name         string
		current      []string
		desired      []string
		expectChange bool
	}{
		{
			name:         "no changes when topics match",
			current:      []string{"go", "cli"},
			desired:      []string{"go", "cli"},
			expectChange: false,
		},
		{
			name:         "no changes when order differs",
			current:      []string{"cli", "go"},
			desired:      []string{"go", "cli"},
			expectChange: false,
		},
		{
			name:         "change when topics added",
			current:      []string{"go"},
			desired:      []string{"go", "cli"},
			expectChange: true,
		},
		{
			name:         "change when topics removed",
			current:      []string{"go", "cli"},
			desired:      []string{"go"},
			expectChange: true,
		},
		{
			name:         "change when topics different",
			current:      []string{"go", "cli"},
			desired:      []string{"go", "github"},
			expectChange: true,
		},
		{
			name:         "empty to non-empty",
			current:      []string{},
			desired:      []string{"go"},
			expectChange: true,
		},
		{
			name:         "nil current treated as empty",
			current:      nil,
			desired:      []string{"go"},
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			if tt.current != nil {
				mock.RepoData = &github.RepoData{Topics: &tt.current}
			} else {
				mock.RepoData = &github.RepoData{Topics: nil}
			}

			comparator := NewTopicsComparator(mock, tt.desired)
			plan, err := comparator.Compare(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasChange := plan.HasChanges()
			if hasChange != tt.expectChange {
				t.Errorf("expected change=%v, got %v", tt.expectChange, hasChange)
			}

			if hasChange {
				change := plan.Changes()[0]
				if change.Category != model.CategoryTopics {
					t.Errorf("expected category %s, got %s", model.CategoryTopics, change.Category)
				}
				if change.Key != "topics" {
					t.Errorf("expected key 'topics', got %s", change.Key)
				}
			}
		})
	}
}
