import { z } from 'zod'

export const branchProtectionSchema = z
  .object({
    // Pull request reviews
    required_reviews: z.number().int().min(0).max(6).optional(),
    dismiss_stale_reviews: z.boolean().optional(),
    require_code_owner: z.boolean().optional(),

    // Status checks
    require_status_checks: z.boolean().optional(),
    status_checks: z.array(z.string()).optional(),
    strict_status_checks: z.boolean().optional(),

    // Deployments
    required_deployments: z.array(z.string()).optional(),

    // Commit requirements
    require_signed_commits: z.boolean().optional(),
    require_linear_history: z.boolean().optional(),

    // Push/merge restrictions
    enforce_admins: z.boolean().optional(),
    restrict_creations: z.boolean().optional(),
    restrict_pushes: z.boolean().optional(),
    allow_force_pushes: z.boolean().optional(),
    allow_deletions: z.boolean().optional(),
  })
  .strict()
