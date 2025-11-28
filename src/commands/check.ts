import type { Config, DiffItem } from '~/types'
import { colors } from '~/utils/colors'
import { loadConfig, printValidationErrors } from '~/utils/config'
import { printDiff } from '~/utils/diff'
import { getRepoInfo, ghApiGet } from '~/utils/gh'
import { validateConfig } from '~/utils/schema'

interface CheckOptions {
  repo?: string
  config?: string
  dir?: string
  secrets?: boolean
  env?: boolean
  schemaOnly?: boolean
}

interface GitHubRepo {
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

interface GitHubLabel {
  name: string
  color: string
  description: string | null
}

interface GitHubSecret {
  name: string
}

interface GitHubEnvVariable {
  name: string
}

interface GitHubBranchProtection {
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

export async function checkCommand(options: CheckOptions): Promise<void> {
  // Load config first (without validation to show all errors)
  const rawConfig = loadConfig({
    dir: options.dir,
    config: options.config,
  })

  // Step 1: Schema validation
  console.log(colors.blue('Validating YAML schema...'))
  const validationResult = validateConfig(rawConfig)

  if (!validationResult.valid) {
    printValidationErrors(validationResult)
    process.exit(1)
  }

  console.log(colors.green('Schema validation passed.\n'))

  // If schema-only mode, stop here
  if (options.schemaOnly) {
    return
  }

  const repoInfo = getRepoInfo(options.repo)
  const { owner, name } = repoInfo

  console.log(colors.blue(`Checking settings for ${owner}/${name}...`))

  const diffs: DiffItem[] = []

  // Check secrets only
  if (options.secrets) {
    const secretDiffs = await checkSecrets(owner, name, rawConfig)
    diffs.push(...secretDiffs)
  }
  // Check env only
  else if (options.env) {
    const envDiffs = await checkEnv(owner, name, rawConfig)
    diffs.push(...envDiffs)
  }
  // Full check
  else {
    const allDiffs = await calculateDetailedDiffs(owner, name, rawConfig)
    diffs.push(...allDiffs)
  }

  if (diffs.length === 0) {
    console.log(colors.green('\nAll settings are in sync!'))
    return
  }

  console.log(colors.yellow(`\nFound ${diffs.length} difference(s):\n`))
  printDiff(diffs)

  // Exit with error if there are issues
  const hasErrors = diffs.some(
    (d) => d.action === 'check' || d.action === 'create',
  )
  if (hasErrors) {
    process.exit(1)
  }
}

async function checkSecrets(
  owner: string,
  name: string,
  config: Config,
): Promise<DiffItem[]> {
  const diffs: DiffItem[] = []

  if (!config.secrets?.required || config.secrets.required.length === 0) {
    console.log(colors.gray('No required secrets defined in config.'))
    return diffs
  }

  try {
    const secretsResponse = ghApiGet<{ secrets: GitHubSecret[] }>(
      `/repos/${owner}/${name}/actions/secrets`,
    )
    const existingSecrets = new Set(
      secretsResponse.secrets?.map((s) => s.name) || [],
    )

    for (const secretName of config.secrets.required) {
      if (!existingSecrets.has(secretName)) {
        diffs.push({
          type: 'secrets',
          action: 'check',
          details: `Missing required secret: ${secretName}`,
        })
      }
    }

    if (diffs.length === 0) {
      console.log(colors.green('All required secrets are present.'))
    }
  } catch {
    diffs.push({
      type: 'secrets',
      action: 'check',
      details: `Could not verify secrets (API access may be restricted)`,
    })
  }

  return diffs
}

async function checkEnv(
  owner: string,
  name: string,
  config: Config,
): Promise<DiffItem[]> {
  const diffs: DiffItem[] = []

  if (!config.env?.required || config.env.required.length === 0) {
    console.log(
      colors.gray('No required environment variables defined in config.'),
    )
    return diffs
  }

  try {
    const envResponse = ghApiGet<{ variables: GitHubEnvVariable[] }>(
      `/repos/${owner}/${name}/actions/variables`,
    )
    const existingVars = new Set(
      envResponse.variables?.map((v) => v.name) || [],
    )

    for (const varName of config.env.required) {
      if (!existingVars.has(varName)) {
        diffs.push({
          type: 'env',
          action: 'check',
          details: `Missing required environment variable: ${varName}`,
        })
      }
    }

    if (diffs.length === 0) {
      console.log(
        colors.green('All required environment variables are present.'),
      )
    }
  } catch {
    diffs.push({
      type: 'env',
      action: 'check',
      details: `Could not verify environment variables (API access may be restricted)`,
    })
  }

  return diffs
}

async function calculateDetailedDiffs(
  owner: string,
  name: string,
  config: Config,
): Promise<DiffItem[]> {
  const diffs: DiffItem[] = []

  // Check repo metadata
  if (config.repo) {
    const currentRepo = ghApiGet<GitHubRepo>(`/repos/${owner}/${name}`)

    for (const [key, value] of Object.entries(config.repo)) {
      if (value === undefined) continue

      const currentValue = currentRepo[key as keyof GitHubRepo]
      if (currentValue !== value) {
        diffs.push({
          type: 'repo',
          action: 'update',
          details: `${key}: current="${currentValue}" expected="${value}"`,
        })
      }
    }
  }

  // Check topics
  if (config.topics) {
    const currentRepo = ghApiGet<GitHubRepo>(`/repos/${owner}/${name}`)
    const currentTopics = new Set(currentRepo.topics || [])
    const expectedTopics = new Set(config.topics)

    const missing = config.topics.filter((t) => !currentTopics.has(t))
    const extra = (currentRepo.topics || []).filter(
      (t) => !expectedTopics.has(t),
    )

    if (missing.length > 0) {
      diffs.push({
        type: 'topics',
        action: 'create',
        details: `Missing topics: ${missing.join(', ')}`,
      })
    }

    if (extra.length > 0) {
      diffs.push({
        type: 'topics',
        action: 'delete',
        details: `Extra topics (not in config): ${extra.join(', ')}`,
      })
    }
  }

  // Check labels
  if (config.labels) {
    const currentLabels = ghApiGet<GitHubLabel[]>(
      `/repos/${owner}/${name}/labels`,
    )
    const currentLabelMap = new Map(currentLabels.map((l) => [l.name, l]))
    const configLabelMap = new Map(config.labels.items.map((l) => [l.name, l]))

    // Check for missing labels
    for (const label of config.labels.items) {
      const current = currentLabelMap.get(label.name)

      if (!current) {
        diffs.push({
          type: 'labels',
          action: 'create',
          details: `Missing label: ${label.name} (#${label.color})`,
        })
      } else {
        const colorDiff = current.color !== label.color
        const descDiff =
          (current.description || '') !== (label.description || '')

        if (colorDiff || descDiff) {
          const changes: string[] = []
          if (colorDiff) {
            changes.push(`color: #${current.color} -> #${label.color}`)
          }
          if (descDiff) {
            changes.push(
              `description: "${current.description || ''}" -> "${label.description || ''}"`,
            )
          }
          diffs.push({
            type: 'labels',
            action: 'update',
            details: `Label "${label.name}" differs: ${changes.join(', ')}`,
          })
        }
      }
    }

    // Check for extra labels (if replace_default)
    if (config.labels.replace_default) {
      for (const label of currentLabels) {
        if (!configLabelMap.has(label.name)) {
          diffs.push({
            type: 'labels',
            action: 'delete',
            details: `Extra label (will be deleted): ${label.name}`,
          })
        }
      }
    }
  }

  // Check branch protection
  if (config.branch_protection?.main) {
    const bp = config.branch_protection.main

    try {
      const protection = ghApiGet<GitHubBranchProtection>(
        `/repos/${owner}/${name}/branches/main/protection`,
      )

      // Check required reviews
      if (bp.required_reviews !== undefined) {
        const currentReviews =
          protection.required_pull_request_reviews
            ?.required_approving_review_count ?? 0
        if (currentReviews !== bp.required_reviews) {
          diffs.push({
            type: 'branch_protection',
            action: 'update',
            details: `required_reviews: ${currentReviews} -> ${bp.required_reviews}`,
          })
        }
      }

      // Check dismiss_stale_reviews
      if (bp.dismiss_stale_reviews !== undefined) {
        const currentDismiss =
          protection.required_pull_request_reviews?.dismiss_stale_reviews ??
          false
        if (currentDismiss !== bp.dismiss_stale_reviews) {
          diffs.push({
            type: 'branch_protection',
            action: 'update',
            details: `dismiss_stale_reviews: ${currentDismiss} -> ${bp.dismiss_stale_reviews}`,
          })
        }
      }

      // Check status checks
      if (bp.require_status_checks !== undefined) {
        const hasStatusChecks = !!protection.required_status_checks
        if (hasStatusChecks !== bp.require_status_checks) {
          diffs.push({
            type: 'branch_protection',
            action: 'update',
            details: `require_status_checks: ${hasStatusChecks} -> ${bp.require_status_checks}`,
          })
        } else if (bp.status_checks && protection.required_status_checks) {
          const currentChecks = new Set(
            protection.required_status_checks.contexts || [],
          )
          const expectedChecks = new Set(bp.status_checks)

          const missing = bp.status_checks.filter((c) => !currentChecks.has(c))
          const extra = (
            protection.required_status_checks.contexts || []
          ).filter((c) => !expectedChecks.has(c))

          if (missing.length > 0 || extra.length > 0) {
            diffs.push({
              type: 'branch_protection',
              action: 'update',
              details: `status_checks differ - missing: [${missing.join(', ')}], extra: [${extra.join(', ')}]`,
            })
          }
        }
      }

      // Check enforce_admins
      if (bp.enforce_admins !== undefined) {
        const currentEnforce = protection.enforce_admins?.enabled ?? false
        if (currentEnforce !== bp.enforce_admins) {
          diffs.push({
            type: 'branch_protection',
            action: 'update',
            details: `enforce_admins: ${currentEnforce} -> ${bp.enforce_admins}`,
          })
        }
      }
    } catch {
      // No branch protection exists
      diffs.push({
        type: 'branch_protection',
        action: 'create',
        details: `Branch protection for main does not exist`,
      })
    }
  }

  // Check secrets
  const secretDiffs = await checkSecrets(owner, name, config)
  diffs.push(...secretDiffs)

  // Check env
  const envDiffs = await checkEnv(owner, name, config)
  diffs.push(...envDiffs)

  return diffs
}
