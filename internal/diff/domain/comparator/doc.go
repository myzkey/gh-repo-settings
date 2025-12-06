// Package comparator contains application services that orchestrate
// domain logic for comparing repository settings.
//
// Comparators are responsible for:
//
//   - Fetching current state from GitHub via Gateway interfaces
//   - Mapping configuration to domain models
//   - Delegating comparison logic to domain services
//   - Aggregating results into a Plan
//
// # Available Comparators
//
//   - RepoComparator: Repository settings (description, visibility, merge options)
//   - TopicsComparator: Repository topics
//   - LabelsComparator: Issue and PR labels
//   - BranchProtectionComparator: Branch protection rules
//   - EnvComparator: Environment variables and secrets
//   - ActionsComparator: GitHub Actions permissions
//   - PagesComparator: GitHub Pages settings
//
// # Gateway Pattern
//
// Comparators use Gateway interfaces to abstract infrastructure dependencies.
// This allows for easy testing and keeps domain logic isolated:
//
//	type BranchProtectionGateway interface {
//	    GetBranchProtection(ctx context.Context, branch string) (model.BranchProtectionCurrent, error)
//	}
//
// # Usage
//
// Comparators are typically created and called by the Calculator:
//
//	comparator := comparator.NewLabelsComparator(client, cfg.Labels)
//	plan, err := comparator.Compare(ctx)
package comparator
