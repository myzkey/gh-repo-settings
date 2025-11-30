package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorkflowFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("job with name", func(t *testing.T) {
		content := `name: Test
jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest
`
		path := filepath.Join(tmpDir, "test.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := parseWorkflowFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 1 || names[0] != "Run tests" {
			t.Errorf("expected [Run tests], got %v", names)
		}
	})

	t.Run("job without name", func(t *testing.T) {
		content := `name: Build
jobs:
  build:
    runs-on: ubuntu-latest
`
		path := filepath.Join(tmpDir, "build.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := parseWorkflowFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 1 || names[0] != "build" {
			t.Errorf("expected [build], got %v", names)
		}
	})

	t.Run("multiple jobs", func(t *testing.T) {
		content := `name: CI
jobs:
  lint:
    name: golangci-lint
    runs-on: ubuntu-latest
  test:
    runs-on: ubuntu-latest
  deploy:
    name: Deploy to production
    runs-on: ubuntu-latest
`
		path := filepath.Join(tmpDir, "ci.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := parseWorkflowFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 3 {
			t.Errorf("expected 3 names, got %d", len(names))
		}

		nameSet := make(map[string]bool)
		for _, n := range names {
			nameSet[n] = true
		}

		expected := []string{"golangci-lint", "test", "Deploy to production"}
		for _, e := range expected {
			if !nameSet[e] {
				t.Errorf("expected %q in names", e)
			}
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		content := `invalid: [yaml`
		path := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := parseWorkflowFile(path)
		if err == nil {
			t.Error("expected error for invalid yaml")
		}
	})
}

func TestGetCheckNames(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("empty directory", func(t *testing.T) {
		names, err := GetCheckNames(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("expected empty, got %v", names)
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		names, err := GetCheckNames(filepath.Join(tmpDir, "nonexistent"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if names != nil {
			t.Errorf("expected nil, got %v", names)
		}
	})

	t.Run("with workflow files", func(t *testing.T) {
		workflowDir := filepath.Join(tmpDir, "workflows")
		if err := os.MkdirAll(workflowDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// Create test.yml
		testContent := `name: Test
jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest
`
		if err := os.WriteFile(filepath.Join(workflowDir, "test.yml"), []byte(testContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create lint.yaml
		lintContent := `name: Lint
jobs:
  lint:
    name: golangci-lint
    runs-on: ubuntu-latest
`
		if err := os.WriteFile(filepath.Join(workflowDir, "lint.yaml"), []byte(lintContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create non-yaml file (should be ignored)
		if err := os.WriteFile(filepath.Join(workflowDir, "README.md"), []byte("# Workflows"), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := GetCheckNames(workflowDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 2 {
			t.Errorf("expected 2 names, got %d: %v", len(names), names)
		}

		nameSet := make(map[string]bool)
		for _, n := range names {
			nameSet[n] = true
		}

		if !nameSet["Run tests"] {
			t.Error("expected 'Run tests' in names")
		}
		if !nameSet["golangci-lint"] {
			t.Error("expected 'golangci-lint' in names")
		}
	})

	t.Run("skips directories", func(t *testing.T) {
		workflowDir := filepath.Join(tmpDir, "workflows2")
		if err := os.MkdirAll(filepath.Join(workflowDir, "subdir"), 0o755); err != nil {
			t.Fatal(err)
		}

		content := `name: Test
jobs:
  test:
    runs-on: ubuntu-latest
`
		if err := os.WriteFile(filepath.Join(workflowDir, "test.yaml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := GetCheckNames(workflowDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 1 {
			t.Errorf("expected 1 name, got %d", len(names))
		}
	})

	t.Run("deduplicates names", func(t *testing.T) {
		workflowDir := filepath.Join(tmpDir, "workflows3")
		if err := os.MkdirAll(workflowDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// Two files with same job name
		content := `name: CI
jobs:
  build:
    runs-on: ubuntu-latest
`
		if err := os.WriteFile(filepath.Join(workflowDir, "ci1.yaml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(workflowDir, "ci2.yaml"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		names, err := GetCheckNames(workflowDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(names) != 1 {
			t.Errorf("expected 1 unique name, got %d: %v", len(names), names)
		}
	})
}

func TestValidateStatusChecks(t *testing.T) {
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `name: CI
jobs:
  lint:
    name: golangci-lint
    runs-on: ubuntu-latest
  test:
    name: Run tests
    runs-on: ubuntu-latest
  build:
    runs-on: ubuntu-latest
`
	if err := os.WriteFile(filepath.Join(workflowDir, "ci.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("all checks valid", func(t *testing.T) {
		checks := []string{"golangci-lint", "Run tests", "build"}
		unknown, available, err := ValidateStatusChecks(checks, workflowDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(unknown) != 0 {
			t.Errorf("expected no unknown checks, got %v", unknown)
		}

		if len(available) != 3 {
			t.Errorf("expected 3 available checks, got %d", len(available))
		}
	})

	t.Run("some checks invalid", func(t *testing.T) {
		checks := []string{"golangci-lint", "lint", "test"}
		unknown, available, err := ValidateStatusChecks(checks, workflowDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(unknown) != 2 {
			t.Errorf("expected 2 unknown checks, got %v", unknown)
		}

		unknownSet := make(map[string]bool)
		for _, u := range unknown {
			unknownSet[u] = true
		}

		if !unknownSet["lint"] || !unknownSet["test"] {
			t.Errorf("expected lint and test to be unknown, got %v", unknown)
		}

		if len(available) != 3 {
			t.Errorf("expected 3 available checks, got %d", len(available))
		}
	})

	t.Run("no workflows directory", func(t *testing.T) {
		checks := []string{"lint", "test"}
		unknown, available, err := ValidateStatusChecks(checks, filepath.Join(tmpDir, "nonexistent"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if unknown != nil || available != nil {
			t.Errorf("expected nil results for non-existent directory")
		}
	})

	t.Run("empty workflow directory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		if err := os.MkdirAll(emptyDir, 0o755); err != nil {
			t.Fatal(err)
		}

		checks := []string{"lint", "test"}
		unknown, available, err := ValidateStatusChecks(checks, emptyDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if unknown != nil || available != nil {
			t.Errorf("expected nil results for empty directory")
		}
	})
}
