import { describe, expect, it } from 'vitest'
import type { GitHubRepoData, IGitHubClient } from '../ports'
import type { Config } from '../types'
import { calculateRepoDiffs } from './repo'
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

describe('calculateRepoDiffs', () => {
  it('should return empty array when config.repo is undefined', () => {
    const config: Config = {}
    const currentRepo = createMockRepo({ description: 'Test' })

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when all values match', () => {
    const config: Config = {
      repo: {
        description: 'Test description',
        visibility: 'public',
      },
    }
    const currentRepo = createMockRepo({ description: 'Test description' })

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })

  it('should detect description change', () => {
    const config: Config = {
      repo: {
        description: 'New description',
      },
    }
    const currentRepo = createMockRepo({ description: 'Old description' })

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('repo')
    expect(diffs[0].action).toBe('update')
    expect(diffs[0].details).toContain('description')
    expect(diffs[0].apiCall).toBe('PATCH /repos/test-owner/test-repo')
  })

  it('should detect visibility change', () => {
    const config: Config = {
      repo: {
        visibility: 'private',
      },
    }
    const currentRepo = createMockRepo()

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('visibility')
    expect(diffs[0].details).toContain('"public"')
    expect(diffs[0].details).toContain('"private"')
  })

  it('should detect multiple changes', () => {
    const config: Config = {
      repo: {
        description: 'New description',
        visibility: 'private',
        allow_merge_commit: false,
      },
    }
    const currentRepo = createMockRepo({ description: 'Old description' })

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(3)
  })

  it('should ignore undefined values in config', () => {
    const config: Config = {
      repo: {
        description: 'Test',
        homepage: undefined,
      },
    }
    const currentRepo = createMockRepo({
      description: 'Test',
      homepage: 'https://example.com',
    })

    const diffs = calculateRepoDiffs(config, currentRepo, context)

    expect(diffs).toHaveLength(0)
  })
})
