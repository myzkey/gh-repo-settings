import { z } from 'zod'

export const topicsSchema = z
  .array(
    z
      .string()
      .min(1)
      .max(50)
      .regex(/^[a-z0-9-]+$/),
  )
  .max(20)
