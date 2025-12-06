// Package diff provides use cases and domain models to calculate differences
// between the desired repository settings (YAML) and the current state
// fetched from GitHub.
//
// # Architecture
//
// This package follows Domain-Driven Design (DDD) principles with clear layer separation:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                     diff (Application Layer)                     │
//	│  calculator.go - Orchestrates domain services to build a Plan   │
//	│  types.go      - Re-exports domain types for external use       │
//	│  json.go       - JSON rendering for Plan output                 │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                  domain/comparator (Use Cases)                   │
//	│  Application services that orchestrate domain logic              │
//	│  - Uses Gateway interfaces to fetch current state               │
//	│  - Maps config to domain models                                  │
//	│  - Delegates comparison to domain services                       │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                    ┌─────────┴─────────┐
//	                    ▼                   ▼
//	┌────────────────────────┐  ┌────────────────────────┐
//	│   domain/model         │  │   domain/service       │
//	│   (Domain Models)      │  │   (Domain Services)    │
//	│                        │  │                        │
//	│ - Change, ChangeType   │  │ - CompareBranchRule    │
//	│ - ChangeCategory       │  │ - Pure comparison      │
//	│ - Plan                 │  │   logic with no        │
//	│ - BranchProtection*    │  │   infrastructure deps  │
//	│ - Helper functions     │  │                        │
//	└────────────────────────┘  └────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                  presentation (Presentation Layer)               │
//	│  Formatting logic for human-readable output                      │
//	│  - FormatBranchRule                                              │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Key Design Decisions
//
//   - Domain models (Change, Plan) are rich with behavior (Filter, Merge, Invert)
//   - ChangeCategory is a typed enum to prevent typos and enable safe filtering
//   - Gateway pattern isolates infrastructure (GitHub API) from domain logic
//   - Comparators are application services, not domain services
//   - Domain services (CompareBranchRule) are pure functions with no side effects
//
// # Usage
//
// External packages should primarily interact with the diff package:
//
//	calc := diff.NewCalculator(client, cfg)
//	plan, err := calc.Calculate(ctx)
//	if plan.HasChanges() {
//	    for _, change := range plan.Changes() {
//	        // process change
//	    }
//	}
//
// Domain models are re-exported via types.go for backward compatibility.
package diff
