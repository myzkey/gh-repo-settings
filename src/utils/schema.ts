import { z } from "zod";
import type { ValidationResult, Config } from "~/types";

// Hex color validation (6 characters, no #)
const hexColorSchema = z
  .string()
  .regex(/^[0-9a-fA-F]{6}$/, "Must be a 6-character hex color (without #)");

// Repo settings schema
const repoSettingsSchema = z
  .object({
    description: z.string().optional(),
    homepage: z.string().url().optional().or(z.literal("")),
    visibility: z.enum(["public", "private", "internal"]).optional(),
    allow_merge_commit: z.boolean().optional(),
    allow_rebase_merge: z.boolean().optional(),
    allow_squash_merge: z.boolean().optional(),
    delete_branch_on_merge: z.boolean().optional(),
    allow_update_branch: z.boolean().optional(),
  })
  .strict();

// Label schema
const labelSchema = z
  .object({
    name: z.string().min(1, "Label name is required"),
    color: hexColorSchema,
    description: z.string().optional(),
  })
  .strict();

// Labels config schema
const labelsConfigSchema = z
  .object({
    replace_default: z.boolean().optional(),
    items: z.array(labelSchema),
  })
  .strict();

// Branch protection schema
const branchProtectionSchema = z
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
  .strict();

// Secrets config schema
const secretsConfigSchema = z
  .object({
    required: z.array(z.string().min(1)).optional(),
  })
  .strict();

// Env config schema
const envConfigSchema = z
  .object({
    required: z.array(z.string().min(1)).optional(),
  })
  .strict();

// Main config schema
export const configSchema = z
  .object({
    repo: repoSettingsSchema.optional(),
    topics: z
      .array(z.string().min(1).max(50).regex(/^[a-z0-9-]+$/))
      .max(20)
      .optional(),
    labels: labelsConfigSchema.optional(),
    branch_protection: z.record(z.string(), branchProtectionSchema).optional(),
    secrets: secretsConfigSchema.optional(),
    env: envConfigSchema.optional(),
  })
  .strict();

export function validateConfig(config: unknown): ValidationResult {
  const result = configSchema.safeParse(config);

  if (result.success) {
    return { valid: true, errors: [] };
  }

  const errors = result.error.issues.map((issue) => ({
    path: issue.path.join("."),
    message: issue.message,
  }));

  return { valid: false, errors };
}

export function validateConfigOrThrow(config: unknown): Config {
  return configSchema.parse(config) as Config;
}
