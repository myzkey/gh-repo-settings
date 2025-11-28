import type { GitHubBranchProtectionData } from '../ports'
import type { Config, DiffItem } from '../types'
import type { DiffContext } from './types'

export function calculateBranchProtectionDiffs(
  config: Config,
  protection: GitHubBranchProtectionData | null,
  branch: string,
  context: DiffContext,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = context

  const bp = config.branch_protection?.[branch]
  if (!bp) return diffs

  if (!protection) {
    diffs.push({
      type: 'branch_protection',
      action: 'create',
      details: `Branch protection for ${branch} does not exist`,
      apiCall: `PUT /repos/${owner}/${name}/branches/${branch}/protection`,
    })
    return diffs
  }

  // Check required reviews
  if (bp.required_reviews !== undefined) {
    const currentReviews =
      protection.required_pull_request_reviews
        ?.required_approving_review_count ?? 0
    if (currentReviews !== bp.required_reviews) {
      diffs.push({
        type: 'branch_protection',
        action: 'update',
        details: `required_reviews: ${currentReviews} -> ${bp.required_reviews}`,
        apiCall: `PUT /repos/${owner}/${name}/branches/${branch}/protection`,
      })
    }
  }

  // Check dismiss_stale_reviews
  if (bp.dismiss_stale_reviews !== undefined) {
    const currentDismiss =
      protection.required_pull_request_reviews?.dismiss_stale_reviews ?? false
    if (currentDismiss !== bp.dismiss_stale_reviews) {
      diffs.push({
        type: 'branch_protection',
        action: 'update',
        details: `dismiss_stale_reviews: ${currentDismiss} -> ${bp.dismiss_stale_reviews}`,
      })
    }
  }

  // Check status checks
  if (bp.require_status_checks !== undefined) {
    const hasStatusChecks = !!protection.required_status_checks
    if (hasStatusChecks !== bp.require_status_checks) {
      diffs.push({
        type: 'branch_protection',
        action: 'update',
        details: `require_status_checks: ${hasStatusChecks} -> ${bp.require_status_checks}`,
      })
    } else if (bp.status_checks && protection.required_status_checks) {
      const currentChecks = new Set(
        protection.required_status_checks.contexts || [],
      )
      const expectedChecks = new Set(bp.status_checks)

      const missing = bp.status_checks.filter((c) => !currentChecks.has(c))
      const extra = (protection.required_status_checks.contexts || []).filter(
        (c) => !expectedChecks.has(c),
      )

      if (missing.length > 0 || extra.length > 0) {
        diffs.push({
          type: 'branch_protection',
          action: 'update',
          details: `status_checks differ - missing: [${missing.join(', ')}], extra: [${extra.join(', ')}]`,
        })
      }
    }
  }

  // Check enforce_admins
  if (bp.enforce_admins !== undefined) {
    const currentEnforce = protection.enforce_admins?.enabled ?? false
    if (currentEnforce !== bp.enforce_admins) {
      diffs.push({
        type: 'branch_protection',
        action: 'update',
        details: `enforce_admins: ${currentEnforce} -> ${bp.enforce_admins}`,
      })
    }
  }

  return diffs
}
