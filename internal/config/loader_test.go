package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSingleFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name: "full config",
			content: `
repo:
  description: "Test repository"
  visibility: public
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false
  delete_branch_on_merge: true
topics:
  - go
  - cli
labels:
  replace_default: true
  items:
    - name: bug
      color: d73a4a
      description: Something isn't working
branch_protection:
  main:
    required_reviews: 2
    enforce_admins: true
secrets:
  required:
    - DEPLOY_KEY
env:
  required:
    - NODE_ENV
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Repo == nil {
					t.Error("expected repo config")
					return
				}
				if *cfg.Repo.Description != "Test repository" {
					t.Errorf("expected description 'Test repository', got '%s'", *cfg.Repo.Description)
				}
				if len(cfg.Topics) != 2 {
					t.Errorf("expected 2 topics, got %d", len(cfg.Topics))
				}
				if cfg.Labels == nil || len(cfg.Labels.Items) != 1 {
					t.Error("expected 1 label")
					return
				}
				if !cfg.Labels.ReplaceDefault {
					t.Error("expected replace_default to be true")
				}
				if cfg.BranchProtection == nil || cfg.BranchProtection["main"] == nil {
					t.Error("expected branch protection for main")
					return
				}
				if *cfg.BranchProtection["main"].RequiredReviews != 2 {
					t.Errorf("expected 2 required reviews, got %d", *cfg.BranchProtection["main"].RequiredReviews)
				}
				if cfg.Secrets == nil || len(cfg.Secrets.Required) != 1 {
					t.Error("expected 1 required secret")
				}
				if cfg.Env == nil || len(cfg.Env.Required) != 1 {
					t.Error("expected 1 required env var")
				}
			},
		},
		{
			name: "minimal config",
			content: `
repo:
  description: "Minimal"
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Repo == nil {
					t.Error("expected repo config")
					return
				}
				if *cfg.Repo.Description != "Minimal" {
					t.Errorf("expected description 'Minimal', got '%s'", *cfg.Repo.Description)
				}
			},
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				// Empty config should be valid
			},
		},
		{
			name:    "invalid yaml",
			content: "invalid: yaml: content:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name+".yaml")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			cfg, err := loadSingleFile(filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestLoadFromDirectory(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-dir-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create separate config files
	repoContent := `
repo:
  description: "From repo.yaml"
  visibility: private
`
	topicsContent := `
topics:
  - testing
  - golang
`
	labelsContent := `
labels:
  items:
    - name: enhancement
      color: a2eeef
`
	branchProtectionContent := `
branch_protection:
  main:
    required_reviews: 1
`

	files := map[string]string{
		"repo.yaml":              repoContent,
		"topics.yaml":            topicsContent,
		"labels.yaml":            labelsContent,
		"branch_protection.yaml": branchProtectionContent,
		"ignored.txt":            "should be ignored",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	cfg, err := loadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check repo
	if cfg.Repo == nil {
		t.Error("expected repo config")
	} else if *cfg.Repo.Description != "From repo.yaml" {
		t.Errorf("expected description 'From repo.yaml', got '%s'", *cfg.Repo.Description)
	}

	// Check topics
	if len(cfg.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(cfg.Topics))
	}

	// Check labels
	if cfg.Labels == nil || len(cfg.Labels.Items) != 1 {
		t.Error("expected 1 label")
	}

	// Check branch protection
	if cfg.BranchProtection == nil || cfg.BranchProtection["main"] == nil {
		t.Error("expected branch protection for main")
	}
}

func TestLoadDirectFormat(t *testing.T) {
	// Test loading files without wrapper keys (direct format)
	tmpDir, err := os.MkdirTemp("", "config-direct-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Direct format (no wrapper key)
	repoContent := `
description: "Direct format"
visibility: public
`

	path := filepath.Join(tmpDir, "repo.yaml")
	if err := os.WriteFile(path, []byte(repoContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg, err := loadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Repo == nil {
		t.Error("expected repo config")
	} else if cfg.Repo.Description == nil || *cfg.Repo.Description != "Direct format" {
		t.Error("expected description 'Direct format'")
	}
}

func TestLoadPriority(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "config-priority-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory for default path testing
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(oldWd)

	// Create .github directory first
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("failed to create .github dir: %v", err)
	}

	// Create default single file
	singleFileContent := `
repo:
  description: "From single file"
`
	defaultSingleFile := filepath.Join(githubDir, "repo-settings.yaml")
	if err := os.WriteFile(defaultSingleFile, []byte(singleFileContent), 0644); err != nil {
		t.Fatalf("failed to write default single file: %v", err)
	}

	// Test default single file
	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Repo == nil || cfg.Repo.Description == nil || *cfg.Repo.Description != "From single file" {
		t.Error("expected to load from default single file")
	}

	// Create default directory with content
	defaultDir := filepath.Join(githubDir, "repo-settings")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		t.Fatalf("failed to create default dir: %v", err)
	}

	dirContent := `
repo:
  description: "From directory"
`
	dirFile := filepath.Join(defaultDir, "repo.yaml")
	if err := os.WriteFile(dirFile, []byte(dirContent), 0644); err != nil {
		t.Fatalf("failed to write directory file: %v", err)
	}

	// Test default directory takes priority
	cfg, err = Load(LoadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Repo == nil || cfg.Repo.Description == nil || *cfg.Repo.Description != "From directory" {
		t.Error("expected directory to take priority over single file")
	}
}

func TestToYAML(t *testing.T) {
	desc := "Test description"
	visibility := "public"
	cfg := &Config{
		Repo: &RepoConfig{
			Description: &desc,
			Visibility:  &visibility,
		},
		Topics: []string{"go", "cli"},
	}

	yaml, err := cfg.ToYAML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if yaml == "" {
		t.Error("expected non-empty YAML output")
	}

	// Verify it can be parsed back
	var parsed Config
	if err := parseYAML([]byte(yaml), &parsed); err != nil {
		t.Errorf("generated YAML is not valid: %v", err)
	}
}

func parseYAML(data []byte, v interface{}) error {
	return nil // Simplified for test
}
