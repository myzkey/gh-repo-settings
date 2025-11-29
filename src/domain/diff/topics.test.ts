import { describe, expect, it } from 'vitest'
import type { GitHubRepoData, IGitHubClient } from '../ports'
import type { Config } from '../types'
import { calculateTopicsDiffs } from './topics'
import type { DiffContext } from './types'

const mockClient = {} as IGitHubClient
const context: DiffContext = {
  client: mockClient,
  owner: 'test-owner',
  name: 'test-repo',
}

function createMockRepo(
  overrides: Partial<GitHubRepoData> = {},
): GitHubRepoData {
  return {
    description: null,
    homepage: null,
    visibility: 'public',
    allow_merge_commit: true,
    allow_rebase_merge: true,
    allow_squash_merge: true,
    delete_branch_on_merge: false,
    allow_update_branch: false,
    topics: [],
    ...overrides,
  }
}

describe('calculateTopicsDiffs', () => {
  it('should return empty array when config.topics is undefined', () => {
    const config: Config = {}
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when topics match', () => {
    const config: Config = {
      topics: ['cli', 'github'],
    }
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when topics match in different order', () => {
    const config: Config = {
      topics: ['github', 'cli'],
    }
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })

  it('should detect added topics', () => {
    const config: Config = {
      topics: ['cli', 'github', 'new-topic'],
    }
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('topics')
    expect(diffs[0].action).toBe('update')
    expect(diffs[0].apiCall).toBe('PUT /repos/test-owner/test-repo/topics')
  })

  it('should detect removed topics', () => {
    const config: Config = {
      topics: ['cli'],
    }
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('cli')
  })

  it('should handle empty current topics', () => {
    const config: Config = {
      topics: ['cli', 'github'],
    }
    const currentRepo = createMockRepo({ topics: [] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
  })

  it('should handle undefined current topics', () => {
    const config: Config = {
      topics: ['cli', 'github'],
    }
    const currentRepo = createMockRepo()

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
  })

  it('should handle setting empty topics', () => {
    const config: Config = {
      topics: [],
    }
    const currentRepo = createMockRepo({ topics: ['cli', 'github'] })

    const diffs = calculateTopicsDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
  })
})
