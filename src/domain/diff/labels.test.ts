import { describe, expect, it } from 'vitest'
import type { GitHubLabelData, IGitHubClient } from '../ports'
import type { Config } from '../types'
import { calculateLabelsDiffs } from './labels'
import type { DiffContext } from './types'

const mockClient = {} as IGitHubClient
const context: DiffContext = {
  client: mockClient,
  owner: 'test-owner',
  name: 'test-repo',
}

describe('calculateLabelsDiffs', () => {
  it('should return empty array when config.labels is undefined', () => {
    const config: Config = {}
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: 'Bug report' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when all labels match', () => {
    const config: Config = {
      labels: {
        items: [{ name: 'bug', color: 'ff0000', description: 'Bug report' }],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: 'Bug report' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(0)
  })

  it('should detect new label to create', () => {
    const config: Config = {
      labels: {
        items: [{ name: 'bug', color: 'ff0000' }],
      },
    }
    const currentLabels: GitHubLabelData[] = []

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('labels')
    expect(diffs[0].action).toBe('create')
    expect(diffs[0].details).toContain('bug')
    expect(diffs[0].apiCall).toBe('POST /repos/test-owner/test-repo/labels')
  })

  it('should detect label color change', () => {
    const config: Config = {
      labels: {
        items: [{ name: 'bug', color: '00ff00' }],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].action).toBe('update')
    expect(diffs[0].apiCall).toBe(
      'PATCH /repos/test-owner/test-repo/labels/bug',
    )
  })

  it('should detect label description change', () => {
    const config: Config = {
      labels: {
        items: [
          { name: 'bug', color: 'ff0000', description: 'New description' },
        ],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: 'Old description' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].action).toBe('update')
  })

  it('should handle empty description as equivalent to undefined', () => {
    const config: Config = {
      labels: {
        items: [{ name: 'bug', color: 'ff0000' }],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(0)
  })

  it('should not delete labels when replace_default is false', () => {
    const config: Config = {
      labels: {
        replace_default: false,
        items: [{ name: 'bug', color: 'ff0000' }],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: '' },
      { name: 'enhancement', color: '00ff00', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(0)
  })

  it('should delete labels when replace_default is true', () => {
    const config: Config = {
      labels: {
        replace_default: true,
        items: [{ name: 'bug', color: 'ff0000' }],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: '' },
      { name: 'enhancement', color: '00ff00', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].action).toBe('delete')
    expect(diffs[0].details).toContain('enhancement')
  })

  it('should handle label name with special characters', () => {
    const config: Config = {
      labels: {
        replace_default: true,
        items: [],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug/critical', color: 'ff0000', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].apiCall).toContain(encodeURIComponent('bug/critical'))
  })

  it('should handle multiple operations', () => {
    const config: Config = {
      labels: {
        replace_default: true,
        items: [
          { name: 'bug', color: '00ff00' },
          { name: 'feature', color: '0000ff' },
        ],
      },
    }
    const currentLabels: GitHubLabelData[] = [
      { name: 'bug', color: 'ff0000', description: '' },
      { name: 'wontfix', color: 'ffffff', description: '' },
    ]

    const diffs = calculateLabelsDiffs(config, currentLabels, context)

    expect(diffs).toHaveLength(3)
    expect(diffs.filter((d) => d.action === 'delete')).toHaveLength(1)
    expect(diffs.filter((d) => d.action === 'update')).toHaveLength(1)
    expect(diffs.filter((d) => d.action === 'create')).toHaveLength(1)
  })
})
