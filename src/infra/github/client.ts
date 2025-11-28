import type {
  BranchProtectionConfig,
  GitHubBranchProtectionData,
  GitHubLabelData,
  GitHubRepoData,
  IGitHubClient,
  Label,
  RepoIdentifier,
  RepoSettings,
} from '~/domain'
import { ghApi } from './api'
import type {
  GitHubBranchProtection,
  GitHubLabel,
  GitHubRepo,
  GitHubSecretsResponse,
  GitHubVariablesResponse,
} from './types'

export class GitHubClient implements IGitHubClient {
  constructor(private repo: RepoIdentifier) {}

  private endpoint(path: string): string {
    return `/repos/${this.repo.owner}/${this.repo.name}${path}`
  }

  // Repository
  getRepo(): GitHubRepoData {
    return ghApi<GitHubRepo>('GET', this.endpoint(''))
  }

  updateRepo(data: Partial<RepoSettings>): void {
    ghApi('PATCH', this.endpoint(''), data as Record<string, unknown>)
  }

  // Topics
  setTopics(topics: string[]): void {
    ghApi('PUT', this.endpoint('/topics'), { names: topics })
  }

  // Labels
  getLabels(): GitHubLabelData[] {
    return ghApi<GitHubLabel[]>('GET', this.endpoint('/labels'))
  }

  createLabel(data: Label): void {
    ghApi('POST', this.endpoint('/labels'), {
      name: data.name,
      color: data.color,
      description: data.description,
    })
  }

  updateLabel(name: string, data: Label): void {
    ghApi('PATCH', this.endpoint(`/labels/${encodeURIComponent(name)}`), {
      name: data.name,
      color: data.color,
      description: data.description,
    })
  }

  deleteLabel(name: string): void {
    ghApi('DELETE', this.endpoint(`/labels/${encodeURIComponent(name)}`))
  }

  // Branch Protection
  getBranchProtection(branch: string): GitHubBranchProtectionData {
    return ghApi<GitHubBranchProtection>(
      'GET',
      this.endpoint(`/branches/${branch}/protection`),
    )
  }

  setBranchProtection(branch: string, config: BranchProtectionConfig): void {
    const data: Record<string, unknown> = {
      required_status_checks: config.require_status_checks
        ? {
            strict: true,
            contexts: config.status_checks || [],
          }
        : null,
      enforce_admins: config.enforce_admins ?? false,
      required_pull_request_reviews: config.required_reviews
        ? {
            required_approving_review_count: config.required_reviews,
            dismiss_stale_reviews: config.dismiss_stale_reviews ?? false,
          }
        : null,
      restrictions: null,
      required_linear_history: config.require_linear_history ?? false,
      allow_force_pushes: config.allow_force_pushes ?? false,
      allow_deletions: config.allow_deletions ?? false,
    }
    ghApi('PUT', this.endpoint(`/branches/${branch}/protection`), data)
  }

  // Secrets
  getSecretNames(): string[] {
    const response = ghApi<GitHubSecretsResponse>(
      'GET',
      this.endpoint('/actions/secrets'),
    )
    return response.secrets?.map((s) => s.name) || []
  }

  // Variables
  getVariableNames(): string[] {
    const response = ghApi<GitHubVariablesResponse>(
      'GET',
      this.endpoint('/actions/variables'),
    )
    return response.variables?.map((v) => v.name) || []
  }
}
