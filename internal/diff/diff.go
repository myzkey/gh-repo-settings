package diff

import (
	"context"
	"fmt"
	"reflect"

	"github.com/myzkey/gh-repo-settings/internal/config"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/github"
)

// Change represents a single configuration change
type Change struct {
	Type     ChangeType
	Category string
	Key      string
	Old      interface{}
	New      interface{}
}

// ChangeType represents the type of change
type ChangeType int

const (
	ChangeAdd ChangeType = iota
	ChangeUpdate
	ChangeDelete
	ChangeMissing // For secrets/env that are required but missing
)

func (c ChangeType) String() string {
	switch c {
	case ChangeAdd:
		return "add"
	case ChangeUpdate:
		return "update"
	case ChangeDelete:
		return "delete"
	case ChangeMissing:
		return "missing"
	default:
		return "unknown"
	}
}

// Plan represents the execution plan
type Plan struct {
	Changes []Change
}

// HasChanges returns true if there are any changes
func (p *Plan) HasChanges() bool {
	return len(p.Changes) > 0
}

// HasMissingSecrets returns true if there are missing secrets
func (p *Plan) HasMissingSecrets() bool {
	for _, c := range p.Changes {
		if c.Category == "secrets" && c.Type == ChangeMissing {
			return true
		}
	}
	return false
}

// HasMissingVariables returns true if there are missing variables
func (p *Plan) HasMissingVariables() bool {
	for _, c := range p.Changes {
		if c.Category == "env" && c.Type == ChangeMissing {
			return true
		}
	}
	return false
}

// Calculator calculates diff between config and current state
type Calculator struct {
	client github.GitHubClient
	config *config.Config
}

// NewCalculator creates a new diff calculator
func NewCalculator(client github.GitHubClient, cfg *config.Config) *Calculator {
	return &Calculator{
		client: client,
		config: cfg,
	}
}

// CalculateOptions contains options for calculating diff
type CalculateOptions struct {
	CheckSecrets bool
	CheckEnv     bool
}

// Calculate calculates the diff with default options
func (c *Calculator) Calculate(ctx context.Context) (*Plan, error) {
	return c.CalculateWithOptions(ctx, CalculateOptions{})
}

// CalculateWithOptions calculates the diff with specified options
func (c *Calculator) CalculateWithOptions(ctx context.Context, opts CalculateOptions) (*Plan, error) {
	plan := &Plan{}

	// Compare repo settings
	if c.config.Repo != nil {
		changes, err := c.compareRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare repo settings: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare topics
	if c.config.Topics != nil {
		changes, err := c.compareTopics(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare topics: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare labels
	if c.config.Labels != nil {
		changes, err := c.compareLabels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare labels: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare branch protection
	if c.config.BranchProtection != nil {
		changes, err := c.compareBranchProtection(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare branch protection: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Check secrets (if requested)
	if opts.CheckSecrets && c.config.Secrets != nil {
		changes, err := c.checkSecrets(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check secrets: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Check env variables (if requested)
	if opts.CheckEnv && c.config.Env != nil {
		changes, err := c.checkVariables(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check variables: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	return plan, nil
}

func (c *Calculator) compareRepo(ctx context.Context) ([]Change, error) {
	current, err := c.client.GetRepo(ctx)
	if err != nil {
		return nil, err
	}

	var changes []Change
	cfg := c.config.Repo

	if cfg.Description != nil && !ptrEqual(cfg.Description, current.Description) {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "description",
			Old:      ptrVal(current.Description),
			New:      ptrVal(cfg.Description),
		})
	}

	if cfg.Homepage != nil && !ptrEqual(cfg.Homepage, current.Homepage) {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "homepage",
			Old:      ptrVal(current.Homepage),
			New:      ptrVal(cfg.Homepage),
		})
	}

	if cfg.Visibility != nil && *cfg.Visibility != current.Visibility {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "visibility",
			Old:      current.Visibility,
			New:      *cfg.Visibility,
		})
	}

	if cfg.AllowMergeCommit != nil && *cfg.AllowMergeCommit != current.AllowMergeCommit {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "allow_merge_commit",
			Old:      current.AllowMergeCommit,
			New:      *cfg.AllowMergeCommit,
		})
	}

	if cfg.AllowRebaseMerge != nil && *cfg.AllowRebaseMerge != current.AllowRebaseMerge {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "allow_rebase_merge",
			Old:      current.AllowRebaseMerge,
			New:      *cfg.AllowRebaseMerge,
		})
	}

	if cfg.AllowSquashMerge != nil && *cfg.AllowSquashMerge != current.AllowSquashMerge {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "allow_squash_merge",
			Old:      current.AllowSquashMerge,
			New:      *cfg.AllowSquashMerge,
		})
	}

	if cfg.DeleteBranchOnMerge != nil && *cfg.DeleteBranchOnMerge != current.DeleteBranchOnMerge {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "delete_branch_on_merge",
			Old:      current.DeleteBranchOnMerge,
			New:      *cfg.DeleteBranchOnMerge,
		})
	}

	if cfg.AllowUpdateBranch != nil && *cfg.AllowUpdateBranch != current.AllowUpdateBranch {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "repo",
			Key:      "allow_update_branch",
			Old:      current.AllowUpdateBranch,
			New:      *cfg.AllowUpdateBranch,
		})
	}

	return changes, nil
}

func (c *Calculator) compareTopics(ctx context.Context) ([]Change, error) {
	current, err := c.client.GetRepo(ctx)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(c.config.Topics, current.Topics) {
		return []Change{{
			Type:     ChangeUpdate,
			Category: "topics",
			Key:      "topics",
			Old:      current.Topics,
			New:      c.config.Topics,
		}}, nil
	}

	return nil, nil
}

func (c *Calculator) compareLabels(ctx context.Context) ([]Change, error) {
	currentLabels, err := c.client.GetLabels(ctx)
	if err != nil {
		return nil, err
	}

	var changes []Change
	currentMap := make(map[string]github.LabelData)
	for _, l := range currentLabels {
		currentMap[l.Name] = l
	}

	configMap := make(map[string]config.Label)
	for _, l := range c.config.Labels.Items {
		configMap[l.Name] = l
	}

	// Check for additions and updates
	for _, cfgLabel := range c.config.Labels.Items {
		if current, exists := currentMap[cfgLabel.Name]; exists {
			// Check for updates
			if cfgLabel.Color != current.Color || cfgLabel.Description != current.Description {
				changes = append(changes, Change{
					Type:     ChangeUpdate,
					Category: "labels",
					Key:      cfgLabel.Name,
					Old:      fmt.Sprintf("color=%s, description=%s", current.Color, current.Description),
					New:      fmt.Sprintf("color=%s, description=%s", cfgLabel.Color, cfgLabel.Description),
				})
			}
		} else {
			// Addition
			changes = append(changes, Change{
				Type:     ChangeAdd,
				Category: "labels",
				Key:      cfgLabel.Name,
				New:      fmt.Sprintf("color=%s, description=%s", cfgLabel.Color, cfgLabel.Description),
			})
		}
	}

	// Check for deletions (only if replace_default is true)
	if c.config.Labels.ReplaceDefault {
		for _, currentLabel := range currentLabels {
			if _, exists := configMap[currentLabel.Name]; !exists {
				changes = append(changes, Change{
					Type:     ChangeDelete,
					Category: "labels",
					Key:      currentLabel.Name,
					Old:      fmt.Sprintf("color=%s, description=%s", currentLabel.Color, currentLabel.Description),
				})
			}
		}
	}

	return changes, nil
}

func (c *Calculator) compareBranchProtection(ctx context.Context) ([]Change, error) {
	var changes []Change

	for branchName, rule := range c.config.BranchProtection {
		current, err := c.client.GetBranchProtection(ctx, branchName)
		if err != nil {
			if apperrors.Is(err, apperrors.ErrBranchNotProtected) {
				// Branch protection doesn't exist, will be added
				changes = append(changes, Change{
					Type:     ChangeAdd,
					Category: "branch_protection",
					Key:      branchName,
					New:      formatBranchRule(rule),
				})
				continue
			}
			return nil, err
		}

		// Compare individual settings
		branchChanges := compareBranchRule(branchName, current, rule)
		changes = append(changes, branchChanges...)
	}

	return changes, nil
}

func compareBranchRule(branch string, current *github.BranchProtectionData, desired *config.BranchRule) []Change {
	var changes []Change
	prefix := fmt.Sprintf("%s.", branch)

	// Required reviews
	if desired.RequiredReviews != nil {
		currentReviews := 0
		if current.RequiredPullRequestReviews != nil {
			currentReviews = current.RequiredPullRequestReviews.RequiredApprovingReviewCount
		}
		if *desired.RequiredReviews != currentReviews {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "required_reviews",
				Old:      currentReviews,
				New:      *desired.RequiredReviews,
			})
		}
	}

	// Dismiss stale reviews
	if desired.DismissStaleReviews != nil {
		currentVal := false
		if current.RequiredPullRequestReviews != nil {
			currentVal = current.RequiredPullRequestReviews.DismissStaleReviews
		}
		if *desired.DismissStaleReviews != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "dismiss_stale_reviews",
				Old:      currentVal,
				New:      *desired.DismissStaleReviews,
			})
		}
	}

	// Require code owner
	if desired.RequireCodeOwner != nil {
		currentVal := false
		if current.RequiredPullRequestReviews != nil {
			currentVal = current.RequiredPullRequestReviews.RequireCodeOwnerReviews
		}
		if *desired.RequireCodeOwner != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "require_code_owner",
				Old:      currentVal,
				New:      *desired.RequireCodeOwner,
			})
		}
	}

	// Strict status checks
	if desired.StrictStatusChecks != nil {
		currentVal := false
		if current.RequiredStatusChecks != nil {
			currentVal = current.RequiredStatusChecks.Strict
		}
		if *desired.StrictStatusChecks != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "strict_status_checks",
				Old:      currentVal,
				New:      *desired.StrictStatusChecks,
			})
		}
	}

	// Status checks
	if desired.StatusChecks != nil {
		var currentChecks []string
		if current.RequiredStatusChecks != nil {
			currentChecks = current.RequiredStatusChecks.Contexts
		}
		if !reflect.DeepEqual(desired.StatusChecks, currentChecks) {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "status_checks",
				Old:      currentChecks,
				New:      desired.StatusChecks,
			})
		}
	}

	// Enforce admins
	if desired.EnforceAdmins != nil {
		currentVal := false
		if current.EnforceAdmins != nil {
			currentVal = current.EnforceAdmins.Enabled
		}
		if *desired.EnforceAdmins != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "enforce_admins",
				Old:      currentVal,
				New:      *desired.EnforceAdmins,
			})
		}
	}

	// Require linear history
	if desired.RequireLinearHistory != nil {
		currentVal := false
		if current.RequiredLinearHistory != nil {
			currentVal = current.RequiredLinearHistory.Enabled
		}
		if *desired.RequireLinearHistory != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "require_linear_history",
				Old:      currentVal,
				New:      *desired.RequireLinearHistory,
			})
		}
	}

	// Allow force pushes
	if desired.AllowForcePushes != nil {
		currentVal := false
		if current.AllowForcePushes != nil {
			currentVal = current.AllowForcePushes.Enabled
		}
		if *desired.AllowForcePushes != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "allow_force_pushes",
				Old:      currentVal,
				New:      *desired.AllowForcePushes,
			})
		}
	}

	// Allow deletions
	if desired.AllowDeletions != nil {
		currentVal := false
		if current.AllowDeletions != nil {
			currentVal = current.AllowDeletions.Enabled
		}
		if *desired.AllowDeletions != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "allow_deletions",
				Old:      currentVal,
				New:      *desired.AllowDeletions,
			})
		}
	}

	// Require signed commits
	if desired.RequireSignedCommits != nil {
		currentVal := false
		if current.RequiredSignatures != nil {
			currentVal = current.RequiredSignatures.Enabled
		}
		if *desired.RequireSignedCommits != currentVal {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "branch_protection",
				Key:      prefix + "require_signed_commits",
				Old:      currentVal,
				New:      *desired.RequireSignedCommits,
			})
		}
	}

	return changes
}

func formatBranchRule(rule *config.BranchRule) string {
	parts := []string{}
	if rule.RequiredReviews != nil {
		parts = append(parts, fmt.Sprintf("required_reviews=%d", *rule.RequiredReviews))
	}
	if rule.EnforceAdmins != nil && *rule.EnforceAdmins {
		parts = append(parts, "enforce_admins=true")
	}
	if rule.RequireLinearHistory != nil && *rule.RequireLinearHistory {
		parts = append(parts, "require_linear_history=true")
	}
	if len(parts) == 0 {
		return "new protection"
	}
	return fmt.Sprintf("{%s}", joinParts(parts))
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

func (c *Calculator) checkSecrets(ctx context.Context) ([]Change, error) {
	if c.config.Secrets == nil || len(c.config.Secrets.Required) == 0 {
		return nil, nil
	}

	currentSecrets, err := c.client.GetSecrets(ctx)
	if err != nil {
		return nil, err
	}

	secretSet := make(map[string]bool)
	for _, s := range currentSecrets {
		secretSet[s] = true
	}

	var changes []Change
	for _, required := range c.config.Secrets.Required {
		if !secretSet[required] {
			changes = append(changes, Change{
				Type:     ChangeMissing,
				Category: "secrets",
				Key:      required,
				New:      "required but not configured",
			})
		}
	}

	return changes, nil
}

func (c *Calculator) checkVariables(ctx context.Context) ([]Change, error) {
	if c.config.Env == nil || len(c.config.Env.Required) == 0 {
		return nil, nil
	}

	currentVars, err := c.client.GetVariables(ctx)
	if err != nil {
		return nil, err
	}

	varSet := make(map[string]bool)
	for _, v := range currentVars {
		varSet[v] = true
	}

	var changes []Change
	for _, required := range c.config.Env.Required {
		if !varSet[required] {
			changes = append(changes, Change{
				Type:     ChangeMissing,
				Category: "env",
				Key:      required,
				New:      "required but not configured",
			})
		}
	}

	return changes, nil
}

func ptrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
