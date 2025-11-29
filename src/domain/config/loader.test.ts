import { beforeEach, describe, expect, it, vi } from 'vitest'
import { existsSync, readdirSync, readFileSync } from '~/infra/fs'
import { configToYaml, loadConfig } from './loader'

vi.mock('~/infra/fs', () => ({
  existsSync: vi.fn(),
  readFileSync: vi.fn(),
  readdirSync: vi.fn(),
  join: (...paths: string[]) => paths.join('/'),
}))

vi.mock('~/infra/yaml', () => ({
  parse: vi.fn(),
  stringify: vi.fn((data: unknown) => JSON.stringify(data)),
}))

import { parse } from '~/infra/yaml'

const mockParse = vi.mocked(parse)

const mockExistsSync = vi.mocked(existsSync)
const mockReadFileSync = vi.mocked(readFileSync)
const mockReaddirSync = vi.mocked(readdirSync)

describe('loadConfig', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('with --dir option', () => {
    it('should load config from specified directory', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['repo.yaml'])
      mockReadFileSync.mockReturnValue('repo:\n  description: Test repo')
      mockParse.mockReturnValue({ repo: { description: 'Test repo' } })

      const config = loadConfig({ dir: 'custom/dir' })

      expect(mockExistsSync).toHaveBeenCalledWith('custom/dir')
      expect(mockReaddirSync).toHaveBeenCalledWith('custom/dir')
      expect(config.repo).toBeDefined()
    })

    it('should throw error when directory not found', () => {
      mockExistsSync.mockReturnValue(false)

      expect(() => loadConfig({ dir: 'nonexistent' })).toThrow(
        'Config directory not found',
      )
    })

    it('should throw error when no YAML files in directory', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['readme.md', 'config.json'])

      expect(() => loadConfig({ dir: 'empty-dir' })).toThrow(
        'No YAML files found',
      )
    })
  })

  describe('with --config option', () => {
    it('should load config from specified file', () => {
      mockExistsSync.mockReturnValue(true)
      mockReadFileSync.mockReturnValue('repo:\n  description: Test repo')
      mockParse.mockReturnValue({ repo: { description: 'Test repo' } })

      const config = loadConfig({ config: 'custom/config.yaml' })

      expect(mockExistsSync).toHaveBeenCalledWith('custom/config.yaml')
      expect(config.repo).toBeDefined()
    })

    it('should throw error when file not found', () => {
      mockExistsSync.mockReturnValue(false)

      expect(() => loadConfig({ config: 'nonexistent.yaml' })).toThrow(
        'Config file not found',
      )
    })

    it('should throw error when config is invalid', () => {
      mockExistsSync.mockReturnValue(true)
      mockReadFileSync.mockReturnValue('invalid')
      mockParse.mockReturnValue(null)

      expect(() => loadConfig({ config: 'invalid.yaml' })).toThrow(
        'Invalid config',
      )
    })
  })

  describe('with default locations', () => {
    it('should prefer default directory over default file', () => {
      mockExistsSync.mockImplementation((path) => {
        return path === '.github/repo-settings'
      })
      mockReaddirSync.mockReturnValue(['repo.yaml'])
      mockReadFileSync.mockReturnValue('repo:\n  description: Test repo')
      mockParse.mockReturnValue({ repo: { description: 'Test repo' } })

      const config = loadConfig({})

      expect(mockReaddirSync).toHaveBeenCalledWith('.github/repo-settings')
      expect(config.repo).toBeDefined()
    })

    it('should fall back to default single file', () => {
      mockExistsSync.mockImplementation((path) => {
        return path === '.github/repo-settings.yaml'
      })
      mockReadFileSync.mockReturnValue('repo:\n  description: Test repo')
      mockParse.mockReturnValue({ repo: { description: 'Test repo' } })

      const config = loadConfig({})

      expect(config.repo).toBeDefined()
    })

    it('should throw error when no config found', () => {
      mockExistsSync.mockReturnValue(false)

      expect(() => loadConfig({})).toThrow('No config found')
    })
  })

  describe('directory loading with multiple files', () => {
    it('should load labels.yaml', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['labels.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({
        labels: { items: [{ name: 'bug', color: 'ff0000' }] },
      })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.labels).toBeDefined()
      expect(config.labels?.items).toHaveLength(1)
    })

    it('should load branch-protection.yaml', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['branch-protection.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({
        branch_protection: { main: { required_reviews: 2 } },
      })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.branch_protection).toBeDefined()
    })

    it('should load branch_protection.yaml with underscore', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['branch_protection.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ main: { required_reviews: 2 } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.branch_protection).toBeDefined()
    })

    it('should load secrets.yaml', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['secrets.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ secrets: { required: ['API_KEY'] } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.secrets).toBeDefined()
      expect(config.secrets?.required).toContain('API_KEY')
    })

    it('should load env.yaml', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['env.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ env: { required: ['NODE_ENV'] } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.env).toBeDefined()
      expect(config.env?.required).toContain('NODE_ENV')
    })

    it('should load topics.yaml with topics array', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['topics.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ topics: ['cli', 'github'] })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.topics).toEqual(['cli', 'github'])
    })

    it('should load topics.yaml when file contains raw array', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['topics.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue(['cli', 'github'])

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.topics).toEqual(['cli', 'github'])
    })

    it('should merge unknown files at top level', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['custom.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ repo: { description: 'Custom' } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.repo?.description).toBe('Custom')
    })

    it('should skip invalid parsed content', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['invalid.yaml', 'repo.yaml'])
      mockReadFileSync.mockReturnValue('')
      mockParse
        .mockReturnValueOnce(null)
        .mockReturnValueOnce({ repo: { description: 'Test' } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.repo).toBeDefined()
    })

    it('should handle .yml extension', () => {
      mockExistsSync.mockReturnValue(true)
      mockReaddirSync.mockReturnValue(['repo.yml'])
      mockReadFileSync.mockReturnValue('')
      mockParse.mockReturnValue({ repo: { description: 'Test repo' } })

      const config = loadConfig({ dir: 'test-dir' })

      expect(config.repo).toBeDefined()
    })
  })
})

describe('configToYaml', () => {
  it('should convert config to YAML string', () => {
    const config = { repo: { description: 'Test' } }

    const result = configToYaml(config)

    expect(result).toBeDefined()
    expect(typeof result).toBe('string')
  })
})
