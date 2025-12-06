package model

// Plan represents the execution plan containing all changes
type Plan struct {
	changes []Change
}

// NewPlan creates an empty plan
func NewPlan() *Plan {
	return &Plan{
		changes: make([]Change, 0),
	}
}

// NewPlanFromChanges creates a plan from existing changes
func NewPlanFromChanges(changes []Change) *Plan {
	return &Plan{
		changes: changes,
	}
}

// Add adds a change to the plan
func (p *Plan) Add(change Change) {
	p.changes = append(p.changes, change)
}

// AddAll adds multiple changes to the plan
func (p *Plan) AddAll(changes []Change) {
	p.changes = append(p.changes, changes...)
}

// Changes returns all changes in the plan
func (p *Plan) Changes() []Change {
	return p.changes
}

// IsEmpty returns true if the plan has no changes
func (p *Plan) IsEmpty() bool {
	return len(p.changes) == 0
}

// HasChanges returns true if there are any changes
func (p *Plan) HasChanges() bool {
	return len(p.changes) > 0
}

// Size returns the number of changes
func (p *Plan) Size() int {
	return len(p.changes)
}

// Merge combines two plans into a new plan
func (p *Plan) Merge(other *Plan) *Plan {
	if other == nil {
		return NewPlanFromChanges(p.changes)
	}
	merged := make([]Change, 0, len(p.changes)+len(other.changes))
	merged = append(merged, p.changes...)
	merged = append(merged, other.changes...)
	return NewPlanFromChanges(merged)
}

// Filter returns a new plan containing only changes that match the predicate
func (p *Plan) Filter(predicate func(Change) bool) *Plan {
	filtered := make([]Change, 0)
	for _, c := range p.changes {
		if predicate(c) {
			filtered = append(filtered, c)
		}
	}
	return NewPlanFromChanges(filtered)
}

// FilterByCategory returns a new plan containing only changes in the given category
func (p *Plan) FilterByCategory(category ChangeCategory) *Plan {
	return p.Filter(func(c Change) bool {
		return c.Category == category
	})
}

// FilterByType returns a new plan containing only changes of the given type
func (p *Plan) FilterByType(changeType ChangeType) *Plan {
	return p.Filter(func(c Change) bool {
		return c.Type == changeType
	})
}

// HasMissingSecrets returns true if there are missing secrets
func (p *Plan) HasMissingSecrets() bool {
	return !p.FilterByCategory(CategorySecrets).FilterByType(ChangeMissing).IsEmpty()
}

// HasMissingVariables returns true if there are missing variables
func (p *Plan) HasMissingVariables() bool {
	return !p.FilterByCategory(CategoryEnv).FilterByType(ChangeMissing).IsEmpty()
}

// HasDeletes returns true if there are any delete changes
func (p *Plan) HasDeletes() bool {
	for _, c := range p.changes {
		if c.Type == ChangeDelete {
			return true
		}
	}
	return false
}

// CountByType returns the count of changes by type
func (p *Plan) CountByType() map[ChangeType]int {
	counts := make(map[ChangeType]int)
	for _, c := range p.changes {
		counts[c.Type]++
	}
	return counts
}

// CountByCategory returns the count of changes by category
func (p *Plan) CountByCategory() map[ChangeCategory]int {
	counts := make(map[ChangeCategory]int)
	for _, c := range p.changes {
		counts[c.Category]++
	}
	return counts
}

// Invert returns a new plan with all changes inverted
func (p *Plan) Invert() *Plan {
	inverted := make([]Change, len(p.changes))
	for i, c := range p.changes {
		inverted[i] = c.Invert()
	}
	return NewPlanFromChanges(inverted)
}

// Categories returns all unique categories in the plan
func (p *Plan) Categories() []ChangeCategory {
	seen := make(map[ChangeCategory]bool)
	categories := make([]ChangeCategory, 0)
	for _, c := range p.changes {
		if !seen[c.Category] {
			seen[c.Category] = true
			categories = append(categories, c.Category)
		}
	}
	return categories
}
