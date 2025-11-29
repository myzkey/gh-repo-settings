import { beforeEach, describe, expect, it, vi } from 'vitest'
import { exportCommand } from './export'

vi.mock('~/domain', () => ({}))

vi.mock('~/infra/fs', () => ({
  existsSync: vi.fn(),
  mkdirSync: vi.fn(),
  writeFileSync: vi.fn(),
  join: (...paths: string[]) => paths.join('/'),
}))

vi.mock('~/infra/github', () => ({
  getRepoInfo: vi.fn(),
  createClient: vi.fn(),
}))

vi.mock('~/infra/yaml', () => ({
  stringify: vi.fn((data) => `yaml:${JSON.stringify(data)}`),
}))

vi.mock('~/utils/logger', () => ({
  logger: {
    info: vi.fn(),
    success: vi.fn(),
    log: vi.fn(),
  },
}))

import { existsSync, mkdirSync, writeFileSync } from '~/infra/fs'
import { createClient, getRepoInfo } from '~/infra/github'
import { logger } from '~/utils/logger'

const mockGetRepoInfo = vi.mocked(getRepoInfo)
const mockCreateClient = vi.mocked(createClient)
const mockExistsSync = vi.mocked(existsSync)
const mockMkdirSync = vi.mocked(mkdirSync)
const mockWriteFileSync = vi.mocked(writeFileSync)

const mockClient = {
  getRepo: vi.fn(),
  getLabels: vi.fn(),
  getBranchProtection: vi.fn(),
  getSecretNames: vi.fn(),
}

describe('exportCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetRepoInfo.mockReturnValue({ owner: 'test', name: 'repo' })
    mockCreateClient.mockReturnValue(mockClient as never)
    mockClient.getRepo.mockReturnValue({
      description: 'Test repo',
      homepage: '',
      visibility: 'public',
      allow_merge_commit: true,
      allow_rebase_merge: true,
      allow_squash_merge: true,
      delete_branch_on_merge: false,
      allow_update_branch: false,
      topics: ['cli'],
    })
    mockClient.getLabels.mockReturnValue([
      { name: 'bug', color: 'ff0000', description: 'Bug report' },
    ])
    mockClient.getBranchProtection.mockReturnValue({
      required_pull_request_reviews: {
        required_approving_review_count: 2,
        dismiss_stale_reviews: true,
      },
      enforce_admins: { enabled: true },
    })
    mockClient.getSecretNames.mockReturnValue(['API_KEY'])
    mockExistsSync.mockReturnValue(false)
  })

  it('should export to stdout by default', async () => {
    await exportCommand({})

    expect(logger.log).toHaveBeenCalled()
  })

  it('should export to single file', async () => {
    await exportCommand({ single: 'output.yaml' })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'output.yaml',
      expect.any(String),
    )
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('output.yaml'),
    )
  })

  it('should export to directory', async () => {
    await exportCommand({ dir: 'output-dir' })

    expect(mockMkdirSync).toHaveBeenCalledWith('output-dir', {
      recursive: true,
    })
    expect(mockWriteFileSync).toHaveBeenCalled()
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('output-dir'),
    )
  })

  it('should not create directory if exists', async () => {
    mockExistsSync.mockReturnValue(true)

    await exportCommand({ dir: 'existing-dir' })

    expect(mockMkdirSync).not.toHaveBeenCalled()
  })

  it('should export repo settings', async () => {
    await exportCommand({ dir: 'out' })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'out/repo.yaml',
      expect.stringContaining('repo'),
    )
  })

  it('should export topics when present', async () => {
    await exportCommand({ dir: 'out' })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'out/topics.yaml',
      expect.stringContaining('topics'),
    )
  })

  it('should not export topics when empty', async () => {
    mockClient.getRepo.mockReturnValue({
      description: 'Test',
      visibility: 'public',
      topics: [],
    })

    await exportCommand({ dir: 'out' })

    const topicsCall = mockWriteFileSync.mock.calls.find(
      (call) => typeof call[0] === 'string' && call[0].includes('topics'),
    )
    expect(topicsCall).toBeUndefined()
  })

  it('should export labels', async () => {
    await exportCommand({ dir: 'out' })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'out/labels.yaml',
      expect.stringContaining('labels'),
    )
  })

  it('should not export labels when empty', async () => {
    mockClient.getLabels.mockReturnValue([])

    await exportCommand({ dir: 'out' })

    const labelsCall = mockWriteFileSync.mock.calls.find(
      (call) => typeof call[0] === 'string' && call[0].includes('labels'),
    )
    expect(labelsCall).toBeUndefined()
  })

  it('should export branch protection', async () => {
    await exportCommand({ dir: 'out' })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'out/branch-protection.yaml',
      expect.stringContaining('branch_protection'),
    )
  })

  it('should handle missing branch protection', async () => {
    mockClient.getBranchProtection.mockImplementation(() => {
      throw new Error('Not found')
    })

    await exportCommand({ dir: 'out' })

    const bpCall = mockWriteFileSync.mock.calls.find(
      (call) =>
        typeof call[0] === 'string' && call[0].includes('branch-protection'),
    )
    expect(bpCall).toBeUndefined()
  })

  it('should export secrets when includeSecrets is true', async () => {
    await exportCommand({ dir: 'out', includeSecrets: true })

    expect(mockWriteFileSync).toHaveBeenCalledWith(
      'out/secrets.yaml',
      expect.stringContaining('secrets'),
    )
  })

  it('should not export secrets by default', async () => {
    await exportCommand({ dir: 'out' })

    const secretsCall = mockWriteFileSync.mock.calls.find(
      (call) => typeof call[0] === 'string' && call[0].includes('secrets'),
    )
    expect(secretsCall).toBeUndefined()
  })

  it('should handle secrets API error', async () => {
    mockClient.getSecretNames.mockImplementation(() => {
      throw new Error('API error')
    })

    await exportCommand({ dir: 'out', includeSecrets: true })

    const secretsCall = mockWriteFileSync.mock.calls.find(
      (call) => typeof call[0] === 'string' && call[0].includes('secrets'),
    )
    expect(secretsCall).toBeUndefined()
  })

  it('should handle empty secrets', async () => {
    mockClient.getSecretNames.mockReturnValue([])

    await exportCommand({ dir: 'out', includeSecrets: true })

    const secretsCall = mockWriteFileSync.mock.calls.find(
      (call) => typeof call[0] === 'string' && call[0].includes('secrets'),
    )
    expect(secretsCall).toBeUndefined()
  })
})
