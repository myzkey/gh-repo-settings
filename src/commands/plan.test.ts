import { beforeEach, describe, expect, it, vi } from 'vitest'
import { planCommand } from './plan'

vi.mock('~/domain', () => ({
  loadConfig: vi.fn(),
  validateConfig: vi.fn(),
  printValidationErrors: vi.fn(),
  printDiff: vi.fn(),
  calculateRepoDiffs: vi.fn(),
  calculateTopicsDiffs: vi.fn(),
  calculateLabelsDiffs: vi.fn(),
  calculateBranchProtectionDiffs: vi.fn(),
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
    debug: vi.fn(),
  },
}))

import {
  calculateBranchProtectionDiffs,
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
const mockPrintDiff = vi.mocked(printDiff)
const mockPrintValidationErrors = vi.mocked(printValidationErrors)
const mockCalculateRepoDiffs = vi.mocked(calculateRepoDiffs)
const mockCalculateTopicsDiffs = vi.mocked(calculateTopicsDiffs)
const mockCalculateLabelsDiffs = vi.mocked(calculateLabelsDiffs)
const mockCalculateBranchProtectionDiffs = vi.mocked(
  calculateBranchProtectionDiffs,
)
const mockCalculateSecretsDiffs = vi.mocked(calculateSecretsDiffs)
const mockCalculateEnvDiffs = vi.mocked(calculateEnvDiffs)

const mockClient = {
  getRepo: vi.fn(),
  getLabels: vi.fn(),
  getBranchProtection: vi.fn(),
  getSecretNames: vi.fn(),
  getVariableNames: vi.fn(),
}

describe('planCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetRepoInfo.mockReturnValue({ owner: 'test', name: 'repo' })
    mockCreateClient.mockReturnValue(mockClient as never)
    mockLoadConfig.mockReturnValue({})
    mockValidateConfig.mockReturnValue({ valid: true, errors: [] })
    mockClient.getRepo.mockReturnValue({
      description: '',
      visibility: 'public',
      topics: [],
    })
    mockClient.getLabels.mockReturnValue([])
    mockClient.getBranchProtection.mockReturnValue({})
    mockClient.getSecretNames.mockReturnValue([])
    mockClient.getVariableNames.mockReturnValue([])
    mockCalculateRepoDiffs.mockReturnValue([])
    mockCalculateTopicsDiffs.mockReturnValue([])
    mockCalculateLabelsDiffs.mockReturnValue([])
    mockCalculateBranchProtectionDiffs.mockReturnValue([])
    mockCalculateSecretsDiffs.mockReturnValue([])
    mockCalculateEnvDiffs.mockReturnValue([])
  })

  it('should validate schema first', async () => {
    await planCommand({ config: 'test.yaml' })

    expect(mockLoadConfig).toHaveBeenCalled()
    expect(mockValidateConfig).toHaveBeenCalled()
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('validation passed'),
    )
  })

  it('should exit on validation failure', async () => {
    mockValidateConfig.mockReturnValue({
      valid: false,
      errors: [{ message: 'Invalid' }],
    })

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await planCommand({})

    expect(mockPrintValidationErrors).toHaveBeenCalled()
    expect(mockExit).toHaveBeenCalledWith(1)

    mockExit.mockRestore()
  })

  it('should stop at schema-only mode', async () => {
    await planCommand({ schemaOnly: true })

    expect(mockCreateClient).not.toHaveBeenCalled()
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('validation passed'),
    )
  })

  it('should show all settings in sync', async () => {
    await planCommand({})

    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('in sync'),
    )
  })

  it('should show diffs when changes detected', async () => {
    mockLoadConfig.mockReturnValue({ repo: { description: 'New' } })
    mockCalculateRepoDiffs.mockReturnValue([
      { type: 'repo', action: 'update', details: 'description changed' },
    ])

    await planCommand({})

    expect(logger.warn).toHaveBeenCalledWith(expect.stringContaining('1'))
    expect(mockPrintDiff).toHaveBeenCalled()
  })

  it('should check only secrets when --secrets flag', async () => {
    mockLoadConfig.mockReturnValue({
      secrets: { required: ['API_KEY'] },
    })
    mockCalculateSecretsDiffs.mockReturnValue([])

    await planCommand({ secrets: true })

    expect(mockCalculateRepoDiffs).not.toHaveBeenCalled()
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('secrets'),
    )
  })

  it('should check only env when --env flag', async () => {
    mockLoadConfig.mockReturnValue({
      env: { required: ['NODE_ENV'] },
    })
    mockCalculateEnvDiffs.mockReturnValue([])

    await planCommand({ env: true })

    expect(mockCalculateRepoDiffs).not.toHaveBeenCalled()
    expect(logger.success).toHaveBeenCalledWith(
      expect.stringContaining('environment'),
    )
  })

  it('should handle branch protection fetch error', async () => {
    mockLoadConfig.mockReturnValue({
      branch_protection: { main: { required_reviews: 2 } },
    })
    mockClient.getBranchProtection.mockImplementation(() => {
      throw new Error('Not found')
    })
    mockCalculateBranchProtectionDiffs.mockReturnValue([
      { type: 'branch_protection', action: 'create', details: 'Create' },
    ])

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await planCommand({})

    expect(mockCalculateBranchProtectionDiffs).toHaveBeenCalledWith(
      expect.any(Object),
      null,
      'main',
      expect.any(Object),
    )

    mockExit.mockRestore()
  })

  it('should handle secrets API error', async () => {
    mockLoadConfig.mockReturnValue({
      secrets: { required: ['API_KEY'] },
    })
    mockClient.getSecretNames.mockImplementation(() => {
      throw new Error('API error')
    })

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await planCommand({ secrets: true })

    expect(mockPrintDiff).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          type: 'secrets',
          action: 'check',
        }),
      ]),
    )

    mockExit.mockRestore()
  })

  it('should handle env API error', async () => {
    mockLoadConfig.mockReturnValue({
      env: { required: ['NODE_ENV'] },
    })
    mockClient.getVariableNames.mockImplementation(() => {
      throw new Error('API error')
    })

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await planCommand({ env: true })

    expect(mockPrintDiff).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          type: 'env',
          action: 'check',
        }),
      ]),
    )

    mockExit.mockRestore()
  })

  it('should exit with error when issues found', async () => {
    mockLoadConfig.mockReturnValue({
      secrets: { required: ['MISSING'] },
    })
    mockCalculateSecretsDiffs.mockReturnValue([
      { type: 'secrets', action: 'check', details: 'Missing MISSING' },
    ])

    const mockExit = vi
      .spyOn(process, 'exit')
      .mockImplementation(() => undefined as never)

    await planCommand({ secrets: true })

    expect(mockExit).toHaveBeenCalledWith(1)

    mockExit.mockRestore()
  })
})
