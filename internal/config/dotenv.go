package config

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/myzkey/gh-repo-settings/internal/logger"
	"github.com/myzkey/gh-repo-settings/internal/provider"
)

// DotEnvValues holds parsed values from .github/.env file
type DotEnvValues struct {
	Values map[string]string
}

// LoadDotEnv loads and parses the .github/.env file
// Returns empty DotEnvValues if file doesn't exist (not an error)
func LoadDotEnv(configPath string) (*DotEnvValues, error) {
	// Determine .env path based on config path
	envPath := resolveDotEnvPath(configPath)

	values := &DotEnvValues{
		Values: make(map[string]string),
	}

	file, err := os.Open(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			// .env file doesn't exist, return empty values
			return values, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logger.Warn(".env:%d: skipping malformed line (missing '='): %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = unquote(value)

		values.Values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}

// resolveDotEnvPath determines the .env file path based on config path
func resolveDotEnvPath(configPath string) string {
	// If configPath is a directory, look for .env in that directory
	info, err := os.Stat(configPath)
	if err == nil && info.IsDir() {
		return filepath.Join(configPath, ".env")
	}

	// If configPath is a file, look for .env in .github directory
	dir := filepath.Dir(configPath)
	if filepath.Base(dir) == ".github" {
		return filepath.Join(dir, ".env")
	}

	// Default: look for .github/.env relative to current directory
	return filepath.Join(".github", ".env")
}

// unquote removes surrounding quotes from a string
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}

	// Check for double quotes
	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	// Check for single quotes
	if s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}

	return s
}

// GetVariable returns the value for a variable, checking .env first, then YAML default
func (d *DotEnvValues) GetVariable(name, yamlDefault string) string {
	if val, ok := d.Values[name]; ok {
		return val
	}
	return yamlDefault
}

// GetSecret returns the value for a secret from .env, or empty string if not found
func (d *DotEnvValues) GetSecret(name string) (string, bool) {
	val, ok := d.Values[name]
	return val, ok
}

// HasValue checks if a key exists in the .env file
func (d *DotEnvValues) HasValue(name string) bool {
	_, ok := d.Values[name]
	return ok
}

// Merge merges values from another DotEnvValues into this one
// Existing values are NOT overwritten
func (d *DotEnvValues) Merge(other *DotEnvValues) {
	if other == nil {
		return
	}
	for k, v := range other.Values {
		if _, exists := d.Values[k]; !exists {
			d.Values[k] = v
		}
	}
}

// ProviderResult holds the result of loading from a provider
type ProviderResult struct {
	Values      map[string]string
	WrittenFile bool
}

// LoadFromProvider loads secrets from an external provider
// If output is "file", writes to .env file. If "memory", returns values for in-memory merge.
// If keys is empty, all keys from the provider will be loaded.
func LoadFromProvider(ctx context.Context, cfg *ProviderConfig, keys []string, configPath string) (*ProviderResult, error) {
	if cfg == nil {
		return &ProviderResult{Values: make(map[string]string)}, nil
	}

	logger.Info("Loading secrets from provider: %s", cfg.Name)

	providerCfg := &provider.Config{
		Name:   cfg.Name,
		Secret: cfg.Secret,
		Region: cfg.Region,
	}

	values, err := provider.LoadSecrets(ctx, providerCfg, keys)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return &ProviderResult{Values: make(map[string]string)}, nil
	}

	result := &ProviderResult{Values: values}

	// Determine output mode (default: file)
	outputMode := cfg.Output
	if outputMode == "" {
		outputMode = "file"
	}

	if outputMode == "file" {
		// Write to .env file
		envPath := resolveDotEnvPath(configPath)
		if err := writeToEnvFile(envPath, values); err != nil {
			return nil, err
		}
		logger.Info("Wrote %d secrets to %s", len(values), envPath)
		result.WrittenFile = true
	} else {
		logger.Info("Loaded %d secrets into memory", len(values))
	}

	return result, nil
}

// writeToEnvFile writes or updates values in .env file
// Existing values are preserved, new values are appended
func writeToEnvFile(envPath string, values map[string]string) error {
	// Read existing content
	existing := make(map[string]string)
	var lines []string

	file, err := os.Open(envPath)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)

			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				existing[strings.TrimSpace(parts[0])] = ""
			}
		}
		_ = file.Close()
	}

	// Append new values that don't exist
	var newLines []string
	for key, value := range values {
		if _, exists := existing[key]; !exists {
			// Quote value if it contains special characters
			quotedValue := value
			if strings.ContainsAny(value, " \t\n\"'") {
				quotedValue = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
			}
			newLines = append(newLines, key+"="+quotedValue)
		}
	}

	if len(newLines) == 0 {
		return nil // Nothing new to add
	}

	// Ensure directory exists
	dir := filepath.Dir(envPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write file
	output := strings.Join(lines, "\n")
	if len(lines) > 0 && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	output += "# Added by provider: " + time.Now().Format("2006-01-02 15:04:05") + "\n"
	output += strings.Join(newLines, "\n") + "\n"

	return os.WriteFile(envPath, []byte(output), 0o600)
}
