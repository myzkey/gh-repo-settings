export interface RepoSettings {
  description?: string;
  homepage?: string;
  visibility?: "public" | "private" | "internal";
  allow_merge_commit?: boolean;
  allow_rebase_merge?: boolean;
  allow_squash_merge?: boolean;
  delete_branch_on_merge?: boolean;
  allow_update_branch?: boolean;
}

export interface Label {
  name: string;
  color: string;
  description?: string;
}

export interface LabelsConfig {
  replace_default?: boolean;
  items: Label[];
}

export interface BranchProtectionConfig {
  required_reviews?: number;
  dismiss_stale_reviews?: boolean;
  require_status_checks?: boolean;
  status_checks?: string[];
  enforce_admins?: boolean;
  require_linear_history?: boolean;
  allow_force_pushes?: boolean;
  allow_deletions?: boolean;
}

export interface SecretsConfig {
  required?: string[];
}

export interface Config {
  repo?: RepoSettings;
  topics?: string[];
  labels?: LabelsConfig;
  branch_protection?: {
    [branch: string]: BranchProtectionConfig;
  };
  secrets?: SecretsConfig;
}

export interface RepoInfo {
  owner: string;
  name: string;
}

export interface DiffItem {
  type: "repo" | "topics" | "labels" | "branch_protection" | "secrets";
  action: "create" | "update" | "delete" | "check";
  details: string;
  apiCall?: string;
}
