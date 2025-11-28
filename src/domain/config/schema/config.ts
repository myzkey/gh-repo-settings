import { z } from 'zod'
import type { Config, ValidationResult } from '../../types'
import { branchProtectionSchema } from './branch-protection'
import { envConfigSchema } from './env'
import { labelsConfigSchema } from './labels'
import { repoSettingsSchema } from './repo'
import { secretsConfigSchema } from './secrets'
import { topicsSchema } from './topics'

export const configSchema = z
  .object({
    repo: repoSettingsSchema.optional(),
    topics: topicsSchema.optional(),
    labels: labelsConfigSchema.optional(),
    branch_protection: z.record(z.string(), branchProtectionSchema).optional(),
    secrets: secretsConfigSchema.optional(),
    env: envConfigSchema.optional(),
  })
  .strict()

export function validateConfig(config: unknown): ValidationResult {
  const result = configSchema.safeParse(config)

  if (result.success) {
    return { valid: true, errors: [] }
  }

  const errors = result.error.issues.map((issue) => ({
    path: issue.path.join('.'),
    message: issue.message,
  }))

  return { valid: false, errors }
}

export function validateConfigOrThrow(config: unknown): Config {
  return configSchema.parse(config) as Config
}
