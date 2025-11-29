import { describe, expect, it } from 'vitest'
import { parse, stringify } from './parser'

describe('parse', () => {
  it('should parse simple YAML', () => {
    const yaml = 'name: test\nvalue: 123'

    const result = parse<{ name: string; value: number }>(yaml)

    expect(result).toEqual({ name: 'test', value: 123 })
  })

  it('should parse nested YAML', () => {
    const yaml = `
repo:
  description: Test
  visibility: public
`

    const result = parse<{ repo: { description: string; visibility: string } }>(
      yaml,
    )

    expect(result.repo.description).toBe('Test')
    expect(result.repo.visibility).toBe('public')
  })

  it('should parse arrays', () => {
    const yaml = `
topics:
  - cli
  - github
`

    const result = parse<{ topics: string[] }>(yaml)

    expect(result.topics).toEqual(['cli', 'github'])
  })

  it('should return undefined for empty content', () => {
    const result = parse('')

    expect(result).toBeUndefined()
  })
})

describe('stringify', () => {
  it('should convert object to YAML', () => {
    const data = { name: 'test', value: 123 }

    const result = stringify(data)

    expect(result).toContain('name: test')
    expect(result).toContain('value: 123')
  })

  it('should use 2-space indentation by default', () => {
    const data = { repo: { description: 'Test' } }

    const result = stringify(data)

    expect(result).toBe('repo:\n  description: Test\n')
  })

  it('should respect custom indent option', () => {
    const data = { repo: { description: 'Test' } }

    const result = stringify(data, { indent: 4 })

    expect(result).toBe('repo:\n    description: Test\n')
  })

  it('should handle arrays', () => {
    const data = { topics: ['cli', 'github'] }

    const result = stringify(data)

    expect(result).toContain('topics:')
    expect(result).toContain('- cli')
    expect(result).toContain('- github')
  })

  it('should not sort keys by default', () => {
    const data = { zebra: 1, apple: 2 }

    const result = stringify(data)

    expect(result.indexOf('zebra')).toBeLessThan(result.indexOf('apple'))
  })

  it('should sort keys when option is set', () => {
    const data = { zebra: 1, apple: 2 }

    const result = stringify(data, { sortKeys: true })

    expect(result.indexOf('apple')).toBeLessThan(result.indexOf('zebra'))
  })
})
