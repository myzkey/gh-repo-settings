import { beforeEach, describe, expect, it, vi } from 'vitest'
import { loadAndValidateConfig, printValidationErrors } from './validator'

vi.mock('~/utils/logger', () => ({
  logger: {
    error: vi.fn(),
    log: vi.fn(),
  },
}))

vi.mock('./loader', () => ({
  loadConfig: vi.fn(),
}))

vi.mock('./schema/config', () => ({
  validateConfig: vi.fn(),
}))

import { logger } from '~/utils/logger'
import { loadConfig } from './loader'
import { validateConfig } from './schema/config'

const mockLoadConfig = vi.mocked(loadConfig)
const mockValidateConfig = vi.mocked(validateConfig)

describe('loadAndValidateConfig', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return config when validation passes', () => {
    const config = { repo: { description: 'Test' } }
    mockLoadConfig.mockReturnValue(config)
    mockValidateConfig.mockReturnValue({ valid: true, errors: [] })

    const result = loadAndValidateConfig({ config: 'test.yaml' })

    expect(result).toEqual(config)
    expect(mockLoadConfig).toHaveBeenCalledWith({ config: 'test.yaml' })
    expect(mockValidateConfig).toHaveBeenCalledWith(config)
  })

  it('should throw error when validation fails', () => {
    const config = { repo: { visibility: 'invalid' } }
    mockLoadConfig.mockReturnValue(config)
    mockValidateConfig.mockReturnValue({
      valid: false,
      errors: [{ path: 'repo.visibility', message: 'Invalid value' }],
    })

    expect(() => loadAndValidateConfig({ config: 'test.yaml' })).toThrow(
      'Config validation failed',
    )
  })

  it('should print validation errors when validation fails', () => {
    const config = { repo: { visibility: 'invalid' } }
    mockLoadConfig.mockReturnValue(config)
    mockValidateConfig.mockReturnValue({
      valid: false,
      errors: [{ path: 'repo.visibility', message: 'Invalid value' }],
    })

    try {
      loadAndValidateConfig({ config: 'test.yaml' })
    } catch {
      // Expected
    }

    expect(logger.error).toHaveBeenCalled()
  })
})

describe('printValidationErrors', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should print error header', () => {
    const result = {
      valid: false,
      errors: [{ path: 'repo', message: 'Invalid' }],
    }

    printValidationErrors(result)

    expect(logger.error).toHaveBeenCalledWith('\nConfig validation failed:\n')
  })

  it('should print each error with path', () => {
    const result = {
      valid: false,
      errors: [
        { path: 'repo.visibility', message: 'Must be public or private' },
        { path: 'labels.items', message: 'Invalid color format' },
      ],
    }

    printValidationErrors(result)

    expect(logger.error).toHaveBeenCalledWith(
      '  - repo.visibility: Must be public or private',
    )
    expect(logger.error).toHaveBeenCalledWith(
      '  - labels.items: Invalid color format',
    )
  })

  it('should use (root) for errors without path', () => {
    const result = {
      valid: false,
      errors: [{ message: 'Invalid config' }],
    }

    printValidationErrors(result)

    expect(logger.error).toHaveBeenCalledWith('  - (root): Invalid config')
  })

  it('should print trailing newline', () => {
    const result = {
      valid: false,
      errors: [{ path: 'repo', message: 'Invalid' }],
    }

    printValidationErrors(result)

    expect(logger.log).toHaveBeenCalledWith('')
  })
})
