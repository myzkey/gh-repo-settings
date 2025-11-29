import { describe, expect, it } from 'vitest'
import type { GitHubBranchProtectionData, IGitHubClient } from '../ports'
import type { Config } from '../types'
import { calculateBranchProtectionDiffs } from './branch-protection'
import type { DiffContext } from './types'

const mockClient = {} as IGitHubClient
const context: DiffContext = {
  client: mockClient,
  owner: 'test-owner',
  name: 'test-repo',
}

describe('calculateBranchProtectionDiffs', () => {
  it('should return empty array when config has no branch_protection', () => {
    const config: Config = {}
    const protection: GitHubBranchProtectionData = {}

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(0)
  })

  it('should return empty array when branch config is undefined', () => {
    const config: Config = {
      branch_protection: {},
    }
    const protection: GitHubBranchProtectionData = {}

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(0)
  })

  it('should detect missing branch protection', () => {
    const config: Config = {
      branch_protection: {
        main: {
          required_reviews: 2,
        },
      },
    }

    const diffs = calculateBranchProtectionDiffs(config, null, 'main', context)

    expect(diffs).toHaveLength(1)
    expect(diffs[0].type).toBe('branch_protection')
    expect(diffs[0].action).toBe('create')
    expect(diffs[0].details).toContain('does not exist')
    expect(diffs[0].apiCall).toContain('PUT /repos/test-owner/test-repo')
  })

  it('should detect required_reviews change', () => {
    const config: Config = {
      branch_protection: {
        main: {
          required_reviews: 2,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_pull_request_reviews: {
        required_approving_review_count: 1,
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].action).toBe('update')
    expect(diffs[0].details).toContain('required_reviews: 1 -> 2')
  })

  it('should detect required_reviews when currently 0', () => {
    const config: Config = {
      branch_protection: {
        main: {
          required_reviews: 1,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {}

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('required_reviews: 0 -> 1')
  })

  it('should not detect change when required_reviews matches', () => {
    const config: Config = {
      branch_protection: {
        main: {
          required_reviews: 2,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_pull_request_reviews: {
        required_approving_review_count: 2,
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(0)
  })

  it('should detect dismiss_stale_reviews change', () => {
    const config: Config = {
      branch_protection: {
        main: {
          dismiss_stale_reviews: true,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_pull_request_reviews: {
        dismiss_stale_reviews: false,
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('dismiss_stale_reviews: false -> true')
  })

  it('should detect require_status_checks change', () => {
    const config: Config = {
      branch_protection: {
        main: {
          require_status_checks: true,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {}

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('require_status_checks: false -> true')
  })

  it('should detect status_checks difference', () => {
    const config: Config = {
      branch_protection: {
        main: {
          require_status_checks: true,
          status_checks: ['ci', 'lint'],
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_status_checks: {
        contexts: ['ci'],
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('missing: [lint]')
  })

  it('should detect extra status checks', () => {
    const config: Config = {
      branch_protection: {
        main: {
          require_status_checks: true,
          status_checks: ['ci'],
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_status_checks: {
        contexts: ['ci', 'old-check'],
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('extra: [old-check]')
  })

  it('should not detect change when status_checks match', () => {
    const config: Config = {
      branch_protection: {
        main: {
          require_status_checks: true,
          status_checks: ['ci', 'lint'],
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_status_checks: {
        contexts: ['ci', 'lint'],
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(0)
  })

  it('should detect enforce_admins change', () => {
    const config: Config = {
      branch_protection: {
        main: {
          enforce_admins: true,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      enforce_admins: {
        enabled: false,
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(1)
    expect(diffs[0].details).toContain('enforce_admins: false -> true')
  })

  it('should detect multiple changes', () => {
    const config: Config = {
      branch_protection: {
        main: {
          required_reviews: 2,
          dismiss_stale_reviews: true,
          enforce_admins: true,
        },
      },
    }
    const protection: GitHubBranchProtectionData = {
      required_pull_request_reviews: {
        required_approving_review_count: 1,
        dismiss_stale_reviews: false,
      },
      enforce_admins: {
        enabled: false,
      },
    }

    const diffs = calculateBranchProtectionDiffs(
      config,
      protection,
      'main',
      context,
    )

    expect(diffs).toHaveLength(3)
  })
})
