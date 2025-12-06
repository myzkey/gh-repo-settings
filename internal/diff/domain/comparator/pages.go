package comparator

import (
	"context"
	"fmt"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// PagesComparator compares GitHub Pages settings
type PagesComparator struct {
	client github.GitHubClient
	config *config.PagesConfig
}

// NewPagesComparator creates a new PagesComparator
func NewPagesComparator(client github.GitHubClient, cfg *config.PagesConfig) *PagesComparator {
	return &PagesComparator{
		client: client,
		config: cfg,
	}
}

// Compare compares the current pages settings with the desired configuration
func (c *PagesComparator) Compare(ctx context.Context) (*model.Plan, error) {
	plan := model.NewPlan()

	current, err := c.client.GetPages(ctx)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrPagesNotEnabled) {
			// Pages not enabled, will be created
			buildType := "workflow"
			if c.config.BuildType != nil {
				buildType = *c.config.BuildType
			}
			plan.Add(model.NewAddChange(
				model.CategoryPages,
				"pages",
				fmt.Sprintf("build_type=%s", buildType),
			))
			return plan, nil
		}
		return nil, err
	}

	// Compare build_type
	if c.config.BuildType != nil {
		currentBuildType := ""
		if current.BuildType.IsSpecified() && !current.BuildType.IsNull() {
			currentBuildType = string(current.BuildType.MustGet())
		}
		if *c.config.BuildType != currentBuildType {
			plan.Add(model.NewUpdateChange(
				model.CategoryPages,
				"build_type",
				currentBuildType,
				*c.config.BuildType,
			))
		}
	}

	// Compare source (only for legacy build type)
	if c.config.Source != nil && current.Source != nil {
		if c.config.Source.Branch != nil && *c.config.Source.Branch != current.Source.Branch {
			plan.Add(model.NewUpdateChange(
				model.CategoryPages,
				"source.branch",
				current.Source.Branch,
				*c.config.Source.Branch,
			))
		}
		if c.config.Source.Path != nil && *c.config.Source.Path != current.Source.Path {
			plan.Add(model.NewUpdateChange(
				model.CategoryPages,
				"source.path",
				current.Source.Path,
				*c.config.Source.Path,
			))
		}
	}

	return plan, nil
}
