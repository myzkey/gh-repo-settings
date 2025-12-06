// Package presentation contains formatting logic for human-readable output.
//
// This package is responsible for converting domain objects into
// user-friendly string representations. It should have no business logic
// and only concern itself with display formatting.
//
// # Available Functions
//
//   - FormatBranchRule: Formats a branch rule configuration for display
//
// # Usage
//
// Presentation functions are typically called when creating Change objects
// that need human-readable descriptions:
//
//	plan.Add(model.NewAddChange(
//	    model.CategoryBranchProtection,
//	    branchName,
//	    presentation.FormatBranchRule(rule),
//	))
package presentation
