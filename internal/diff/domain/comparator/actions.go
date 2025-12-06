package comparator

import (
	"context"
	"reflect"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// ActionsComparator compares GitHub Actions permissions
type ActionsComparator struct {
	client github.GitHubClient
	config *config.ActionsConfig
}

// NewActionsComparator creates a new ActionsComparator
func NewActionsComparator(client github.GitHubClient, cfg *config.ActionsConfig) *ActionsComparator {
	return &ActionsComparator{
		client: client,
		config: cfg,
	}
}

// Compare compares the current actions settings with the desired configuration
func (c *ActionsComparator) Compare(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	// Compare permissions
	permsPlan, err := c.comparePermissions(ctx)
	if err != nil {
		return nil, err
	}
	plan.AddAll(permsPlan.Changes())

	// Compare selected actions
	if c.config.SelectedActions != nil {
		selectedPlan, err := c.compareSelectedActions(ctx)
		if err != nil {
			return nil, err
		}
		plan.AddAll(selectedPlan.Changes())
	}

	// Compare workflow permissions
	workflowPlan, err := c.compareWorkflowPermissions(ctx)
	if err != nil {
		return nil, err
	}
	plan.AddAll(workflowPlan.Changes())

	return plan, nil
}

func (c *ActionsComparator) comparePermissions(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	currentPerms, err := c.client.GetActionsPermissions(ctx)
	if err != nil {
		return nil, err
	}

	// Compare enabled
	if c.config.Enabled != nil && *c.config.Enabled != currentPerms.Enabled {
		plan.Add(model.NewUpdateChange(
			model.CategoryActions,
			"enabled",
			currentPerms.Enabled,
			*c.config.Enabled,
		))
	}

	// Compare allowed_actions
	if c.config.AllowedActions != nil {
		currentAllowed := ""
		if currentPerms.AllowedActions != nil {
			currentAllowed = string(*currentPerms.AllowedActions)
		}
		if *c.config.AllowedActions != currentAllowed {
			plan.Add(model.NewUpdateChange(
				model.CategoryActions,
				"allowed_actions",
				currentAllowed,
				*c.config.AllowedActions,
			))
		}
	}

	return plan, nil
}

func (c *ActionsComparator) compareSelectedActions(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	currentSelected, err := c.client.GetActionsSelectedActions(ctx)
	if err != nil {
		// Ignore error if not applicable
		currentSelected = &github.ActionsSelectedData{}
	}

	cfg := c.config.SelectedActions

	if cfg.GithubOwnedAllowed != nil && !model.PtrBoolEqual(cfg.GithubOwnedAllowed, currentSelected.GithubOwnedAllowed) {
		plan.Add(model.NewUpdateChange(
			model.CategoryActions,
			"github_owned_allowed",
			model.PtrBoolVal(currentSelected.GithubOwnedAllowed),
			*cfg.GithubOwnedAllowed,
		))
	}

	if cfg.VerifiedAllowed != nil && !model.PtrBoolEqual(cfg.VerifiedAllowed, currentSelected.VerifiedAllowed) {
		plan.Add(model.NewUpdateChange(
			model.CategoryActions,
			"verified_allowed",
			model.PtrBoolVal(currentSelected.VerifiedAllowed),
			*cfg.VerifiedAllowed,
		))
	}

	if len(cfg.PatternsAllowed) > 0 {
		var currentPatterns []string
		if currentSelected.PatternsAllowed != nil {
			currentPatterns = *currentSelected.PatternsAllowed
		}
		if !reflect.DeepEqual(cfg.PatternsAllowed, currentPatterns) {
			plan.Add(model.NewUpdateChange(
				model.CategoryActions,
				"patterns_allowed",
				currentPatterns,
				cfg.PatternsAllowed,
			))
		}
	}

	return plan, nil
}

func (c *ActionsComparator) compareWorkflowPermissions(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	currentWorkflow, err := c.client.GetActionsWorkflowPermissions(ctx)
	if err != nil {
		return nil, err
	}

	if c.config.DefaultWorkflowPermissions != nil && *c.config.DefaultWorkflowPermissions != string(currentWorkflow.DefaultWorkflowPermissions) {
		plan.Add(model.NewUpdateChange(
			model.CategoryActions,
			"default_workflow_permissions",
			string(currentWorkflow.DefaultWorkflowPermissions),
			*c.config.DefaultWorkflowPermissions,
		))
	}

	if c.config.CanApprovePullRequestReviews != nil && *c.config.CanApprovePullRequestReviews != currentWorkflow.CanApprovePullRequestReviews {
		plan.Add(model.NewUpdateChange(
			model.CategoryActions,
			"can_approve_pull_request_reviews",
			currentWorkflow.CanApprovePullRequestReviews,
			*c.config.CanApprovePullRequestReviews,
		))
	}

	return plan, nil
}
