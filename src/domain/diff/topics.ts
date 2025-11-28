import type { GitHubRepoData } from '../ports'
import type { Config, DiffItem } from '../types'
import type { DiffContext } from './types'

export function calculateTopicsDiffs(
  config: Config,
  currentRepo: GitHubRepoData,
  context: DiffContext,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = context

  if (!config.topics) return diffs

  const currentTopics = currentRepo.topics || []
  const newTopics = config.topics

  const added = newTopics.filter((t) => !currentTopics.includes(t))
  const removed = currentTopics.filter((t) => !newTopics.includes(t))

  if (added.length > 0 || removed.length > 0) {
    diffs.push({
      type: 'topics',
      action: 'update',
      details: `Set topics to: [${newTopics.join(', ')}]`,
      apiCall: `PUT /repos/${owner}/${name}/topics`,
    })
  }

  return diffs
}
