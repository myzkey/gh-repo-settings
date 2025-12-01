package config

import (
	"fmt"
	"regexp"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// GitHub Actions variable/secret name pattern
// Must start with a letter or underscore, contain only alphanumeric characters and underscores
// Cannot start with GITHUB_ prefix (reserved)
var envNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	if c.Env != nil {
		if err := c.Env.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates the EnvConfig
func (e *EnvConfig) Validate() error {
	// Validate variable names
	for name := range e.Variables {
		if err := validateEnvName(name, "variable"); err != nil {
			return err
		}
	}

	// Validate secret names
	for _, name := range e.Secrets {
		if err := validateEnvName(name, "secret"); err != nil {
			return err
		}
	}

	return nil
}

// validateEnvName validates a variable or secret name
func validateEnvName(name, kind string) error {
	if name == "" {
		return apperrors.NewValidationError(kind, "name cannot be empty")
	}

	if !envNameRegex.MatchString(name) {
		return apperrors.NewValidationError(
			kind,
			fmt.Sprintf("invalid name %q: must start with a letter or underscore, and contain only alphanumeric characters and underscores", name),
		)
	}

	// Check for reserved GITHUB_ prefix
	if len(name) >= 7 && (name[:7] == "GITHUB_" || name[:7] == "github_") {
		return apperrors.NewValidationError(
			kind,
			fmt.Sprintf("invalid name %q: names starting with GITHUB_ are reserved", name),
		)
	}

	return nil
}
