package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/spf13/cobra"
)

var initOutput string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file interactively",
	Long:  `Create a new repository settings configuration file with interactive prompts.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "", "Output file path (default: .github/repo-settings.yaml)")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("gh-repo-settings configuration wizard")
	fmt.Println()

	cfg := &config.Config{}

	// Repository settings
	var configureRepo bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure repository settings?",
		Default: true,
	}, &configureRepo); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if configureRepo {
		cfg.Repo = &config.RepoConfig{}

		var description string
		if err := survey.AskOne(&survey.Input{
			Message: "Repository description:",
		}, &description); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		if description != "" {
			cfg.Repo.Description = &description
		}

		var visibility string
		if err := survey.AskOne(&survey.Select{
			Message: "Visibility:",
			Options: []string{"public", "private", "internal"},
			Default: "public",
		}, &visibility); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		cfg.Repo.Visibility = &visibility

		var mergeOptions []string
		if err := survey.AskOne(&survey.MultiSelect{
			Message: "Allowed merge methods:",
			Options: []string{"merge commit", "squash merge", "rebase merge"},
			Default: []string{"merge commit", "squash merge"},
		}, &mergeOptions); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		allowMerge := contains(mergeOptions, "merge commit")
		allowSquash := contains(mergeOptions, "squash merge")
		allowRebase := contains(mergeOptions, "rebase merge")
		cfg.Repo.AllowMergeCommit = &allowMerge
		cfg.Repo.AllowSquashMerge = &allowSquash
		cfg.Repo.AllowRebaseMerge = &allowRebase

		var deleteBranch bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Delete branch on merge?",
			Default: true,
		}, &deleteBranch); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		cfg.Repo.DeleteBranchOnMerge = &deleteBranch

		var allowUpdate bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Allow update branch button?",
			Default: true,
		}, &allowUpdate); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		cfg.Repo.AllowUpdateBranch = &allowUpdate
	}

	// Topics
	var configureTopics bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure topics?",
		Default: false,
	}, &configureTopics); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if configureTopics {
		var topics string
		if err := survey.AskOne(&survey.Input{
			Message: "Topics (comma-separated):",
		}, &topics); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		if topics != "" {
			cfg.Topics = splitAndTrim(topics)
		}
	}

	// Labels
	var configureLabels bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure labels?",
		Default: false,
	}, &configureLabels); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if configureLabels {
		var labelPreset string
		if err := survey.AskOne(&survey.Select{
			Message: "Label preset:",
			Options: []string{"none", "semantic", "priority", "custom"},
			Default: "none",
		}, &labelPreset); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		switch labelPreset {
		case "semantic":
			cfg.Labels = &config.LabelsConfig{
				Items: []config.Label{
					{Name: "feat", Color: "0e8a16", Description: "New feature"},
					{Name: "fix", Color: "d73a4a", Description: "Bug fix"},
					{Name: "docs", Color: "0075ca", Description: "Documentation"},
					{Name: "refactor", Color: "cfd3d7", Description: "Code refactoring"},
					{Name: "test", Color: "fbca04", Description: "Tests"},
					{Name: "chore", Color: "fef2c0", Description: "Maintenance"},
				},
			}
		case "priority":
			cfg.Labels = &config.LabelsConfig{
				Items: []config.Label{
					{Name: "priority: critical", Color: "b60205", Description: "Critical priority"},
					{Name: "priority: high", Color: "d93f0b", Description: "High priority"},
					{Name: "priority: medium", Color: "fbca04", Description: "Medium priority"},
					{Name: "priority: low", Color: "0e8a16", Description: "Low priority"},
				},
			}
		}

		if cfg.Labels != nil {
			var replaceDefault bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "Replace default GitHub labels?",
				Default: false,
			}, &replaceDefault); err != nil {
				return fmt.Errorf("prompt failed: %w", err)
			}
			cfg.Labels.ReplaceDefault = replaceDefault
		}
	}

	// Branch protection
	var configureBranch bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure branch protection for 'main'?",
		Default: false,
	}, &configureBranch); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if configureBranch {
		cfg.BranchProtection = make(map[string]*config.BranchRule)
		rule := &config.BranchRule{}

		var requiredReviewsStr string
		if err := survey.AskOne(&survey.Select{
			Message: "Required approving reviews:",
			Options: []string{"0", "1", "2", "3"},
			Default: "1",
		}, &requiredReviewsStr); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		requiredReviews, _ := strconv.Atoi(requiredReviewsStr)
		if requiredReviews > 0 {
			rule.RequiredReviews = &requiredReviews
		}

		var dismissStale bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Dismiss stale reviews?",
			Default: true,
		}, &dismissStale); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		rule.DismissStaleReviews = &dismissStale

		var enforceAdmins bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Enforce rules for administrators?",
			Default: false,
		}, &enforceAdmins); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		rule.EnforceAdmins = &enforceAdmins

		cfg.BranchProtection["main"] = rule
	}

	// Determine output path
	outputPath := initOutput
	if outputPath == "" {
		var outputChoice string
		if err := survey.AskOne(&survey.Select{
			Message: "Output format:",
			Options: []string{
				".github/repo-settings.yaml (single file)",
				".github/repo-settings/ (directory)",
			},
			Default: ".github/repo-settings.yaml (single file)",
		}, &outputChoice); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if outputChoice == ".github/repo-settings.yaml (single file)" {
			outputPath = ".github/repo-settings.yaml"
		} else {
			outputPath = ".github/repo-settings/"
		}
	}

	// Write config
	if filepath.Ext(outputPath) == "" || outputPath[len(outputPath)-1] == '/' {
		return writeConfigToDirectory(cfg, outputPath)
	}
	return writeConfigToFile(cfg, outputPath)
}

func writeConfigToFile(cfg *config.Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := marshalYAML(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("\n✓ Configuration written to %s\n", path)
	return nil
}

func writeConfigToDirectory(cfg *config.Config, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if cfg.Repo != nil {
		data, err := marshalYAML(map[string]interface{}{"repo": cfg.Repo})
		if err != nil {
			return fmt.Errorf("failed to marshal repo config: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "repo.yaml"), data, 0o644); err != nil {
			return fmt.Errorf("failed to write repo.yaml: %w", err)
		}
	}

	if len(cfg.Topics) > 0 {
		data, err := marshalYAML(map[string]interface{}{"topics": cfg.Topics})
		if err != nil {
			return fmt.Errorf("failed to marshal topics config: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "topics.yaml"), data, 0o644); err != nil {
			return fmt.Errorf("failed to write topics.yaml: %w", err)
		}
	}

	if cfg.Labels != nil {
		data, err := marshalYAML(map[string]interface{}{"labels": cfg.Labels})
		if err != nil {
			return fmt.Errorf("failed to marshal labels config: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "labels.yaml"), data, 0o644); err != nil {
			return fmt.Errorf("failed to write labels.yaml: %w", err)
		}
	}

	if cfg.BranchProtection != nil {
		data, err := marshalYAML(map[string]interface{}{"branch_protection": cfg.BranchProtection})
		if err != nil {
			return fmt.Errorf("failed to marshal branch_protection config: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "branch-protection.yaml"), data, 0o644); err != nil {
			return fmt.Errorf("failed to write branch-protection.yaml: %w", err)
		}
	}

	fmt.Printf("\n✓ Configuration written to %s\n", dir)
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range splitString(s, ",") {
		trimmed := trimString(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
