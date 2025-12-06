package diff

import "github.com/myzkey/gh-repo-settings/internal/diff/domain/model"

// Type aliases for backward compatibility
// These allow existing code to continue using diff.Change, diff.Plan, etc.
type (
	Change         = model.Change
	ChangeType     = model.ChangeType
	ChangeCategory = model.ChangeCategory
	Plan           = model.Plan
)

// Re-export ChangeType constants for backward compatibility
const (
	ChangeAdd     = model.ChangeAdd
	ChangeUpdate  = model.ChangeUpdate
	ChangeDelete  = model.ChangeDelete
	ChangeMissing = model.ChangeMissing
)

// Re-export ChangeCategory constants for backward compatibility
const (
	CategoryRepo             = model.CategoryRepo
	CategoryTopics           = model.CategoryTopics
	CategoryLabels           = model.CategoryLabels
	CategoryBranchProtection = model.CategoryBranchProtection
	CategoryVariables        = model.CategoryVariables
	CategorySecrets          = model.CategorySecrets
	CategoryActions          = model.CategoryActions
	CategoryPages            = model.CategoryPages
	CategoryEnv              = model.CategoryEnv // Alias for CategoryVariables
)
