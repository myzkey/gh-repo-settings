import { beforeEach, describe, expect, it, vi } from 'vitest'
import { applyCommand } from './apply'

vi.mock('~/domain', () => ({
  loadConfig: vi.fn(),
  validateConfig: vi.fn(),
  printValidationErrors: vi.fn(),
  printDiff: vi.fn(),
  calculateRepoDiffs: vi.fn(),
  calculateTopicsDiffs: vi.fn(),
  calculateLabelsDiffs: vi.fn(),
  calculateSecretsDiffs: vi.fn(),
  calculateEnvDiffs: vi.fn(),
}))

vi.mock('~/infra/github', () => ({
  getRepoInfo: vi.fn(),
  createClient: vi.fn(),
}))

vi.mock('~/utils/logger', () => ({
  logger: {
    info: vi.fn(),
    success: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}))

import {
  calculateEnvDiffs,
  calculateLabelsDiffs,
  calculateRepoDiffs,
  calculateSecretsDiffs,
  calculateTopicsDiffs,
  loadConfig,
  printDiff,
  printValidationErrors,
  validateConfig,
} from '~/domain'
import { createClient, getRepoInfo } from '~/infra/github'
import { logger } from '~/utils/logger'

const mockLoadConfig = vi.mocked(loadConfig)
const mockValidateConfig = vi.mocked(validateConfig)
const mockGetRepoInfo = vi.mocked(getRepoInfo)
const mockCreateClient = vi.mocked(createClient)
const mockCalculateRepoDiffs = vi.mocked(calculateRepoDiffs)
const mockCalculateTopicsDiffs = vi.mocked(calculateTopicsDiffs)
const mockCalculateLabelsDiffs = vi.mocked(calculateLabelsDiffs)
const mockCalculateSecretsDiffs = vi.mocked(calculateSecretsDiffs)
const mockCalculateEnvDiffs = vi.mocked(calculateEnvDiffs)
const mockPrintDiff = vi.mocked(printDiff)
const mockPrintValidationErrors = vi.mocked(printValidationErrors)

const mockClient = {
  getRepo: vi.fn(),
  updateRepo: vi.fn(),
  setTopics: vi.fn(),
  getLabels: vi.fn(),
  createLabel: vi.fn(),
  updateLabel: vi.fn(),
  deleteLabel: vi.fn(),
  getBranchProtection: vi.fn(),
  setBranchProtection: vi.fn(),
  getSecretNames: vi.fn(),
  getVariableNames: vi.fn(),
}

describe('applyCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetRepoInfo.mockReturnValue({ owner: 'test', name: 'repo' })
    mockCreateClient.mockReturnValue(mockClient as never)
    mockLoadConfig.mockReturnValue({})
    mockValidateConfig.mockReturnValue({ valid: true, errors: [] })
    mockClient.getRepo.mockReturnValue({
      description: '',
      visibility: 'public',
    })
    mockClient.getLabels.mockReturnValue([])
    mockClient.getSecretNames.mockReturnValue([])
    mockClient.getVariableNames.mockReturnValue([])
    mockCalculateRepoDiffs.mockReturnValue([])
    mockCalculateTopicsDiffs.mockReturnValue([])
    mockCalculateLabelsDiffs.mockReturnValue([])
    mockCalculateSecretsDiffs.mockReturnValue([])
    mockCalculateEnvDiffs.mockReturnValue([])
  })

  it('should load config and validate', async () => {
    await applyCommand({ config: 'test.yaml' })

    expect(mockLoadConfig).toHaveBeenCalledWith({
      dir: undefined,
      config: 'test.yaml',
    })
    expect(mockValidateConfig).toHaveBeenCalled()
  })

  it('should exit with error on validation failure', async () => {
    mockValidateConfig.mockReturnValue({
      valid: false,
      errors: [{ path: 'repo', message: 'Invalid' }],
    })

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await applyCommand({})

    expect(mockPrintValidationErrors).toHaveBeenCalled()
    expect(mockExit).toHaveBeenCalledWith(1)

    mockExit.mockRestore()
  })

  it('should print diffs in dry run mode', async () => {
    mockLoadConfig.mockReturnValue({ repo: { description: 'Test' } })
    mockCalculateRepoDiffs.mockReturnValue([
      { type: 'repo', action: 'update', details: 'description changed' },
    ])

    await applyCommand({ dryRun: true })

    expect(mockPrintDiff).toHaveBeenCalled()
    expect(logger.warn).toHaveBeenCalledWith(expect.stringContaining('DRY RUN'))
    expect(mockClient.updateRepo).not.toHaveBeenCalled()
  })

  it('should skip when no changes', async () => {
    await applyCommand({})

    expect(logger.success).toHaveBeenCalledWith('No changes to apply.')
  })

  it('should apply repo changes', async () => {
    mockLoadConfig.mockReturnValue({ repo: { description: 'New' } })
    mockCalculateRepoDiffs.mockReturnValue([
      { type: 'repo', action: 'update', details: 'description' },
    ])

    await applyCommand({})

    expect(mockClient.updateRepo).toHaveBeenCalledWith({ description: 'New' })
  })

  it('should apply topic changes', async () => {
    mockLoadConfig.mockReturnValue({ topics: ['cli', 'github'] })
    mockCalculateTopicsDiffs.mockReturnValue([
      { type: 'topics', action: 'update', details: 'topics' },
    ])

    await applyCommand({})

    expect(mockClient.setTopics).toHaveBeenCalledWith(['cli', 'github'])
  })

  it('should create new labels', async () => {
    mockLoadConfig.mockReturnValue({
      labels: { items: [{ name: 'bug', color: 'ff0000' }] },
    })
    mockClient.getLabels.mockReturnValue([])
    mockCalculateLabelsDiffs.mockReturnValue([
      { type: 'labels', action: 'create', details: 'Create bug' },
    ])

    await applyCommand({})

    expect(mockClient.createLabel).toHaveBeenCalledWith({
      name: 'bug',
      color: 'ff0000',
    })
  })

  it('should update existing labels', async () => {
    mockLoadConfig.mockReturnValue({
      labels: { items: [{ name: 'bug', color: '00ff00' }] },
    })
    mockClient.getLabels.mockReturnValue([
      { name: 'bug', color: 'ff0000', description: '' },
    ])
    mockCalculateLabelsDiffs.mockReturnValue([
      { type: 'labels', action: 'update', details: 'Update bug' },
    ])

    await applyCommand({})

    expect(mockClient.updateLabel).toHaveBeenCalledWith('bug', {
      name: 'bug',
      color: '00ff00',
    })
  })

  it('should delete labels when replace_default is true', async () => {
    mockLoadConfig.mockReturnValue({
      labels: { replace_default: true, items: [] },
    })
    mockClient.getLabels.mockReturnValue([
      { name: 'old', color: 'ff0000', description: '' },
    ])
    mockCalculateLabelsDiffs.mockReturnValue([
      { type: 'labels', action: 'delete', details: 'Delete old' },
    ])

    await applyCommand({})

    expect(mockClient.deleteLabel).toHaveBeenCalledWith('old')
  })

  it('should apply branch protection', async () => {
    mockLoadConfig.mockReturnValue({
      branch_protection: { main: { required_reviews: 2 } },
    })

    await applyCommand({})

    expect(mockClient.setBranchProtection).toHaveBeenCalledWith('main', {
      required_reviews: 2,
    })
  })

  it('should warn about missing secrets', async () => {
    mockLoadConfig.mockReturnValue({
      secrets: { required: ['API_KEY'] },
    })
    mockCalculateSecretsDiffs.mockReturnValue([
      { type: 'secrets', action: 'check', details: 'Missing API_KEY' },
    ])

    await applyCommand({})

    expect(logger.warn).toHaveBeenCalledWith(expect.stringContaining('Secret'))
  })

  it('should warn about missing env variables', async () => {
    mockLoadConfig.mockReturnValue({
      env: { required: ['NODE_ENV'] },
    })
    mockCalculateEnvDiffs.mockReturnValue([
      { type: 'env', action: 'check', details: 'Missing NODE_ENV' },
    ])

    await applyCommand({})

    expect(logger.warn).toHaveBeenCalledWith(
      expect.stringContaining('Environment'),
    )
  })

  it('should handle secret API errors', async () => {
    mockLoadConfig.mockReturnValue({
      secrets: { required: ['API_KEY'] },
    })
    mockClient.getSecretNames.mockImplementation(() => {
      throw new Error('API error')
    })

    await applyCommand({})

    expect(logger.warn).toHaveBeenCalled()
  })

  it('should handle env API errors', async () => {
    mockLoadConfig.mockReturnValue({
      env: { required: ['NODE_ENV'] },
    })
    mockClient.getVariableNames.mockImplementation(() => {
      throw new Error('API error')
    })

    await applyCommand({})

    expect(logger.warn).toHaveBeenCalled()
  })
})
