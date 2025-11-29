package diff

import (
	"fmt"
	"reflect"

	"github.com/myzkey/gh-repo-settings/internal/config"
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
)

func (c ChangeType) String() string {
	switch c {
	case ChangeAdd:
		return "add"
	case ChangeUpdate:
		return "update"
	case ChangeDelete:
		return "delete"
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

// Calculator calculates diff between config and current state
type Calculator struct {
	client *github.Client
	config *config.Config
}

// NewCalculator creates a new diff calculator
func NewCalculator(client *github.Client, cfg *config.Config) *Calculator {
	return &Calculator{
		client: client,
		config: cfg,
	}
}

// Calculate calculates the diff
func (c *Calculator) Calculate() (*Plan, error) {
	plan := &Plan{}

	// Compare repo settings
	if c.config.Repo != nil {
		changes, err := c.compareRepo()
		if err != nil {
			return nil, fmt.Errorf("failed to compare repo settings: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare topics
	if c.config.Topics != nil {
		changes, err := c.compareTopics()
		if err != nil {
			return nil, fmt.Errorf("failed to compare topics: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	// Compare labels
	if c.config.Labels != nil {
		changes, err := c.compareLabels()
		if err != nil {
			return nil, fmt.Errorf("failed to compare labels: %w", err)
		}
		plan.Changes = append(plan.Changes, changes...)
	}

	return plan, nil
}

func (c *Calculator) compareRepo() ([]Change, error) {
	current, err := c.client.GetRepo()
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

func (c *Calculator) compareTopics() ([]Change, error) {
	current, err := c.client.GetRepo()
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

func (c *Calculator) compareLabels() ([]Change, error) {
	currentLabels, err := c.client.GetLabels()
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
