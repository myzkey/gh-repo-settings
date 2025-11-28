import { z } from 'zod'

export const branchProtectionSchema = z
  .object({
    required_reviews: z.number().int().min(0).max(6).optional(),
    dismiss_stale_reviews: z.boolean().optional(),
    require_status_checks: z.boolean().optional(),
    status_checks: z.array(z.string()).optional(),
    enforce_admins: z.boolean().optional(),
    require_linear_history: z.boolean().optional(),
    allow_force_pushes: z.boolean().optional(),
    allow_deletions: z.boolean().optional(),
  })
  .strict()
