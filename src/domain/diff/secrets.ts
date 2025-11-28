import type { Config, DiffItem } from '../types'

export function calculateSecretsDiffs(
  config: Config,
  existingSecretNames: string[],
): DiffItem[] {
  const diffs: DiffItem[] = []

  if (!config.secrets?.required || config.secrets.required.length === 0) {
    return diffs
  }

  const existingSecrets = new Set(existingSecretNames)

  for (const secretName of config.secrets.required) {
    if (!existingSecrets.has(secretName)) {
      diffs.push({
        type: 'secrets',
        action: 'check',
        details: `Missing secret: ${secretName} (use 'gh secret set ${secretName}' to add)`,
      })
    }
  }

  return diffs
}
