package comparator

import (
	"context"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// EnvComparatorOptions contains options for comparing environment settings
type EnvComparatorOptions struct {
	CheckSecrets bool
	CheckVars    bool
	SyncDelete   bool
}

// EnvComparator compares environment variables and secrets
type EnvComparator struct {
	client       github.GitHubClient
	config       *config.EnvConfig
	dotEnvValues *config.DotEnvValues
	options      EnvComparatorOptions
}

// NewEnvComparator creates a new EnvComparator
func NewEnvComparator(client github.GitHubClient, cfg *config.EnvConfig, dotEnv *config.DotEnvValues, opts EnvComparatorOptions) *EnvComparator {
	return &EnvComparator{
		client:       client,
		config:       cfg,
		dotEnvValues: dotEnv,
		options:      opts,
	}
}

// Compare compares the current environment with the desired configuration
func (c *EnvComparator) Compare(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	if c.options.CheckSecrets {
		secretsPlan, err := c.compareSecrets(ctx)
		if err != nil {
			return nil, err
		}
		plan.AddAll(secretsPlan.Changes())
	}

	if c.options.CheckVars {
		varsPlan, err := c.compareVariables(ctx)
		if err != nil {
			return nil, err
		}
		plan.AddAll(varsPlan.Changes())
	}

	return plan, nil
}

func (c *EnvComparator) compareSecrets(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	currentSecrets, err := c.client.GetSecrets(ctx)
	if err != nil {
		return nil, err
	}

	secretSet := model.ToStringSet(currentSecrets)

	// Check for secrets that need to be added
	for _, secretName := range c.config.Secrets {
		if !secretSet[secretName] {
			// Check if value exists in .env
			hasValue := false
			if c.dotEnvValues != nil {
				_, hasValue = c.dotEnvValues.GetSecret(secretName)
			}
			if hasValue {
				plan.Add(model.NewAddChange(
					model.CategorySecrets,
					secretName,
					"(will be set from .env)",
				))
			} else {
				plan.Add(model.NewMissingChange(
					model.CategorySecrets,
					secretName,
					"not in .github/.env (will prompt)",
				))
			}
		}
	}

	// Check for secrets to delete (if syncDelete)
	if c.options.SyncDelete {
		configSecretSet := model.ToStringSet(c.config.Secrets)
		for _, s := range currentSecrets {
			if !configSecretSet[s] {
				plan.Add(model.NewDeleteChange(
					model.CategorySecrets,
					s,
					nil,
				))
			}
		}
	}

	return plan, nil
}

func (c *EnvComparator) compareVariables(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	currentVars, err := c.client.GetVariables(ctx)
	if err != nil {
		return nil, err
	}

	currentVarMap := make(map[string]string)
	for _, v := range currentVars {
		currentVarMap[v.Name] = v.Value
	}

	// Check for variables that need to be added or updated
	for name, defaultValue := range c.config.Variables {
		// Get final value (.env overrides YAML default)
		finalValue := defaultValue
		if c.dotEnvValues != nil {
			finalValue = c.dotEnvValues.GetVariable(name, defaultValue)
		}

		currentValue, exists := currentVarMap[name]
		if !exists {
			plan.Add(model.NewAddChange(
				model.CategoryVariables,
				name,
				finalValue,
			))
		} else if currentValue != finalValue {
			plan.Add(model.NewUpdateChange(
				model.CategoryVariables,
				name,
				currentValue,
				finalValue,
			))
		}
	}

	// Check for variables to delete (if syncDelete)
	if c.options.SyncDelete {
		for _, v := range currentVars {
			if _, exists := c.config.Variables[v.Name]; !exists {
				plan.Add(model.NewDeleteChange(
					model.CategoryVariables,
					v.Name,
					v.Value,
				))
			}
		}
	}

	return plan, nil
}
