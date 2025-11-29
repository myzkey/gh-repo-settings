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
	planDir      string
	planConfig   string
	checkSecrets bool
	checkEnv     bool
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

	fmt.Printf("Run %s to apply these changes.\n", cyan("gh repo-settings apply"))

	if deletes > 0 {
		os.Exit(2)
	}
}
