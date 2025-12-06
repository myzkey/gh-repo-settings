// Package model contains the core domain models for the diff package.
//
// These models represent the fundamental concepts in the repository settings
// comparison domain:
//
//   - Change: A single configuration difference with type, category, key, and values
//   - ChangeType: Enumeration of change types (Add, Update, Delete, Missing)
//   - ChangeCategory: Typed enumeration of resource categories (repo, labels, etc.)
//   - Plan: A collection of changes with rich query and transformation methods
//   - BranchProtectionCurrent/Desired: Domain models for branch protection state
//
// # Design Principles
//
//   - Models are independent of infrastructure (no GitHub API dependencies)
//   - Models are independent of configuration format (no config package dependencies)
//   - Rich behavior is encapsulated within models (not anemic data structures)
//   - Immutable operations return new instances (Filter, Merge, Invert)
//
// # Usage
//
// Most external usage should go through the parent diff package, which re-exports
// these types for convenience:
//
//	import "github.com/myzkey/gh-repo-settings/internal/diff"
//
//	plan := diff.NewPlan()
//	plan.Add(diff.NewAddChange(diff.CategoryLabels, "bug", "color=red"))
package model
