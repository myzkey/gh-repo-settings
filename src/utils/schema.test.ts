import { describe, expect, it } from 'vitest'
import { validateConfig } from '~/utils/schema'

describe('validateConfig', () => {
  it('should validate empty config', () => {
    const result = validateConfig({})
    expect(result.valid).toBe(true)
    expect(result.errors).toHaveLength(0)
  })

  it('should validate valid repo config', () => {
    const result = validateConfig({
      repo: {
        description: 'Test repo',
        visibility: 'public',
      },
    })
    expect(result.valid).toBe(true)
  })

  it('should reject invalid visibility', () => {
    const result = validateConfig({
      repo: {
        visibility: 'invalid',
      },
    })
    expect(result.valid).toBe(false)
    expect(result.errors.length).toBeGreaterThan(0)
  })

  it('should validate valid label config', () => {
    const result = validateConfig({
      labels: {
        items: [{ name: 'bug', color: 'ff0000' }],
      },
    })
    expect(result.valid).toBe(true)
  })

  it('should reject invalid hex color', () => {
    const result = validateConfig({
      labels: {
        items: [{ name: 'bug', color: '#ff0000' }],
      },
    })
    expect(result.valid).toBe(false)
  })

  it('should validate topics', () => {
    const result = validateConfig({
      topics: ['cli', 'github'],
    })
    expect(result.valid).toBe(true)
  })

  it('should reject invalid topic format', () => {
    const result = validateConfig({
      topics: ['Invalid Topic'],
    })
    expect(result.valid).toBe(false)
  })
})
