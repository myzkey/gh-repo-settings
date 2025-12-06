package model

import (
	"testing"
)

// TestPlanInvariants tests the invariants of the Plan type
func TestPlanInvariants(t *testing.T) {
	t.Run("NewPlan creates empty plan", func(t *testing.T) {
		plan := NewPlan()

		if !plan.IsEmpty() {
			t.Error("NewPlan should create empty plan")
		}
		if plan.Size() != 0 {
			t.Errorf("NewPlan.Size() = %d, want 0", plan.Size())
		}
		if plan.HasChanges() {
			t.Error("NewPlan should not have changes")
		}
	})

	t.Run("NewPlanFromChanges preserves all changes", func(t *testing.T) {
		changes := []Change{
			NewAddChange(CategoryLabels, "bug", "red"),
			NewUpdateChange(CategoryRepo, "desc", "old", "new"),
		}
		plan := NewPlanFromChanges(changes)

		if plan.Size() != len(changes) {
			t.Errorf("Size() = %d, want %d", plan.Size(), len(changes))
		}
	})

	t.Run("Add increases size by 1", func(t *testing.T) {
		plan := NewPlan()
		initialSize := plan.Size()

		plan.Add(NewAddChange(CategoryLabels, "bug", "red"))

		if plan.Size() != initialSize+1 {
			t.Errorf("Add should increase size by 1")
		}
	})

	t.Run("AddAll increases size by number of changes", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "existing", "value"))
		initialSize := plan.Size()

		newChanges := []Change{
			NewAddChange(CategoryLabels, "bug", "red"),
			NewAddChange(CategoryLabels, "feature", "blue"),
		}
		plan.AddAll(newChanges)

		if plan.Size() != initialSize+len(newChanges) {
			t.Errorf("AddAll should increase size by %d", len(newChanges))
		}
	})

	t.Run("Changes returns all added changes in order", func(t *testing.T) {
		plan := NewPlan()
		c1 := NewAddChange(CategoryLabels, "first", "1")
		c2 := NewAddChange(CategoryLabels, "second", "2")
		c3 := NewAddChange(CategoryLabels, "third", "3")

		plan.Add(c1)
		plan.Add(c2)
		plan.Add(c3)

		changes := plan.Changes()
		if len(changes) != 3 {
			t.Fatalf("expected 3 changes, got %d", len(changes))
		}
		if changes[0].Key != "first" || changes[1].Key != "second" || changes[2].Key != "third" {
			t.Error("Changes should preserve insertion order")
		}
	})
}

// TestPlanIsEmptyHasChangesConsistency tests IsEmpty and HasChanges are always opposite
func TestPlanIsEmptyHasChangesConsistency(t *testing.T) {
	t.Run("empty plan: IsEmpty=true, HasChanges=false", func(t *testing.T) {
		plan := NewPlan()

		if !plan.IsEmpty() {
			t.Error("empty plan should have IsEmpty() = true")
		}
		if plan.HasChanges() {
			t.Error("empty plan should have HasChanges() = false")
		}
		if plan.IsEmpty() == plan.HasChanges() {
			t.Error("IsEmpty and HasChanges should always be opposite")
		}
	})

	t.Run("non-empty plan: IsEmpty=false, HasChanges=true", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "bug", "red"))

		if plan.IsEmpty() {
			t.Error("non-empty plan should have IsEmpty() = false")
		}
		if !plan.HasChanges() {
			t.Error("non-empty plan should have HasChanges() = true")
		}
		if plan.IsEmpty() == plan.HasChanges() {
			t.Error("IsEmpty and HasChanges should always be opposite")
		}
	})
}

// TestPlanMergeInvariants tests the invariants of Plan.Merge()
func TestPlanMergeInvariants(t *testing.T) {
	t.Run("Merge preserves all changes from both plans", func(t *testing.T) {
		plan1 := NewPlan()
		plan1.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan1.Add(NewAddChange(CategoryLabels, "b", "2"))

		plan2 := NewPlan()
		plan2.Add(NewAddChange(CategoryRepo, "c", "3"))

		merged := plan1.Merge(plan2)

		if merged.Size() != plan1.Size()+plan2.Size() {
			t.Errorf("Merge should have %d changes, got %d", plan1.Size()+plan2.Size(), merged.Size())
		}
	})

	t.Run("Merge with nil returns copy of original", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))

		merged := plan.Merge(nil)

		if merged.Size() != plan.Size() {
			t.Errorf("Merge(nil) should preserve original size")
		}
	})

	t.Run("Merge does not mutate original plans", func(t *testing.T) {
		plan1 := NewPlan()
		plan1.Add(NewAddChange(CategoryLabels, "a", "1"))
		originalSize1 := plan1.Size()

		plan2 := NewPlan()
		plan2.Add(NewAddChange(CategoryRepo, "b", "2"))
		originalSize2 := plan2.Size()

		_ = plan1.Merge(plan2)

		if plan1.Size() != originalSize1 {
			t.Error("Merge should not mutate first plan")
		}
		if plan2.Size() != originalSize2 {
			t.Error("Merge should not mutate second plan")
		}
	})

	t.Run("Merge is associative: (a.Merge(b)).Merge(c) == a.Merge(b.Merge(c))", func(t *testing.T) {
		a := NewPlanFromChanges([]Change{NewAddChange(CategoryLabels, "a", "1")})
		b := NewPlanFromChanges([]Change{NewAddChange(CategoryLabels, "b", "2")})
		c := NewPlanFromChanges([]Change{NewAddChange(CategoryLabels, "c", "3")})

		left := a.Merge(b).Merge(c)
		right := a.Merge(b.Merge(c))

		if left.Size() != right.Size() {
			t.Error("Merge should be associative")
		}
	})

	t.Run("Merge with empty plan returns equivalent plan", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{NewAddChange(CategoryLabels, "a", "1")})
		empty := NewPlan()

		mergedRight := plan.Merge(empty)
		mergedLeft := empty.Merge(plan)

		if mergedRight.Size() != plan.Size() {
			t.Error("plan.Merge(empty) should have same size as plan")
		}
		if mergedLeft.Size() != plan.Size() {
			t.Error("empty.Merge(plan) should have same size as plan")
		}
	})
}

// TestPlanFilterInvariants tests the invariants of Plan.Filter()
func TestPlanFilterInvariants(t *testing.T) {
	t.Run("Filter returns subset of original", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan.Add(NewUpdateChange(CategoryRepo, "b", "old", "new"))
		plan.Add(NewDeleteChange(CategoryLabels, "c", "val"))

		filtered := plan.Filter(func(c Change) bool {
			return c.Category == CategoryLabels
		})

		if filtered.Size() > plan.Size() {
			t.Error("Filter result should be <= original size")
		}
	})

	t.Run("Filter with always-true predicate returns all", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryLabels, "b", "2"),
		})

		filtered := plan.Filter(func(c Change) bool { return true })

		if filtered.Size() != plan.Size() {
			t.Error("Filter(always-true) should return all changes")
		}
	})

	t.Run("Filter with always-false predicate returns empty", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryLabels, "b", "2"),
		})

		filtered := plan.Filter(func(c Change) bool { return false })

		if !filtered.IsEmpty() {
			t.Error("Filter(always-false) should return empty plan")
		}
	})

	t.Run("Filter does not mutate original", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
		})
		originalSize := plan.Size()

		_ = plan.Filter(func(c Change) bool {
			return c.Category == CategoryLabels
		})

		if plan.Size() != originalSize {
			t.Error("Filter should not mutate original plan")
		}
	})

	t.Run("FilterByCategory filters correctly", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "label1", "1"))
		plan.Add(NewAddChange(CategoryRepo, "repo1", "2"))
		plan.Add(NewAddChange(CategoryLabels, "label2", "3"))

		filtered := plan.FilterByCategory(CategoryLabels)

		if filtered.Size() != 2 {
			t.Errorf("FilterByCategory should return 2 labels, got %d", filtered.Size())
		}
		for _, c := range filtered.Changes() {
			if c.Category != CategoryLabels {
				t.Errorf("FilterByCategory returned wrong category: %s", c.Category)
			}
		}
	})

	t.Run("FilterByType filters correctly", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan.Add(NewUpdateChange(CategoryRepo, "b", "old", "new"))
		plan.Add(NewAddChange(CategoryLabels, "c", "3"))

		filtered := plan.FilterByType(ChangeAdd)

		if filtered.Size() != 2 {
			t.Errorf("FilterByType should return 2 adds, got %d", filtered.Size())
		}
		for _, c := range filtered.Changes() {
			if c.Type != ChangeAdd {
				t.Errorf("FilterByType returned wrong type: %v", c.Type)
			}
		}
	})
}

// TestPlanInvertInvariants tests the invariants of Plan.Invert()
func TestPlanInvertInvariants(t *testing.T) {
	t.Run("Invert preserves size", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewUpdateChange(CategoryRepo, "b", "old", "new"),
			NewDeleteChange(CategoryLabels, "c", "val"),
		})

		inverted := plan.Invert()

		if inverted.Size() != plan.Size() {
			t.Error("Invert should preserve size")
		}
	})

	t.Run("Double invert is identity", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewUpdateChange(CategoryRepo, "b", "old", "new"),
		})

		doubleInverted := plan.Invert().Invert()

		originalChanges := plan.Changes()
		invertedChanges := doubleInverted.Changes()

		if len(originalChanges) != len(invertedChanges) {
			t.Fatal("double invert should have same number of changes")
		}

		for i := range originalChanges {
			if originalChanges[i].Type != invertedChanges[i].Type {
				t.Errorf("double invert should preserve type at index %d", i)
			}
			if originalChanges[i].Key != invertedChanges[i].Key {
				t.Errorf("double invert should preserve key at index %d", i)
			}
		}
	})

	t.Run("Invert does not mutate original", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
		})
		originalType := plan.Changes()[0].Type

		_ = plan.Invert()

		if plan.Changes()[0].Type != originalType {
			t.Error("Invert should not mutate original plan")
		}
	})
}

// TestPlanCountInvariants tests counting methods
func TestPlanCountInvariants(t *testing.T) {
	t.Run("CountByType sums to Size", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryLabels, "b", "2"),
			NewUpdateChange(CategoryRepo, "c", "old", "new"),
			NewDeleteChange(CategoryLabels, "d", "val"),
		})

		counts := plan.CountByType()
		total := 0
		for _, count := range counts {
			total += count
		}

		if total != plan.Size() {
			t.Errorf("CountByType sum (%d) should equal Size (%d)", total, plan.Size())
		}
	})

	t.Run("CountByCategory sums to Size", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
			NewAddChange(CategoryLabels, "c", "3"),
		})

		counts := plan.CountByCategory()
		total := 0
		for _, count := range counts {
			total += count
		}

		if total != plan.Size() {
			t.Errorf("CountByCategory sum (%d) should equal Size (%d)", total, plan.Size())
		}
	})

	t.Run("Categories returns unique categories", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
			NewAddChange(CategoryLabels, "c", "3"),
			NewAddChange(CategorySecrets, "d", "4"),
		})

		categories := plan.Categories()
		seen := make(map[ChangeCategory]bool)

		for _, cat := range categories {
			if seen[cat] {
				t.Errorf("Categories returned duplicate: %s", cat)
			}
			seen[cat] = true
		}

		if len(categories) != 3 {
			t.Errorf("expected 3 unique categories, got %d", len(categories))
		}
	})
}

// TestPlanHasMethodsInvariants tests Has* methods
func TestPlanHasMethodsInvariants(t *testing.T) {
	t.Run("HasDeletes is true only when delete exists", func(t *testing.T) {
		planWithDelete := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewDeleteChange(CategoryLabels, "b", "2"),
		})
		planWithoutDelete := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewUpdateChange(CategoryRepo, "b", "old", "new"),
		})

		if !planWithDelete.HasDeletes() {
			t.Error("HasDeletes should be true when delete exists")
		}
		if planWithoutDelete.HasDeletes() {
			t.Error("HasDeletes should be false when no deletes")
		}
	})

	t.Run("HasMissingSecrets requires both category and type", func(t *testing.T) {
		onlyCategory := NewPlanFromChanges([]Change{
			NewAddChange(CategorySecrets, "a", "1"),
		})
		onlyType := NewPlanFromChanges([]Change{
			NewMissingChange(CategoryVariables, "a", "1"),
		})
		both := NewPlanFromChanges([]Change{
			NewMissingChange(CategorySecrets, "a", "1"),
		})

		if onlyCategory.HasMissingSecrets() {
			t.Error("HasMissingSecrets should be false for add to secrets")
		}
		if onlyType.HasMissingSecrets() {
			t.Error("HasMissingSecrets should be false for missing variables")
		}
		if !both.HasMissingSecrets() {
			t.Error("HasMissingSecrets should be true for missing secrets")
		}
	})

	t.Run("HasMissingVariables requires both category and type", func(t *testing.T) {
		onlyCategory := NewPlanFromChanges([]Change{
			NewAddChange(CategoryVariables, "a", "1"),
		})
		onlyType := NewPlanFromChanges([]Change{
			NewMissingChange(CategorySecrets, "a", "1"),
		})
		both := NewPlanFromChanges([]Change{
			NewMissingChange(CategoryVariables, "a", "1"),
		})

		if onlyCategory.HasMissingVariables() {
			t.Error("HasMissingVariables should be false for add to variables")
		}
		if onlyType.HasMissingVariables() {
			t.Error("HasMissingVariables should be false for missing secrets")
		}
		if !both.HasMissingVariables() {
			t.Error("HasMissingVariables should be true for missing variables")
		}
	})
}

// TestPlanCategoriesInvariants tests additional properties of Categories()
func TestPlanCategoriesInvariants(t *testing.T) {
	t.Run("Categories returns no duplicates", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryLabels, "b", "2"),
			NewAddChange(CategoryLabels, "c", "3"),
			NewAddChange(CategoryRepo, "d", "4"),
			NewAddChange(CategoryLabels, "e", "5"),
		})

		categories := plan.Categories()
		seen := make(map[ChangeCategory]bool)

		for _, cat := range categories {
			if seen[cat] {
				t.Errorf("Categories should not contain duplicates, found duplicate: %s", cat)
			}
			seen[cat] = true
		}
	})

	t.Run("Categories only contains categories that exist in changes", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
		})

		categories := plan.Categories()
		changeCategories := make(map[ChangeCategory]bool)
		for _, c := range plan.Changes() {
			changeCategories[c.Category] = true
		}

		for _, cat := range categories {
			if !changeCategories[cat] {
				t.Errorf("Categories returned category %s that doesn't exist in changes", cat)
			}
		}
	})

	t.Run("all change categories appear in Categories()", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
			NewAddChange(CategorySecrets, "c", "3"),
		})

		categories := plan.Categories()
		categorySet := make(map[ChangeCategory]bool)
		for _, cat := range categories {
			categorySet[cat] = true
		}

		for _, c := range plan.Changes() {
			if !categorySet[c.Category] {
				t.Errorf("change category %s not found in Categories()", c.Category)
			}
		}
	})

	t.Run("empty plan returns empty categories", func(t *testing.T) {
		plan := NewPlan()
		categories := plan.Categories()

		if len(categories) != 0 {
			t.Errorf("empty plan should return empty categories, got %d", len(categories))
		}
	})

	t.Run("Categories preserves insertion order", func(t *testing.T) {
		plan := NewPlan()
		plan.Add(NewAddChange(CategoryLabels, "a", "1"))
		plan.Add(NewAddChange(CategoryRepo, "b", "2"))
		plan.Add(NewAddChange(CategorySecrets, "c", "3"))

		categories := plan.Categories()

		// First occurrence order should be: Labels, Repo, Secrets
		if len(categories) != 3 {
			t.Fatalf("expected 3 categories, got %d", len(categories))
		}
		if categories[0] != CategoryLabels {
			t.Errorf("first category should be Labels, got %s", categories[0])
		}
		if categories[1] != CategoryRepo {
			t.Errorf("second category should be Repo, got %s", categories[1])
		}
		if categories[2] != CategorySecrets {
			t.Errorf("third category should be Secrets, got %s", categories[2])
		}
	})

	t.Run("len(Categories) equals len(CountByCategory)", func(t *testing.T) {
		plan := NewPlanFromChanges([]Change{
			NewAddChange(CategoryLabels, "a", "1"),
			NewAddChange(CategoryRepo, "b", "2"),
			NewAddChange(CategoryLabels, "c", "3"),
			NewAddChange(CategorySecrets, "d", "4"),
		})

		categories := plan.Categories()
		counts := plan.CountByCategory()

		if len(categories) != len(counts) {
			t.Errorf("len(Categories)=%d should equal len(CountByCategory)=%d",
				len(categories), len(counts))
		}
	})
}
