import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ghApi } from './api'

vi.mock('node:child_process', () => ({
  spawnSync: vi.fn(),
}))

import { spawnSync } from 'node:child_process'

const mockSpawnSync = vi.mocked(spawnSync)

describe('ghApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call gh api with correct arguments', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: '{"name": "test"}',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    ghApi('GET', '/repos/owner/repo')

    expect(mockSpawnSync).toHaveBeenCalledWith(
      'gh',
      ['api', '-X', 'GET', '/repos/owner/repo'],
      expect.objectContaining({ encoding: 'utf-8' }),
    )
  })

  it('should pass data as JSON input', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: '{}',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    ghApi('PATCH', '/repos/owner/repo', { description: 'Test' })

    expect(mockSpawnSync).toHaveBeenCalledWith(
      'gh',
      expect.arrayContaining(['--input', '-']),
      expect.objectContaining({ input: '{"description":"Test"}' }),
    )
  })

  it('should parse JSON response', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: '{"name": "test", "value": 123}',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    const result = ghApi<{ name: string; value: number }>('GET', '/test')

    expect(result).toEqual({ name: 'test', value: 123 })
  })

  it('should return empty object for empty response', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: '',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    const result = ghApi('DELETE', '/test')

    expect(result).toEqual({})
  })

  it('should return raw string for non-JSON response', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: 'plain text response',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    const result = ghApi<string>('GET', '/test')

    expect(result).toBe('plain text response')
  })

  it('should throw error when gh execution fails', () => {
    mockSpawnSync.mockReturnValue({
      status: 0,
      stdout: '',
      stderr: '',
      error: new Error('Command not found'),
      pid: 1,
      output: [],
      signal: null,
    })

    expect(() => ghApi('GET', '/test')).toThrow('Failed to execute gh')
  })

  it('should throw error when gh api returns non-zero status', () => {
    mockSpawnSync.mockReturnValue({
      status: 1,
      stdout: '',
      stderr: 'Not found',
      pid: 1,
      output: [],
      signal: null,
    })

    expect(() => ghApi('GET', '/test')).toThrow('gh api failed: Not found')
  })

  it('should use stdout for error message if stderr is empty', () => {
    mockSpawnSync.mockReturnValue({
      status: 1,
      stdout: 'Error in stdout',
      stderr: '',
      pid: 1,
      output: [],
      signal: null,
    })

    expect(() => ghApi('GET', '/test')).toThrow(
      'gh api failed: Error in stdout',
    )
  })
})
