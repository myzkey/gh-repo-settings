import { describe, expect, it } from 'vitest'
import type { Config } from '../types'
import { calculateSecretsDiffs } from './secrets'

describe('calculateSecretsDiffs', () => {
  it('should return empty array when config.secrets is undefined', () => {
    const config: Config = {}
    const existingSecrets: string[] = ['SECRET_1']

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when required secrets is empty', () => {
    const config: Config = {
      secrets: {
        required: [],
      },
    }
    const existingSecrets: string[] = ['SECRET_1']

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when all required secrets exist', () => {
    const config: Config = {
      secrets: {
        required: ['SECRET_1', 'SECRET_2'],
      },
    }
    const existingSecrets: string[] = ['SECRET_1', 'SECRET_2', 'SECRET_3']

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(0)
  })

  it('should detect missing secret', () => {
    const config: Config = {
      secrets: {
        required: ['SECRET_1', 'MISSING_SECRET'],
      },
    }
    const existingSecrets: string[] = ['SECRET_1']

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('secrets')
    expect(diffs[0].action).toBe('check')
    expect(diffs[0].details).toContain('MISSING_SECRET')
    expect(diffs[0].details).toContain('gh secret set')
  })

  it('should detect multiple missing secrets', () => {
    const config: Config = {
      secrets: {
        required: ['SECRET_1', 'SECRET_2', 'SECRET_3'],
      },
    }
    const existingSecrets: string[] = ['SECRET_1']

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(2)
    expect(diffs[0].details).toContain('SECRET_2')
    expect(diffs[1].details).toContain('SECRET_3')
  })

  it('should handle empty existing secrets', () => {
    const config: Config = {
      secrets: {
        required: ['SECRET_1'],
      },
    }
    const existingSecrets: string[] = []

    const diffs = calculateSecretsDiffs(config, existingSecrets)

    expect(diffs).toHaveLength(1)
  })
})
