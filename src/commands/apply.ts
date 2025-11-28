import type { Config, DiffItem, IGitHubClient, RepoIdentifier } from '~/domain'
import {
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

interface ApplyOptions {
  repo?: string
  config?: string
  dir?: string
  dryRun?: boolean
}

export async function applyCommand(options: ApplyOptions): Promise<void> {
  const repoInfo = getRepoInfo(options.repo)
  const client = createClient(options.repo)
  const { owner, name } = repoInfo

  logger.info(`Loading config for ${owner}/${name}...`)

  const config = loadConfig({
    dir: options.dir,
    config: options.config,
  })

  // Validate config before applying
  logger.info('Validating config schema...')
  const validationResult = validateConfig(config)

  if (!validationResult.valid) {
    printValidationErrors(validationResult)
    process.exit(1)
  }

  logger.success('Schema validation passed.\n')

  const diffs = calculateDiffs(client, repoInfo, config)

  if (options.dryRun) {
    logger.warn('\n[DRY RUN] No changes will be made.\n')
    printDiff(diffs)
    return
  }

  if (diffs.length === 0) {
    logger.success('No changes to apply.')
    return
  }

  printDiff(diffs)
  logger.info('\nApplying changes...\n')

  applyChanges(client, config, diffs)

  logger.success('\nAll changes applied successfully!')
}

function calculateDiffs(
  client: IGitHubClient,
  repoInfo: RepoIdentifier,
  config: Config,
): DiffItem[] {
  const diffs: DiffItem[] = []
  const { owner, name } = repoInfo
  const diffContext = { client, owner, name }

  // Repo metadata
  if (config.repo) {
    const currentRepo = client.getRepo()
    const repoDiffs = calculateRepoDiffs(config, currentRepo, diffContext)
    diffs.push(...repoDiffs)
  }

  // Topics
  if (config.topics) {
    const currentRepo = client.getRepo()
    const topicsDiffs = calculateTopicsDiffs(config, currentRepo, diffContext)
    diffs.push(...topicsDiffs)
  }

  // Labels
  if (config.labels) {
    const currentLabels = client.getLabels()
    const labelsDiffs = calculateLabelsDiffs(config, currentLabels, diffContext)
    diffs.push(...labelsDiffs)
  }

  // Branch protection (main only for v0)
  if (config.branch_protection?.main) {
    diffs.push({
      type: 'branch_protection',
      action: 'update',
      details: 'Set branch protection for main',
      apiCall: `PUT /repos/${owner}/${name}/branches/main/protection`,
    })
  }

  // Secrets (existence check only)
  if (config.secrets?.required) {
    try {
      const existingSecretNames = client.getSecretNames()
      const secretDiffs = calculateSecretsDiffs(config, existingSecretNames)
      diffs.push(...secretDiffs)
    } catch {
      diffs.push({
        type: 'secrets',
        action: 'check',
        details: 'Could not verify secrets (API access may be restricted)',
      })
    }
  }

  // Env (existence check only)
  if (config.env?.required) {
    try {
      const existingVarNames = client.getVariableNames()
      const envDiffs = calculateEnvDiffs(config, existingVarNames)
      diffs.push(...envDiffs)
    } catch {
      diffs.push({
        type: 'env',
        action: 'check',
        details:
          'Could not verify env variables (API access may be restricted)',
      })
    }
  }

  return diffs
}

function applyChanges(
  client: IGitHubClient,
  config: Config,
  diffs: DiffItem[],
): void {
  // 1. Apply repo metadata
  if (config.repo && diffs.some((d) => d.type === 'repo')) {
    logger.info('Updating repository settings...')
    client.updateRepo(config.repo)
    logger.success('  Repository settings updated')
  }

  // 2. Apply topics
  if (config.topics && diffs.some((d) => d.type === 'topics')) {
    logger.info('Updating topics...')
    client.setTopics(config.topics)
    logger.success('  Topics updated')
  }

  // 3. Apply labels
  if (config.labels && diffs.some((d) => d.type === 'labels')) {
    logger.info('Updating labels...')

    const currentLabels = client.getLabels()
    const currentLabelMap = new Map(currentLabels.map((l) => [l.name, l]))
    const configLabelMap = new Map(config.labels.items.map((l) => [l.name, l]))

    // Delete labels if replace_default
    if (config.labels.replace_default) {
      for (const label of currentLabels) {
        if (!configLabelMap.has(label.name)) {
          client.deleteLabel(label.name)
          logger.error(`  Deleted label: ${label.name}`)
        }
      }
    }

    // Create or update labels
    for (const label of config.labels.items) {
      const current = currentLabelMap.get(label.name)

      if (!current) {
        client.createLabel(label)
        logger.success(`  Created label: ${label.name}`)
      } else if (
        current.color !== label.color ||
        (current.description || '') !== (label.description || '')
      ) {
        client.updateLabel(label.name, label)
        logger.warn(`  Updated label: ${label.name}`)
      }
    }
  }

  // 4. Apply branch protection
  if (config.branch_protection?.main) {
    logger.info('Updating branch protection for main...')
    client.setBranchProtection('main', config.branch_protection.main)
    logger.success('  Branch protection updated')
  }

  // 5. Check secrets (no changes, just warnings)
  const secretDiffs = diffs.filter((d) => d.type === 'secrets')
  if (secretDiffs.length > 0) {
    logger.warn('\nSecret warnings:')
    for (const diff of secretDiffs) {
      logger.warn(`  ${diff.details}`)
    }
  }

  // 6. Check env variables (no changes, just warnings)
  const envDiffs = diffs.filter((d) => d.type === 'env')
  if (envDiffs.length > 0) {
    logger.warn('\nEnvironment variable warnings:')
    for (const diff of envDiffs) {
      logger.warn(`  ${diff.details}`)
    }
  }
}
