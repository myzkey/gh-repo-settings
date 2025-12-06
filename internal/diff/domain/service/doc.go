// Package service contains pure domain services for the diff package.
//
// Domain services encapsulate domain logic that doesn't naturally fit within
// a single entity or value object. They are:
//
//   - Pure functions with no side effects
//   - Independent of infrastructure (no GitHub API calls)
//   - Independent of application concerns (no orchestration logic)
//   - Focused on a single domain operation
//
// # Available Services
//
//   - CompareBranchRule: Compares current and desired branch protection states
//
// # Usage
//
// Domain services are typically called by application services (comparators)
// rather than directly by external code:
//
//	changes := service.CompareBranchRule(branchName, current, desired)
package service
