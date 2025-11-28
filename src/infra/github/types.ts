// GitHub API response types

export interface GitHubRepo {
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

export interface GitHubLabel {
  name: string
  color: string
  description: string | null
}

export interface GitHubBranchProtection {
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

export interface GitHubSecret {
  name: string
}

export interface GitHubVariable {
  name: string
}

export interface GitHubSecretsResponse {
  secrets: GitHubSecret[]
}

export interface GitHubVariablesResponse {
  variables: GitHubVariable[]
}
