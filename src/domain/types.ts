// Domain types - Configuration and comparison

export interface RepoSettings {
  description?: string
  homepage?: string
  visibility?: 'public' | 'private' | 'internal'
  allow_merge_commit?: boolean
  allow_rebase_merge?: boolean
  allow_squash_merge?: boolean
  delete_branch_on_merge?: boolean
  allow_update_branch?: boolean
}

export interface Label {
  name: string
  color: string
  description?: string
}

export interface LabelsConfig {
  replace_default?: boolean
  items: Label[]
}

export interface BranchProtectionConfig {
  required_reviews?: number
  dismiss_stale_reviews?: boolean
  require_status_checks?: boolean
  status_checks?: string[]
  enforce_admins?: boolean
  require_linear_history?: boolean
  allow_force_pushes?: boolean
  allow_deletions?: boolean
}

export interface SecretsConfig {
  required?: string[]
}

export interface EnvConfig {
  required?: string[]
}

export interface Config {
  repo?: RepoSettings
  topics?: string[]
  labels?: LabelsConfig
  branch_protection?: {
    [branch: string]: BranchProtectionConfig
  }
  secrets?: SecretsConfig
  env?: EnvConfig
}

export type DiffType =
  | 'repo'
  | 'topics'
  | 'labels'
  | 'branch_protection'
  | 'secrets'
  | 'env'

export type DiffAction = 'create' | 'update' | 'delete' | 'check'

export interface DiffItem {
  type: DiffType
  action: DiffAction
  details: string
  apiCall?: string
}

export interface ValidationError {
  path: string
  message: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
}
