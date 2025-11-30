package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/myzkey/gh-repo-settings/internal/logger"
	"github.com/spf13/cobra"
)

var (
	applyDir          string
	applyConfig       string
	autoApprove       bool
	applyCheckSecrets bool
	applyCheckEnv     bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration changes to the repository",
	Long:  `Apply local YAML configuration to GitHub repository settings.`,
	RunE:  runApply,
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&applyDir, "dir", "d", "", "Config directory")
	applyCmd.Flags().StringVarP(&applyConfig, "config", "c", "", "Config file path")
	applyCmd.Flags().BoolVarP(&autoApprove, "yes", "y", false, "Auto-approve changes")
	applyCmd.Flags().BoolVar(&applyCheckSecrets, "secrets", false, "Check for required secrets")
	applyCmd.Flags().BoolVar(&applyCheckEnv, "env", false, "Check for required environment variables")
}

func runApply(cmd *cobra.Command, args []string) error {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("Received interrupt, cancelling...")
		cancel()
	}()

	logger.Debug("Starting apply command")

	client, err := github.NewClientWithContext(ctx, repo)
	if err != nil {
		return err
	}

	logger.Debug("Connected to repository: %s/%s", client.RepoOwner(), client.RepoName())

	cfg, err := config.Load(config.LoadOptions{
		Dir:    applyDir,
		Config: applyConfig,
	})
	if err != nil {
		return err
	}

	logger.Debug("Loaded configuration")

	logger.Info("Applying changes to %s/%s...\n", client.RepoOwner(), client.RepoName())

	calculator := diff.NewCalculator(client, cfg)
	plan, err := calculator.CalculateWithOptions(ctx, diff.CalculateOptions{
		CheckSecrets: applyCheckSecrets,
		CheckEnv:     applyCheckEnv,
	})
	if err != nil {
		return err
	}

	if !plan.HasChanges() {
		logger.Success("No changes to apply. Repository is up to date.")
		return nil
	}

	// Check for missing secrets/env before proceeding
	if plan.HasMissingSecrets() || plan.HasMissingVariables() {
		printPlan(plan)
		return fmt.Errorf("cannot apply: required secrets or environment variables are missing")
	}

	printPlan(plan)

	if !autoApprove {
		fmt.Print("Do you want to apply these changes? (yes/no): ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if answer != "yes" && answer != "y" {
			logger.Info("Apply cancelled.")
			return nil
		}
	}

	fmt.Println()
	logger.Info("Applying changes...")
	fmt.Println()

	return applyChanges(ctx, client, cfg, plan)
}

func applyChanges(ctx context.Context, client *github.Client, cfg *config.Config, plan *diff.Plan) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Group changes by category
	repoChanges := make(map[string]interface{})
	var topicsChanged bool
	var labelChanges []diff.Change
	branchProtectionChanges := make(map[string][]diff.Change)
	var actionsChanges []diff.Change
	var pagesChanges []diff.Change

	for _, change := range plan.Changes {
		switch change.Category {
		case "repo":
			repoChanges[change.Key] = change.New
		case "topics":
			topicsChanged = true
		case "labels":
			labelChanges = append(labelChanges, change)
		case "branch_protection":
			// Extract branch name from key (format: "branch.setting")
			branchName := extractBranchName(change.Key)
			branchProtectionChanges[branchName] = append(branchProtectionChanges[branchName], change)
		case "actions":
			actionsChanges = append(actionsChanges, change)
		case "pages":
			pagesChanges = append(pagesChanges, change)
		}
	}

	// Apply repo changes
	if len(repoChanges) > 0 {
		fmt.Print("  Updating repository settings... ")
		if err := client.UpdateRepo(ctx, repoChanges); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update repo: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// Apply topics
	if topicsChanged {
		fmt.Print("  Updating topics... ")
		if err := client.SetTopics(ctx, cfg.Topics); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update topics: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// Apply label changes
	for _, change := range labelChanges {
		switch change.Type {
		case diff.ChangeAdd:
			fmt.Printf("  Creating label '%s'... ", change.Key)
			label := findLabel(cfg.Labels.Items, change.Key)
			if err := client.CreateLabel(ctx, label.Name, label.Color, label.Description); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to create label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))

		case diff.ChangeUpdate:
			fmt.Printf("  Updating label '%s'... ", change.Key)
			label := findLabel(cfg.Labels.Items, change.Key)
			if err := client.UpdateLabel(ctx, change.Key, label.Name, label.Color, label.Description); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to update label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))

		case diff.ChangeDelete:
			fmt.Printf("  Deleting label '%s'... ", change.Key)
			if err := client.DeleteLabel(ctx, change.Key); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to delete label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))
		}
	}

	// Apply branch protection changes
	for branchName := range branchProtectionChanges {
		fmt.Printf("  Updating branch protection for '%s'... ", branchName)

		rule := cfg.BranchProtection[branchName]
		settings := &github.BranchProtectionSettings{
			RequiredReviews:         rule.RequiredReviews,
			DismissStaleReviews:     rule.DismissStaleReviews,
			RequireCodeOwnerReviews: rule.RequireCodeOwner,
			RequireStatusChecks:     rule.RequireStatusChecks,
			StatusChecks:            rule.StatusChecks,
			StrictStatusChecks:      rule.StrictStatusChecks,
			EnforceAdmins:           rule.EnforceAdmins,
			RequireLinearHistory:    rule.RequireLinearHistory,
			AllowForcePushes:        rule.AllowForcePushes,
			AllowDeletions:          rule.AllowDeletions,
			RequireSignedCommits:    rule.RequireSignedCommits,
		}

		if err := client.UpdateBranchProtection(ctx, branchName, settings); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update branch protection for %s: %w", branchName, err)
		}
		fmt.Println(green("✓"))
	}

	// Apply actions changes
	if len(actionsChanges) > 0 && cfg.Actions != nil {
		if err := applyActionsChanges(ctx, client, cfg, actionsChanges, green, red); err != nil {
			return err
		}
	}

	// Apply pages changes
	if len(pagesChanges) > 0 && cfg.Pages != nil {
		if err := applyPagesChanges(ctx, client, cfg, pagesChanges, green, red); err != nil {
			return err
		}
	}

	fmt.Println()
	logger.Success("Apply complete!")

	return nil
}

func applyActionsChanges(ctx context.Context, client *github.Client, cfg *config.Config, changes []diff.Change, green, red func(a ...interface{}) string) error {
	// Check which settings need updating
	needsPermissionsUpdate := false
	needsSelectedUpdate := false
	needsWorkflowUpdate := false

	for _, change := range changes {
		switch change.Key {
		case "enabled", "allowed_actions":
			needsPermissionsUpdate = true
		case "github_owned_allowed", "verified_allowed", "patterns_allowed":
			needsSelectedUpdate = true
		case "default_workflow_permissions", "can_approve_pull_request_reviews":
			needsWorkflowUpdate = true
		}
	}

	// Update actions permissions
	if needsPermissionsUpdate {
		fmt.Print("  Updating actions permissions... ")
		enabled := true
		if cfg.Actions.Enabled != nil {
			enabled = *cfg.Actions.Enabled
		}
		allowedActions := "all"
		if cfg.Actions.AllowedActions != nil {
			allowedActions = *cfg.Actions.AllowedActions
		}
		if err := client.UpdateActionsPermissions(ctx, enabled, allowedActions); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update actions permissions: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// Update selected actions
	if needsSelectedUpdate && cfg.Actions.SelectedActions != nil {
		fmt.Print("  Updating selected actions... ")
		settings := &github.ActionsSelectedData{}
		if cfg.Actions.SelectedActions.GithubOwnedAllowed != nil {
			settings.GithubOwnedAllowed = *cfg.Actions.SelectedActions.GithubOwnedAllowed
		}
		if cfg.Actions.SelectedActions.VerifiedAllowed != nil {
			settings.VerifiedAllowed = *cfg.Actions.SelectedActions.VerifiedAllowed
		}
		settings.PatternsAllowed = cfg.Actions.SelectedActions.PatternsAllowed
		if err := client.UpdateActionsSelectedActions(ctx, settings); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update selected actions: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// Update workflow permissions
	if needsWorkflowUpdate {
		fmt.Print("  Updating workflow permissions... ")
		permissions := "read"
		if cfg.Actions.DefaultWorkflowPermissions != nil {
			permissions = *cfg.Actions.DefaultWorkflowPermissions
		}
		canApprove := false
		if cfg.Actions.CanApprovePullRequestReviews != nil {
			canApprove = *cfg.Actions.CanApprovePullRequestReviews
		}
		if err := client.UpdateActionsWorkflowPermissions(ctx, permissions, canApprove); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update workflow permissions: %w", err)
		}
		fmt.Println(green("✓"))
	}

	return nil
}

func findLabel(labels []config.Label, name string) config.Label {
	for _, l := range labels {
		if l.Name == name {
			return l
		}
	}
	return config.Label{}
}

func extractBranchName(key string) string {
	// Key format is "branchName.setting"
	for i, c := range key {
		if c == '.' {
			return key[:i]
		}
	}
	return key
}

func applyPagesChanges(ctx context.Context, client *github.Client, cfg *config.Config, changes []diff.Change, green, red func(a ...interface{}) string) error {
	// Check if pages needs to be created or updated
	needsCreate := false
	needsUpdate := false

	for _, change := range changes {
		if change.Type == diff.ChangeAdd && change.Key == "pages" {
			needsCreate = true
		} else {
			needsUpdate = true
		}
	}

	buildType := "workflow"
	if cfg.Pages.BuildType != nil {
		buildType = *cfg.Pages.BuildType
	}

	var source *github.PagesSourceData
	if cfg.Pages.Source != nil {
		source = &github.PagesSourceData{}
		if cfg.Pages.Source.Branch != nil {
			source.Branch = *cfg.Pages.Source.Branch
		}
		if cfg.Pages.Source.Path != nil {
			source.Path = *cfg.Pages.Source.Path
		}
	}

	if needsCreate {
		fmt.Print("  Creating GitHub Pages... ")
		if err := client.CreatePages(ctx, buildType, source); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to create pages: %w", err)
		}
		fmt.Println(green("✓"))
	} else if needsUpdate {
		fmt.Print("  Updating GitHub Pages... ")
		if err := client.UpdatePages(ctx, buildType, source); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update pages: %w", err)
		}
		fmt.Println(green("✓"))
	}

	return nil
}
