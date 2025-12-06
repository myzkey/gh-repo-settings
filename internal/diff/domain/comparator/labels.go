package comparator

import (
	"context"
	"fmt"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

// LabelsComparator compares repository labels
type LabelsComparator struct {
	client github.GitHubClient
	config *config.LabelsConfig
}

// NewLabelsComparator creates a new LabelsComparator
func NewLabelsComparator(client github.GitHubClient, cfg *config.LabelsConfig) *LabelsComparator {
	return &LabelsComparator{
		client: client,
		config: cfg,
	}
}

// Compare compares the current labels with the desired configuration
func (c *LabelsComparator) Compare(ctx context.Context) (*model.Plan, error) {
	currentLabels, err := c.client.GetLabels(ctx)
	if err != nil {
		return nil, err
	}

	plan := model.NewPlan()

	currentMap := make(map[string]github.LabelData)
	for _, l := range currentLabels {
		currentMap[l.Name] = l
	}

	configMap := make(map[string]config.Label)
	for _, l := range c.config.Items {
		configMap[l.Name] = l
	}

	// Check for additions and updates
	for _, cfgLabel := range c.config.Items {
		if current, exists := currentMap[cfgLabel.Name]; exists {
			// Check for updates
			currentDesc := model.NullableStringVal(current.Description)
			if cfgLabel.Color != current.Color || cfgLabel.Description != currentDesc {
				plan.Add(model.NewUpdateChange(
					model.CategoryLabels,
					cfgLabel.Name,
					formatLabel(current.Color, currentDesc),
					formatLabel(cfgLabel.Color, cfgLabel.Description),
				))
			}
		} else {
			// Addition
			plan.Add(model.NewAddChange(
				model.CategoryLabels,
				cfgLabel.Name,
				formatLabel(cfgLabel.Color, cfgLabel.Description),
			))
		}
	}

	// Check for deletions (only if replace_default is true)
	if c.config.ReplaceDefault {
		for _, currentLabel := range currentLabels {
			if _, exists := configMap[currentLabel.Name]; !exists {
				plan.Add(model.NewDeleteChange(
					model.CategoryLabels,
					currentLabel.Name,
					formatLabel(currentLabel.Color, model.NullableStringVal(currentLabel.Description)),
				))
			}
		}
	}

	return plan, nil
}

func formatLabel(color, description string) string {
	return fmt.Sprintf("color=%s, description=%s", color, description)
}
