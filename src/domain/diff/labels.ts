import type { GitHubLabelData } from '../ports'
import type { Config, DiffItem } from '../types'
import type { DiffContext } from './types'

export function calculateLabelsDiffs(
  config: Config,
  currentLabels: GitHubLabelData[],
  context: DiffContext,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = context

  if (!config.labels) return diffs

  const currentLabelMap = new Map(currentLabels.map((l) => [l.name, l]))
  const configLabelMap = new Map(config.labels.items.map((l) => [l.name, l]))

  // Delete labels if replace_default
  if (config.labels.replace_default) {
    for (const label of currentLabels) {
      if (!configLabelMap.has(label.name)) {
        diffs.push({
          type: 'labels',
          action: 'delete',
          details: `Delete label: ${label.name}`,
          apiCall: `DELETE /repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`,
        })
      }
    }
  }

  // Create or update labels
  for (const label of config.labels.items) {
    const current = currentLabelMap.get(label.name)

    if (!current) {
      diffs.push({
        type: 'labels',
        action: 'create',
        details: `Create label: ${label.name} (#${label.color})`,
        apiCall: `POST /repos/${owner}/${name}/labels`,
      })
    } else if (
      current.color !== label.color ||
      (current.description || '') !== (label.description || '')
    ) {
      diffs.push({
        type: 'labels',
        action: 'update',
        details: `Update label: ${label.name}`,
        apiCall: `PATCH /repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`,
      })
    }
  }

  return diffs
}
