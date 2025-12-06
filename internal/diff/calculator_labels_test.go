package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

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
				{Name: "bug", Color: "d73a4a", Description: nullStr("Old description")},
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
			for _, c := range plan.Changes() {
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
