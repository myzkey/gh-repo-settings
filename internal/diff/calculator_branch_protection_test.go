package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
	"github.com/myzkey/gh-repo-settings/internal/infra/githubopenapi"
)

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
					RequiredPullRequestReviews: &githubopenapi.ProtectedBranchPullRequestReview{
						RequiredApprovingReviewCount: ptr(1),
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
					RequiredPullRequestReviews: &githubopenapi.ProtectedBranchPullRequestReview{
						RequiredApprovingReviewCount: ptr(2),
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
			for _, c := range plan.Changes() {
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
