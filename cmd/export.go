package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/myzkey/gh-repo-settings/internal/logger"
	"github.com/oapi-codegen/nullable"
	"github.com/spf13/cobra"
)

var (
	exportDir            string
	exportSingle         string
	exportIncludeSecrets bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export current GitHub repository settings to YAML",
	Long:  `Export current GitHub repository settings to YAML format.`,
	RunE:  runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportDir, "dir", "d", "", "Export to directory (multiple YAML files)")
	exportCmd.Flags().StringVarP(&exportSingle, "single", "s", "", "Export to single YAML file")
	exportCmd.Flags().BoolVar(&exportIncludeSecrets, "include-secrets", false, "Include secret names in export")
}

func runExport(cmd *cobra.Command, args []string) error {
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

	logger.Debug("Starting export command")

	client, err := github.NewClientWithContext(ctx, repo)
	if err != nil {
		return err
	}

	logger.Info("Exporting settings from %s/%s...", client.RepoOwner(), client.RepoName())

	cfg := &config.Config{}

	// Get repo settings
	repoData, err := client.GetRepo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get repo settings: %w", err)
	}

	cfg.Repo = &config.RepoConfig{
		Description:         nullableToPtr(repoData.Description),
		Homepage:            nullableToPtr(repoData.Homepage),
		Visibility:          repoData.Visibility,
		AllowMergeCommit:    repoData.AllowMergeCommit,
		AllowRebaseMerge:    repoData.AllowRebaseMerge,
		AllowSquashMerge:    repoData.AllowSquashMerge,
		DeleteBranchOnMerge: repoData.DeleteBranchOnMerge,
		AllowUpdateBranch:   repoData.AllowUpdateBranch,
	}

	// Get topics
	if repoData.Topics != nil && len(*repoData.Topics) > 0 {
		cfg.Topics = *repoData.Topics
	}

	// Get labels
	labels, err := client.GetLabels(ctx)
	if err == nil && len(labels) > 0 {
		cfg.Labels = &config.LabelsConfig{
			ReplaceDefault: false,
			Items:          make([]config.Label, len(labels)),
		}
		for i, l := range labels {
			cfg.Labels.Items[i] = config.Label{
				Name:        l.Name,
				Color:       l.Color,
				Description: nullableStringVal(l.Description),
			}
		}
	}

	// Get secrets and variables if requested
	if exportIncludeSecrets {
		cfg.Env = &config.EnvConfig{}

		secrets, err := client.GetSecrets(ctx)
		if err == nil && len(secrets) > 0 {
			cfg.Env.Secrets = secrets
		}

		vars, err := client.GetVariables(ctx)
		if err == nil && len(vars) > 0 {
			cfg.Env.Variables = make(map[string]string)
			for _, v := range vars {
				cfg.Env.Variables[v.Name] = v.Value
			}
		}

		// Don't export empty env config
		if len(cfg.Env.Secrets) == 0 && len(cfg.Env.Variables) == 0 {
			cfg.Env = nil
		}
	}

	// Get actions permissions
	actionsPerms, err := client.GetActionsPermissions(ctx)
	if err == nil {
		enabled := bool(actionsPerms.Enabled)
		var allowedActions *string
		if actionsPerms.AllowedActions != nil {
			s := string(*actionsPerms.AllowedActions)
			allowedActions = &s
		}
		cfg.Actions = &config.ActionsConfig{
			Enabled:        &enabled,
			AllowedActions: allowedActions,
		}

		// Get selected actions if applicable
		if actionsPerms.AllowedActions != nil && *actionsPerms.AllowedActions == "selected" {
			selected, err := client.GetActionsSelectedActions(ctx)
			if err == nil {
				cfg.Actions.SelectedActions = &config.SelectedActionsConfig{
					GithubOwnedAllowed: selected.GithubOwnedAllowed,
					VerifiedAllowed:    selected.VerifiedAllowed,
				}
				if selected.PatternsAllowed != nil {
					cfg.Actions.SelectedActions.PatternsAllowed = *selected.PatternsAllowed
				}
			}
		}

		// Get workflow permissions
		workflowPerms, err := client.GetActionsWorkflowPermissions(ctx)
		if err == nil {
			perms := string(workflowPerms.DefaultWorkflowPermissions)
			cfg.Actions.DefaultWorkflowPermissions = &perms
			canApprove := bool(workflowPerms.CanApprovePullRequestReviews)
			cfg.Actions.CanApprovePullRequestReviews = &canApprove
		}
	}

	// Output
	if exportDir != "" {
		return exportToDirectory(cfg, exportDir)
	}

	if exportSingle != "" {
		return exportToSingleFile(cfg, exportSingle)
	}

	// Default: stdout
	yamlData, err := marshalYAML(cfg)
	if err != nil {
		return err
	}
	fmt.Print(string(yamlData))
	return nil
}

func exportToDirectory(cfg *config.Config, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Export repo settings
	if cfg.Repo != nil {
		if err := writeYAMLFile(filepath.Join(dir, "repo.yaml"), map[string]interface{}{"repo": cfg.Repo}); err != nil {
			return err
		}
	}

	// Export topics
	if len(cfg.Topics) > 0 {
		if err := writeYAMLFile(filepath.Join(dir, "topics.yaml"), map[string]interface{}{"topics": cfg.Topics}); err != nil {
			return err
		}
	}

	// Export labels
	if cfg.Labels != nil && len(cfg.Labels.Items) > 0 {
		if err := writeYAMLFile(filepath.Join(dir, "labels.yaml"), map[string]interface{}{"labels": cfg.Labels}); err != nil {
			return err
		}
	}

	// Export env (includes both variables and secrets)
	if cfg.Env != nil && (len(cfg.Env.Variables) > 0 || len(cfg.Env.Secrets) > 0) {
		if err := writeYAMLFile(filepath.Join(dir, "env.yaml"), map[string]interface{}{"env": cfg.Env}); err != nil {
			return err
		}
	}

	// Export actions
	if cfg.Actions != nil {
		if err := writeYAMLFile(filepath.Join(dir, "actions.yaml"), map[string]interface{}{"actions": cfg.Actions}); err != nil {
			return err
		}
	}

	logger.Success("Exported to %s/", dir)
	return nil
}

func exportToSingleFile(cfg *config.Config, path string) error {
	if err := writeYAMLFile(path, cfg); err != nil {
		return err
	}
	logger.Success("Exported to %s", path)
	return nil
}

func writeYAMLFile(path string, data interface{}) error {
	yamlData, err := marshalYAML(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, yamlData, 0o644)
}

// nullableToPtr converts a nullable.Nullable[string] to *string
func nullableToPtr(n nullable.Nullable[string]) *string {
	if !n.IsSpecified() || n.IsNull() {
		return nil
	}
	s := n.MustGet()
	return &s
}

// nullableStringVal returns the string value from a nullable.Nullable[string]
func nullableStringVal(n nullable.Nullable[string]) string {
	if !n.IsSpecified() || n.IsNull() {
		return ""
	}
	return n.MustGet()
}
