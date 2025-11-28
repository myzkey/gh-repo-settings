import type { BranchProtectionConfig, Label, RepoSettings } from './types'

export interface RepoIdentifier {
  owner: string
  name: string
}

export interface GitHubRepoData {
  description: string | null
  homepage: string | null
  visibility: string
  allow_merge_commit: boolean
  allow_rebase_merge: boolean
  allow_squash_merge: boolean
  delete_branch_on_merge: boolean
  allow_update_branch: boolean
  topics: string[]
}

export interface GitHubLabelData {
  name: string
  color: string
  description: string | null
}

export interface GitHubBranchProtectionData {
  required_pull_request_reviews?: {
    required_approving_review_count?: number
    dismiss_stale_reviews?: boolean
  }
  required_status_checks?: {
    strict?: boolean
    contexts?: string[]
  }
  enforce_admins?: {
    enabled: boolean
  }
  required_linear_history?: {
    enabled: boolean
  }
  allow_force_pushes?: {
    enabled: boolean
  }
  allow_deletions?: {
    enabled: boolean
  }
}

export interface IGitHubClient {
  // Repository
  getRepo(): GitHubRepoData
  updateRepo(data: Partial<RepoSettings>): void

  // Topics
  setTopics(topics: string[]): void

  // Labels
  getLabels(): GitHubLabelData[]
  createLabel(data: Label): void
  updateLabel(name: string, data: Label): void
  deleteLabel(name: string): void

  // Branch Protection
  getBranchProtection(branch: string): GitHubBranchProtectionData
  setBranchProtection(branch: string, data: BranchProtectionConfig): void

  // Secrets
  getSecretNames(): string[]

  // Variables
  getVariableNames(): string[]
}
