package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/config.yaml", true},
		{"http://example.com/config.yaml", true},
		{"./base.yaml", false},
		{"../base.yaml", false},
		{"/absolute/path/config.yaml", false},
		{"base.yaml", false},
		{"ftp://example.com/config.yaml", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isURL(tt.input)
			if result != tt.expected {
				t.Errorf("isURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeRef(t *testing.T) {
	tests := []struct {
		ref      string
		basePath string
		expected string
	}{
		{"https://example.com/config.yaml", "/some/path", "https://example.com/config.yaml"},
		{"http://example.com/config.yaml", "/some/path", "http://example.com/config.yaml"},
		{"./base.yaml", "/some/path", "/some/path/base.yaml"},
		{"../base.yaml", "/some/path", "/some/base.yaml"},
		{"/absolute/config.yaml", "/some/path", "/absolute/config.yaml"},
		{"base.yaml", "/some/path", "/some/path/base.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			result := normalizeRef(tt.ref, tt.basePath)
			if result != tt.expected {
				t.Errorf("normalizeRef(%q, %q) = %q, want %q", tt.ref, tt.basePath, result, tt.expected)
			}
		})
	}
}

func TestLoadFromURL(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/yaml")
			w.Write([]byte(`
repo:
  visibility: public
  allow_merge_commit: false
`))
		}))
		defer server.Close()

		cfg, err := loadFromURL(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Repo == nil || cfg.Repo.Visibility == nil || *cfg.Repo.Visibility != "public" {
			t.Error("expected visibility 'public'")
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := loadFromURL(server.URL)
		if err == nil {
			t.Error("expected error for 404 response")
		}
		if !strings.Contains(err.Error(), "status 404") {
			t.Errorf("expected status 404 in error, got: %v", err)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/yaml")
			w.Write([]byte("invalid: yaml: content:"))
		}))
		defer server.Close()

		_, err := loadFromURL(server.URL)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/yaml")
		}))
		defer server.Close()

		cfg, err := loadFromURL(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Error("expected non-nil config")
		}
	})
}

func TestResolveExtendsLocalFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-local-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create base config
	baseContent := `
repo:
  visibility: public
  allow_merge_commit: false
branch_protection:
  main:
    required_reviews: 2
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	if err := os.WriteFile(basePath, []byte(baseContent), 0o644); err != nil {
		t.Fatalf("failed to write base file: %v", err)
	}

	// Create config that extends base
	config := &Config{
		Extends: []string{"./base.yaml"},
		Repo: &RepoConfig{
			Description: ptr("My App"),
			Visibility:  ptr("private"), // Override base
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, tmpDir, visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Repo == nil {
		t.Fatal("expected repo config")
	}

	// Local overrides base
	if result.Repo.Visibility == nil || *result.Repo.Visibility != "private" {
		t.Errorf("expected visibility 'private', got %v", result.Repo.Visibility)
	}

	// Local adds new field
	if result.Repo.Description == nil || *result.Repo.Description != "My App" {
		t.Errorf("expected description 'My App', got %v", result.Repo.Description)
	}

	// Base value preserved
	if result.Repo.AllowMergeCommit == nil || *result.Repo.AllowMergeCommit != false {
		t.Errorf("expected allow_merge_commit false, got %v", result.Repo.AllowMergeCommit)
	}

	// Branch protection from base
	if result.BranchProtection == nil || result.BranchProtection["main"] == nil {
		t.Fatal("expected branch protection for main")
	}
}

func TestResolveExtendsMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-multi-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create first base config
	base1Content := `
repo:
  visibility: public
  allow_merge_commit: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, "base1.yaml"), []byte(base1Content), 0o644); err != nil {
		t.Fatalf("failed to write base1 file: %v", err)
	}

	// Create second base config (overrides first)
	base2Content := `
repo:
  allow_merge_commit: false
  allow_squash_merge: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, "base2.yaml"), []byte(base2Content), 0o644); err != nil {
		t.Fatalf("failed to write base2 file: %v", err)
	}

	config := &Config{
		Extends: []string{"./base1.yaml", "./base2.yaml"},
		Repo: &RepoConfig{
			Description: ptr("My App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, tmpDir, visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// From base1 (not overridden)
	if result.Repo.Visibility == nil || *result.Repo.Visibility != "public" {
		t.Errorf("expected visibility 'public', got %v", result.Repo.Visibility)
	}

	// From base2 (overrides base1)
	if result.Repo.AllowMergeCommit == nil || *result.Repo.AllowMergeCommit != false {
		t.Errorf("expected allow_merge_commit false, got %v", result.Repo.AllowMergeCommit)
	}

	// From base2
	if result.Repo.AllowSquashMerge == nil || *result.Repo.AllowSquashMerge != true {
		t.Errorf("expected allow_squash_merge true, got %v", result.Repo.AllowSquashMerge)
	}

	// From local
	if result.Repo.Description == nil || *result.Repo.Description != "My App" {
		t.Errorf("expected description 'My App', got %v", result.Repo.Description)
	}
}

func TestResolveExtendsNested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-nested-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// grandparent -> parent -> child
	grandparentContent := `
repo:
  visibility: public
  allow_squash_merge: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, "grandparent.yaml"), []byte(grandparentContent), 0o644); err != nil {
		t.Fatalf("failed to write grandparent file: %v", err)
	}

	parentContent := `
extends:
  - ./grandparent.yaml
repo:
  allow_merge_commit: false
  delete_branch_on_merge: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, "parent.yaml"), []byte(parentContent), 0o644); err != nil {
		t.Fatalf("failed to write parent file: %v", err)
	}

	config := &Config{
		Extends: []string{"./parent.yaml"},
		Repo: &RepoConfig{
			Description: ptr("Child App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, tmpDir, visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// From grandparent
	if result.Repo.Visibility == nil || *result.Repo.Visibility != "public" {
		t.Errorf("expected visibility 'public', got %v", result.Repo.Visibility)
	}
	if result.Repo.AllowSquashMerge == nil || *result.Repo.AllowSquashMerge != true {
		t.Errorf("expected allow_squash_merge true, got %v", result.Repo.AllowSquashMerge)
	}

	// From parent
	if result.Repo.AllowMergeCommit == nil || *result.Repo.AllowMergeCommit != false {
		t.Errorf("expected allow_merge_commit false, got %v", result.Repo.AllowMergeCommit)
	}
	if result.Repo.DeleteBranchOnMerge == nil || *result.Repo.DeleteBranchOnMerge != true {
		t.Errorf("expected delete_branch_on_merge true, got %v", result.Repo.DeleteBranchOnMerge)
	}

	// From child
	if result.Repo.Description == nil || *result.Repo.Description != "Child App" {
		t.Errorf("expected description 'Child App', got %v", result.Repo.Description)
	}
}

func TestResolveExtendsCircularReference(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-circular-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// A extends B, B extends A
	configAContent := `
extends:
  - ./b.yaml
repo:
  description: "Config A"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte(configAContent), 0o644); err != nil {
		t.Fatalf("failed to write config A: %v", err)
	}

	configBContent := `
extends:
  - ./a.yaml
repo:
  description: "Config B"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "b.yaml"), []byte(configBContent), 0o644); err != nil {
		t.Fatalf("failed to write config B: %v", err)
	}

	config := &Config{
		Extends: []string{"./a.yaml"},
	}

	visited := make(map[string]bool)
	_, err = resolveExtends(config, tmpDir, visited)
	if err == nil {
		t.Error("expected circular reference error")
	}
	if !strings.Contains(err.Error(), "circular reference") {
		t.Errorf("expected circular reference error, got: %v", err)
	}
}

func TestResolveExtendsSelfReference(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-self-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	selfContent := `
extends:
  - ./self.yaml
repo:
  description: "Self reference"
`
	selfPath := filepath.Join(tmpDir, "self.yaml")
	if err := os.WriteFile(selfPath, []byte(selfContent), 0o644); err != nil {
		t.Fatalf("failed to write self file: %v", err)
	}

	config := &Config{
		Extends: []string{"./self.yaml"},
	}

	visited := make(map[string]bool)
	_, err = resolveExtends(config, tmpDir, visited)
	if err == nil {
		t.Error("expected circular reference error for self-reference")
	}
	if !strings.Contains(err.Error(), "circular reference") {
		t.Errorf("expected circular reference error, got: %v", err)
	}
}

func TestResolveExtendsFileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-notfound-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Extends: []string{"./nonexistent.yaml"},
	}

	visited := make(map[string]bool)
	_, err = resolveExtends(config, tmpDir, visited)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestResolveExtendsFromURL(t *testing.T) {
	baseConfig := `
repo:
  visibility: public
  allow_merge_commit: false
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(baseConfig))
	}))
	defer server.Close()

	config := &Config{
		Extends: []string{server.URL},
		Repo: &RepoConfig{
			Description: ptr("My App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, "", visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Repo.Visibility == nil || *result.Repo.Visibility != "public" {
		t.Errorf("expected visibility 'public', got %v", result.Repo.Visibility)
	}
	if result.Repo.Description == nil || *result.Repo.Description != "My App" {
		t.Errorf("expected description 'My App', got %v", result.Repo.Description)
	}
}

func TestResolveExtendsAbsolutePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-abs-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	baseContent := `
repo:
  visibility: public
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	if err := os.WriteFile(basePath, []byte(baseContent), 0o644); err != nil {
		t.Fatalf("failed to write base file: %v", err)
	}

	config := &Config{
		Extends: []string{basePath}, // Absolute path
		Repo: &RepoConfig{
			Description: ptr("My App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, "/different/path", visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Repo.Visibility == nil || *result.Repo.Visibility != "public" {
		t.Errorf("expected visibility 'public', got %v", result.Repo.Visibility)
	}
}

func TestResolveExtendsEmptyList(t *testing.T) {
	config := &Config{
		Extends: []string{},
		Repo: &RepoConfig{
			Description: ptr("My App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, "", visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the same config since no extends
	if result.Repo.Description == nil || *result.Repo.Description != "My App" {
		t.Errorf("expected description 'My App', got %v", result.Repo.Description)
	}
}

func TestLoadExtendedConfigURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(`repo:
  visibility: public
`))
	}))
	defer server.Close()

	cfg, basePath, err := loadExtendedConfig(server.URL, "/some/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if basePath != "" {
		t.Errorf("expected empty basePath for URL, got %q", basePath)
	}
	if cfg.Repo == nil || cfg.Repo.Visibility == nil {
		t.Error("expected repo config")
	}
}

func TestLoadExtendedConfigRelativePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-relative-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := `repo:
  visibility: public
`
	if err := os.WriteFile(filepath.Join(tmpDir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg, basePath, err := loadExtendedConfig("./base.yaml", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if basePath != tmpDir {
		t.Errorf("expected basePath %q, got %q", tmpDir, basePath)
	}
	if cfg.Repo == nil || cfg.Repo.Visibility == nil {
		t.Error("expected repo config")
	}
}

func TestResolveExtendsInvalidYAMLInBase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-invalid-yaml-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidContent := `invalid: yaml: content:`
	if err := os.WriteFile(filepath.Join(tmpDir, "base.yaml"), []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("failed to write base file: %v", err)
	}

	config := &Config{
		Extends: []string{"./base.yaml"},
	}

	visited := make(map[string]bool)
	_, err = resolveExtends(config, tmpDir, visited)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestResolveExtendsMixedURLAndFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extends-mixed-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Local file
	localContent := `
repo:
  allow_squash_merge: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, "local.yaml"), []byte(localContent), 0o644); err != nil {
		t.Fatalf("failed to write local file: %v", err)
	}

	// Remote URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(`
repo:
  visibility: public
`))
	}))
	defer server.Close()

	config := &Config{
		Extends: []string{"./local.yaml", server.URL},
		Repo: &RepoConfig{
			Description: ptr("My App"),
		},
	}

	visited := make(map[string]bool)
	result, err := resolveExtends(config, tmpDir, visited)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// From local file
	if result.Repo.AllowSquashMerge == nil || *result.Repo.AllowSquashMerge != true {
		t.Error("expected allow_squash_merge from local file")
	}

	// From URL
	if result.Repo.Visibility == nil || *result.Repo.Visibility != "public" {
		t.Error("expected visibility from URL")
	}

	// From config
	if result.Repo.Description == nil || *result.Repo.Description != "My App" {
		t.Error("expected description from config")
	}
}

func TestResolveExtendsURLCircular(t *testing.T) {
	// This tests that we detect circular references even with URLs
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		// Return config that extends itself via URL
		w.Write([]byte(fmt.Sprintf(`
extends:
  - %s
repo:
  visibility: public
`, server.URL)))
	}))
	defer server.Close()

	config := &Config{
		Extends: []string{server.URL},
	}

	visited := make(map[string]bool)
	_, err := resolveExtends(config, "", visited)
	if err == nil {
		t.Error("expected circular reference error for URL")
	}
	if !strings.Contains(err.Error(), "circular reference") {
		t.Errorf("expected circular reference error, got: %v", err)
	}
}
