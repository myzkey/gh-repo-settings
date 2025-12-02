package diff

import (
	"encoding/json"
	"testing"
)

func TestPlanToJSON(t *testing.T) {
	tests := []struct {
		name     string
		plan     *Plan
		expected *JSONPlan
	}{
		{
			name: "empty plan",
			plan: &Plan{Changes: []Change{}},
			expected: &JSONPlan{
				Summary: JSONSummary{Add: 0, Update: 0, Delete: 0, Missing: 0},
			},
		},
		{
			name: "single repo change",
			plan: &Plan{
				Changes: []Change{
					{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
				},
			},
			expected: &JSONPlan{
				Repo: []JSONChange{
					{Type: "update", Key: "description", Old: "old", New: "new"},
				},
				Summary: JSONSummary{Add: 0, Update: 1, Delete: 0, Missing: 0},
			},
		},
		{
			name: "multiple categories",
			plan: &Plan{
				Changes: []Change{
					{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
					{Type: ChangeAdd, Category: "labels", Key: "bug", New: "color=d73a4a"},
					{Type: ChangeDelete, Category: "labels", Key: "old-label", Old: "color=000000"},
					{Type: ChangeUpdate, Category: "branch_protection", Key: "main.required_reviews", Old: 1, New: 2},
					{Type: ChangeMissing, Category: "secrets", Key: "API_KEY", New: "not in .env"},
				},
			},
			expected: &JSONPlan{
				Repo: []JSONChange{
					{Type: "update", Key: "description", Old: "old", New: "new"},
				},
				Labels: []JSONChange{
					{Type: "add", Key: "bug", New: "color=d73a4a"},
					{Type: "delete", Key: "old-label", Old: "color=000000"},
				},
				BranchProtection: []JSONChange{
					{Type: "update", Key: "main.required_reviews", Old: 1, New: 2},
				},
				Secrets: []JSONChange{
					{Type: "missing", Key: "API_KEY", New: "not in .env"},
				},
				Summary: JSONSummary{Add: 1, Update: 2, Delete: 1, Missing: 1},
			},
		},
		{
			name: "all categories",
			plan: &Plan{
				Changes: []Change{
					{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
					{Type: ChangeUpdate, Category: "topics", Key: "topics", Old: []string{"go"}, New: []string{"go", "cli"}},
					{Type: ChangeAdd, Category: "labels", Key: "feature", New: "color=a2eeef"},
					{Type: ChangeAdd, Category: "branch_protection", Key: "main", New: "{required_reviews=2}"},
					{Type: ChangeUpdate, Category: "actions", Key: "enabled", Old: false, New: true},
					{Type: ChangeAdd, Category: "pages", Key: "pages", New: "build_type=workflow"},
					{Type: ChangeAdd, Category: "variables", Key: "NODE_ENV", New: "production"},
					{Type: ChangeMissing, Category: "secrets", Key: "DEPLOY_KEY", New: "not in .env"},
				},
			},
			expected: &JSONPlan{
				Repo: []JSONChange{
					{Type: "update", Key: "description", Old: "old", New: "new"},
				},
				Topics: []JSONChange{
					{Type: "update", Key: "topics", Old: []string{"go"}, New: []string{"go", "cli"}},
				},
				Labels: []JSONChange{
					{Type: "add", Key: "feature", New: "color=a2eeef"},
				},
				BranchProtection: []JSONChange{
					{Type: "add", Key: "main", New: "{required_reviews=2}"},
				},
				Actions: []JSONChange{
					{Type: "update", Key: "enabled", Old: false, New: true},
				},
				Pages: []JSONChange{
					{Type: "add", Key: "pages", New: "build_type=workflow"},
				},
				Variables: []JSONChange{
					{Type: "add", Key: "NODE_ENV", New: "production"},
				},
				Secrets: []JSONChange{
					{Type: "missing", Key: "DEPLOY_KEY", New: "not in .env"},
				},
				Summary: JSONSummary{Add: 4, Update: 3, Delete: 0, Missing: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.ToJSON()

			// Check summary
			if result.Summary != tt.expected.Summary {
				t.Errorf("summary mismatch: got %+v, want %+v", result.Summary, tt.expected.Summary)
			}

			// Check each category
			checkJSONChanges(t, "repo", result.Repo, tt.expected.Repo)
			checkJSONChanges(t, "topics", result.Topics, tt.expected.Topics)
			checkJSONChanges(t, "labels", result.Labels, tt.expected.Labels)
			checkJSONChanges(t, "branch_protection", result.BranchProtection, tt.expected.BranchProtection)
			checkJSONChanges(t, "actions", result.Actions, tt.expected.Actions)
			checkJSONChanges(t, "pages", result.Pages, tt.expected.Pages)
			checkJSONChanges(t, "variables", result.Variables, tt.expected.Variables)
			checkJSONChanges(t, "secrets", result.Secrets, tt.expected.Secrets)
		})
	}
}

func checkJSONChanges(t *testing.T, category string, got, want []JSONChange) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %d changes, want %d", category, len(got), len(want))
		return
	}
	for i := range got {
		if got[i].Type != want[i].Type {
			t.Errorf("%s[%d].Type: got %q, want %q", category, i, got[i].Type, want[i].Type)
		}
		if got[i].Key != want[i].Key {
			t.Errorf("%s[%d].Key: got %q, want %q", category, i, got[i].Key, want[i].Key)
		}
	}
}

func TestPlanMarshalIndent(t *testing.T) {
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
			{Type: ChangeAdd, Category: "labels", Key: "bug", New: "color=d73a4a"},
		},
	}

	jsonBytes, err := plan.MarshalIndent()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var result JSONPlan
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Check that it's pretty-printed (has newlines and indentation)
	jsonStr := string(jsonBytes)
	if jsonStr[0] != '{' {
		t.Error("expected JSON to start with '{'")
	}
	if len(jsonStr) < 10 {
		t.Error("JSON output seems too short")
	}

	// Verify content
	if result.Summary.Add != 1 || result.Summary.Update != 1 {
		t.Errorf("unexpected summary: %+v", result.Summary)
	}
	if len(result.Repo) != 1 || len(result.Labels) != 1 {
		t.Errorf("unexpected changes count: repo=%d, labels=%d", len(result.Repo), len(result.Labels))
	}
}

func TestJSONPlanOmitempty(t *testing.T) {
	// Test that empty categories are omitted from JSON output
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
		},
	}

	jsonBytes, err := plan.MarshalIndent()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Should contain "repo" but not other empty categories
	if !containsString(jsonStr, `"repo"`) {
		t.Error("expected JSON to contain 'repo'")
	}
	if containsString(jsonStr, `"labels"`) {
		t.Error("expected JSON to not contain 'labels' (empty)")
	}
	if containsString(jsonStr, `"branch_protection"`) {
		t.Error("expected JSON to not contain 'branch_protection' (empty)")
	}
	if containsString(jsonStr, `"secrets"`) {
		t.Error("expected JSON to not contain 'secrets' (empty)")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPlanHasDeletes(t *testing.T) {
	tests := []struct {
		name     string
		changes  []Change
		expected bool
	}{
		{
			name:     "no changes",
			changes:  []Change{},
			expected: false,
		},
		{
			name:     "no deletes",
			changes:  []Change{{Type: ChangeAdd}, {Type: ChangeUpdate}},
			expected: false,
		},
		{
			name:     "has delete",
			changes:  []Change{{Type: ChangeAdd}, {Type: ChangeDelete}},
			expected: true,
		},
		{
			name:     "only delete",
			changes:  []Change{{Type: ChangeDelete}},
			expected: true,
		},
		{
			name:     "missing is not delete",
			changes:  []Change{{Type: ChangeMissing}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Changes: tt.changes}
			if plan.HasDeletes() != tt.expected {
				t.Errorf("expected HasDeletes() = %v", tt.expected)
			}
		})
	}
}

func TestJSONChangeOldNewValues(t *testing.T) {
	// Test that Old and New values are correctly preserved in JSON
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old desc", New: "new desc"},
			{Type: ChangeAdd, Category: "labels", Key: "bug", Old: nil, New: "color=d73a4a"},
			{Type: ChangeDelete, Category: "labels", Key: "wontfix", Old: "color=ffffff", New: nil},
			{Type: ChangeUpdate, Category: "branch_protection", Key: "main.required_reviews", Old: 1, New: 2},
			{Type: ChangeUpdate, Category: "actions", Key: "enabled", Old: false, New: true},
		},
	}

	jsonPlan := plan.ToJSON()

	// Check repo change
	if len(jsonPlan.Repo) != 1 {
		t.Fatalf("expected 1 repo change, got %d", len(jsonPlan.Repo))
	}
	if jsonPlan.Repo[0].Old != "old desc" {
		t.Errorf("repo old value: got %v, want 'old desc'", jsonPlan.Repo[0].Old)
	}
	if jsonPlan.Repo[0].New != "new desc" {
		t.Errorf("repo new value: got %v, want 'new desc'", jsonPlan.Repo[0].New)
	}

	// Check label add (Old should be nil/omitted)
	if len(jsonPlan.Labels) != 2 {
		t.Fatalf("expected 2 label changes, got %d", len(jsonPlan.Labels))
	}
	if jsonPlan.Labels[0].Old != nil {
		t.Errorf("label add should have nil Old, got %v", jsonPlan.Labels[0].Old)
	}

	// Check label delete (New should be nil/omitted)
	if jsonPlan.Labels[1].New != nil {
		t.Errorf("label delete should have nil New, got %v", jsonPlan.Labels[1].New)
	}

	// Check branch protection with int values
	if len(jsonPlan.BranchProtection) != 1 {
		t.Fatalf("expected 1 branch protection change, got %d", len(jsonPlan.BranchProtection))
	}
	if jsonPlan.BranchProtection[0].Old != 1 {
		t.Errorf("branch protection old value: got %v, want 1", jsonPlan.BranchProtection[0].Old)
	}
	if jsonPlan.BranchProtection[0].New != 2 {
		t.Errorf("branch protection new value: got %v, want 2", jsonPlan.BranchProtection[0].New)
	}

	// Check actions with bool values
	if len(jsonPlan.Actions) != 1 {
		t.Fatalf("expected 1 actions change, got %d", len(jsonPlan.Actions))
	}
	if jsonPlan.Actions[0].Old != false {
		t.Errorf("actions old value: got %v, want false", jsonPlan.Actions[0].Old)
	}
	if jsonPlan.Actions[0].New != true {
		t.Errorf("actions new value: got %v, want true", jsonPlan.Actions[0].New)
	}
}

func TestJSONSummaryCountsAllTypes(t *testing.T) {
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeAdd, Category: "labels", Key: "a"},
			{Type: ChangeAdd, Category: "labels", Key: "b"},
			{Type: ChangeAdd, Category: "labels", Key: "c"},
			{Type: ChangeUpdate, Category: "repo", Key: "x"},
			{Type: ChangeUpdate, Category: "repo", Key: "y"},
			{Type: ChangeDelete, Category: "variables", Key: "z"},
			{Type: ChangeMissing, Category: "secrets", Key: "m1"},
			{Type: ChangeMissing, Category: "secrets", Key: "m2"},
		},
	}

	jsonPlan := plan.ToJSON()

	if jsonPlan.Summary.Add != 3 {
		t.Errorf("expected 3 adds, got %d", jsonPlan.Summary.Add)
	}
	if jsonPlan.Summary.Update != 2 {
		t.Errorf("expected 2 updates, got %d", jsonPlan.Summary.Update)
	}
	if jsonPlan.Summary.Delete != 1 {
		t.Errorf("expected 1 delete, got %d", jsonPlan.Summary.Delete)
	}
	if jsonPlan.Summary.Missing != 2 {
		t.Errorf("expected 2 missing, got %d", jsonPlan.Summary.Missing)
	}
}

func TestJSONMarshalIndentFormat(t *testing.T) {
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
		},
	}

	jsonBytes, err := plan.MarshalIndent()
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Should be pretty-printed with indentation
	if !containsString(jsonStr, "\n") {
		t.Error("expected JSON to contain newlines (pretty-printed)")
	}
	if !containsString(jsonStr, "  ") {
		t.Error("expected JSON to contain indentation")
	}

	// Should contain expected fields
	if !containsString(jsonStr, `"repo"`) {
		t.Error("expected JSON to contain 'repo' field")
	}
	if !containsString(jsonStr, `"summary"`) {
		t.Error("expected JSON to contain 'summary' field")
	}
	if !containsString(jsonStr, `"type"`) {
		t.Error("expected JSON to contain 'type' field")
	}
	if !containsString(jsonStr, `"update"`) {
		t.Error("expected JSON to contain 'update' value")
	}
}

func TestJSONUnknownCategoryIgnored(t *testing.T) {
	// Test that unknown categories don't cause panic
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeAdd, Category: "unknown_category", Key: "test"},
			{Type: ChangeUpdate, Category: "repo", Key: "description", Old: "old", New: "new"},
		},
	}

	jsonPlan := plan.ToJSON()

	// Unknown category should not appear in any field
	if len(jsonPlan.Repo) != 1 {
		t.Errorf("expected 1 repo change, got %d", len(jsonPlan.Repo))
	}

	// Summary should still count it as an add
	if jsonPlan.Summary.Add != 1 {
		t.Errorf("expected 1 add (unknown category still counted), got %d", jsonPlan.Summary.Add)
	}
}
