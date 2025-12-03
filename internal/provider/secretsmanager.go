package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/myzkey/gh-repo-settings/internal/logger"
)

// SecretsManagerProvider fetches secrets from AWS Secrets Manager
type SecretsManagerProvider struct {
	secret string
	region string
}

// NewSecretsManagerProvider creates a new Secrets Manager provider
func NewSecretsManagerProvider(cfg *Config) (*SecretsManagerProvider, error) {
	if cfg.Secret == "" {
		return nil, fmt.Errorf("secret is required for secretsmanager provider")
	}
	return &SecretsManagerProvider{
		secret: cfg.Secret,
		region: cfg.Region,
	}, nil
}

// Name returns the provider name
func (p *SecretsManagerProvider) Name() string {
	return "secretsmanager"
}

// Load fetches secrets from AWS Secrets Manager
// If keys is empty, returns all keys from the secret JSON
// If keys is specified, returns only those keys
func (p *SecretsManagerProvider) Load(ctx context.Context, keys []string) (map[string]string, error) {
	// Check if AWS CLI is available
	if _, err := exec.LookPath("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	// Fetch the secret value
	rawValue, err := p.getSecretValue(ctx, p.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret %s: %w", p.secret, err)
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawValue), &data); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	result := make(map[string]string)

	// If no keys specified, return all keys
	if len(keys) == 0 {
		for k, v := range data {
			strVal, err := toString(v)
			if err != nil {
				logger.Warn("Skipping key %s: %v", k, err)
				continue
			}
			result[k] = strVal
		}
		logger.Debug("Loaded all %d keys from secret: %s", len(result), p.secret)
		return result, nil
	}

	// Return only specified keys
	var errors []string
	for _, key := range keys {
		v, ok := data[key]
		if !ok {
			logger.Warn("Key %s not found in secret %s", key, p.secret)
			errors = append(errors, fmt.Sprintf("%s: not found", key))
			continue
		}
		strVal, err := toString(v)
		if err != nil {
			logger.Warn("Failed to convert key %s: %v", key, err)
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
			continue
		}
		result[key] = strVal
	}

	if len(errors) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("failed to load secrets:\n  %s", strings.Join(errors, "\n  "))
	}

	logger.Debug("Loaded %d keys from secret: %s", len(result), p.secret)
	return result, nil
}

// getSecretValue fetches a secret from Secrets Manager
func (p *SecretsManagerProvider) getSecretValue(ctx context.Context, secretName string) (string, error) {
	args := []string{
		"secretsmanager", "get-secret-value",
		"--secret-id", secretName,
		"--query", "SecretString",
		"--output", "text",
	}

	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// toString converts an interface{} to string
func toString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		return fmt.Sprintf("%v", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	default:
		// For complex types, marshal to JSON
		b, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("failed to convert to string: %w", err)
		}
		return string(b), nil
	}
}
