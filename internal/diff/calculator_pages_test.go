package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestCalculatorComparePages(t *testing.T) {
	tests := []struct {
		name          string
		currentPages  *github.PagesData
		getPagesError error
		config        *config.PagesConfig
		expectChanges int
		expectAdd     bool
	}{
		{
			name:          "pages not enabled - create new",
			currentPages:  nil,
			getPagesError: apperrors.ErrPagesNotEnabled,
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectChanges: 1,
			expectAdd:     true,
		},
		{
			name: "no changes",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("workflow"),
			},
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectChanges: 0,
		},
		{
			name: "build type change",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("legacy"),
			},
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectChanges: 1,
		},
		{
			name: "source branch change",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("legacy"),
				Source: &github.PagesSourceData{
					Branch: "main",
					Path:   "/",
				},
			},
			config: &config.PagesConfig{
				BuildType: ptr("legacy"),
				Source: &config.PagesSourceConfig{
					Branch: ptr("gh-pages"),
					Path:   ptr("/"),
				},
			},
			expectChanges: 1,
		},
		{
			name: "source path change",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("legacy"),
				Source: &github.PagesSourceData{
					Branch: "main",
					Path:   "/",
				},
			},
			config: &config.PagesConfig{
				BuildType: ptr("legacy"),
				Source: &config.PagesSourceConfig{
					Branch: ptr("main"),
					Path:   ptr("/docs"),
				},
			},
			expectChanges: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.PagesData = tt.currentPages
			if tt.getPagesError != nil {
				mock.GetPagesError = tt.getPagesError
			}

			cfg := &config.Config{Pages: tt.config}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			pagesChanges := 0
			hasAdd := false
			for _, c := range plan.Changes() {
				if c.Category == "pages" {
					pagesChanges++
					if c.Type == ChangeAdd {
						hasAdd = true
					}
				}
			}

			if pagesChanges != tt.expectChanges {
				t.Errorf("expected %d pages changes, got %d", tt.expectChanges, pagesChanges)
			}
			if tt.expectAdd && !hasAdd {
				t.Error("expected add change, got none")
			}
		})
	}
}

func TestCalculatorGetPagesError(t *testing.T) {
	mock := github.NewMockClient()
	mock.GetPagesError = apperrors.ErrPermissionDenied

	cfg := &config.Config{Pages: &config.PagesConfig{BuildType: ptr("workflow")}}
	calc := NewCalculator(mock, cfg)

	_, err := calc.Calculate(context.Background())
	if err == nil {
		t.Error("expected error")
	}
}
