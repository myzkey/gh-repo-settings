import type { GitHubRepoData } from '../ports'
import type { Config, DiffItem } from '../types'
import type { DiffContext } from './types'

export function calculateRepoDiffs(
  config: Config,
  currentRepo: GitHubRepoData,
  context: DiffContext,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = context

  if (!config.repo) return diffs

  for (const [key, value] of Object.entries(config.repo)) {
    if (value === undefined) continue

    const currentValue = currentRepo[key as keyof GitHubRepoData]
    if (currentValue !== value) {
      diffs.push({
        type: 'repo',
        action: 'update',
        details: `${key}: ${JSON.stringify(currentValue)} -> ${JSON.stringify(value)}`,
        apiCall: `PATCH /repos/${owner}/${name}`,
      })
    }
  }

  return diffs
}
