// Package provider handles loading secrets from external secret managers
package provider

import (
	"context"
	"fmt"
)

// Provider defines the interface for secret providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Load fetches secrets from the provider and returns them as key-value pairs
	// The keys parameter specifies which secrets to fetch
	Load(ctx context.Context, keys []string) (map[string]string, error)
}

// Config represents the configuration for a secret provider
type Config struct {
	// Name is the provider name (e.g., "secretsmanager")
	Name string

	// Secret is the secret name/path in AWS Secrets Manager
	Secret string

	// Region is the AWS region
	Region string
}

// New creates a new provider based on configuration
func New(cfg *Config) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider config is nil")
	}

	switch cfg.Name {
	case "secretsmanager":
		return NewSecretsManagerProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Name)
	}
}

// LoadSecrets loads secrets using the configured provider
func LoadSecrets(ctx context.Context, cfg *Config, keys []string) (map[string]string, error) {
	provider, err := New(cfg)
	if err != nil {
		return nil, err
	}

	return provider.Load(ctx, keys)
}
