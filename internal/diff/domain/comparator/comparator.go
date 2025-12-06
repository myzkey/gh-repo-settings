package comparator

import (
	"context"

	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
)

// Comparator is the interface that all resource comparators must implement
type Comparator interface {
	// Compare compares the current state with the desired state and returns a plan
	Compare(ctx context.Context) (*model.Plan, error)
}
