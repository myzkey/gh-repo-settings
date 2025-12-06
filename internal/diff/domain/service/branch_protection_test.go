package service

import (
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
)

// Helper functions for creating pointers
func intPtr(v int) *int       { return &v }
func boolPtr(v bool) *bool    { return &v }

// TestCompareBranchRuleInvariants tests the invariants of CompareBranchRule
func TestCompareBranchRuleInvariants(t *testing.T) {
	t.Run("identical states produce no changes", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			RequiredReviews:      2,
			DismissStaleReviews:  true,
			RequireCodeOwner:     true,
			StrictStatusChecks:   true,
			StatusChecks:         []string{"ci", "lint"},
			EnforceAdmins:        true,
			RequireLinearHistory: false,
			AllowForcePushes:     false,
			AllowDeletions:       false,
			RequireSignedCommits: true,
		}
		desired := model.BranchProtectionDesired{
			RequiredReviews:      intPtr(2),
			DismissStaleReviews:  boolPtr(true),
			RequireCodeOwner:     boolPtr(true),
			StrictStatusChecks:   boolPtr(true),
			StatusChecks:         []string{"ci", "lint"},
			EnforceAdmins:        boolPtr(true),
			RequireLinearHistory: boolPtr(false),
			AllowForcePushes:     boolPtr(false),
			AllowDeletions:       boolPtr(false),
			RequireSignedCommits: boolPtr(true),
		}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 0 {
			t.Errorf("identical states should produce no changes, got %d", len(changes))
		}
	})

	t.Run("nil desired fields produce no changes", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			RequiredReviews:     2,
			DismissStaleReviews: true,
		}
		desired := model.BranchProtectionDesired{
			// All nil - should not compare
		}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 0 {
			t.Errorf("nil desired fields should produce no changes, got %d", len(changes))
		}
	})

	t.Run("all changes have correct category", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			RequiredReviews: 1,
		}
		desired := model.BranchProtectionDesired{
			RequiredReviews:     intPtr(2),
			DismissStaleReviews: boolPtr(true),
		}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Category != model.CategoryBranchProtection {
				t.Errorf("change category should be branch_protection, got %s", c.Category)
			}
		}
	})

	t.Run("all changes have branch prefix in key", func(t *testing.T) {
		current := model.BranchProtectionCurrent{}
		desired := model.BranchProtectionDesired{
			RequiredReviews: intPtr(2),
		}

		branchName := "feature/test"
		changes := CompareBranchRule(branchName, current, desired)

		for _, c := range changes {
			if len(c.Key) < len(branchName)+1 {
				t.Errorf("key should have branch prefix: %s", c.Key)
			}
			if c.Key[:len(branchName)+1] != branchName+"." {
				t.Errorf("key should start with '%s.', got %s", branchName, c.Key)
			}
		}
	})

	t.Run("all changes are update type", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			RequiredReviews:     1,
			DismissStaleReviews: false,
		}
		desired := model.BranchProtectionDesired{
			RequiredReviews:     intPtr(2),
			DismissStaleReviews: boolPtr(true),
		}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Type != model.ChangeUpdate {
				t.Errorf("branch protection changes should be updates, got %v", c.Type)
			}
		}
	})
}

// TestCompareBranchRuleIntFields tests integer field comparison
func TestCompareBranchRuleIntFields(t *testing.T) {
	t.Run("required_reviews change detected", func(t *testing.T) {
		current := model.BranchProtectionCurrent{RequiredReviews: 1}
		desired := model.BranchProtectionDesired{RequiredReviews: intPtr(2)}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d", len(changes))
		}
		if changes[0].Key != "main.required_reviews" {
			t.Errorf("expected key 'main.required_reviews', got %s", changes[0].Key)
		}
		if changes[0].Old != 1 {
			t.Errorf("expected Old = 1, got %v", changes[0].Old)
		}
		if changes[0].New != 2 {
			t.Errorf("expected New = 2, got %v", changes[0].New)
		}
	})

	t.Run("required_reviews 0 to n is detected", func(t *testing.T) {
		current := model.BranchProtectionCurrent{RequiredReviews: 0}
		desired := model.BranchProtectionDesired{RequiredReviews: intPtr(1)}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 1 {
			t.Errorf("0 to 1 change should be detected")
		}
	})

	t.Run("required_reviews n to 0 is detected", func(t *testing.T) {
		current := model.BranchProtectionCurrent{RequiredReviews: 2}
		desired := model.BranchProtectionDesired{RequiredReviews: intPtr(0)}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 1 {
			t.Errorf("n to 0 change should be detected")
		}
	})
}

// TestCompareBranchRuleBoolFields tests boolean field comparison
func TestCompareBranchRuleBoolFields(t *testing.T) {
	boolFields := []struct {
		name    string
		keyName string
		setCurrent func(*model.BranchProtectionCurrent, bool)
		setDesired func(*model.BranchProtectionDesired, *bool)
	}{
		{
			name:    "dismiss_stale_reviews",
			keyName: "dismiss_stale_reviews",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.DismissStaleReviews = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.DismissStaleReviews = v },
		},
		{
			name:    "require_code_owner",
			keyName: "require_code_owner",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.RequireCodeOwner = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.RequireCodeOwner = v },
		},
		{
			name:    "strict_status_checks",
			keyName: "strict_status_checks",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.StrictStatusChecks = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.StrictStatusChecks = v },
		},
		{
			name:    "enforce_admins",
			keyName: "enforce_admins",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.EnforceAdmins = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.EnforceAdmins = v },
		},
		{
			name:    "require_linear_history",
			keyName: "require_linear_history",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.RequireLinearHistory = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.RequireLinearHistory = v },
		},
		{
			name:    "allow_force_pushes",
			keyName: "allow_force_pushes",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.AllowForcePushes = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.AllowForcePushes = v },
		},
		{
			name:    "allow_deletions",
			keyName: "allow_deletions",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.AllowDeletions = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.AllowDeletions = v },
		},
		{
			name:    "require_signed_commits",
			keyName: "require_signed_commits",
			setCurrent: func(c *model.BranchProtectionCurrent, v bool) { c.RequireSignedCommits = v },
			setDesired: func(d *model.BranchProtectionDesired, v *bool) { d.RequireSignedCommits = v },
		},
	}

	for _, field := range boolFields {
		t.Run(field.name+" false to true detected", func(t *testing.T) {
			current := model.BranchProtectionCurrent{}
			desired := model.BranchProtectionDesired{}
			field.setCurrent(&current, false)
			field.setDesired(&desired, boolPtr(true))

			changes := CompareBranchRule("main", current, desired)

			if len(changes) != 1 {
				t.Fatalf("expected 1 change for %s, got %d", field.name, len(changes))
			}
			if changes[0].Key != "main."+field.keyName {
				t.Errorf("expected key 'main.%s', got %s", field.keyName, changes[0].Key)
			}
			if changes[0].Old != false {
				t.Errorf("expected Old = false")
			}
			if changes[0].New != true {
				t.Errorf("expected New = true")
			}
		})

		t.Run(field.name+" true to false detected", func(t *testing.T) {
			current := model.BranchProtectionCurrent{}
			desired := model.BranchProtectionDesired{}
			field.setCurrent(&current, true)
			field.setDesired(&desired, boolPtr(false))

			changes := CompareBranchRule("main", current, desired)

			if len(changes) != 1 {
				t.Fatalf("expected 1 change for %s, got %d", field.name, len(changes))
			}
		})

		t.Run(field.name+" nil desired produces no change", func(t *testing.T) {
			current := model.BranchProtectionCurrent{}
			desired := model.BranchProtectionDesired{}
			field.setCurrent(&current, true)
			// desired field remains nil

			changes := CompareBranchRule("main", current, desired)

			if len(changes) != 0 {
				t.Errorf("nil desired should produce no change for %s", field.name)
			}
		})

		t.Run(field.name+" same value produces no change", func(t *testing.T) {
			current := model.BranchProtectionCurrent{}
			desired := model.BranchProtectionDesired{}
			field.setCurrent(&current, true)
			field.setDesired(&desired, boolPtr(true))

			changes := CompareBranchRule("main", current, desired)

			if len(changes) != 0 {
				t.Errorf("same value should produce no change for %s", field.name)
			}
		})
	}
}

// TestCompareBranchRuleStatusChecks tests status checks slice comparison
func TestCompareBranchRuleStatusChecks(t *testing.T) {
	t.Run("status_checks change detected", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			StatusChecks: []string{"ci"},
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: []string{"ci", "lint"},
		}

		changes := CompareBranchRule("main", current, desired)

		found := false
		for _, c := range changes {
			if c.Key == "main.status_checks" {
				found = true
			}
		}
		if !found {
			t.Error("status_checks change should be detected")
		}
	})

	t.Run("status_checks nil desired produces no change", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			StatusChecks: []string{"ci"},
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: nil,
		}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("nil desired status_checks should produce no change")
			}
		}
	})

	t.Run("status_checks order matters", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			StatusChecks: []string{"a", "b"},
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: []string{"b", "a"},
		}

		changes := CompareBranchRule("main", current, desired)

		// Order difference should be detected
		found := false
		for _, c := range changes {
			if c.Key == "main.status_checks" {
				found = true
			}
		}
		if !found {
			t.Error("status_checks order change should be detected")
		}
	})

	t.Run("status_checks empty to non-empty detected", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			StatusChecks: []string{},
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: []string{"ci"},
		}

		changes := CompareBranchRule("main", current, desired)

		found := false
		for _, c := range changes {
			if c.Key == "main.status_checks" {
				found = true
			}
		}
		if !found {
			t.Error("empty to non-empty status_checks should be detected")
		}
	})

	t.Run("status_checks empty slice equals nil current", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			StatusChecks: nil,
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: []string{},
		}

		changes := CompareBranchRule("main", current, desired)

		// nil and empty slice should be considered equal
		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("nil and empty status_checks should be equal")
			}
		}
	})
}

// TestStatusChecksNilVsEmptySpec documents the intended behavior for nil vs empty slice
// This is a specification test - changing this behavior is a breaking change
func TestStatusChecksNilVsEmptySpec(t *testing.T) {
	t.Run("SPEC: nil desired means 'do not manage this field'", func(t *testing.T) {
		// When desired.StatusChecks is nil, we don't want to change it
		// regardless of what current value is
		testCases := []struct {
			name    string
			current []string
		}{
			{"current is nil", nil},
			{"current is empty", []string{}},
			{"current has values", []string{"ci", "lint"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				current := model.BranchProtectionCurrent{StatusChecks: tc.current}
				desired := model.BranchProtectionDesired{StatusChecks: nil}

				changes := CompareBranchRule("main", current, desired)

				for _, c := range changes {
					if c.Key == "main.status_checks" {
						t.Errorf("nil desired should not produce change, current was: %v", tc.current)
					}
				}
			})
		}
	})

	t.Run("SPEC: empty slice desired means 'remove all status checks'", func(t *testing.T) {
		// When desired.StatusChecks is []string{}, we want to clear status checks
		// (but nil current is considered equal to empty)
		current := model.BranchProtectionCurrent{
			StatusChecks: []string{"ci", "lint"},
		}
		desired := model.BranchProtectionDesired{
			StatusChecks: []string{},
		}

		changes := CompareBranchRule("main", current, desired)

		found := false
		for _, c := range changes {
			if c.Key == "main.status_checks" {
				found = true
				// Verify old and new values
				if old, ok := c.Old.([]string); ok {
					if len(old) != 2 {
						t.Errorf("old value should be [ci, lint]")
					}
				}
				if newVal, ok := c.New.([]string); ok {
					if len(newVal) != 0 {
						t.Errorf("new value should be empty slice")
					}
				}
			}
		}
		if !found {
			t.Error("empty desired should clear existing status checks")
		}
	})

	t.Run("SPEC: nil current and empty desired are equivalent (no change)", func(t *testing.T) {
		// This is the key specification: nil and [] are semantically equivalent
		current := model.BranchProtectionCurrent{StatusChecks: nil}
		desired := model.BranchProtectionDesired{StatusChecks: []string{}}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("nil current and empty desired should be equal - no change expected")
			}
		}
	})

	t.Run("SPEC: empty current and nil desired produces no change", func(t *testing.T) {
		current := model.BranchProtectionCurrent{StatusChecks: []string{}}
		desired := model.BranchProtectionDesired{StatusChecks: nil}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("empty current with nil desired should produce no change")
			}
		}
	})

	t.Run("SPEC: both nil produces no change", func(t *testing.T) {
		current := model.BranchProtectionCurrent{StatusChecks: nil}
		desired := model.BranchProtectionDesired{StatusChecks: nil}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("both nil should produce no change")
			}
		}
	})

	t.Run("SPEC: both empty produces no change", func(t *testing.T) {
		current := model.BranchProtectionCurrent{StatusChecks: []string{}}
		desired := model.BranchProtectionDesired{StatusChecks: []string{}}

		changes := CompareBranchRule("main", current, desired)

		for _, c := range changes {
			if c.Key == "main.status_checks" {
				t.Error("both empty should produce no change")
			}
		}
	})
}

// TestCompareBranchRuleMultipleChanges tests multiple simultaneous changes
func TestCompareBranchRuleMultipleChanges(t *testing.T) {
	t.Run("multiple changes detected independently", func(t *testing.T) {
		current := model.BranchProtectionCurrent{
			RequiredReviews:     1,
			DismissStaleReviews: false,
			EnforceAdmins:       false,
		}
		desired := model.BranchProtectionDesired{
			RequiredReviews:     intPtr(2),
			DismissStaleReviews: boolPtr(true),
			EnforceAdmins:       boolPtr(true),
		}

		changes := CompareBranchRule("main", current, desired)

		if len(changes) != 3 {
			t.Errorf("expected 3 changes, got %d", len(changes))
		}

		keys := make(map[string]bool)
		for _, c := range changes {
			keys[c.Key] = true
		}

		expectedKeys := []string{
			"main.required_reviews",
			"main.dismiss_stale_reviews",
			"main.enforce_admins",
		}
		for _, k := range expectedKeys {
			if !keys[k] {
				t.Errorf("expected change for key %s", k)
			}
		}
	})
}

// TestCompareBranchRuleBranchNames tests various branch name formats
func TestCompareBranchRuleBranchNames(t *testing.T) {
	branchNames := []string{
		"main",
		"develop",
		"feature/my-feature",
		"release/v1.0.0",
		"hotfix/critical-bug",
		"refs/heads/main",
	}

	for _, branchName := range branchNames {
		t.Run("branch name: "+branchName, func(t *testing.T) {
			current := model.BranchProtectionCurrent{RequiredReviews: 1}
			desired := model.BranchProtectionDesired{RequiredReviews: intPtr(2)}

			changes := CompareBranchRule(branchName, current, desired)

			if len(changes) != 1 {
				t.Fatal("expected 1 change")
			}

			expectedKey := branchName + ".required_reviews"
			if changes[0].Key != expectedKey {
				t.Errorf("expected key '%s', got '%s'", expectedKey, changes[0].Key)
			}
		})
	}
}
