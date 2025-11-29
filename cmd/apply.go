package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/spf13/cobra"
)

var (
	applyDir    string
	applyConfig string
	autoApprove bool
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
}

func runApply(cmd *cobra.Command, args []string) error {
	client, err := github.NewClient(repo)
	if err != nil {
		return err
	}

	cfg, err := config.Load(config.LoadOptions{
		Dir:    applyDir,
		Config: applyConfig,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Applying changes to %s/%s...\n\n", client.Repo.Owner, client.Repo.Name)

	calculator := diff.NewCalculator(client, cfg)
	plan, err := calculator.Calculate()
	if err != nil {
		return err
	}

	if !plan.HasChanges() {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Println(green("✓ No changes to apply. Repository is up to date."))
		return nil
	}

	printPlan(plan)

	if !autoApprove {
		fmt.Print("Do you want to apply these changes? (yes/no): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "yes" && answer != "y" {
			fmt.Println("Apply cancelled.")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("Applying changes...")
	fmt.Println()

	return applyChanges(client, cfg, plan)
}

func applyChanges(client *github.Client, cfg *config.Config, plan *diff.Plan) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Group changes by category
	repoChanges := make(map[string]interface{})
	var topicsChanged bool
	var labelChanges []diff.Change

	for _, change := range plan.Changes {
		switch change.Category {
		case "repo":
			repoChanges[change.Key] = change.New
		case "topics":
			topicsChanged = true
		case "labels":
			labelChanges = append(labelChanges, change)
		}
	}

	// Apply repo changes
	if len(repoChanges) > 0 {
		fmt.Print("  Updating repository settings... ")
		if err := client.UpdateRepo(repoChanges); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("failed to update repo: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// Apply topics
	if topicsChanged {
		fmt.Print("  Updating topics... ")
		if err := client.SetTopics(cfg.Topics); err != nil {
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
			if err := client.CreateLabel(label.Name, label.Color, label.Description); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to create label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))

		case diff.ChangeUpdate:
			fmt.Printf("  Updating label '%s'... ", change.Key)
			label := findLabel(cfg.Labels.Items, change.Key)
			if err := client.UpdateLabel(change.Key, label.Name, label.Color, label.Description); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to update label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))

		case diff.ChangeDelete:
			fmt.Printf("  Deleting label '%s'... ", change.Key)
			if err := client.DeleteLabel(change.Key); err != nil {
				fmt.Println(red("✗"))
				return fmt.Errorf("failed to delete label %s: %w", change.Key, err)
			}
			fmt.Println(green("✓"))
		}
	}

	fmt.Println()
	fmt.Println(green("✓ Apply complete!"))

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
