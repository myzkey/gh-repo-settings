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

// toStringSet converts a slice of strings to a set (map[string]bool)
func toStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

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

// HasDeletes returns true if there are any delete changes
func (p *Plan) HasDeletes() bool {
	for _, c := range p.Changes {
		if c.Type == ChangeDelete {
			return true
		}
	}
	return false
}

// Calculator calculates diff between config and current state
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

	// Compare secrets and variables (if requested)
	if (opts.CheckSecrets || opts.CheckEnv) && c.config.Env != nil {
		changes, err := c.compareEnv(ctx, opts.CheckSecrets, opts.CheckEnv, opts.SyncDelete)
		if err != nil {
			return nil, fmt.Errorf("failed to compare env: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare actions permissions
	if c.config.Actions != nil {
		changes, err := c.compareActions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare actions permissions: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare pages settings
	if c.config.Pages != nil {
		changes, err := c.comparePages(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to compare pages settings: %w", err)
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

	if !stringSliceEqualIgnoreOrder(c.config.Topics, current.Topics) {
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

// stringSliceEqualIgnoreOrder compares two string slices ignoring order
func stringSliceEqualIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]int)
	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		aMap[v]--
		if aMap[v] < 0 {
			return false
		}
	}
	return true
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
	if rule.DismissStaleReviews != nil && *rule.DismissStaleReviews {
		parts = append(parts, "dismiss_stale_reviews=true")
	}
	if rule.RequireCodeOwner != nil && *rule.RequireCodeOwner {
		parts = append(parts, "require_code_owner=true")
	}
	if rule.StrictStatusChecks != nil && *rule.StrictStatusChecks {
		parts = append(parts, "strict_status_checks=true")
	}
	if len(rule.StatusChecks) > 0 {
		parts = append(parts, fmt.Sprintf("status_checks=%v", rule.StatusChecks))
	}
	if rule.EnforceAdmins != nil && *rule.EnforceAdmins {
		parts = append(parts, "enforce_admins=true")
	}
	if rule.RequireLinearHistory != nil && *rule.RequireLinearHistory {
		parts = append(parts, "require_linear_history=true")
	}
	if rule.RequireSignedCommits != nil && *rule.RequireSignedCommits {
		parts = append(parts, "require_signed_commits=true")
	}
	if rule.AllowForcePushes != nil && *rule.AllowForcePushes {
		parts = append(parts, "allow_force_pushes=true")
	}
	if rule.AllowDeletions != nil && *rule.AllowDeletions {
		parts = append(parts, "allow_deletions=true")
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

func (c *Calculator) compareEnv(ctx context.Context, checkSecrets, checkVars, syncDelete bool) ([]Change, error) {
	var changes []Change

	// Compare secrets
	if checkSecrets {
		currentSecrets, err := c.client.GetSecrets(ctx)
		if err != nil {
			return nil, err
		}

		secretSet := toStringSet(currentSecrets)

		// Check for secrets that need to be added
		for _, secretName := range c.config.Env.Secrets {
			if !secretSet[secretName] {
				// Check if value exists in .env
				hasValue := false
				if c.dotEnvValues != nil {
					_, hasValue = c.dotEnvValues.GetSecret(secretName)
				}
				if hasValue {
					changes = append(changes, Change{
						Type:     ChangeAdd,
						Category: "secrets",
						Key:      secretName,
						New:      "(will be set from .env)",
					})
				} else {
					changes = append(changes, Change{
						Type:     ChangeMissing,
						Category: "secrets",
						Key:      secretName,
						New:      "not in .github/.env (will prompt)",
					})
				}
			}
		}

		// Check for secrets to delete (if syncDelete)
		if syncDelete {
			configSecretSet := toStringSet(c.config.Env.Secrets)
			for _, s := range currentSecrets {
				if !configSecretSet[s] {
					changes = append(changes, Change{
						Type:     ChangeDelete,
						Category: "secrets",
						Key:      s,
					})
				}
			}
		}
	}

	// Check variables
	if checkVars {
		currentVars, err := c.client.GetVariables(ctx)
		if err != nil {
			return nil, err
		}

		currentVarMap := make(map[string]string)
		for _, v := range currentVars {
			currentVarMap[v.Name] = v.Value
		}

		// Check for variables that need to be added or updated
		for name, defaultValue := range c.config.Env.Variables {
			// Get final value (.env overrides YAML default)
			finalValue := defaultValue
			if c.dotEnvValues != nil {
				finalValue = c.dotEnvValues.GetVariable(name, defaultValue)
			}

			currentValue, exists := currentVarMap[name]
			if !exists {
				changes = append(changes, Change{
					Type:     ChangeAdd,
					Category: "variables",
					Key:      name,
					New:      finalValue,
				})
			} else if currentValue != finalValue {
				changes = append(changes, Change{
					Type:     ChangeUpdate,
					Category: "variables",
					Key:      name,
					Old:      currentValue,
					New:      finalValue,
				})
			}
		}

		// Check for variables to delete (if syncDelete)
		if syncDelete {
			for _, v := range currentVars {
				if _, exists := c.config.Env.Variables[v.Name]; !exists {
					changes = append(changes, Change{
						Type:     ChangeDelete,
						Category: "variables",
						Key:      v.Name,
						Old:      v.Value,
					})
				}
			}
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

func (c *Calculator) compareActions(ctx context.Context) ([]Change, error) {
	var changes []Change
	cfg := c.config.Actions

	// Get current permissions
	currentPerms, err := c.client.GetActionsPermissions(ctx)
	if err != nil {
		return nil, err
	}

	// Compare enabled
	if cfg.Enabled != nil && *cfg.Enabled != currentPerms.Enabled {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "actions",
			Key:      "enabled",
			Old:      currentPerms.Enabled,
			New:      *cfg.Enabled,
		})
	}

	// Compare allowed_actions
	if cfg.AllowedActions != nil && *cfg.AllowedActions != currentPerms.AllowedActions {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "actions",
			Key:      "allowed_actions",
			Old:      currentPerms.AllowedActions,
			New:      *cfg.AllowedActions,
		})
	}

	// Compare selected actions (only if allowed_actions is "selected")
	if cfg.SelectedActions != nil {
		currentSelected, err := c.client.GetActionsSelectedActions(ctx)
		if err != nil {
			// Ignore error if not applicable
			currentSelected = &github.ActionsSelectedData{}
		}

		if cfg.SelectedActions.GithubOwnedAllowed != nil && *cfg.SelectedActions.GithubOwnedAllowed != currentSelected.GithubOwnedAllowed {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "actions",
				Key:      "github_owned_allowed",
				Old:      currentSelected.GithubOwnedAllowed,
				New:      *cfg.SelectedActions.GithubOwnedAllowed,
			})
		}

		if cfg.SelectedActions.VerifiedAllowed != nil && *cfg.SelectedActions.VerifiedAllowed != currentSelected.VerifiedAllowed {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "actions",
				Key:      "verified_allowed",
				Old:      currentSelected.VerifiedAllowed,
				New:      *cfg.SelectedActions.VerifiedAllowed,
			})
		}

		if len(cfg.SelectedActions.PatternsAllowed) > 0 && !reflect.DeepEqual(cfg.SelectedActions.PatternsAllowed, currentSelected.PatternsAllowed) {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "actions",
				Key:      "patterns_allowed",
				Old:      currentSelected.PatternsAllowed,
				New:      cfg.SelectedActions.PatternsAllowed,
			})
		}
	}

	// Compare workflow permissions
	currentWorkflow, err := c.client.GetActionsWorkflowPermissions(ctx)
	if err != nil {
		return nil, err
	}

	if cfg.DefaultWorkflowPermissions != nil && *cfg.DefaultWorkflowPermissions != currentWorkflow.DefaultWorkflowPermissions {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "actions",
			Key:      "default_workflow_permissions",
			Old:      currentWorkflow.DefaultWorkflowPermissions,
			New:      *cfg.DefaultWorkflowPermissions,
		})
	}

	if cfg.CanApprovePullRequestReviews != nil && *cfg.CanApprovePullRequestReviews != currentWorkflow.CanApprovePullRequestReviews {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "actions",
			Key:      "can_approve_pull_request_reviews",
			Old:      currentWorkflow.CanApprovePullRequestReviews,
			New:      *cfg.CanApprovePullRequestReviews,
		})
	}

	return changes, nil
}

func (c *Calculator) comparePages(ctx context.Context) ([]Change, error) {
	var changes []Change
	cfg := c.config.Pages

	current, err := c.client.GetPages(ctx)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrPagesNotEnabled) {
			// Pages not enabled, will be created
			buildType := "workflow"
			if cfg.BuildType != nil {
				buildType = *cfg.BuildType
			}
			changes = append(changes, Change{
				Type:     ChangeAdd,
				Category: "pages",
				Key:      "pages",
				New:      fmt.Sprintf("build_type=%s", buildType),
			})
			return changes, nil
		}
		return nil, err
	}

	// Compare build_type
	if cfg.BuildType != nil && *cfg.BuildType != current.BuildType {
		changes = append(changes, Change{
			Type:     ChangeUpdate,
			Category: "pages",
			Key:      "build_type",
			Old:      current.BuildType,
			New:      *cfg.BuildType,
		})
	}

	// Compare source (only for legacy build type)
	if cfg.Source != nil && current.Source != nil {
		if cfg.Source.Branch != nil && *cfg.Source.Branch != current.Source.Branch {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "pages",
				Key:      "source.branch",
				Old:      current.Source.Branch,
				New:      *cfg.Source.Branch,
			})
		}
		if cfg.Source.Path != nil && *cfg.Source.Path != current.Source.Path {
			changes = append(changes, Change{
				Type:     ChangeUpdate,
				Category: "pages",
				Key:      "source.path",
				Old:      current.Source.Path,
				New:      *cfg.Source.Path,
			})
		}
	}

	return changes, nil
}
