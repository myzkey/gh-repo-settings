import { describe, expect, it } from 'vitest'
import type { Config } from '../types'
import { calculateEnvDiffs } from './env'

describe('calculateEnvDiffs', () => {
  it('should return empty array when config.env is undefined', () => {
    const config: Config = {}
    const existingVars: string[] = ['VAR_1']

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when required env is empty', () => {
    const config: Config = {
      env: {
        required: [],
      },
    }
    const existingVars: string[] = ['VAR_1']

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when all required env vars exist', () => {
    const config: Config = {
      env: {
        required: ['VAR_1', 'VAR_2'],
      },
    }
    const existingVars: string[] = ['VAR_1', 'VAR_2', 'VAR_3']

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(0)
  })

  it('should detect missing env variable', () => {
    const config: Config = {
      env: {
        required: ['VAR_1', 'MISSING_VAR'],
      },
    }
    const existingVars: string[] = ['VAR_1']

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('env')
    expect(diffs[0].action).toBe('check')
    expect(diffs[0].details).toContain('MISSING_VAR')
    expect(diffs[0].details).toContain('gh variable set')
  })

  it('should detect multiple missing env variables', () => {
    const config: Config = {
      env: {
        required: ['VAR_1', 'VAR_2', 'VAR_3'],
      },
    }
    const existingVars: string[] = ['VAR_1']

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(2)
    expect(diffs[0].details).toContain('VAR_2')
    expect(diffs[1].details).toContain('VAR_3')
  })

  it('should handle empty existing env variables', () => {
    const config: Config = {
      env: {
        required: ['VAR_1'],
      },
    }
    const existingVars: string[] = []

    const diffs = calculateEnvDiffs(config, existingVars)

    expect(diffs).toHaveLength(1)
  })
})
