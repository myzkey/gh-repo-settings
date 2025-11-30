package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/myzkey/gh-repo-settings/internal/logger"
	"github.com/myzkey/gh-repo-settings/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	planDir      string
	planConfig   string
	checkSecrets bool
	checkEnv     bool
	showCurrent  bool
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show planned changes without applying them",
	Long:  `Compare local YAML configuration with current GitHub repository settings and show planned changes.`,
	RunE:  runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVarP(&planDir, "dir", "d", "", "Config directory")
	planCmd.Flags().StringVarP(&planConfig, "config", "c", "", "Config file path")
	planCmd.Flags().BoolVar(&checkSecrets, "secrets", false, "Check for required secrets")
	planCmd.Flags().BoolVar(&checkEnv, "env", false, "Check for required environment variables")
	planCmd.Flags().BoolVar(&showCurrent, "show-current", false, "Show current GitHub settings")
}

func runPlan(cmd *cobra.Command, args []string) error {
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

	logger.Debug("Starting plan command")
	logger.Debug("Config dir: %s, Config file: %s", planDir, planConfig)

	client, err := github.NewClientWithContext(ctx, repo)
	if err != nil {
		return err
	}

	logger.Debug("Connected to repository: %s/%s", client.RepoOwner(), client.RepoName())

	cfg, err := config.Load(config.LoadOptions{
		Dir:    planDir,
		Config: planConfig,
	})
	if err != nil {
		return err
	}

	logger.Debug("Loaded configuration")

	// Validate status checks against workflow files
	validateStatusChecks(cfg)

	// Show current GitHub settings if requested
	if showCurrent {
		if err := printCurrentSettings(ctx, client); err != nil {
			return err
		}
		return nil
	}

	logger.Info("Planning changes for %s/%s...\n", client.RepoOwner(), client.RepoName())

	calculator := diff.NewCalculator(client, cfg)
	plan, err := calculator.CalculateWithOptions(ctx, diff.CalculateOptions{
		CheckSecrets: checkSecrets,
		CheckEnv:     checkEnv,
	})
	if err != nil {
		return err
	}

	if !plan.HasChanges() {
		logger.Success("No changes detected. Repository is up to date.")
		return nil
	}

	printPlan(plan)

	// Exit with code 3 if missing secrets/env
	if plan.HasMissingSecrets() || plan.HasMissingVariables() {
		os.Exit(3)
	}

	return nil
}

func printPlan(plan *diff.Plan) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var adds, updates, deletes, missing int

	fmt.Println("Planned changes:")
	fmt.Println()

	currentCategory := ""
	for _, change := range plan.Changes {
		if change.Category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("%s:\n", cyan(change.Category))
			currentCategory = change.Category
		}

		switch change.Type {
		case diff.ChangeAdd:
			fmt.Printf("  %s %s\n", green("+"), change.Key)
			if change.New != nil {
				fmt.Printf("      → %v\n", change.New)
			}
			adds++
		case diff.ChangeUpdate:
			fmt.Printf("  %s %s\n", yellow("~"), change.Key)
			fmt.Printf("      %v → %v\n", change.Old, change.New)
			updates++
		case diff.ChangeDelete:
			fmt.Printf("  %s %s\n", red("-"), change.Key)
			if change.Old != nil {
				fmt.Printf("      ← %v\n", change.Old)
			}
			deletes++
		case diff.ChangeMissing:
			fmt.Printf("  %s %s\n", magenta("!"), change.Key)
			if change.New != nil {
				fmt.Printf("      %v\n", change.New)
			}
			missing++
		}
	}

	fmt.Println()
	fmt.Printf("Plan: %s to add, %s to change, %s to destroy",
		green(fmt.Sprintf("%d", adds)),
		yellow(fmt.Sprintf("%d", updates)),
		red(fmt.Sprintf("%d", deletes)),
	)
	if missing > 0 {
		fmt.Printf(", %s missing", magenta(fmt.Sprintf("%d", missing)))
	}
	fmt.Println(".")
	fmt.Println()

	if missing > 0 {
		fmt.Printf("%s Some required secrets or environment variables are not configured.\n", magenta("Warning:"))
		fmt.Println()
	}

	fmt.Printf("Run %s to apply these changes.\n", cyan("gh rset apply"))

	if deletes > 0 {
		os.Exit(2)
	}
}

func validateStatusChecks(cfg *config.Config) {
	if cfg.BranchProtection == nil {
		return
	}

	// Collect all status checks from branch protection rules
	var allStatusChecks []string
	for _, rule := range cfg.BranchProtection {
		if rule != nil && len(rule.StatusChecks) > 0 {
			allStatusChecks = append(allStatusChecks, rule.StatusChecks...)
		}
	}

	if len(allStatusChecks) == 0 {
		return
	}

	unknown, available, err := workflow.ValidateStatusChecks(allStatusChecks, "")
	if err != nil {
		logger.Debug("Failed to validate status checks: %v", err)
		return
	}

	if len(unknown) > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Println()
		for _, check := range unknown {
			fmt.Printf("%s status check %s not found in workflows\n", yellow("⚠"), yellow(check))
		}
		if len(available) > 0 {
			fmt.Printf("  Available checks: %s\n", strings.Join(available, ", "))
		}
		fmt.Println()
	}
}

func printCurrentSettings(ctx context.Context, client *github.Client) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	fmt.Printf("Current GitHub settings for %s/%s:\n\n", client.RepoOwner(), client.RepoName())

	// Repo settings
	repo, err := client.GetRepo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get repo: %w", err)
	}

	fmt.Printf("%s:\n", cyan("repo"))
	if repo.Description != nil && *repo.Description != "" {
		fmt.Printf("  description: %s\n", *repo.Description)
	}
	if repo.Homepage != nil && *repo.Homepage != "" {
		fmt.Printf("  homepage: %s\n", *repo.Homepage)
	}
	fmt.Printf("  visibility: %s\n", repo.Visibility)
	fmt.Printf("  allow_merge_commit: %v\n", repo.AllowMergeCommit)
	fmt.Printf("  allow_rebase_merge: %v\n", repo.AllowRebaseMerge)
	fmt.Printf("  allow_squash_merge: %v\n", repo.AllowSquashMerge)
	fmt.Printf("  delete_branch_on_merge: %v\n", repo.DeleteBranchOnMerge)
	fmt.Printf("  allow_update_branch: %v\n", repo.AllowUpdateBranch)

	// Topics
	if len(repo.Topics) > 0 {
		fmt.Printf("\n%s:\n", cyan("topics"))
		for _, t := range repo.Topics {
			fmt.Printf("  - %s\n", t)
		}
	}

	// Branch protection for main
	fmt.Printf("\n%s:\n", cyan("branch_protection"))
	bp, err := client.GetBranchProtection(ctx, "main")
	if err != nil {
		fmt.Printf("  main: %s\n", gray("(not configured)"))
	} else {
		fmt.Printf("  main:\n")
		if bp.RequiredPullRequestReviews != nil {
			fmt.Printf("    required_reviews: %d\n", bp.RequiredPullRequestReviews.RequiredApprovingReviewCount)
			fmt.Printf("    dismiss_stale_reviews: %v\n", bp.RequiredPullRequestReviews.DismissStaleReviews)
			fmt.Printf("    require_code_owner: %v\n", bp.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		} else {
			fmt.Printf("    required_reviews: %s\n", gray("(not set)"))
		}
		if bp.RequiredStatusChecks != nil {
			fmt.Printf("    require_status_checks: true\n")
			fmt.Printf("    strict_status_checks: %v\n", bp.RequiredStatusChecks.Strict)
			if len(bp.RequiredStatusChecks.Contexts) > 0 {
				fmt.Printf("    status_checks:\n")
				for _, c := range bp.RequiredStatusChecks.Contexts {
					fmt.Printf("      - %s\n", c)
				}
			}
		} else {
			fmt.Printf("    require_status_checks: false\n")
		}
		if bp.EnforceAdmins != nil {
			fmt.Printf("    enforce_admins: %v\n", bp.EnforceAdmins.Enabled)
		}
		if bp.RequiredLinearHistory != nil {
			fmt.Printf("    require_linear_history: %v\n", bp.RequiredLinearHistory.Enabled)
		}
		if bp.AllowForcePushes != nil {
			fmt.Printf("    allow_force_pushes: %v\n", bp.AllowForcePushes.Enabled)
		}
		if bp.AllowDeletions != nil {
			fmt.Printf("    allow_deletions: %v\n", bp.AllowDeletions.Enabled)
		}
	}

	// Actions
	fmt.Printf("\n%s:\n", cyan("actions"))
	actionsPerms, err := client.GetActionsPermissions(ctx)
	if err != nil {
		fmt.Printf("  %s\n", gray("(failed to get)"))
	} else {
		fmt.Printf("  enabled: %v\n", actionsPerms.Enabled)
		fmt.Printf("  allowed_actions: %s\n", actionsPerms.AllowedActions)
	}

	workflowPerms, err := client.GetActionsWorkflowPermissions(ctx)
	if err == nil {
		fmt.Printf("  default_workflow_permissions: %s\n", workflowPerms.DefaultWorkflowPermissions)
		fmt.Printf("  can_approve_pull_request_reviews: %v\n", workflowPerms.CanApprovePullRequestReviews)
	}

	return nil
}
