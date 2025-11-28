import chalk from "chalk";
import type {
  Config,
  DiffItem,
  RepoInfo,
  Label,
  BranchProtectionConfig,
} from "../types.js";
import { loadConfig, printValidationErrors } from "../utils/config.js";
import { validateConfig } from "../utils/schema.js";
import { printDiff } from "../utils/diff.js";
import {
  ghApiGet,
  ghApiPatch,
  ghApiPut,
  ghApiPost,
  ghApiDelete,
  getRepoInfo,
} from "../utils/gh.js";

interface ApplyOptions {
  repo?: string;
  config?: string;
  dir?: string;
  dryRun?: boolean;
}

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

interface GitHubSecret {
  name: string;
}

interface GitHubEnvVariable {
  name: string;
}

export async function applyCommand(options: ApplyOptions): Promise<void> {
  const repoInfo = getRepoInfo(options.repo);
  const { owner, name } = repoInfo;

  console.log(chalk.blue(`Loading config for ${owner}/${name}...`));

  const config = loadConfig({
    dir: options.dir,
    config: options.config,
  });

  // Validate config before applying
  console.log(chalk.blue("Validating config schema..."));
  const validationResult = validateConfig(config);

  if (!validationResult.valid) {
    printValidationErrors(validationResult);
    process.exit(1);
  }

  console.log(chalk.green("Schema validation passed.\n"));

  const diffs = await calculateDiffs(owner, name, config);

  if (options.dryRun) {
    console.log(chalk.yellow("\n[DRY RUN] No changes will be made.\n"));
    printDiff(diffs);
    return;
  }

  if (diffs.length === 0) {
    console.log(chalk.green("No changes to apply."));
    return;
  }

  printDiff(diffs);
  console.log(chalk.blue("\nApplying changes...\n"));

  await applyChanges(repoInfo, config, diffs);

  console.log(chalk.green("\nAll changes applied successfully!"));
}

async function calculateDiffs(
  owner: string,
  name: string,
  config: Config
): Promise<DiffItem[]> {
  const diffs: DiffItem[] = [];

  // Repo metadata
  if (config.repo) {
    const currentRepo = ghApiGet<GitHubRepo>(`/repos/${owner}/${name}`);

    for (const [key, value] of Object.entries(config.repo)) {
      if (value === undefined) continue;

      const currentValue = currentRepo[key as keyof GitHubRepo];
      if (currentValue !== value) {
        diffs.push({
          type: "repo",
          action: "update",
          details: `${key}: ${JSON.stringify(currentValue)} -> ${JSON.stringify(value)}`,
          apiCall: `PATCH /repos/${owner}/${name}`,
        });
      }
    }
  }

  // Topics
  if (config.topics) {
    const currentRepo = ghApiGet<GitHubRepo>(`/repos/${owner}/${name}`);
    const currentTopics = currentRepo.topics || [];
    const newTopics = config.topics;

    const added = newTopics.filter((t) => !currentTopics.includes(t));
    const removed = currentTopics.filter((t) => !newTopics.includes(t));

    if (added.length > 0 || removed.length > 0) {
      diffs.push({
        type: "topics",
        action: "update",
        details: `Set topics to: [${newTopics.join(", ")}]`,
        apiCall: `PUT /repos/${owner}/${name}/topics`,
      });
    }
  }

  // Labels
  if (config.labels) {
    const currentLabels = ghApiGet<GitHubLabel[]>(
      `/repos/${owner}/${name}/labels`
    );
    const currentLabelMap = new Map(currentLabels.map((l) => [l.name, l]));
    const configLabelMap = new Map(config.labels.items.map((l) => [l.name, l]));

    if (config.labels.replace_default) {
      // Delete labels not in config
      for (const label of currentLabels) {
        if (!configLabelMap.has(label.name)) {
          diffs.push({
            type: "labels",
            action: "delete",
            details: `Delete label: ${label.name}`,
            apiCall: `DELETE /repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`,
          });
        }
      }
    }

    // Create or update labels
    for (const label of config.labels.items) {
      const current = currentLabelMap.get(label.name);

      if (!current) {
        diffs.push({
          type: "labels",
          action: "create",
          details: `Create label: ${label.name} (#${label.color})`,
          apiCall: `POST /repos/${owner}/${name}/labels`,
        });
      } else if (
        current.color !== label.color ||
        (current.description || "") !== (label.description || "")
      ) {
        diffs.push({
          type: "labels",
          action: "update",
          details: `Update label: ${label.name}`,
          apiCall: `PATCH /repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`,
        });
      }
    }
  }

  // Branch protection (main only for v0)
  if (config.branch_protection?.main) {
    diffs.push({
      type: "branch_protection",
      action: "update",
      details: `Set branch protection for main`,
      apiCall: `PUT /repos/${owner}/${name}/branches/main/protection`,
    });
  }

  // Secrets (existence check only)
  if (config.secrets?.required) {
    try {
      const secretsResponse = ghApiGet<{ secrets: GitHubSecret[] }>(
        `/repos/${owner}/${name}/actions/secrets`
      );
      const existingSecrets = new Set(
        secretsResponse.secrets?.map((s) => s.name) || []
      );

      for (const secretName of config.secrets.required) {
        if (!existingSecrets.has(secretName)) {
          diffs.push({
            type: "secrets",
            action: "check",
            details: `Missing secret: ${secretName} (use 'gh secret set ${secretName}' to add)`,
          });
        }
      }
    } catch {
      diffs.push({
        type: "secrets",
        action: "check",
        details: `Could not verify secrets (API access may be restricted)`,
      });
    }
  }

  // Env (existence check only)
  if (config.env?.required) {
    try {
      const envResponse = ghApiGet<{ variables: GitHubEnvVariable[] }>(
        `/repos/${owner}/${name}/actions/variables`
      );
      const existingVars = new Set(
        envResponse.variables?.map((v) => v.name) || []
      );

      for (const varName of config.env.required) {
        if (!existingVars.has(varName)) {
          diffs.push({
            type: "env",
            action: "check",
            details: `Missing env variable: ${varName} (use 'gh variable set ${varName}' to add)`,
          });
        }
      }
    } catch {
      diffs.push({
        type: "env",
        action: "check",
        details: `Could not verify env variables (API access may be restricted)`,
      });
    }
  }

  return diffs;
}

async function applyChanges(
  repoInfo: RepoInfo,
  config: Config,
  diffs: DiffItem[]
): Promise<void> {
  const { owner, name } = repoInfo;

  // 1. Apply repo metadata
  if (config.repo && diffs.some((d) => d.type === "repo")) {
    console.log(chalk.blue("Updating repository settings..."));
    ghApiPatch(`/repos/${owner}/${name}`, config.repo as Record<string, unknown>);
    console.log(chalk.green("  Repository settings updated"));
  }

  // 2. Apply topics
  if (config.topics && diffs.some((d) => d.type === "topics")) {
    console.log(chalk.blue("Updating topics..."));
    ghApiPut(`/repos/${owner}/${name}/topics`, { names: config.topics });
    console.log(chalk.green("  Topics updated"));
  }

  // 3. Apply labels
  if (config.labels && diffs.some((d) => d.type === "labels")) {
    console.log(chalk.blue("Updating labels..."));

    const currentLabels = ghApiGet<GitHubLabel[]>(
      `/repos/${owner}/${name}/labels`
    );
    const currentLabelMap = new Map(currentLabels.map((l) => [l.name, l]));
    const configLabelMap = new Map(config.labels.items.map((l) => [l.name, l]));

    // Delete labels if replace_default
    if (config.labels.replace_default) {
      for (const label of currentLabels) {
        if (!configLabelMap.has(label.name)) {
          ghApiDelete(
            `/repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`
          );
          console.log(chalk.red(`  Deleted label: ${label.name}`));
        }
      }
    }

    // Create or update labels
    for (const label of config.labels.items) {
      const current = currentLabelMap.get(label.name);
      const labelData: Record<string, string> = {
        name: label.name,
        color: label.color,
      };
      if (label.description) {
        labelData.description = label.description;
      }

      if (!current) {
        ghApiPost(`/repos/${owner}/${name}/labels`, labelData);
        console.log(chalk.green(`  Created label: ${label.name}`));
      } else if (
        current.color !== label.color ||
        (current.description || "") !== (label.description || "")
      ) {
        ghApiPatch(
          `/repos/${owner}/${name}/labels/${encodeURIComponent(label.name)}`,
          labelData
        );
        console.log(chalk.yellow(`  Updated label: ${label.name}`));
      }
    }
  }

  // 4. Apply branch protection
  if (config.branch_protection?.main) {
    console.log(chalk.blue("Updating branch protection for main..."));
    const bp = config.branch_protection.main;

    const protectionData: Record<string, unknown> = {
      required_status_checks: bp.require_status_checks
        ? {
            strict: true,
            contexts: bp.status_checks || [],
          }
        : null,
      enforce_admins: bp.enforce_admins ?? false,
      required_pull_request_reviews: bp.required_reviews
        ? {
            required_approving_review_count: bp.required_reviews,
            dismiss_stale_reviews: bp.dismiss_stale_reviews ?? false,
          }
        : null,
      restrictions: null,
      required_linear_history: bp.require_linear_history ?? false,
      allow_force_pushes: bp.allow_force_pushes ?? false,
      allow_deletions: bp.allow_deletions ?? false,
    };

    ghApiPut(`/repos/${owner}/${name}/branches/main/protection`, protectionData);
    console.log(chalk.green("  Branch protection updated"));
  }

  // 5. Check secrets (no changes, just warnings)
  const secretDiffs = diffs.filter((d) => d.type === "secrets");
  if (secretDiffs.length > 0) {
    console.log(chalk.yellow("\nSecret warnings:"));
    for (const diff of secretDiffs) {
      console.log(chalk.yellow(`  ${diff.details}`));
    }
  }

  // 6. Check env variables (no changes, just warnings)
  const envDiffs = diffs.filter((d) => d.type === "env");
  if (envDiffs.length > 0) {
    console.log(chalk.yellow("\nEnvironment variable warnings:"));
    for (const diff of envDiffs) {
      console.log(chalk.yellow(`  ${diff.details}`));
    }
  }
}
