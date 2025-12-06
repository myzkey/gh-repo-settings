package diff

import (
	"context"
	"fmt"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/comparator"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// Calculator orchestrates the comparison of all repository settings
// It delegates to domain comparators and aggregates their results
type Calculator struct {
	client       github.GitHubClient
	config       *config.Config
	dotEnvValues *config.DotEnvValues
}

// NewCalculator creates a new diff calculator
func NewCalculator(client github.GitHubClient, cfg *config.Config) *Calculator {
	return &Calculator{
		client: client,
		config: cfg,
	}
}

// NewCalculatorWithEnv creates a new diff calculator with .env values
func NewCalculatorWithEnv(client github.GitHubClient, cfg *config.Config, dotEnv *config.DotEnvValues) *Calculator {
	return &Calculator{
		client:       client,
		config:       cfg,
		dotEnvValues: dotEnv,
	}
}

// CalculateOptions contains options for calculating diff
type CalculateOptions struct {
	CheckSecrets bool
	CheckEnv     bool
	SyncDelete   bool // If true, show variables/secrets to delete that are not in config
}

// Calculate calculates the diff with default options
func (c *Calculator) Calculate(ctx context.Context) (*model.Plan, error) {
	return c.CalculateWithOptions(ctx, CalculateOptions{})
}

// CalculateWithOptions calculates the diff with specified options
func (c *Calculator) CalculateWithOptions(ctx context.Context, opts CalculateOptions) (*model.Plan, error) {
	plan := model.NewPlan()

	// Compare repo settings
	if c.config.Repo != nil {
		repoComparator := comparator.NewRepoComparator(c.client, c.config.Repo)
		repoPlan, err := repoComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare repo settings: %w", err)
		}
		plan.AddAll(repoPlan.Changes())
	}

	// Compare topics
	if c.config.Topics != nil {
		topicsComparator := comparator.NewTopicsComparator(c.client, c.config.Topics)
		topicsPlan, err := topicsComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare topics: %w", err)
		}
		plan.AddAll(topicsPlan.Changes())
	}

	// Compare labels
	if c.config.Labels != nil {
		labelsComparator := comparator.NewLabelsComparator(c.client, c.config.Labels)
		labelsPlan, err := labelsComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare labels: %w", err)
		}
		plan.AddAll(labelsPlan.Changes())
	}

	// Compare branch protection
	if c.config.BranchProtection != nil {
		branchComparator := comparator.NewBranchProtectionComparatorWithClient(c.client, c.config.BranchProtection)
		branchPlan, err := branchComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare branch protection: %w", err)
		}
		plan.AddAll(branchPlan.Changes())
	}

	// Compare secrets and variables (if requested)
	if (opts.CheckSecrets || opts.CheckEnv) && c.config.Env != nil {
		envComparator := comparator.NewEnvComparator(c.client, c.config.Env, c.dotEnvValues, comparator.EnvComparatorOptions{
			CheckSecrets: opts.CheckSecrets,
			CheckVars:    opts.CheckEnv,
			SyncDelete:   opts.SyncDelete,
		})
		envPlan, err := envComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare env: %w", err)
		}
		plan.AddAll(envPlan.Changes())
	}

	// Compare actions permissions
	if c.config.Actions != nil {
		actionsComparator := comparator.NewActionsComparator(c.client, c.config.Actions)
		actionsPlan, err := actionsComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare actions permissions: %w", err)
		}
		plan.AddAll(actionsPlan.Changes())
	}

	// Compare pages settings
	if c.config.Pages != nil {
		pagesComparator := comparator.NewPagesComparator(c.client, c.config.Pages)
		pagesPlan, err := pagesComparator.Compare(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare pages settings: %w", err)
		}
		plan.AddAll(pagesPlan.Changes())
	}

	return plan, nil
}
