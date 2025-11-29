package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDir        = ".github/repo-settings"
	DefaultSingleFile = ".github/repo-settings.yaml"
)

// LoadOptions represents options for loading config
type LoadOptions struct {
	Dir    string
	Config string
}

// Load loads configuration from file or directory
func Load(opts LoadOptions) (*Config, error) {
	// Priority: --dir > --config > default dir > default single file
	if opts.Dir != "" {
		return loadFromDirectory(opts.Dir)
	}

	if opts.Config != "" {
		return loadSingleFile(opts.Config)
	}

	// Check default directory
	if info, err := os.Stat(DefaultDir); err == nil && info.IsDir() {
		return loadFromDirectory(DefaultDir)
	}

	// Check default single file
	if _, err := os.Stat(DefaultSingleFile); err == nil {
		return loadSingleFile(DefaultSingleFile)
	}

	return nil, fmt.Errorf("no config found. Create %s/ or %s", DefaultDir, DefaultSingleFile)
}

func loadSingleFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	return &config, nil
}

func loadFromDirectory(dirPath string) (*Config, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory %s: %w", dirPath, err)
	}

	config := &Config{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		filePath := filepath.Join(dirPath, name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		baseName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")

		switch baseName {
		case "repo":
			var wrapper struct {
				Repo *RepoConfig `yaml:"repo"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Repo != nil {
				config.Repo = wrapper.Repo
			} else {
				var repo RepoConfig
				if err := yaml.Unmarshal(data, &repo); err == nil {
					config.Repo = &repo
				}
			}
		case "topics":
			var wrapper struct {
				Topics []string `yaml:"topics"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Topics != nil {
				config.Topics = wrapper.Topics
			} else {
				var topics []string
				if err := yaml.Unmarshal(data, &topics); err == nil {
					config.Topics = topics
				}
			}
		case "labels":
			var wrapper struct {
				Labels *LabelsConfig `yaml:"labels"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Labels != nil {
				config.Labels = wrapper.Labels
			} else {
				var labels LabelsConfig
				if err := yaml.Unmarshal(data, &labels); err == nil {
					config.Labels = &labels
				}
			}
		case "branch-protection", "branch_protection":
			var wrapper struct {
				BranchProtection map[string]*BranchRule `yaml:"branch_protection"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.BranchProtection != nil {
				config.BranchProtection = wrapper.BranchProtection
			} else {
				var bp map[string]*BranchRule
				if err := yaml.Unmarshal(data, &bp); err == nil {
					config.BranchProtection = bp
				}
			}
		case "secrets":
			var wrapper struct {
				Secrets *SecretsConfig `yaml:"secrets"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Secrets != nil {
				config.Secrets = wrapper.Secrets
			} else {
				var secrets SecretsConfig
				if err := yaml.Unmarshal(data, &secrets); err == nil {
					config.Secrets = &secrets
				}
			}
		case "env":
			var wrapper struct {
				Env *EnvConfig `yaml:"env"`
			}
			if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Env != nil {
				config.Env = wrapper.Env
			} else {
				var env EnvConfig
				if err := yaml.Unmarshal(data, &env); err == nil {
					config.Env = &env
				}
			}
		default:
			// For any other file, try to merge at top level
			var partial Config
			if err := yaml.Unmarshal(data, &partial); err == nil {
				mergeConfig(config, &partial)
			}
		}
	}

	return config, nil
}

func mergeConfig(dst, src *Config) {
	if src.Repo != nil {
		dst.Repo = src.Repo
	}
	if src.Topics != nil {
		dst.Topics = src.Topics
	}
	if src.Labels != nil {
		dst.Labels = src.Labels
	}
	if src.BranchProtection != nil {
		dst.BranchProtection = src.BranchProtection
	}
	if src.Secrets != nil {
		dst.Secrets = src.Secrets
	}
	if src.Env != nil {
		dst.Env = src.Env
	}
}

// ToYAML converts config to YAML string
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
