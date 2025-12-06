package comparator

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestPagesComparator_Compare(t *testing.T) {
	tests := []struct {
		name          string
		currentPages  *github.PagesData
		getPagesError error
		config        *config.PagesConfig
		expectAdds    int
		expectUpdates int
	}{
		{
			name: "no changes when config matches",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("workflow"),
			},
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectAdds:    0,
			expectUpdates: 0,
		},
		{
			name:          "add pages when not enabled",
			currentPages:  nil,
			getPagesError: apperrors.ErrPagesNotEnabled,
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectAdds:    1,
			expectUpdates: 0,
		},
		{
			name: "update build_type",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("legacy"),
			},
			config: &config.PagesConfig{
				BuildType: ptr("workflow"),
			},
			expectAdds:    0,
			expectUpdates: 1,
		},
		{
			name: "update source branch",
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
			expectAdds:    0,
			expectUpdates: 1,
		},
		{
			name: "update source path",
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
			expectAdds:    0,
			expectUpdates: 1,
		},
		{
			name: "nil config fields produce no changes",
			currentPages: &github.PagesData{
				BuildType: nullBuildType("workflow"),
			},
			config:        &config.PagesConfig{},
			expectAdds:    0,
			expectUpdates: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.PagesData = tt.currentPages
			if tt.getPagesError != nil {
				mock.GetPagesError = tt.getPagesError
			}

			comparator := NewPagesComparator(mock, tt.config)
			plan, err := comparator.Compare(context.Background())

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var adds, updates int
			for _, c := range plan.Changes() {
				if c.Category != model.CategoryPages {
					t.Errorf("expected category %s, got %s", model.CategoryPages, c.Category)
				}
				switch c.Type {
				case model.ChangeAdd:
					adds++
				case model.ChangeUpdate:
					updates++
				}
			}

			if adds != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, adds)
			}
			if updates != tt.expectUpdates {
				t.Errorf("expected %d updates, got %d", tt.expectUpdates, updates)
			}
		})
	}
}

func TestPagesComparator_AddPagesDefaultBuildType(t *testing.T) {
	mock := github.NewMockClient()
	mock.GetPagesError = apperrors.ErrPagesNotEnabled

	// Config without explicit build_type should default to "workflow"
	comparator := NewPagesComparator(mock, &config.PagesConfig{})
	plan, err := comparator.Compare(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Size() != 1 {
		t.Fatalf("expected 1 change, got %d", plan.Size())
	}

	change := plan.Changes()[0]
	if change.Type != model.ChangeAdd {
		t.Errorf("expected add change, got %v", change.Type)
	}

	// Check that default build_type is "workflow"
	newVal, ok := change.New.(string)
	if !ok {
		t.Fatalf("expected string value, got %T", change.New)
	}
	if newVal != "build_type=workflow" {
		t.Errorf("expected 'build_type=workflow', got %q", newVal)
	}
}

func TestPagesComparator_GetPagesError(t *testing.T) {
	mock := github.NewMockClient()
	mock.GetPagesError = apperrors.ErrPermissionDenied

	comparator := NewPagesComparator(mock, &config.PagesConfig{
		BuildType: ptr("workflow"),
	})

	_, err := comparator.Compare(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
