package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
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
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
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
