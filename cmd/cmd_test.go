package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff"
)

// Test utility functions from init.go

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		slice  []string
		item   string
		expect bool
	}{
		{"found first", []string{"a", "b", "c"}, "a", true},
		{"found middle", []string{"a", "b", "c"}, "b", true},
		{"found last", []string{"a", "b", "c"}, "c", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"empty item", []string{"a", "b", "c"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.expect {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.expect)
			}
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"simple", "a,b,c", []string{"a", "b", "c"}},
		{"with spaces", "a, b , c", []string{"a", "b", "c"}},
		{"with tabs", "a,\tb,\tc", []string{"a", "b", "c"}},
		{"empty parts", "a,,b", []string{"a", "b"}},
		{"single item", "a", []string{"a"}},
		{"empty string", "", []string{}},
		{"only comma", ",", []string{}},
		{"spaces only", "  ,  ", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAndTrim(tt.input)
			if len(got) != len(tt.expect) {
				t.Errorf("splitAndTrim(%q) = %v (len=%d), want %v (len=%d)",
					tt.input, got, len(got), tt.expect, len(tt.expect))
				return
			}
			for i, v := range got {
				if v != tt.expect[i] {
					t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, v, tt.expect[i])
				}
			}
		})
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		sep    string
		expect []string
	}{
		{"comma", "a,b,c", ",", []string{"a", "b", "c"}},
		{"dash", "a-b-c", "-", []string{"a", "b", "c"}},
		{"multi-char sep", "a::b::c", "::", []string{"a", "b", "c"}},
		{"no sep", "abc", ",", []string{"abc"}},
		{"empty string", "", ",", []string{""}},
		{"trailing sep", "a,b,", ",", []string{"a", "b", ""}},
		{"leading sep", ",a,b", ",", []string{"", "a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitString(tt.input, tt.sep)
			if len(got) != len(tt.expect) {
				t.Errorf("splitString(%q, %q) = %v, want %v", tt.input, tt.sep, got, tt.expect)
				return
			}
			for i, v := range got {
				if v != tt.expect[i] {
					t.Errorf("splitString(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, v, tt.expect[i])
				}
			}
		})
	}
}

func TestTrimString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"no trim needed", "abc", "abc"},
		{"leading spaces", "  abc", "abc"},
		{"trailing spaces", "abc  ", "abc"},
		{"both spaces", "  abc  ", "abc"},
		{"tabs", "\tabc\t", "abc"},
		{"mixed", " \t abc \t ", "abc"},
		{"empty", "", ""},
		{"only whitespace", "   ", ""},
		{"inner spaces preserved", "a b c", "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimString(tt.input)
			if got != tt.expect {
				t.Errorf("trimString(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

// Test utility functions from apply.go

func TestFindLabel(t *testing.T) {
	labels := []config.Label{
		{Name: "bug", Color: "d73a4a", Description: "Bug fix"},
		{Name: "feature", Color: "0e8a16", Description: "New feature"},
		{Name: "docs", Color: "0075ca", Description: "Documentation"},
	}

	tests := []struct {
		name       string
		labels     []config.Label
		searchName string
		expectName string
	}{
		{"found first", labels, "bug", "bug"},
		{"found middle", labels, "feature", "feature"},
		{"found last", labels, "docs", "docs"},
		{"not found", labels, "invalid", ""},
		{"empty labels", []config.Label{}, "bug", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findLabel(tt.labels, tt.searchName)
			if got.Name != tt.expectName {
				t.Errorf("findLabel(%q) = %q, want %q", tt.searchName, got.Name, tt.expectName)
			}
		})
	}
}

func TestExtractBranchName(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		expect string
	}{
		{"with dot", "main.required_reviews", "main"},
		{"multiple dots", "release/v1.0.strict", "release/v1"},
		{"no dot", "main", "main"},
		{"empty", "", ""},
		{"starts with dot", ".setting", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBranchName(tt.key)
			if got != tt.expect {
				t.Errorf("extractBranchName(%q) = %q, want %q", tt.key, got, tt.expect)
			}
		})
	}
}

// Test init.go file writing functions

func TestWriteConfigToFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("write single file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "config.yaml")
		cfg := &config.Config{
			Topics: []string{"go", "cli"},
		}

		err := writeConfigToFile(cfg, path)
		if err != nil {
			t.Fatalf("writeConfigToFile() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Verify content
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}
		content := string(data)
		if !containsStr(content, "topics:") {
			t.Errorf("config file does not contain topics: %s", content)
		}
	})

	t.Run("creates parent directory", func(t *testing.T) {
		path := filepath.Join(tmpDir, "nested", "dir", "config.yaml")
		cfg := &config.Config{}

		err := writeConfigToFile(cfg, path)
		if err != nil {
			t.Fatalf("writeConfigToFile() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("nested config file was not created")
		}
	})
}

func TestWriteConfigToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("write directory structure", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "repo-settings")
		visibility := "public"
		cfg := &config.Config{
			Repo: &config.RepoConfig{
				Visibility: &visibility,
			},
			Topics: []string{"go"},
			Labels: &config.LabelsConfig{
				Items: []config.Label{{Name: "bug", Color: "red"}},
			},
			BranchProtection: map[string]*config.BranchRule{
				"main": {},
			},
		}

		err := writeConfigToDirectory(cfg, dir)
		if err != nil {
			t.Fatalf("writeConfigToDirectory() error = %v", err)
		}

		// Verify files exist
		expectedFiles := []string{"repo.yaml", "topics.yaml", "labels.yaml", "branch-protection.yaml"}
		for _, file := range expectedFiles {
			path := filepath.Join(dir, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("expected file %s was not created", file)
			}
		}
	})

	t.Run("skip empty sections", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "minimal")
		cfg := &config.Config{
			Topics: []string{"go"},
		}

		err := writeConfigToDirectory(cfg, dir)
		if err != nil {
			t.Fatalf("writeConfigToDirectory() error = %v", err)
		}

		// Topics should exist
		if _, err := os.Stat(filepath.Join(dir, "topics.yaml")); os.IsNotExist(err) {
			t.Error("topics.yaml should exist")
		}

		// Repo should not exist
		if _, err := os.Stat(filepath.Join(dir, "repo.yaml")); !os.IsNotExist(err) {
			t.Error("repo.yaml should not exist for empty config")
		}
	})
}

// Test export.go utility functions

func TestWriteYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("write map", func(t *testing.T) {
		path := filepath.Join(tmpDir, "map.yaml")
		data := map[string]interface{}{
			"key": "value",
			"nested": map[string]string{
				"a": "b",
			},
		}

		err := writeYAMLFile(path, data)
		if err != nil {
			t.Fatalf("writeYAMLFile() error = %v", err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if !containsStr(string(content), "key: value") {
			t.Errorf("file content = %s, want to contain 'key: value'", content)
		}
	})

	t.Run("write struct", func(t *testing.T) {
		path := filepath.Join(tmpDir, "struct.yaml")
		data := struct {
			Name  string `yaml:"name"`
			Count int    `yaml:"count"`
		}{
			Name:  "test",
			Count: 42,
		}

		err := writeYAMLFile(path, data)
		if err != nil {
			t.Fatalf("writeYAMLFile() error = %v", err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if !containsStr(string(content), "name: test") {
			t.Errorf("file content = %s, want to contain 'name: test'", content)
		}
	})
}

// Test root.go command structure

func TestRootCommand(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		if rootCmd == nil {
			t.Fatal("rootCmd is nil")
		}
	})

	t.Run("has expected flags", func(t *testing.T) {
		vFlag := rootCmd.PersistentFlags().Lookup("verbose")
		if vFlag == nil {
			t.Error("missing --verbose flag")
		}

		qFlag := rootCmd.PersistentFlags().Lookup("quiet")
		if qFlag == nil {
			t.Error("missing --quiet flag")
		}

		rFlag := rootCmd.PersistentFlags().Lookup("repo")
		if rFlag == nil {
			t.Error("missing --repo flag")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		subCmds := rootCmd.Commands()
		cmdNames := make(map[string]bool)
		for _, cmd := range subCmds {
			cmdNames[cmd.Name()] = true
		}

		expectedCmds := []string{"plan", "apply", "init", "export"}
		for _, name := range expectedCmds {
			if !cmdNames[name] {
				t.Errorf("missing subcommand: %s", name)
			}
		}
	})
}

// Test plan.go command structure

func TestPlanCommand(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		if planCmd == nil {
			t.Fatal("planCmd is nil")
		}
	})

	t.Run("has expected flags", func(t *testing.T) {
		dFlag := planCmd.Flags().Lookup("dir")
		if dFlag == nil {
			t.Error("missing --dir flag")
		}

		cFlag := planCmd.Flags().Lookup("config")
		if cFlag == nil {
			t.Error("missing --config flag")
		}

		secretsFlag := planCmd.Flags().Lookup("secrets")
		if secretsFlag == nil {
			t.Error("missing --secrets flag")
		}

		envFlag := planCmd.Flags().Lookup("env")
		if envFlag == nil {
			t.Error("missing --env flag")
		}

		showCurrentFlag := planCmd.Flags().Lookup("show-current")
		if showCurrentFlag == nil {
			t.Error("missing --show-current flag")
		}

		syncFlag := planCmd.Flags().Lookup("sync")
		if syncFlag == nil {
			t.Error("missing --sync flag")
		}

		jsonFlag := planCmd.Flags().Lookup("json")
		if jsonFlag == nil {
			t.Error("missing --json flag")
		}
	})
}

// Test apply.go command structure

func TestApplyCommand(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		if applyCmd == nil {
			t.Fatal("applyCmd is nil")
		}
	})

	t.Run("has expected flags", func(t *testing.T) {
		dFlag := applyCmd.Flags().Lookup("dir")
		if dFlag == nil {
			t.Error("missing --dir flag")
		}

		cFlag := applyCmd.Flags().Lookup("config")
		if cFlag == nil {
			t.Error("missing --config flag")
		}

		yFlag := applyCmd.Flags().Lookup("yes")
		if yFlag == nil {
			t.Error("missing --yes flag")
		}
	})
}

// Test export.go command structure

func TestExportCommand(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		if exportCmd == nil {
			t.Fatal("exportCmd is nil")
		}
	})

	t.Run("has expected flags", func(t *testing.T) {
		dFlag := exportCmd.Flags().Lookup("dir")
		if dFlag == nil {
			t.Error("missing --dir flag")
		}

		sFlag := exportCmd.Flags().Lookup("single")
		if sFlag == nil {
			t.Error("missing --single flag")
		}

		secretsFlag := exportCmd.Flags().Lookup("include-secrets")
		if secretsFlag == nil {
			t.Error("missing --include-secrets flag")
		}
	})
}

// Test init.go command structure

func TestInitCommand(t *testing.T) {
	t.Run("command exists", func(t *testing.T) {
		if initCmd == nil {
			t.Fatal("initCmd is nil")
		}
	})

	t.Run("has expected flags", func(t *testing.T) {
		oFlag := initCmd.Flags().Lookup("output")
		if oFlag == nil {
			t.Error("missing --output flag")
		}

		fromRepoFlag := initCmd.Flags().Lookup("from-repo")
		if fromRepoFlag == nil {
			t.Error("missing --from-repo flag")
		}

		singleFileFlag := initCmd.Flags().Lookup("single-file")
		if singleFileFlag == nil {
			t.Error("missing --single-file flag")
		}

		directoryFlag := initCmd.Flags().Lookup("directory")
		if directoryFlag == nil {
			t.Error("missing --directory flag")
		}
	})
}

func TestInitFromRepoFlagValidation(t *testing.T) {
	// Test that --single-file and --directory are mutually exclusive
	// We can't easily test the runtime behavior, but we can verify the flags exist

	t.Run("flags are defined correctly", func(t *testing.T) {
		fromRepoFlag := initCmd.Flags().Lookup("from-repo")
		if fromRepoFlag.DefValue != "" {
			t.Errorf("--from-repo default should be empty, got %q", fromRepoFlag.DefValue)
		}

		singleFileFlag := initCmd.Flags().Lookup("single-file")
		if singleFileFlag.DefValue != "false" {
			t.Errorf("--single-file default should be false, got %q", singleFileFlag.DefValue)
		}

		directoryFlag := initCmd.Flags().Lookup("directory")
		if directoryFlag.DefValue != "false" {
			t.Errorf("--directory default should be false, got %q", directoryFlag.DefValue)
		}
	})
}

// Test printPlan function with mock plan

func TestPrintPlan(t *testing.T) {
	// Create a plan with various change types
	plan := &diff.Plan{
		Changes: []diff.Change{
			{Category: "repo", Key: "description", Type: diff.ChangeUpdate, Old: "old", New: "new"},
			{Category: "labels", Key: "bug", Type: diff.ChangeAdd, New: "new label"},
			{Category: "labels", Key: "old-label", Type: diff.ChangeDelete, Old: "deleted"},
			{Category: "secrets", Key: "API_KEY", Type: diff.ChangeMissing, New: "required"},
		},
	}

	// printPlan writes to stdout and calls os.Exit on deletes, so we just verify it doesn't panic
	// In real testing, we'd capture stdout and verify the output format
	// For now, this is a smoke test that the function handles various change types

	// Note: We can't easily test printPlan as it calls os.Exit
	// A better design would be to pass a writer and not call os.Exit
	_ = plan
}

// Test printPlan output format
func TestPrintPlanOutput(t *testing.T) {
	tests := []struct {
		name        string
		changes     []diff.Change
		wantDeletes bool
	}{
		{
			name:        "empty plan",
			changes:     []diff.Change{},
			wantDeletes: false,
		},
		{
			name: "repo update",
			changes: []diff.Change{
				{Category: "repo", Key: "description", Type: diff.ChangeUpdate, Old: "old", New: "new"},
			},
			wantDeletes: false,
		},
		{
			name: "label add",
			changes: []diff.Change{
				{Category: "labels", Key: "bug", Type: diff.ChangeAdd, New: "color=red"},
			},
			wantDeletes: false,
		},
		{
			name: "label delete",
			changes: []diff.Change{
				{Category: "labels", Key: "old-label", Type: diff.ChangeDelete, Old: "deleted"},
			},
			wantDeletes: true,
		},
		{
			name: "multiple categories",
			changes: []diff.Change{
				{Category: "repo", Key: "visibility", Type: diff.ChangeUpdate, Old: "private", New: "public"},
				{Category: "topics", Key: "topics", Type: diff.ChangeUpdate, Old: []string{"old"}, New: []string{"new"}},
				{Category: "branch_protection", Key: "main.required_reviews", Type: diff.ChangeUpdate, Old: 1, New: 2},
				{Category: "actions", Key: "enabled", Type: diff.ChangeUpdate, Old: false, New: true},
				{Category: "pages", Key: "build_type", Type: diff.ChangeUpdate, Old: "legacy", New: "workflow"},
				{Category: "variables", Key: "NODE_ENV", Type: diff.ChangeAdd, New: "production"},
				{Category: "secrets", Key: "API_KEY", Type: diff.ChangeAdd, New: "***"},
			},
			wantDeletes: false,
		},
		{
			name: "with delete",
			changes: []diff.Change{
				{Category: "variables", Key: "OLD_VAR", Type: diff.ChangeDelete, Old: "value"},
			},
			wantDeletes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &diff.Plan{Changes: tt.changes}
			hasDeletes := printPlan(plan)
			if hasDeletes != tt.wantDeletes {
				t.Errorf("printPlan() hasDeletes = %v, want %v", hasDeletes, tt.wantDeletes)
			}
		})
	}
}

// Test groupChanges helper logic via applyChanges structure
func TestChangeCategoryGrouping(t *testing.T) {
	changes := []diff.Change{
		{Category: "repo", Key: "description", Type: diff.ChangeUpdate},
		{Category: "repo", Key: "visibility", Type: diff.ChangeUpdate},
		{Category: "topics", Key: "topics", Type: diff.ChangeUpdate},
		{Category: "labels", Key: "bug", Type: diff.ChangeAdd},
		{Category: "labels", Key: "feature", Type: diff.ChangeUpdate},
		{Category: "branch_protection", Key: "main.required_reviews", Type: diff.ChangeUpdate},
		{Category: "branch_protection", Key: "develop.required_reviews", Type: diff.ChangeUpdate},
		{Category: "actions", Key: "enabled", Type: diff.ChangeUpdate},
		{Category: "pages", Key: "build_type", Type: diff.ChangeUpdate},
		{Category: "variables", Key: "NODE_ENV", Type: diff.ChangeAdd},
		{Category: "secrets", Key: "API_KEY", Type: diff.ChangeAdd},
	}

	// Group changes by category (same logic as applyChanges)
	repoChanges := make(map[string]interface{})
	var topicsChanged bool
	var labelChanges []diff.Change
	branchProtectionChanges := make(map[string][]diff.Change)
	var actionsChanges []diff.Change
	var pagesChanges []diff.Change
	var variableChanges []diff.Change
	var secretChanges []diff.Change

	for _, change := range changes {
		switch change.Category {
		case "repo":
			repoChanges[change.Key] = change.New
		case "topics":
			topicsChanged = true
		case "labels":
			labelChanges = append(labelChanges, change)
		case "branch_protection":
			branchName := extractBranchName(change.Key)
			branchProtectionChanges[branchName] = append(branchProtectionChanges[branchName], change)
		case "actions":
			actionsChanges = append(actionsChanges, change)
		case "pages":
			pagesChanges = append(pagesChanges, change)
		case "variables":
			variableChanges = append(variableChanges, change)
		case "secrets":
			secretChanges = append(secretChanges, change)
		}
	}

	// Verify grouping
	if len(repoChanges) != 2 {
		t.Errorf("expected 2 repo changes, got %d", len(repoChanges))
	}
	if !topicsChanged {
		t.Error("expected topicsChanged to be true")
	}
	if len(labelChanges) != 2 {
		t.Errorf("expected 2 label changes, got %d", len(labelChanges))
	}
	if len(branchProtectionChanges) != 2 {
		t.Errorf("expected 2 branch protection branches, got %d", len(branchProtectionChanges))
	}
	if len(actionsChanges) != 1 {
		t.Errorf("expected 1 actions change, got %d", len(actionsChanges))
	}
	if len(pagesChanges) != 1 {
		t.Errorf("expected 1 pages change, got %d", len(pagesChanges))
	}
	if len(variableChanges) != 1 {
		t.Errorf("expected 1 variable change, got %d", len(variableChanges))
	}
	if len(secretChanges) != 1 {
		t.Errorf("expected 1 secret change, got %d", len(secretChanges))
	}
}

// Test extractBranchName with more edge cases
func TestExtractBranchNameEdgeCases(t *testing.T) {
	tests := []struct {
		key    string
		expect string
	}{
		{"main.required_reviews", "main"},
		{"develop.dismiss_stale_reviews", "develop"},
		{"feature/test.strict_status_checks", "feature/test"},
		{"release-1.0.enforce_admins", "release-1"},
		{"no-dot-here", "no-dot-here"},
		{"", ""},
		{".", ""},
		{".only-dot", ""},
		{"a.b.c.d", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := extractBranchName(tt.key)
			if got != tt.expect {
				t.Errorf("extractBranchName(%q) = %q, want %q", tt.key, got, tt.expect)
			}
		})
	}
}

// Helper function
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
