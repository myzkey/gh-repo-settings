package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	initOutput string
)

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
	survey.AskOne(&survey.Confirm{
		Message: "Configure repository settings?",
		Default: true,
	}, &configureRepo)

	if configureRepo {
		cfg.Repo = &config.RepoConfig{}

		var description string
		survey.AskOne(&survey.Input{
			Message: "Repository description:",
		}, &description)
		if description != "" {
			cfg.Repo.Description = &description
		}

		var visibility string
		survey.AskOne(&survey.Select{
			Message: "Visibility:",
			Options: []string{"public", "private", "internal"},
			Default: "public",
		}, &visibility)
		cfg.Repo.Visibility = &visibility

		var mergeOptions []string
		survey.AskOne(&survey.MultiSelect{
			Message: "Allowed merge methods:",
			Options: []string{"merge commit", "squash merge", "rebase merge"},
			Default: []string{"merge commit", "squash merge"},
		}, &mergeOptions)

		allowMerge := contains(mergeOptions, "merge commit")
		allowSquash := contains(mergeOptions, "squash merge")
		allowRebase := contains(mergeOptions, "rebase merge")
		cfg.Repo.AllowMergeCommit = &allowMerge
		cfg.Repo.AllowSquashMerge = &allowSquash
		cfg.Repo.AllowRebaseMerge = &allowRebase

		var deleteBranch bool
		survey.AskOne(&survey.Confirm{
			Message: "Delete branch on merge?",
			Default: true,
		}, &deleteBranch)
		cfg.Repo.DeleteBranchOnMerge = &deleteBranch

		var allowUpdate bool
		survey.AskOne(&survey.Confirm{
			Message: "Allow update branch button?",
			Default: true,
		}, &allowUpdate)
		cfg.Repo.AllowUpdateBranch = &allowUpdate
	}

	// Topics
	var configureTopics bool
	survey.AskOne(&survey.Confirm{
		Message: "Configure topics?",
		Default: false,
	}, &configureTopics)

	if configureTopics {
		var topics string
		survey.AskOne(&survey.Input{
			Message: "Topics (comma-separated):",
		}, &topics)
		if topics != "" {
			cfg.Topics = splitAndTrim(topics)
		}
	}

	// Labels
	var configureLabels bool
	survey.AskOne(&survey.Confirm{
		Message: "Configure labels?",
		Default: false,
	}, &configureLabels)

	if configureLabels {
		var labelPreset string
		survey.AskOne(&survey.Select{
			Message: "Label preset:",
			Options: []string{"none", "semantic", "priority", "custom"},
			Default: "none",
		}, &labelPreset)

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
			survey.AskOne(&survey.Confirm{
				Message: "Replace default GitHub labels?",
				Default: false,
			}, &replaceDefault)
			cfg.Labels.ReplaceDefault = replaceDefault
		}
	}

	// Branch protection
	var configureBranch bool
	survey.AskOne(&survey.Confirm{
		Message: "Configure branch protection for 'main'?",
		Default: false,
	}, &configureBranch)

	if configureBranch {
		cfg.BranchProtection = make(map[string]*config.BranchRule)
		rule := &config.BranchRule{}

		var requiredReviews int
		survey.AskOne(&survey.Select{
			Message: "Required approving reviews:",
			Options: []string{"0", "1", "2", "3"},
			Default: "1",
		}, &requiredReviews)
		if requiredReviews > 0 {
			rule.RequiredReviews = &requiredReviews
		}

		var dismissStale bool
		survey.AskOne(&survey.Confirm{
			Message: "Dismiss stale reviews?",
			Default: true,
		}, &dismissStale)
		rule.DismissStaleReviews = &dismissStale

		var enforceAdmins bool
		survey.AskOne(&survey.Confirm{
			Message: "Enforce rules for administrators?",
			Default: false,
		}, &enforceAdmins)
		rule.EnforceAdmins = &enforceAdmins

		cfg.BranchProtection["main"] = rule
	}

	// Determine output path
	outputPath := initOutput
	if outputPath == "" {
		var outputChoice string
		survey.AskOne(&survey.Select{
			Message: "Output format:",
			Options: []string{
				".github/repo-settings.yaml (single file)",
				".github/repo-settings/ (directory)",
			},
			Default: ".github/repo-settings.yaml (single file)",
		}, &outputChoice)

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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	fmt.Printf("\n✓ Configuration written to %s\n", path)
	return nil
}

func writeConfigToDirectory(cfg *config.Config, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if cfg.Repo != nil {
		data, _ := yaml.Marshal(map[string]interface{}{"repo": cfg.Repo})
		os.WriteFile(filepath.Join(dir, "repo.yaml"), data, 0644)
	}

	if len(cfg.Topics) > 0 {
		data, _ := yaml.Marshal(map[string]interface{}{"topics": cfg.Topics})
		os.WriteFile(filepath.Join(dir, "topics.yaml"), data, 0644)
	}

	if cfg.Labels != nil {
		data, _ := yaml.Marshal(map[string]interface{}{"labels": cfg.Labels})
		os.WriteFile(filepath.Join(dir, "labels.yaml"), data, 0644)
	}

	if cfg.BranchProtection != nil {
		data, _ := yaml.Marshal(map[string]interface{}{"branch_protection": cfg.BranchProtection})
		os.WriteFile(filepath.Join(dir, "branch-protection.yaml"), data, 0644)
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
