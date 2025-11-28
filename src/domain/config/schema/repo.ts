import { z } from 'zod'

export const repoSettingsSchema = z
  .object({
    description: z.string().optional(),
    homepage: z.string().url().optional().or(z.literal('')),
    visibility: z.enum(['public', 'private', 'internal']).optional(),
    allow_merge_commit: z.boolean().optional(),
    allow_rebase_merge: z.boolean().optional(),
    allow_squash_merge: z.boolean().optional(),
    delete_branch_on_merge: z.boolean().optional(),
    allow_update_branch: z.boolean().optional(),
  })
  .strict()
