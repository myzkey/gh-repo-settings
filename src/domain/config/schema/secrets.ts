import { z } from 'zod'

export const secretsConfigSchema = z
  .object({
    required: z.array(z.string().min(1)).optional(),
  })
  .strict()
