package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/spf13/cobra"
)

var (
	planDir    string
	planConfig string
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
}

func runPlan(cmd *cobra.Command, args []string) error {
	client, err := github.NewClient(repo)
	if err != nil {
		return err
	}

	cfg, err := config.Load(config.LoadOptions{
		Dir:    planDir,
		Config: planConfig,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Planning changes for %s/%s...\n\n", client.Repo.Owner, client.Repo.Name)

	calculator := diff.NewCalculator(client, cfg)
	plan, err := calculator.Calculate()
	if err != nil {
		return err
	}

	if !plan.HasChanges() {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Println(green("✓ No changes detected. Repository is up to date."))
		return nil
	}

	printPlan(plan)
	return nil
}

func printPlan(plan *diff.Plan) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var adds, updates, deletes int

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
		}
	}

	fmt.Println()
	fmt.Printf("Plan: %s to add, %s to change, %s to destroy.\n",
		green(fmt.Sprintf("%d", adds)),
		yellow(fmt.Sprintf("%d", updates)),
		red(fmt.Sprintf("%d", deletes)),
	)
	fmt.Println()
	fmt.Printf("Run %s to apply these changes.\n", cyan("gh repo-settings apply"))

	if deletes > 0 {
		os.Exit(2)
	}
}
