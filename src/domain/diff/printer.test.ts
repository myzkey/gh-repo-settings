import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { DiffItem } from '../types'
import { printDiff } from './printer'

vi.mock('~/utils/logger', () => ({
  logger: {
    success: vi.fn(),
    heading: vi.fn(),
    section: vi.fn(),
    log: vi.fn(),
    debug: vi.fn(),
  },
  colors: {
    green: (s: string) => `[green]${s}[/green]`,
    yellow: (s: string) => `[yellow]${s}[/yellow]`,
    red: (s: string) => `[red]${s}[/red]`,
    blue: (s: string) => `[blue]${s}[/blue]`,
  },
}))

import { logger } from '~/utils/logger'

describe('printDiff', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should show success message when no diffs', () => {
    const diffs: DiffItem[] = []

    printDiff(diffs)

    expect(logger.success).toHaveBeenCalledWith('No changes detected.')
  })

  it('should print heading for non-empty diffs', () => {
    const diffs: DiffItem[] = [
      { type: 'repo', action: 'update', details: 'description changed' },
    ]

    printDiff(diffs)

    expect(logger.heading).toHaveBeenCalledWith('\nPlanned changes:\n')
  })

  it('should group diffs by type', () => {
    const diffs: DiffItem[] = [
      { type: 'repo', action: 'update', details: 'description changed' },
      { type: 'repo', action: 'update', details: 'visibility changed' },
      { type: 'labels', action: 'create', details: 'new label' },
    ]

    printDiff(diffs)

    expect(logger.section).toHaveBeenCalledWith('[repo]')
    expect(logger.section).toHaveBeenCalledWith('[labels]')
  })

  it('should print create action with green +', () => {
    const diffs: DiffItem[] = [
      { type: 'labels', action: 'create', details: 'Create label: bug' },
    ]

    printDiff(diffs)

    expect(logger.log).toHaveBeenCalledWith(
      expect.stringContaining('[green]+[/green]'),
    )
    expect(logger.log).toHaveBeenCalledWith(
      expect.stringContaining('[green]Create label: bug[/green]'),
    )
  })

  it('should print update action with yellow ~', () => {
    const diffs: DiffItem[] = [
      { type: 'repo', action: 'update', details: 'description changed' },
    ]

    printDiff(diffs)

    expect(logger.log).toHaveBeenCalledWith(
      expect.stringContaining('[yellow]~[/yellow]'),
    )
  })

  it('should print delete action with red -', () => {
    const diffs: DiffItem[] = [
      { type: 'labels', action: 'delete', details: 'Delete label: old' },
    ]

    printDiff(diffs)

    expect(logger.log).toHaveBeenCalledWith(
      expect.stringContaining('[red]-[/red]'),
    )
  })

  it('should print check action with blue ?', () => {
    const diffs: DiffItem[] = [
      { type: 'secrets', action: 'check', details: 'Missing secret' },
    ]

    printDiff(diffs)

    expect(logger.log).toHaveBeenCalledWith(
      expect.stringContaining('[blue]?[/blue]'),
    )
  })

  it('should print apiCall in debug mode', () => {
    const diffs: DiffItem[] = [
      {
        type: 'repo',
        action: 'update',
        details: 'description',
        apiCall: 'PATCH /repos/owner/repo',
      },
    ]

    printDiff(diffs)

    expect(logger.debug).toHaveBeenCalledWith(
      '    API: PATCH /repos/owner/repo',
    )
  })

  it('should not print apiCall when not present', () => {
    const diffs: DiffItem[] = [
      { type: 'secrets', action: 'check', details: 'Missing' },
    ]

    printDiff(diffs)

    expect(logger.debug).not.toHaveBeenCalled()
  })
})
