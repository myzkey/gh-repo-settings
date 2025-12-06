package cmd

import (
	"os"

	"github.com/myzkey/gh-repo-settings/internal/infra/logger"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
	repo    string

	// Version is set by main.go from version.go
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "gh-repo-settings",
	Short: "Manage GitHub repository settings via YAML configuration",
	Long:  `A GitHub CLI extension to manage repository settings via YAML configuration. Inspired by Terraform's workflow.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set log level based on flags
		if quiet {
			logger.SetDefaultLevel(logger.LevelQuiet)
		} else if verbose {
			logger.SetDefaultLevel(logger.LevelVerbose)
		} else {
			logger.SetDefaultLevel(logger.LevelNormal)
		}
	},
}

func Execute() {
	rootCmd.Version = Version
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	err := rootCmd.Execute()
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show debug output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Only show errors")
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "r", "", "Target repository (default: current repo)")
}
