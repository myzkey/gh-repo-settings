import { z } from 'zod'

const hexColorSchema = z
  .string()
  .regex(/^[0-9a-fA-F]{6}$/, 'Must be a 6-character hex color (without #)')

const labelSchema = z
  .object({
    name: z.string().min(1, 'Label name is required'),
    color: hexColorSchema,
    description: z.string().optional(),
  })
  .strict()

export const labelsConfigSchema = z
  .object({
    replace_default: z.boolean().optional(),
    items: z.array(labelSchema),
  })
  .strict()
