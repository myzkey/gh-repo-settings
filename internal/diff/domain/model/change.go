package model

import "fmt"

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

// ChangeCategory represents the category of a change
type ChangeCategory string

const (
	CategoryRepo             ChangeCategory = "repo"
	CategoryTopics           ChangeCategory = "topics"
	CategoryLabels           ChangeCategory = "labels"
	CategoryBranchProtection ChangeCategory = "branch_protection"
	CategoryVariables        ChangeCategory = "variables"
	CategorySecrets          ChangeCategory = "secrets"
	CategoryActions          ChangeCategory = "actions"
	CategoryPages            ChangeCategory = "pages"
)

// CategoryEnv is an alias for CategoryVariables for backward compatibility
const CategoryEnv = CategoryVariables

// String returns the string representation of the category
func (c ChangeCategory) String() string {
	return string(c)
}

// Change represents a single configuration change
type Change struct {
	Type     ChangeType
	Category ChangeCategory
	Key      string
	Old      interface{}
	New      interface{}
}

// NewAddChange creates a new add change
func NewAddChange(category ChangeCategory, key string, newValue interface{}) Change {
	return Change{
		Type:     ChangeAdd,
		Category: category,
		Key:      key,
		New:      newValue,
	}
}

// NewUpdateChange creates a new update change
func NewUpdateChange(category ChangeCategory, key string, oldValue, newValue interface{}) Change {
	return Change{
		Type:     ChangeUpdate,
		Category: category,
		Key:      key,
		Old:      oldValue,
		New:      newValue,
	}
}

// NewDeleteChange creates a new delete change
func NewDeleteChange(category ChangeCategory, key string, oldValue interface{}) Change {
	return Change{
		Type:     ChangeDelete,
		Category: category,
		Key:      key,
		Old:      oldValue,
	}
}

// NewMissingChange creates a new missing change (for secrets/env)
func NewMissingChange(category ChangeCategory, key string, description interface{}) Change {
	return Change{
		Type:     ChangeMissing,
		Category: category,
		Key:      key,
		New:      description,
	}
}

// IsAdd returns true if this is an add change
func (c Change) IsAdd() bool {
	return c.Type == ChangeAdd
}

// IsUpdate returns true if this is an update change
func (c Change) IsUpdate() bool {
	return c.Type == ChangeUpdate
}

// IsDelete returns true if this is a delete change
func (c Change) IsDelete() bool {
	return c.Type == ChangeDelete
}

// IsMissing returns true if this is a missing change
func (c Change) IsMissing() bool {
	return c.Type == ChangeMissing
}

// Invert returns the inverse of this change (add becomes delete, etc.)
func (c Change) Invert() Change {
	inverted := c
	switch c.Type {
	case ChangeAdd:
		inverted.Type = ChangeDelete
		inverted.Old = c.New
		inverted.New = nil
	case ChangeDelete:
		inverted.Type = ChangeAdd
		inverted.New = c.Old
		inverted.Old = nil
	case ChangeUpdate:
		inverted.Old = c.New
		inverted.New = c.Old
	}
	return inverted
}

// WithCategory returns a copy of the change with a different category
func (c Change) WithCategory(category ChangeCategory) Change {
	result := c
	result.Category = category
	return result
}

// WithKeyPrefix returns a copy of the change with a prefixed key
func (c Change) WithKeyPrefix(prefix string) Change {
	result := c
	result.Key = prefix + c.Key
	return result
}

// String returns a human-readable representation of the change
func (c Change) String() string {
	switch c.Type {
	case ChangeAdd:
		return fmt.Sprintf("[ADD] %s.%s = %v", c.Category, c.Key, c.New)
	case ChangeUpdate:
		return fmt.Sprintf("[UPDATE] %s.%s: %v -> %v", c.Category, c.Key, c.Old, c.New)
	case ChangeDelete:
		return fmt.Sprintf("[DELETE] %s.%s (was %v)", c.Category, c.Key, c.Old)
	case ChangeMissing:
		return fmt.Sprintf("[MISSING] %s.%s: %v", c.Category, c.Key, c.New)
	default:
		return fmt.Sprintf("[UNKNOWN] %s.%s", c.Category, c.Key)
	}
}
