package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
	repo    string
)

var rootCmd = &cobra.Command{
	Use:     "gh-repo-settings",
	Short:   "Manage GitHub repository settings via YAML configuration",
	Long:    `A GitHub CLI extension to manage repository settings via YAML configuration. Inspired by Terraform's workflow.`,
	Version: "0.1.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show debug output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Only show errors")
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "r", "", "Target repository (default: current repo)")
}
