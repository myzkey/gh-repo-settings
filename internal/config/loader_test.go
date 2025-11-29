package config

import (
	"os"
	"path/filepath"
	"testing"
)

// E2E tests for Load function - priority and integration

func TestLoadSingleFile(t *testing.T) {
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
actions:
  enabled: true
  allowed_actions: all
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
				if cfg.BranchProtection == nil || cfg.BranchProtection["main"] == nil {
					t.Error("expected branch protection for main")
					return
				}
				if cfg.Secrets == nil || len(cfg.Secrets.Required) != 1 {
					t.Error("expected 1 required secret")
				}
				if cfg.Env == nil || len(cfg.Env.Required) != 1 {
					t.Error("expected 1 required env var")
				}
				if cfg.Actions == nil || !*cfg.Actions.Enabled {
					t.Error("expected actions to be enabled")
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
				if cfg.Repo == nil || *cfg.Repo.Description != "Minimal" {
					t.Error("expected description 'Minimal'")
				}
			},
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
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

			cfg, err := Load(LoadOptions{Config: filePath})
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
	tmpDir, err := os.MkdirTemp("", "config-dir-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"repo.yaml": `
repo:
  description: "From repo.yaml"
  visibility: private
`,
		"topics.yaml": `
topics:
  - testing
  - golang
`,
		"labels.yaml": `
labels:
  items:
    - name: enhancement
      color: a2eeef
`,
		"branch_protection.yaml": `
branch_protection:
  main:
    required_reviews: 1
`,
		"actions.yaml": `
actions:
  enabled: true
`,
		"ignored.txt": "should be ignored",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	cfg, err := Load(LoadOptions{Dir: tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Repo == nil || *cfg.Repo.Description != "From repo.yaml" {
		t.Error("expected description 'From repo.yaml'")
	}
	if len(cfg.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(cfg.Topics))
	}
	if cfg.Labels == nil || len(cfg.Labels.Items) != 1 {
		t.Error("expected 1 label")
	}
	if cfg.BranchProtection == nil || cfg.BranchProtection["main"] == nil {
		t.Error("expected branch protection for main")
	}
	if cfg.Actions == nil || !*cfg.Actions.Enabled {
		t.Error("expected actions to be enabled")
	}
}

func TestLoadDirectFormat(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(tmpDir, "repo.yaml"), []byte(repoContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg, err := Load(LoadOptions{Dir: tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Repo == nil || cfg.Repo.Description == nil || *cfg.Repo.Description != "Direct format" {
		t.Error("expected description 'Direct format'")
	}
}

func TestLoadPriority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-priority-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(oldWd)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("failed to create .github dir: %v", err)
	}

	// Create default single file
	singleFileContent := `
repo:
  description: "From single file"
`
	if err := os.WriteFile(filepath.Join(githubDir, "repo-settings.yaml"), []byte(singleFileContent), 0644); err != nil {
		t.Fatalf("failed to write default single file: %v", err)
	}

	// Test default single file
	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Repo == nil || *cfg.Repo.Description != "From single file" {
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
	if err := os.WriteFile(filepath.Join(defaultDir, "repo.yaml"), []byte(dirContent), 0644); err != nil {
		t.Fatalf("failed to write directory file: %v", err)
	}

	// Test default directory takes priority
	cfg, err = Load(LoadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Repo == nil || *cfg.Repo.Description != "From directory" {
		t.Error("expected directory to take priority over single file")
	}
}

func TestLoadUnknownFieldError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-unknown-field-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := `
repo:
  description: "Test"
unknown_field:
  some_setting: true
`
	filePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = Load(LoadOptions{Config: filePath})
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestLoadUnknownFileInDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-unknown-file-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	unknownContent := `some_config: true`
	if err := os.WriteFile(filepath.Join(tmpDir, "unknown.yaml"), []byte(unknownContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = Load(LoadOptions{Dir: tmpDir})
	if err == nil {
		t.Error("expected error for unknown file, got nil")
	}
}

func TestLoadNoConfigFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-noconfig-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(oldWd)

	_, err = Load(LoadOptions{})
	if err == nil {
		t.Error("expected error when no config found")
	}
}

func TestLoadWithExtends(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-extends-e2e-test")
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
	if err := os.WriteFile(filepath.Join(tmpDir, "base.yaml"), []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write base file: %v", err)
	}

	// Create main config that extends base
	mainContent := `
extends:
  - ./base.yaml
repo:
  description: "My App"
  visibility: private
`
	mainPath := filepath.Join(tmpDir, "main.yaml")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}

	cfg, err := Load(LoadOptions{Config: mainPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Local overrides base
	if cfg.Repo.Visibility == nil || *cfg.Repo.Visibility != "private" {
		t.Errorf("expected visibility 'private', got %v", cfg.Repo.Visibility)
	}

	// Local adds new field
	if cfg.Repo.Description == nil || *cfg.Repo.Description != "My App" {
		t.Errorf("expected description 'My App', got %v", cfg.Repo.Description)
	}

	// Base value preserved
	if cfg.Repo.AllowMergeCommit == nil || *cfg.Repo.AllowMergeCommit != false {
		t.Errorf("expected allow_merge_commit false, got %v", cfg.Repo.AllowMergeCommit)
	}

	// Branch protection from base
	if cfg.BranchProtection == nil || cfg.BranchProtection["main"] == nil {
		t.Fatal("expected branch protection for main")
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
}
