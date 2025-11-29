import { beforeEach, describe, expect, it, vi } from 'vitest'
import { GitHubClient } from './client'

vi.mock('./api', () => ({
  ghApi: vi.fn(),
}))

import { ghApi } from './api'

const mockGhApi = vi.mocked(ghApi)

describe('GitHubClient', () => {
  let client: GitHubClient

  beforeEach(() => {
    vi.clearAllMocks()
    client = new GitHubClient({ owner: 'test-owner', name: 'test-repo' })
  })

  describe('getRepo', () => {
    it('should call GET on repo endpoint', () => {
      mockGhApi.mockReturnValue({
        description: 'Test',
        visibility: 'public',
      })

      const result = client.getRepo()

      expect(mockGhApi).toHaveBeenCalledWith(
        'GET',
        '/repos/test-owner/test-repo',
      )
      expect(result.description).toBe('Test')
    })
  })

  describe('updateRepo', () => {
    it('should call PATCH with repo data', () => {
      client.updateRepo({ description: 'Updated' })

      expect(mockGhApi).toHaveBeenCalledWith(
        'PATCH',
        '/repos/test-owner/test-repo',
        { description: 'Updated' },
      )
    })
  })

  describe('setTopics', () => {
    it('should call PUT with topics', () => {
      client.setTopics(['cli', 'github'])

      expect(mockGhApi).toHaveBeenCalledWith(
        'PUT',
        '/repos/test-owner/test-repo/topics',
        { names: ['cli', 'github'] },
      )
    })
  })

  describe('getLabels', () => {
    it('should return labels from API', () => {
      mockGhApi.mockReturnValue([
        { name: 'bug', color: 'ff0000', description: 'Bug' },
      ])

      const result = client.getLabels()

      expect(mockGhApi).toHaveBeenCalledWith(
        'GET',
        '/repos/test-owner/test-repo/labels',
      )
      expect(result).toHaveLength(1)
      expect(result[0].name).toBe('bug')
    })
  })

  describe('createLabel', () => {
    it('should call POST with label data', () => {
      client.createLabel({ name: 'bug', color: 'ff0000', description: 'Bug' })

      expect(mockGhApi).toHaveBeenCalledWith(
        'POST',
        '/repos/test-owner/test-repo/labels',
        { name: 'bug', color: 'ff0000', description: 'Bug' },
      )
    })
  })

  describe('updateLabel', () => {
    it('should call PATCH with label data', () => {
      client.updateLabel('old-name', {
        name: 'new-name',
        color: '00ff00',
        description: 'Updated',
      })

      expect(mockGhApi).toHaveBeenCalledWith(
        'PATCH',
        '/repos/test-owner/test-repo/labels/old-name',
        { name: 'new-name', color: '00ff00', description: 'Updated' },
      )
    })

    it('should encode special characters in label name', () => {
      client.updateLabel('bug/critical', { name: 'bug', color: 'ff0000' })

      expect(mockGhApi).toHaveBeenCalledWith(
        'PATCH',
        '/repos/test-owner/test-repo/labels/bug%2Fcritical',
        expect.any(Object),
      )
    })
  })

  describe('deleteLabel', () => {
    it('should call DELETE on label endpoint', () => {
      client.deleteLabel('old-label')

      expect(mockGhApi).toHaveBeenCalledWith(
        'DELETE',
        '/repos/test-owner/test-repo/labels/old-label',
      )
    })
  })

  describe('getBranchProtection', () => {
    it('should return branch protection data', () => {
      mockGhApi.mockReturnValue({
        required_pull_request_reviews: {
          required_approving_review_count: 2,
        },
      })

      const result = client.getBranchProtection('main')

      expect(mockGhApi).toHaveBeenCalledWith(
        'GET',
        '/repos/test-owner/test-repo/branches/main/protection',
      )
      expect(
        result.required_pull_request_reviews?.required_approving_review_count,
      ).toBe(2)
    })
  })

  describe('setBranchProtection', () => {
    it('should call PUT with protection config', () => {
      client.setBranchProtection('main', {
        required_reviews: 2,
        dismiss_stale_reviews: true,
        enforce_admins: true,
      })

      expect(mockGhApi).toHaveBeenCalledWith(
        'PUT',
        '/repos/test-owner/test-repo/branches/main/protection',
        expect.objectContaining({
          required_pull_request_reviews: {
            required_approving_review_count: 2,
            dismiss_stale_reviews: true,
          },
          enforce_admins: true,
        }),
      )
    })

    it('should set status checks when enabled', () => {
      client.setBranchProtection('main', {
        require_status_checks: true,
        status_checks: ['ci', 'lint'],
      })

      expect(mockGhApi).toHaveBeenCalledWith(
        'PUT',
        expect.any(String),
        expect.objectContaining({
          required_status_checks: {
            strict: true,
            contexts: ['ci', 'lint'],
          },
        }),
      )
    })

    it('should set null for disabled features', () => {
      client.setBranchProtection('main', {})

      expect(mockGhApi).toHaveBeenCalledWith(
        'PUT',
        expect.any(String),
        expect.objectContaining({
          required_status_checks: null,
          required_pull_request_reviews: null,
        }),
      )
    })
  })

  describe('getSecretNames', () => {
    it('should return secret names', () => {
      mockGhApi.mockReturnValue({
        secrets: [{ name: 'API_KEY' }, { name: 'DB_PASSWORD' }],
      })

      const result = client.getSecretNames()

      expect(mockGhApi).toHaveBeenCalledWith(
        'GET',
        '/repos/test-owner/test-repo/actions/secrets',
      )
      expect(result).toEqual(['API_KEY', 'DB_PASSWORD'])
    })

    it('should return empty array when no secrets', () => {
      mockGhApi.mockReturnValue({})

      const result = client.getSecretNames()

      expect(result).toEqual([])
    })
  })

  describe('getVariableNames', () => {
    it('should return variable names', () => {
      mockGhApi.mockReturnValue({
        variables: [{ name: 'NODE_ENV' }, { name: 'API_URL' }],
      })

      const result = client.getVariableNames()

      expect(mockGhApi).toHaveBeenCalledWith(
        'GET',
        '/repos/test-owner/test-repo/actions/variables',
      )
      expect(result).toEqual(['NODE_ENV', 'API_URL'])
    })

    it('should return empty array when no variables', () => {
      mockGhApi.mockReturnValue({})

      const result = client.getVariableNames()

      expect(result).toEqual([])
    })
  })
})
