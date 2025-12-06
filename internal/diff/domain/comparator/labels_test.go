package comparator

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestLabelsComparator_Compare(t *testing.T) {
	tests := []struct {
		name        string
		current     []github.LabelData
		config      *config.LabelsConfig
		expectAdds  int
		expectUpds  int
		expectDels  int
		expectError bool
	}{
		{
			name: "no changes when labels match",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a", Description: nullStr("Bug report")},
			},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a", Description: "Bug report"},
				},
			},
			expectAdds: 0,
			expectUpds: 0,
			expectDels: 0,
		},
		{
			name:    "add new label",
			current: []github.LabelData{},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a", Description: "Bug report"},
				},
			},
			expectAdds: 1,
			expectUpds: 0,
			expectDels: 0,
		},
		{
			name: "update existing label color",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a", Description: nullStr("Bug report")},
			},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "ff0000", Description: "Bug report"},
				},
			},
			expectAdds: 0,
			expectUpds: 1,
			expectDels: 0,
		},
		{
			name: "update existing label description",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a", Description: nullStr("Old description")},
			},
			config: &config.LabelsConfig{
				Items: []config.Label{
					{Name: "bug", Color: "d73a4a", Description: "New description"},
				},
			},
			expectAdds: 0,
			expectUpds: 1,
			expectDels: 0,
		},
		{
			name: "delete label with replace_default=true",
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
			expectAdds: 0,
			expectUpds: 0,
			expectDels: 1,
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
			expectAdds: 0,
			expectUpds: 0,
			expectDels: 0,
		},
		{
			name: "multiple operations",
			current: []github.LabelData{
				{Name: "bug", Color: "d73a4a"},
				{Name: "to-delete", Color: "000000"},
			},
			config: &config.LabelsConfig{
				ReplaceDefault: true,
				Items: []config.Label{
					{Name: "bug", Color: "ff0000"},       // update
					{Name: "feature", Color: "a2eeef"},   // add
				},
			},
			expectAdds: 1,
			expectUpds: 1,
			expectDels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Labels = tt.current

			comparator := NewLabelsComparator(mock, tt.config)
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

			var adds, upds, dels int
			for _, c := range plan.Changes() {
				if c.Category != model.CategoryLabels {
					t.Errorf("expected category %s, got %s", model.CategoryLabels, c.Category)
				}
				switch c.Type {
				case model.ChangeAdd:
					adds++
				case model.ChangeUpdate:
					upds++
				case model.ChangeDelete:
					dels++
				}
			}

			if adds != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, adds)
			}
			if upds != tt.expectUpds {
				t.Errorf("expected %d updates, got %d", tt.expectUpds, upds)
			}
			if dels != tt.expectDels {
				t.Errorf("expected %d deletes, got %d", tt.expectDels, dels)
			}
		})
	}
}

func TestLabelsComparator_GetLabelsError(t *testing.T) {
	mock := github.NewMockClient()
	mock.GetLabelsError = apperrors.ErrPermissionDenied

	comparator := NewLabelsComparator(mock, &config.LabelsConfig{
		Items: []config.Label{{Name: "bug", Color: "d73a4a"}},
	})

	_, err := comparator.Compare(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFormatLabel(t *testing.T) {
	tests := []struct {
		color       string
		description string
		expected    string
	}{
		{
			color:       "d73a4a",
			description: "Bug report",
			expected:    "color=d73a4a, description=Bug report",
		},
		{
			color:       "ffffff",
			description: "",
			expected:    "color=ffffff, description=",
		},
	}

	for _, tt := range tests {
		result := formatLabel(tt.color, tt.description)
		if result != tt.expected {
			t.Errorf("formatLabel(%q, %q) = %q, want %q", tt.color, tt.description, result, tt.expected)
		}
	}
}
