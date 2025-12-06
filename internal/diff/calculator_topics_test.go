package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestCalculatorCompareTopics(t *testing.T) {
	tests := []struct {
		name          string
		currentTopics []string
		configTopics  []string
		expectChange  bool
	}{
		{
			name:          "no changes",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go", "cli"},
			expectChange:  false,
		},
		{
			name:          "topics changed",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go", "github"},
			expectChange:  true,
		},
		{
			name:          "topics added",
			currentTopics: []string{"go"},
			configTopics:  []string{"go", "cli"},
			expectChange:  true,
		},
		{
			name:          "topics removed",
			currentTopics: []string{"go", "cli"},
			configTopics:  []string{"go"},
			expectChange:  true,
		},
		{
			name:          "empty to some",
			currentTopics: []string{},
			configTopics:  []string{"go"},
			expectChange:  true,
		},
		{
			name:          "order changed - no change",
			currentTopics: []string{"cli", "go"},
			configTopics:  []string{"go", "cli"},
			expectChange:  false, // order is ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.RepoData = &github.RepoData{
				Topics: &tt.currentTopics,
			}

			cfg := &config.Config{Topics: tt.configTopics}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.Calculate(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasTopicsChange := false
			for _, c := range plan.Changes() {
				if c.Category == "topics" {
					hasTopicsChange = true
					break
				}
			}

			if hasTopicsChange != tt.expectChange {
				t.Errorf("expected topics change = %v, got %v", tt.expectChange, hasTopicsChange)
			}
		})
	}
}
