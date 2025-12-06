package comparator

import (
	"context"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// RepoComparator compares repository settings
type RepoComparator struct {
	client github.GitHubClient
	config *config.RepoConfig
}

// NewRepoComparator creates a new RepoComparator
func NewRepoComparator(client github.GitHubClient, cfg *config.RepoConfig) *RepoComparator {
	return &RepoComparator{
		client: client,
		config: cfg,
	}
}

// Compare compares the current repo settings with the desired configuration
func (c *RepoComparator) Compare(ctx context.Context) (*model.Plan, error) {
	current, err := c.client.GetRepo(ctx)
	if err != nil {
		return nil, err
	}

	plan := model.NewPlan()
	cfg := c.config

	if cfg.Description != nil && !model.NullableStringEqual(cfg.Description, current.Description) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"description",
			model.NullableStringVal(current.Description),
			model.PtrVal(cfg.Description),
		))
	}

	if cfg.Homepage != nil && !model.NullableStringEqual(cfg.Homepage, current.Homepage) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"homepage",
			model.NullableStringVal(current.Homepage),
			model.PtrVal(cfg.Homepage),
		))
	}

	if cfg.Visibility != nil && !model.PtrStringEqual(cfg.Visibility, current.Visibility) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"visibility",
			model.PtrVal(current.Visibility),
			*cfg.Visibility,
		))
	}

	if cfg.AllowMergeCommit != nil && !model.PtrBoolEqual(cfg.AllowMergeCommit, current.AllowMergeCommit) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"allow_merge_commit",
			model.PtrBoolVal(current.AllowMergeCommit),
			*cfg.AllowMergeCommit,
		))
	}

	if cfg.AllowRebaseMerge != nil && !model.PtrBoolEqual(cfg.AllowRebaseMerge, current.AllowRebaseMerge) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"allow_rebase_merge",
			model.PtrBoolVal(current.AllowRebaseMerge),
			*cfg.AllowRebaseMerge,
		))
	}

	if cfg.AllowSquashMerge != nil && !model.PtrBoolEqual(cfg.AllowSquashMerge, current.AllowSquashMerge) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"allow_squash_merge",
			model.PtrBoolVal(current.AllowSquashMerge),
			*cfg.AllowSquashMerge,
		))
	}

	if cfg.DeleteBranchOnMerge != nil && !model.PtrBoolEqual(cfg.DeleteBranchOnMerge, current.DeleteBranchOnMerge) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"delete_branch_on_merge",
			model.PtrBoolVal(current.DeleteBranchOnMerge),
			*cfg.DeleteBranchOnMerge,
		))
	}

	if cfg.AllowUpdateBranch != nil && !model.PtrBoolEqual(cfg.AllowUpdateBranch, current.AllowUpdateBranch) {
		plan.Add(model.NewUpdateChange(
			model.CategoryRepo,
			"allow_update_branch",
			model.PtrBoolVal(current.AllowUpdateBranch),
			*cfg.AllowUpdateBranch,
		))
	}

	return plan, nil
}

// TopicsComparator compares repository topics
type TopicsComparator struct {
	client github.GitHubClient
	topics []string
}

// NewTopicsComparator creates a new TopicsComparator
func NewTopicsComparator(client github.GitHubClient, topics []string) *TopicsComparator {
	return &TopicsComparator{
		client: client,
		topics: topics,
	}
}

// Compare compares the current topics with the desired configuration
func (c *TopicsComparator) Compare(ctx context.Context) (*model.Plan, error) {
	current, err := c.client.GetRepo(ctx)
	if err != nil {
		return nil, err
	}

	plan := model.NewPlan()

	var currentTopics []string
	if current.Topics != nil {
		currentTopics = *current.Topics
	}

	if !model.StringSliceEqualIgnoreOrder(c.topics, currentTopics) {
		plan.Add(model.NewUpdateChange(
			model.CategoryTopics,
			"topics",
			currentTopics,
			c.topics,
		))
	}

	return plan, nil
}
