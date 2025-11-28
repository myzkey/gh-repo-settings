import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import yaml from "js-yaml";
import type {
  Config,
  RepoSettings,
  Label,
  BranchProtectionConfig,
} from "~/types";
import { colors } from "~/utils/colors";
import { ghApiGet, getRepoInfo } from "~/utils/gh";

interface GitHubRepo {
  description: string | null;
  homepage: string | null;
  visibility: string;
  allow_merge_commit: boolean;
  allow_rebase_merge: boolean;
  allow_squash_merge: boolean;
  delete_branch_on_merge: boolean;
  allow_update_branch: boolean;
  topics: string[];
}

interface GitHubLabel {
  name: string;
  color: string;
  description: string | null;
}

interface GitHubBranchProtection {
  required_pull_request_reviews?: {
    required_approving_review_count?: number;
    dismiss_stale_reviews?: boolean;
  };
  required_status_checks?: {
    strict?: boolean;
    contexts?: string[];
  };
  enforce_admins?: {
    enabled: boolean;
  };
  required_linear_history?: {
    enabled: boolean;
  };
  allow_force_pushes?: {
    enabled: boolean;
  };
  allow_deletions?: {
    enabled: boolean;
  };
}

interface GitHubSecret {
  name: string;
}

interface ExportOptions {
  repo?: string;
  dir?: string;
  single?: string;
  includeSecrets?: boolean;
}

export async function exportCommand(options: ExportOptions): Promise<void> {
  const repoInfo = getRepoInfo(options.repo);
  const { owner, name } = repoInfo;

  console.log(colors.blue(`Exporting settings from ${owner}/${name}...`));

  const config = await fetchCurrentConfig(owner, name, options.includeSecrets);

  if (options.dir) {
    writeDirectoryConfig(options.dir, config);
    console.log(colors.green(`Settings exported to ${options.dir}/`));
  } else if (options.single) {
    const yamlContent = yaml.dump(config, {
      indent: 2,
      lineWidth: -1,
      noRefs: true,
    });
    writeFileSync(options.single, yamlContent);
    console.log(colors.green(`Settings exported to ${options.single}`));
  } else {
    // Output to stdout
    const yamlContent = yaml.dump(config, {
      indent: 2,
      lineWidth: -1,
      noRefs: true,
    });
    console.log("\n" + yamlContent);
  }
}

async function fetchCurrentConfig(
  owner: string,
  name: string,
  includeSecrets?: boolean
): Promise<Config> {
  const config: Config = {};

  // Fetch repo metadata
  const repoData = ghApiGet<GitHubRepo>(`/repos/${owner}/${name}`);

  config.repo = {
    description: repoData.description || undefined,
    homepage: repoData.homepage || undefined,
    visibility: repoData.visibility as RepoSettings["visibility"],
    allow_merge_commit: repoData.allow_merge_commit,
    allow_rebase_merge: repoData.allow_rebase_merge,
    allow_squash_merge: repoData.allow_squash_merge,
    delete_branch_on_merge: repoData.delete_branch_on_merge,
    allow_update_branch: repoData.allow_update_branch,
  };

  // Remove undefined values
  config.repo = Object.fromEntries(
    Object.entries(config.repo).filter(([, v]) => v !== undefined)
  ) as RepoSettings;

  // Fetch topics
  if (repoData.topics && repoData.topics.length > 0) {
    config.topics = repoData.topics;
  }

  // Fetch labels
  const labels = ghApiGet<GitHubLabel[]>(`/repos/${owner}/${name}/labels`);
  if (labels.length > 0) {
    config.labels = {
      replace_default: false,
      items: labels.map((l) => ({
        name: l.name,
        color: l.color,
        description: l.description || undefined,
      })),
    };
    // Clean up undefined descriptions
    config.labels.items = config.labels.items.map((item) =>
      Object.fromEntries(
        Object.entries(item).filter(([, v]) => v !== undefined)
      )
    ) as Label[];
  }

  // Fetch branch protection for main
  try {
    const protection = ghApiGet<GitHubBranchProtection>(
      `/repos/${owner}/${name}/branches/main/protection`
    );

    const branchConfig: BranchProtectionConfig = {};

    if (protection.required_pull_request_reviews) {
      branchConfig.required_reviews =
        protection.required_pull_request_reviews.required_approving_review_count;
      branchConfig.dismiss_stale_reviews =
        protection.required_pull_request_reviews.dismiss_stale_reviews;
    }

    if (protection.required_status_checks) {
      branchConfig.require_status_checks = true;
      branchConfig.status_checks = protection.required_status_checks.contexts;
    }

    if (protection.enforce_admins) {
      branchConfig.enforce_admins = protection.enforce_admins.enabled;
    }

    if (protection.required_linear_history) {
      branchConfig.require_linear_history =
        protection.required_linear_history.enabled;
    }

    if (protection.allow_force_pushes) {
      branchConfig.allow_force_pushes = protection.allow_force_pushes.enabled;
    }

    if (protection.allow_deletions) {
      branchConfig.allow_deletions = protection.allow_deletions.enabled;
    }

    if (Object.keys(branchConfig).length > 0) {
      config.branch_protection = {
        main: branchConfig,
      };
    }
  } catch {
    // Branch protection might not exist
  }

  // Fetch secrets (names only)
  if (includeSecrets) {
    try {
      const secretsResponse = ghApiGet<{ secrets: GitHubSecret[] }>(
        `/repos/${owner}/${name}/actions/secrets`
      );
      if (secretsResponse.secrets && secretsResponse.secrets.length > 0) {
        config.secrets = {
          required: secretsResponse.secrets.map((s) => s.name),
        };
      }
    } catch {
      // Secrets API might not be accessible
    }
  }

  return config;
}

function writeDirectoryConfig(dirPath: string, config: Config): void {
  if (!existsSync(dirPath)) {
    mkdirSync(dirPath, { recursive: true });
  }

  const dumpOptions = { indent: 2, lineWidth: -1, noRefs: true };

  if (config.repo) {
    writeFileSync(
      join(dirPath, "repo.yaml"),
      yaml.dump({ repo: config.repo }, dumpOptions)
    );
  }

  if (config.topics) {
    writeFileSync(
      join(dirPath, "topics.yaml"),
      yaml.dump({ topics: config.topics }, dumpOptions)
    );
  }

  if (config.labels) {
    writeFileSync(
      join(dirPath, "labels.yaml"),
      yaml.dump({ labels: config.labels }, dumpOptions)
    );
  }

  if (config.branch_protection) {
    writeFileSync(
      join(dirPath, "branch-protection.yaml"),
      yaml.dump({ branch_protection: config.branch_protection }, dumpOptions)
    );
  }

  if (config.secrets) {
    writeFileSync(
      join(dirPath, "secrets.yaml"),
      yaml.dump({ secrets: config.secrets }, dumpOptions)
    );
  }
}
