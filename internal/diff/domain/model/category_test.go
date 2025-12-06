package model

import (
	"testing"
)

// TestChangeCategoryInvariants tests the invariants of ChangeCategory
func TestChangeCategoryInvariants(t *testing.T) {
	t.Run("all categories have non-empty string representation", func(t *testing.T) {
		categories := []ChangeCategory{
			CategoryRepo,
			CategoryTopics,
			CategoryLabels,
			CategoryBranchProtection,
			CategoryVariables,
			CategorySecrets,
			CategoryActions,
			CategoryPages,
		}

		for _, cat := range categories {
			if cat.String() == "" {
				t.Errorf("category %v should have non-empty string", cat)
			}
		}
	})

	t.Run("all categories are unique", func(t *testing.T) {
		categories := []ChangeCategory{
			CategoryRepo,
			CategoryTopics,
			CategoryLabels,
			CategoryBranchProtection,
			CategoryVariables,
			CategorySecrets,
			CategoryActions,
			CategoryPages,
		}

		seen := make(map[ChangeCategory]bool)
		for _, cat := range categories {
			if seen[cat] {
				t.Errorf("duplicate category: %s", cat)
			}
			seen[cat] = true
		}
	})

	t.Run("CategoryEnv is alias for CategoryVariables", func(t *testing.T) {
		if CategoryEnv != CategoryVariables {
			t.Error("CategoryEnv should equal CategoryVariables")
		}
	})

	t.Run("String() returns underlying value", func(t *testing.T) {
		tests := []struct {
			category ChangeCategory
			expected string
		}{
			{CategoryRepo, "repo"},
			{CategoryTopics, "topics"},
			{CategoryLabels, "labels"},
			{CategoryBranchProtection, "branch_protection"},
			{CategoryVariables, "variables"},
			{CategorySecrets, "secrets"},
			{CategoryActions, "actions"},
			{CategoryPages, "pages"},
		}

		for _, tt := range tests {
			if tt.category.String() != tt.expected {
				t.Errorf("category.String() = %s, want %s", tt.category.String(), tt.expected)
			}
		}
	})
}

// TestChangeCategoryEquality tests category comparison
func TestChangeCategoryEquality(t *testing.T) {
	t.Run("same categories are equal", func(t *testing.T) {
		cat1 := CategoryLabels
		cat2 := CategoryLabels

		if cat1 != cat2 {
			t.Error("same categories should be equal")
		}
	})

	t.Run("different categories are not equal", func(t *testing.T) {
		if CategoryLabels == CategoryRepo {
			t.Error("different categories should not be equal")
		}
	})

	t.Run("category can be used as map key", func(t *testing.T) {
		m := make(map[ChangeCategory]int)
		m[CategoryLabels] = 1
		m[CategoryRepo] = 2
		m[CategoryLabels] = 3 // overwrite

		if m[CategoryLabels] != 3 {
			t.Error("category should work as map key")
		}
		if len(m) != 2 {
			t.Error("map should have 2 entries")
		}
	})
}

// TestChangeCategoryUsageInChange tests that categories work correctly in Change
func TestChangeCategoryUsageInChange(t *testing.T) {
	t.Run("Change preserves category through operations", func(t *testing.T) {
		change := NewAddChange(CategoryBranchProtection, "main.required_reviews", 2)

		if change.Category != CategoryBranchProtection {
			t.Error("Change should preserve category")
		}
	})

	t.Run("WithCategory changes category correctly", func(t *testing.T) {
		change := NewAddChange(CategoryLabels, "bug", "red")
		modified := change.WithCategory(CategoryRepo)

		if modified.Category != CategoryRepo {
			t.Error("WithCategory should change category")
		}
		if change.Category != CategoryLabels {
			t.Error("original should not be modified")
		}
	})
}

// TestChangeCategoryUsageInPlan tests that categories work correctly in Plan filtering
func TestChangeCategoryUsageInPlan(t *testing.T) {
	t.Run("FilterByCategory works with enum", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "bug", "red"))
		plan.Add(NewAddChange(CategoryRepo, "desc", "new"))
		plan.Add(NewAddChange(CategoryLabels, "feature", "blue"))

		filtered := plan.FilterByCategory(CategoryLabels)

		if filtered.Size() != 2 {
			t.Errorf("expected 2 labels, got %d", filtered.Size())
		}
	})

	t.Run("CountByCategory returns ChangeCategory keys", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan.Add(NewAddChange(CategoryRepo, "b", "2"))
		plan.Add(NewAddChange(CategoryLabels, "c", "3"))

		counts := plan.CountByCategory()

		if counts[CategoryLabels] != 2 {
			t.Errorf("expected 2 labels, got %d", counts[CategoryLabels])
		}
		if counts[CategoryRepo] != 1 {
			t.Errorf("expected 1 repo, got %d", counts[CategoryRepo])
		}
	})

	t.Run("Categories returns ChangeCategory slice", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan.Add(NewAddChange(CategorySecrets, "b", "2"))

		categories := plan.Categories()

		// Check that all returned values are valid ChangeCategory
		for _, cat := range categories {
			if cat != CategoryLabels && cat != CategorySecrets {
				t.Errorf("unexpected category: %s", cat)
			}
		}
	})
}

// TestChangeCategoryTypeComparison tests that string comparison still works
func TestChangeCategoryTypeComparison(t *testing.T) {
	t.Run("category can be compared with string value", func(t *testing.T) {
		// This tests backward compatibility - string comparison
		if string(CategoryLabels) != "labels" {
			t.Error("category string conversion should work")
		}
	})

	t.Run("category created from string matches constant", func(t *testing.T) {
		cat := ChangeCategory("labels")
		if cat != CategoryLabels {
			t.Error("category from string should match constant")
		}
	})
}
