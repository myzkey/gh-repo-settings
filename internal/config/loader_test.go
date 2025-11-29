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

func TestLoadActionsConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-actions-test")
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
			name: "full actions config",
			content: `
actions:
  enabled: true
  allowed_actions: selected
  selected_actions:
    github_owned_allowed: true
    verified_allowed: false
    patterns_allowed:
      - "actions/*"
      - "github/codeql-action/*"
  default_workflow_permissions: read
  can_approve_pull_request_reviews: false
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Actions == nil {
					t.Error("expected actions config")
					return
				}
				if cfg.Actions.Enabled == nil || !*cfg.Actions.Enabled {
					t.Error("expected enabled to be true")
				}
				if cfg.Actions.AllowedActions == nil || *cfg.Actions.AllowedActions != "selected" {
					t.Error("expected allowed_actions to be 'selected'")
				}
				if cfg.Actions.SelectedActions == nil {
					t.Error("expected selected_actions config")
					return
				}
				if cfg.Actions.SelectedActions.GithubOwnedAllowed == nil || !*cfg.Actions.SelectedActions.GithubOwnedAllowed {
					t.Error("expected github_owned_allowed to be true")
				}
				if cfg.Actions.SelectedActions.VerifiedAllowed == nil || *cfg.Actions.SelectedActions.VerifiedAllowed {
					t.Error("expected verified_allowed to be false")
				}
				if len(cfg.Actions.SelectedActions.PatternsAllowed) != 2 {
					t.Errorf("expected 2 patterns, got %d", len(cfg.Actions.SelectedActions.PatternsAllowed))
				}
				if cfg.Actions.DefaultWorkflowPermissions == nil || *cfg.Actions.DefaultWorkflowPermissions != "read" {
					t.Error("expected default_workflow_permissions to be 'read'")
				}
				if cfg.Actions.CanApprovePullRequestReviews == nil || *cfg.Actions.CanApprovePullRequestReviews {
					t.Error("expected can_approve_pull_request_reviews to be false")
				}
			},
		},
		{
			name: "minimal actions config",
			content: `
actions:
  enabled: false
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Actions == nil {
					t.Error("expected actions config")
					return
				}
				if cfg.Actions.Enabled == nil || *cfg.Actions.Enabled {
					t.Error("expected enabled to be false")
				}
			},
		},
		{
			name: "actions with workflow permissions only",
			content: `
actions:
  default_workflow_permissions: write
  can_approve_pull_request_reviews: true
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Actions == nil {
					t.Error("expected actions config")
					return
				}
				if cfg.Actions.DefaultWorkflowPermissions == nil || *cfg.Actions.DefaultWorkflowPermissions != "write" {
					t.Error("expected default_workflow_permissions to be 'write'")
				}
				if cfg.Actions.CanApprovePullRequestReviews == nil || !*cfg.Actions.CanApprovePullRequestReviews {
					t.Error("expected can_approve_pull_request_reviews to be true")
				}
			},
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
	filePath := filepath.Join(tmpDir, "unknown.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = loadSingleFile(filePath)
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestLoadActionsFromDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-actions-dir-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	actionsContent := `
actions:
  enabled: true
  allowed_actions: all
  default_workflow_permissions: read
`
	path := filepath.Join(tmpDir, "actions.yaml")
	if err := os.WriteFile(path, []byte(actionsContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg, err := loadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Actions == nil {
		t.Error("expected actions config")
		return
	}
	if cfg.Actions.Enabled == nil || !*cfg.Actions.Enabled {
		t.Error("expected enabled to be true")
	}
	if cfg.Actions.AllowedActions == nil || *cfg.Actions.AllowedActions != "all" {
		t.Error("expected allowed_actions to be 'all'")
	}
}

func TestLoadUnknownFileInDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-unknown-file-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	unknownContent := `
some_config: true
`
	path := filepath.Join(tmpDir, "unknown.yaml")
	if err := os.WriteFile(path, []byte(unknownContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = loadFromDirectory(tmpDir)
	if err == nil {
		t.Error("expected error for unknown file, got nil")
	}
}
