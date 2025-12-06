package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

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
				Description: nullStr("Test"),
				Visibility:  ptr("public"),
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
				Description: nullStr("Old"),
			},
			config: &config.RepoConfig{
				Description: ptr("New"),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: CategoryRepo,
					Key:      "description",
					Old:      "Old",
					New:      "New",
				},
			},
		},
		{
			name: "visibility change",
			current: &github.RepoData{
				Visibility: ptr("private"),
			},
			config: &config.RepoConfig{
				Visibility: ptr("public"),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: CategoryRepo,
					Key:      "visibility",
					Old:      "private",
					New:      "public",
				},
			},
		},
		{
			name: "merge options change",
			current: &github.RepoData{
				AllowMergeCommit: ptr(true),
				AllowSquashMerge: ptr(false),
			},
			config: &config.RepoConfig{
				AllowMergeCommit: ptr(false),
				AllowSquashMerge: ptr(true),
			},
			expected: []Change{
				{
					Type:     ChangeUpdate,
					Category: CategoryRepo,
					Key:      "allow_merge_commit",
					Old:      true,
					New:      false,
				},
				{
					Type:     ChangeUpdate,
					Category: CategoryRepo,
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

			if len(plan.Changes()) != len(tt.expected) {
				t.Errorf("expected %d changes, got %d", len(tt.expected), len(plan.Changes()))
				return
			}

			for i, exp := range tt.expected {
				got := plan.Changes()[i]
				if got.Type != exp.Type || got.Category != exp.Category || got.Key != exp.Key {
					t.Errorf("change %d mismatch: expected %+v, got %+v", i, exp, got)
				}
			}
		})
	}
}

func TestCalculatorCompareRepoAllFields(t *testing.T) {
	mock := github.NewMockClient()
	//nolint:unusedwrite // fields are used by Calculator.Calculate()
	mock.RepoData = &github.RepoData{
		Description:         nullStr("old"),
		Homepage:            nullStr("https://old.com"),
		Visibility:          ptr("private"),
		AllowMergeCommit:    ptr(true),
		AllowRebaseMerge:    ptr(true),
		AllowSquashMerge:    ptr(true),
		DeleteBranchOnMerge: ptr(false),
		AllowUpdateBranch:   ptr(false),
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

	for _, c := range plan.Changes() {
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
