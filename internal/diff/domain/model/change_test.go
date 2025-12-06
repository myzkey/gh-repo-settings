package model

import (
	"strings"
	"testing"
)

// TestChangeInvariants tests the invariants of the Change type
func TestChangeInvariants(t *testing.T) {
	t.Run("NewAddChange creates change with ChangeAdd type", func(t *testing.T) {
		change := NewAddChange(CategoryLabels, "bug", "red")

		if change.Type != ChangeAdd {
			t.Errorf("expected Type = ChangeAdd, got %v", change.Type)
		}
		if change.Old != nil {
			t.Errorf("expected Old = nil for add change, got %v", change.Old)
		}
		if change.New != "red" {
			t.Errorf("expected New = 'red', got %v", change.New)
		}
	})

	t.Run("NewUpdateChange creates change with ChangeUpdate type", func(t *testing.T) {
		change := NewUpdateChange(CategoryRepo, "description", "old", "new")

		if change.Type != ChangeUpdate {
			t.Errorf("expected Type = ChangeUpdate, got %v", change.Type)
		}
		if change.Old != "old" {
			t.Errorf("expected Old = 'old', got %v", change.Old)
		}
		if change.New != "new" {
			t.Errorf("expected New = 'new', got %v", change.New)
		}
	})

	t.Run("NewDeleteChange creates change with ChangeDelete type", func(t *testing.T) {
		change := NewDeleteChange(CategoryLabels, "obsolete", "value")

		if change.Type != ChangeDelete {
			t.Errorf("expected Type = ChangeDelete, got %v", change.Type)
		}
		if change.Old != "value" {
			t.Errorf("expected Old = 'value', got %v", change.Old)
		}
		if change.New != nil {
			t.Errorf("expected New = nil for delete change, got %v", change.New)
		}
	})

	t.Run("NewMissingChange creates change with ChangeMissing type", func(t *testing.T) {
		change := NewMissingChange(CategorySecrets, "API_KEY", "required")

		if change.Type != ChangeMissing {
			t.Errorf("expected Type = ChangeMissing, got %v", change.Type)
		}
		if change.Old != nil {
			t.Errorf("expected Old = nil for missing change, got %v", change.Old)
		}
	})
}

// TestChangeInvertInvariants tests the invariants of Change.Invert()
func TestChangeInvertInvariants(t *testing.T) {
	t.Run("Invert(Add) produces Delete", func(t *testing.T) {
		add := NewAddChange(CategoryLabels, "bug", "red")
		inverted := add.Invert()

		if inverted.Type != ChangeDelete {
			t.Errorf("expected inverted Type = ChangeDelete, got %v", inverted.Type)
		}
		if inverted.Old != "red" {
			t.Errorf("expected inverted Old = 'red', got %v", inverted.Old)
		}
		if inverted.New != nil {
			t.Errorf("expected inverted New = nil, got %v", inverted.New)
		}
	})

	t.Run("Invert(Delete) produces Add", func(t *testing.T) {
		del := NewDeleteChange(CategoryLabels, "bug", "red")
		inverted := del.Invert()

		if inverted.Type != ChangeAdd {
			t.Errorf("expected inverted Type = ChangeAdd, got %v", inverted.Type)
		}
		if inverted.Old != nil {
			t.Errorf("expected inverted Old = nil, got %v", inverted.Old)
		}
		if inverted.New != "red" {
			t.Errorf("expected inverted New = 'red', got %v", inverted.New)
		}
	})

	t.Run("Invert(Update) swaps Old and New", func(t *testing.T) {
		update := NewUpdateChange(CategoryRepo, "desc", "old", "new")
		inverted := update.Invert()

		if inverted.Type != ChangeUpdate {
			t.Errorf("expected inverted Type = ChangeUpdate, got %v", inverted.Type)
		}
		if inverted.Old != "new" {
			t.Errorf("expected inverted Old = 'new', got %v", inverted.Old)
		}
		if inverted.New != "old" {
			t.Errorf("expected inverted New = 'old', got %v", inverted.New)
		}
	})

	t.Run("Invert is idempotent for Add/Delete pair", func(t *testing.T) {
		original := NewAddChange(CategoryLabels, "bug", "red")
		doubleInverted := original.Invert().Invert()

		if doubleInverted.Type != original.Type {
			t.Errorf("double invert should preserve Type")
		}
		if doubleInverted.New != original.New {
			t.Errorf("double invert should preserve New value")
		}
	})

	t.Run("Invert is idempotent for Update", func(t *testing.T) {
		original := NewUpdateChange(CategoryRepo, "desc", "old", "new")
		doubleInverted := original.Invert().Invert()

		if doubleInverted.Old != original.Old {
			t.Errorf("double invert should preserve Old value")
		}
		if doubleInverted.New != original.New {
			t.Errorf("double invert should preserve New value")
		}
	})

	t.Run("Invert preserves Category and Key", func(t *testing.T) {
		original := NewAddChange(CategoryBranchProtection, "main.required_reviews", 2)
		inverted := original.Invert()

		if inverted.Category != original.Category {
			t.Errorf("Invert should preserve Category")
		}
		if inverted.Key != original.Key {
			t.Errorf("Invert should preserve Key")
		}
	})
}

// TestChangeTypePredicates tests the type predicate methods
func TestChangeTypePredicates(t *testing.T) {
	tests := []struct {
		name     string
		change   Change
		isAdd    bool
		isUpdate bool
		isDelete bool
		isMissing bool
	}{
		{
			name:     "Add change",
			change:   NewAddChange(CategoryLabels, "bug", "red"),
			isAdd:    true,
			isUpdate: false,
			isDelete: false,
			isMissing: false,
		},
		{
			name:     "Update change",
			change:   NewUpdateChange(CategoryRepo, "desc", "old", "new"),
			isAdd:    false,
			isUpdate: true,
			isDelete: false,
			isMissing: false,
		},
		{
			name:     "Delete change",
			change:   NewDeleteChange(CategoryLabels, "old", "value"),
			isAdd:    false,
			isUpdate: false,
			isDelete: true,
			isMissing: false,
		},
		{
			name:     "Missing change",
			change:   NewMissingChange(CategorySecrets, "KEY", "required"),
			isAdd:    false,
			isUpdate: false,
			isDelete: false,
			isMissing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.change.IsAdd() != tt.isAdd {
				t.Errorf("IsAdd() = %v, want %v", tt.change.IsAdd(), tt.isAdd)
			}
			if tt.change.IsUpdate() != tt.isUpdate {
				t.Errorf("IsUpdate() = %v, want %v", tt.change.IsUpdate(), tt.isUpdate)
			}
			if tt.change.IsDelete() != tt.isDelete {
				t.Errorf("IsDelete() = %v, want %v", tt.change.IsDelete(), tt.isDelete)
			}
			if tt.change.IsMissing() != tt.isMissing {
				t.Errorf("IsMissing() = %v, want %v", tt.change.IsMissing(), tt.isMissing)
			}
		})
	}

	t.Run("exactly one predicate is true", func(t *testing.T) {
		changes := []Change{
			NewAddChange(CategoryLabels, "k", "v"),
			NewUpdateChange(CategoryRepo, "k", "o", "n"),
			NewDeleteChange(CategoryLabels, "k", "v"),
			NewMissingChange(CategorySecrets, "k", "d"),
		}

		for _, c := range changes {
			count := 0
			if c.IsAdd() {
				count++
			}
			if c.IsUpdate() {
				count++
			}
			if c.IsDelete() {
				count++
			}
			if c.IsMissing() {
				count++
			}
			if count != 1 {
				t.Errorf("exactly one predicate should be true, got %d for type %v", count, c.Type)
			}
		}
	})
}

// TestChangeWithMethods tests the With* builder methods preserve other fields
func TestChangeWithMethods(t *testing.T) {
	t.Run("WithCategory preserves other fields", func(t *testing.T) {
		original := NewUpdateChange(CategoryRepo, "key", "old", "new")
		modified := original.WithCategory(CategoryLabels)

		if modified.Category != CategoryLabels {
			t.Errorf("WithCategory should change category")
		}
		if modified.Type != original.Type {
			t.Errorf("WithCategory should preserve Type")
		}
		if modified.Key != original.Key {
			t.Errorf("WithCategory should preserve Key")
		}
		if modified.Old != original.Old {
			t.Errorf("WithCategory should preserve Old")
		}
		if modified.New != original.New {
			t.Errorf("WithCategory should preserve New")
		}
	})

	t.Run("WithKeyPrefix preserves other fields", func(t *testing.T) {
		original := NewUpdateChange(CategoryBranchProtection, "required_reviews", 1, 2)
		modified := original.WithKeyPrefix("main.")

		if modified.Key != "main.required_reviews" {
			t.Errorf("WithKeyPrefix should prepend prefix, got %s", modified.Key)
		}
		if modified.Category != original.Category {
			t.Errorf("WithKeyPrefix should preserve Category")
		}
		if modified.Type != original.Type {
			t.Errorf("WithKeyPrefix should preserve Type")
		}
	})

	t.Run("With methods do not mutate original", func(t *testing.T) {
		original := NewAddChange(CategoryRepo, "key", "value")
		originalCategory := original.Category
		originalKey := original.Key

		_ = original.WithCategory(CategoryLabels)
		_ = original.WithKeyPrefix("prefix.")

		if original.Category != originalCategory {
			t.Errorf("WithCategory mutated original")
		}
		if original.Key != originalKey {
			t.Errorf("WithKeyPrefix mutated original")
		}
	})
}

// TestChangeTypeString tests ChangeType.String() returns valid strings
func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeAdd, "add"},
		{ChangeUpdate, "update"},
		{ChangeDelete, "delete"},
		{ChangeMissing, "missing"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.changeType.String() != tt.expected {
				t.Errorf("ChangeType.String() = %s, want %s", tt.changeType.String(), tt.expected)
			}
		})
	}

	t.Run("unknown type returns 'unknown'", func(t *testing.T) {
		unknown := ChangeType(999)
		if unknown.String() != "unknown" {
			t.Errorf("unknown ChangeType.String() = %s, want 'unknown'", unknown.String())
		}
	})
}

// TestChangeStringContainsEssentialInfo tests Change.String() contains essential information
// These are loose assertions to avoid brittleness from formatting changes
func TestChangeStringContainsEssentialInfo(t *testing.T) {
	t.Run("Add change string contains ADD, category, and key", func(t *testing.T) {
		change := NewAddChange(CategoryLabels, "bug", "red")
		str := change.String()

		if !strings.Contains(strings.ToUpper(str), "ADD") {
			t.Error("Add change string should contain 'ADD'")
		}
		if !strings.Contains(str, "labels") {
			t.Error("Add change string should contain category")
		}
		if !strings.Contains(str, "bug") {
			t.Error("Add change string should contain key")
		}
	})

	t.Run("Update change string contains UPDATE, category, key, old and new", func(t *testing.T) {
		change := NewUpdateChange(CategoryRepo, "description", "old-desc", "new-desc")
		str := change.String()

		if !strings.Contains(strings.ToUpper(str), "UPDATE") {
			t.Error("Update change string should contain 'UPDATE'")
		}
		if !strings.Contains(str, "repo") {
			t.Error("Update change string should contain category")
		}
		if !strings.Contains(str, "description") {
			t.Error("Update change string should contain key")
		}
		if !strings.Contains(str, "old-desc") {
			t.Error("Update change string should contain old value")
		}
		if !strings.Contains(str, "new-desc") {
			t.Error("Update change string should contain new value")
		}
	})

	t.Run("Delete change string contains DELETE, category, and key", func(t *testing.T) {
		change := NewDeleteChange(CategoryLabels, "obsolete", "value")
		str := change.String()

		if !strings.Contains(strings.ToUpper(str), "DELETE") {
			t.Error("Delete change string should contain 'DELETE'")
		}
		if !strings.Contains(str, "labels") {
			t.Error("Delete change string should contain category")
		}
		if !strings.Contains(str, "obsolete") {
			t.Error("Delete change string should contain key")
		}
	})

	t.Run("Missing change string contains MISSING, category, and key", func(t *testing.T) {
		change := NewMissingChange(CategorySecrets, "API_KEY", "required")
		str := change.String()

		if !strings.Contains(strings.ToUpper(str), "MISSING") {
			t.Error("Missing change string should contain 'MISSING'")
		}
		if !strings.Contains(str, "secrets") {
			t.Error("Missing change string should contain category")
		}
		if !strings.Contains(str, "API_KEY") {
			t.Error("Missing change string should contain key")
		}
	})

	t.Run("String is never empty", func(t *testing.T) {
		changes := []Change{
			NewAddChange(CategoryLabels, "k", "v"),
			NewUpdateChange(CategoryRepo, "k", "o", "n"),
			NewDeleteChange(CategoryLabels, "k", "v"),
			NewMissingChange(CategorySecrets, "k", "d"),
		}

		for _, c := range changes {
			if c.String() == "" {
				t.Errorf("Change.String() should never be empty for type %v", c.Type)
			}
		}
	})
}
