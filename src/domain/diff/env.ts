import type { Config, DiffItem } from '../types'

export function calculateEnvDiffs(
  config: Config,
  existingVarNames: string[],
): DiffItem[] {
  const diffs: DiffItem[] = []

  if (!config.env?.required || config.env.required.length === 0) {
    return diffs
  }

  const existingVars = new Set(existingVarNames)

  for (const varName of config.env.required) {
    if (!existingVars.has(varName)) {
      diffs.push({
        type: 'env',
        action: 'check',
        details: `Missing env variable: ${varName} (use 'gh variable set ${varName}' to add)`,
      })
    }
  }

  return diffs
}
