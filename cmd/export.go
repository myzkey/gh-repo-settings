package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	exportDir           string
	exportSingle        string
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
	client, err := github.NewClient(repo)
	if err != nil {
		return err
	}

	fmt.Printf("Exporting settings from %s/%s...\n", client.Repo.Owner, client.Repo.Name)

	cfg := &config.Config{}

	// Get repo settings
	repoData, err := client.GetRepo()
	if err != nil {
		return fmt.Errorf("failed to get repo settings: %w", err)
	}

	cfg.Repo = &config.RepoConfig{
		Description:         repoData.Description,
		Homepage:            repoData.Homepage,
		Visibility:          &repoData.Visibility,
		AllowMergeCommit:    &repoData.AllowMergeCommit,
		AllowRebaseMerge:    &repoData.AllowRebaseMerge,
		AllowSquashMerge:    &repoData.AllowSquashMerge,
		DeleteBranchOnMerge: &repoData.DeleteBranchOnMerge,
		AllowUpdateBranch:   &repoData.AllowUpdateBranch,
	}

	// Get topics
	if len(repoData.Topics) > 0 {
		cfg.Topics = repoData.Topics
	}

	// Get labels
	labels, err := client.GetLabels()
	if err == nil && len(labels) > 0 {
		cfg.Labels = &config.LabelsConfig{
			ReplaceDefault: false,
			Items:          make([]config.Label, len(labels)),
		}
		for i, l := range labels {
			cfg.Labels.Items[i] = config.Label{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
			}
		}
	}

	// Get secrets if requested
	if exportIncludeSecrets {
		secrets, err := client.GetSecrets()
		if err == nil && len(secrets) > 0 {
			cfg.Secrets = &config.SecretsConfig{
				Required: secrets,
			}
		}

		vars, err := client.GetVariables()
		if err == nil && len(vars) > 0 {
			cfg.Env = &config.EnvConfig{
				Required: vars,
			}
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
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlData))
	return nil
}

func exportToDirectory(cfg *config.Config, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
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

	// Export secrets
	if cfg.Secrets != nil && len(cfg.Secrets.Required) > 0 {
		if err := writeYAMLFile(filepath.Join(dir, "secrets.yaml"), map[string]interface{}{"secrets": cfg.Secrets}); err != nil {
			return err
		}
	}

	// Export env
	if cfg.Env != nil && len(cfg.Env.Required) > 0 {
		if err := writeYAMLFile(filepath.Join(dir, "env.yaml"), map[string]interface{}{"env": cfg.Env}); err != nil {
			return err
		}
	}

	fmt.Printf("Exported to %s/\n", dir)
	return nil
}

func exportToSingleFile(cfg *config.Config, path string) error {
	if err := writeYAMLFile(path, cfg); err != nil {
		return err
	}
	fmt.Printf("Exported to %s\n", path)
	return nil
}

func writeYAMLFile(path string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, yamlData, 0644)
}
