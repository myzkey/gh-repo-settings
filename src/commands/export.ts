import type {
  BranchProtectionConfig,
  Config,
  IGitHubClient,
  Label,
  RepoSettings,
} from '~/domain'
import { existsSync, join, mkdirSync, writeFileSync } from '~/infra/fs'
import { createClient, getRepoInfo } from '~/infra/github'
import { stringify } from '~/infra/yaml'
import { logger } from '~/utils/logger'

interface ExportOptions {
  repo?: string
  dir?: string
  single?: string
  includeSecrets?: boolean
}

export async function exportCommand(options: ExportOptions): Promise<void> {
  const repoInfo = getRepoInfo(options.repo)
  const client = createClient(options.repo)
  const { owner, name } = repoInfo

  logger.info(`Exporting settings from ${owner}/${name}...`)

  const config = fetchCurrentConfig(client, options.includeSecrets)

  if (options.dir) {
    writeDirectoryConfig(options.dir, config)
    logger.success(`Settings exported to ${options.dir}/`)
  } else if (options.single) {
    writeFileSync(options.single, stringify(config))
    logger.success(`Settings exported to ${options.single}`)
  } else {
    // Output to stdout
    logger.log(`\n${stringify(config)}`)
  }
}

function fetchCurrentConfig(
  client: IGitHubClient,
  includeSecrets?: boolean,
): Config {
  const config: Config = {}

  // Fetch repo metadata
  const repoData = client.getRepo()

  config.repo = {
    description: repoData.description || undefined,
    homepage: repoData.homepage || undefined,
    visibility: repoData.visibility as RepoSettings['visibility'],
    allow_merge_commit: repoData.allow_merge_commit,
    allow_rebase_merge: repoData.allow_rebase_merge,
    allow_squash_merge: repoData.allow_squash_merge,
    delete_branch_on_merge: repoData.delete_branch_on_merge,
    allow_update_branch: repoData.allow_update_branch,
  }

  // Remove undefined values
  config.repo = Object.fromEntries(
    Object.entries(config.repo).filter(([, v]) => v !== undefined),
  ) as RepoSettings

  // Fetch topics
  if (repoData.topics && repoData.topics.length > 0) {
    config.topics = repoData.topics
  }

  // Fetch labels
  const labels = client.getLabels()
  if (labels.length > 0) {
    config.labels = {
      replace_default: false,
      items: labels.map((l) => ({
        name: l.name,
        color: l.color,
        description: l.description || undefined,
      })),
    }
    // Clean up undefined descriptions
    config.labels.items = config.labels.items.map((item) =>
      Object.fromEntries(
        Object.entries(item).filter(([, v]) => v !== undefined),
      ),
    ) as Label[]
  }

  // Fetch branch protection for main
  try {
    const protection = client.getBranchProtection('main')

    const branchConfig: BranchProtectionConfig = {}

    if (protection.required_pull_request_reviews) {
      branchConfig.required_reviews =
        protection.required_pull_request_reviews.required_approving_review_count
      branchConfig.dismiss_stale_reviews =
        protection.required_pull_request_reviews.dismiss_stale_reviews
    }

    if (protection.required_status_checks) {
      branchConfig.require_status_checks = true
      branchConfig.status_checks = protection.required_status_checks.contexts
    }

    if (protection.enforce_admins) {
      branchConfig.enforce_admins = protection.enforce_admins.enabled
    }

    if (protection.required_linear_history) {
      branchConfig.require_linear_history =
        protection.required_linear_history.enabled
    }

    if (protection.allow_force_pushes) {
      branchConfig.allow_force_pushes = protection.allow_force_pushes.enabled
    }

    if (protection.allow_deletions) {
      branchConfig.allow_deletions = protection.allow_deletions.enabled
    }

    if (Object.keys(branchConfig).length > 0) {
      config.branch_protection = {
        main: branchConfig,
      }
    }
  } catch {
    // Branch protection might not exist
  }

  // Fetch secrets (names only)
  if (includeSecrets) {
    try {
      const secretNames = client.getSecretNames()
      if (secretNames.length > 0) {
        config.secrets = {
          required: secretNames,
        }
      }
    } catch {
      // Secrets API might not be accessible
    }
  }

  return config
}

function writeDirectoryConfig(dirPath: string, config: Config): void {
  if (!existsSync(dirPath)) {
    mkdirSync(dirPath, { recursive: true })
  }

  if (config.repo) {
    writeFileSync(join(dirPath, 'repo.yaml'), stringify({ repo: config.repo }))
  }

  if (config.topics) {
    writeFileSync(
      join(dirPath, 'topics.yaml'),
      stringify({ topics: config.topics }),
    )
  }

  if (config.labels) {
    writeFileSync(
      join(dirPath, 'labels.yaml'),
      stringify({ labels: config.labels }),
    )
  }

  if (config.branch_protection) {
    writeFileSync(
      join(dirPath, 'branch-protection.yaml'),
      stringify({ branch_protection: config.branch_protection }),
    )
  }

  if (config.secrets) {
    writeFileSync(
      join(dirPath, 'secrets.yaml'),
      stringify({ secrets: config.secrets }),
    )
  }
}
