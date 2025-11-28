import type { Config, DiffItem, IGitHubClient, RepoIdentifier } from '~/domain'
import {
  calculateBranchProtectionDiffs,
  calculateEnvDiffs,
  calculateLabelsDiffs,
  calculateRepoDiffs,
  calculateSecretsDiffs,
  calculateTopicsDiffs,
  loadConfig,
  printDiff,
  printValidationErrors,
  validateConfig,
} from '~/domain'
import { createClient, getRepoInfo } from '~/infra/github'
import { logger } from '~/utils/logger'

interface PlanOptions {
  repo?: string
  config?: string
  dir?: string
  secrets?: boolean
  env?: boolean
  schemaOnly?: boolean
}

export async function planCommand(options: PlanOptions): Promise<void> {
  // Load config first (without validation to show all errors)
  const rawConfig = loadConfig({
    dir: options.dir,
    config: options.config,
  })

  // Step 1: Schema validation
  logger.info('Validating YAML schema...')
  const validationResult = validateConfig(rawConfig)

  if (!validationResult.valid) {
    printValidationErrors(validationResult)
    process.exit(1)
  }

  logger.success('Schema validation passed.\n')

  // If schema-only mode, stop here
  if (options.schemaOnly) {
    return
  }

  const repoInfo = getRepoInfo(options.repo)
  const client = createClient(options.repo)
  const { owner, name } = repoInfo

  logger.info(`Planning changes for ${owner}/${name}...`)

  const diffs: DiffItem[] = []

  // Check secrets only
  if (options.secrets) {
    const secretDiffs = verifySecrets(client, rawConfig)
    diffs.push(...secretDiffs)
  }
  // Check env only
  else if (options.env) {
    const envDiffs = verifyEnv(client, rawConfig)
    diffs.push(...envDiffs)
  }
  // Full plan
  else {
    const allDiffs = calculateDetailedDiffs(client, repoInfo, rawConfig)
    diffs.push(...allDiffs)
  }

  if (diffs.length === 0) {
    logger.success('\nAll settings are in sync!')
    return
  }

  logger.warn(`\nFound ${diffs.length} difference(s):\n`)
  printDiff(diffs)

  // Exit with error if there are issues
  const hasErrors = diffs.some(
    (d) => d.action === 'check' || d.action === 'create',
  )
  if (hasErrors) {
    process.exit(1)
  }
}

function verifySecrets(client: IGitHubClient, config: Config): DiffItem[] {
  if (!config.secrets?.required || config.secrets.required.length === 0) {
    logger.debug('No required secrets defined in config.')
    return []
  }

  try {
    const existingSecretNames = client.getSecretNames()
    const diffs = calculateSecretsDiffs(config, existingSecretNames)

    if (diffs.length === 0) {
      logger.success('All required secrets are present.')
    }

    return diffs
  } catch {
    return [
      {
        type: 'secrets',
        action: 'check',
        details: 'Could not verify secrets (API access may be restricted)',
      },
    ]
  }
}

function verifyEnv(client: IGitHubClient, config: Config): DiffItem[] {
  if (!config.env?.required || config.env.required.length === 0) {
    logger.debug('No required environment variables defined in config.')
    return []
  }

  try {
    const existingVarNames = client.getVariableNames()
    const diffs = calculateEnvDiffs(config, existingVarNames)

    if (diffs.length === 0) {
      logger.success('All required environment variables are present.')
    }

    return diffs
  } catch {
    return [
      {
        type: 'env',
        action: 'check',
        details:
          'Could not verify environment variables (API access may be restricted)',
      },
    ]
  }
}

function calculateDetailedDiffs(
  client: IGitHubClient,
  repoInfo: RepoIdentifier,
  config: Config,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = repoInfo
  const diffContext = { client, owner, name }

  // Check repo metadata
  if (config.repo) {
    const currentRepo = client.getRepo()
    const repoDiffs = calculateRepoDiffs(config, currentRepo, diffContext)
    diffs.push(...repoDiffs)
  }

  // Check topics
  if (config.topics) {
    const currentRepo = client.getRepo()
    const topicsDiffs = calculateTopicsDiffs(config, currentRepo, diffContext)
    diffs.push(...topicsDiffs)
  }

  // Check labels
  if (config.labels) {
    const currentLabels = client.getLabels()
    const labelsDiffs = calculateLabelsDiffs(config, currentLabels, diffContext)
    diffs.push(...labelsDiffs)
  }

  // Check branch protection
  if (config.branch_protection?.main) {
    try {
      const protection = client.getBranchProtection('main')
      const bpDiffs = calculateBranchProtectionDiffs(
        config,
        protection,
        'main',
        diffContext,
      )
      diffs.push(...bpDiffs)
    } catch {
      const bpDiffs = calculateBranchProtectionDiffs(
        config,
        null,
        'main',
        diffContext,
      )
      diffs.push(...bpDiffs)
    }
  }

  // Verify secrets
  const secretDiffs = verifySecrets(client, config)
  diffs.push(...secretDiffs)

  // Verify env
  const envDiffs = verifyEnv(client, config)
  diffs.push(...envDiffs)

  return diffs
}
