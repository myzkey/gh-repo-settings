import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createClient, getCurrentRepo, getRepoInfo, parseRepoArg } from './repo'

vi.mock('node:child_process', () => ({
  execSync: vi.fn(),
}))

vi.mock('./client', () => ({
  GitHubClient: vi.fn().mockImplementation(function (
    this: { repo: unknown },
    repo: unknown,
  ) {
    this.repo = repo
  }),
}))

import { execSync } from 'node:child_process'
import { GitHubClient } from './client'

const mockExecSync = vi.mocked(execSync)

describe('getCurrentRepo', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return repo info from gh repo view', () => {
    mockExecSync.mockReturnValue('owner/repo-name\n')

    const result = getCurrentRepo()

    expect(result).toEqual({ owner: 'owner', name: 'repo-name' })
  })

  it('should return null for empty result', () => {
    mockExecSync.mockReturnValue('')

    const result = getCurrentRepo()

    expect(result).toBeNull()
  })

  it('should return null for invalid format', () => {
    mockExecSync.mockReturnValue('invalid')

    const result = getCurrentRepo()

    expect(result).toBeNull()
  })

  it('should return null when gh command fails', () => {
    mockExecSync.mockImplementation(() => {
      throw new Error('Not a git repository')
    })

    const result = getCurrentRepo()

    expect(result).toBeNull()
  })
})

describe('parseRepoArg', () => {
  it('should parse valid owner/name format', () => {
    const result = parseRepoArg('owner/repo-name')

    expect(result).toEqual({ owner: 'owner', name: 'repo-name' })
  })

  it('should throw error for invalid format', () => {
    expect(() => parseRepoArg('invalid')).toThrow(
      'Invalid repo format: invalid. Expected: owner/name',
    )
  })

  it('should throw error for too many slashes', () => {
    expect(() => parseRepoArg('a/b/c')).toThrow('Invalid repo format')
  })
})

describe('getRepoInfo', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should parse repo arg when provided', () => {
    const result = getRepoInfo('owner/repo')

    expect(result).toEqual({ owner: 'owner', name: 'repo' })
  })

  it('should get current repo when no arg provided', () => {
    mockExecSync.mockReturnValue('current-owner/current-repo\n')

    const result = getRepoInfo()

    expect(result).toEqual({ owner: 'current-owner', name: 'current-repo' })
  })

  it('should throw error when no repo can be determined', () => {
    mockExecSync.mockImplementation(() => {
      throw new Error('Not a git repository')
    })

    expect(() => getRepoInfo()).toThrow('Could not determine repository')
  })
})

describe('createClient', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should create GitHubClient with repo info', () => {
    const client = createClient('owner/repo')

    expect(GitHubClient).toHaveBeenCalledWith({ owner: 'owner', name: 'repo' })
    expect(client).toBeDefined()
  })

  it('should use current repo when no arg provided', () => {
    mockExecSync.mockReturnValue('current/repo\n')

    const client = createClient()

    expect(GitHubClient).toHaveBeenCalledWith({
      owner: 'current',
      name: 'repo',
    })
    expect(client).toBeDefined()
  })
})
