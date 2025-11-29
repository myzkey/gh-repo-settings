import { checkbox, confirm, input, select } from '@inquirer/prompts'
import type { Config } from '~/domain'
import { existsSync, writeFileSync } from '~/infra/fs'
import { stringify } from '~/infra/yaml'
import { logger } from '~/utils/logger'

export interface InitOptions {
  output?: string
  force?: boolean
}

export async function initCommand(options: InitOptions): Promise<void> {
  const outputPath = options.output || '.github/repo-settings.yaml'

  if (existsSync(outputPath) && !options.force) {
    logger.error(`File already exists: ${outputPath}`)
    logger.info('Use --force to overwrite')
    process.exit(1)
  }

  logger.heading('gh-repo-settings init')
  logger.info('Answer the following questions to generate your config.\n')

  const config: Config = {}

  // Repository settings
  const configureRepo = await confirm({
    message: 'Configure repository settings?',
    default: true,
  })

  if (configureRepo) {
    const visibility = await select({
      message: 'Repository visibility:',
      choices: [
        { name: 'public', value: 'public' as const },
        { name: 'private', value: 'private' as const },
        { name: 'internal (GitHub Enterprise)', value: 'internal' as const },
      ],
    })

    const description = await input({
      message: 'Repository description (optional):',
    })

    const mergeOptions = await checkbox({
      message: 'Allowed merge methods:',
      choices: [
        { name: 'Squash merge', value: 'squash', checked: true },
        { name: 'Merge commit', value: 'merge', checked: false },
        { name: 'Rebase merge', value: 'rebase', checked: true },
      ],
    })

    const deleteBranchOnMerge = await confirm({
      message: 'Auto-delete head branches after merge?',
      default: true,
    })

    config.repo = {
      visibility,
      ...(description && { description }),
      allow_squash_merge: mergeOptions.includes('squash'),
      allow_merge_commit: mergeOptions.includes('merge'),
      allow_rebase_merge: mergeOptions.includes('rebase'),
      delete_branch_on_merge: deleteBranchOnMerge,
    }
  }

  // Topics
  const configureTopics = await confirm({
    message: 'Add repository topics?',
    default: false,
  })

  if (configureTopics) {
    const topicsInput = await input({
      message: 'Enter topics (comma-separated):',
      default: 'typescript, cli',
    })
    config.topics = topicsInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)
  }

  // Labels
  const configureLabels = await confirm({
    message: 'Configure custom labels?',
    default: true,
  })

  if (configureLabels) {
    const replaceDefault = await confirm({
      message: 'Replace default GitHub labels?',
      default: false,
    })

    const labelPresets = await checkbox({
      message: 'Select label presets:',
      choices: [
        { name: 'bug (red)', value: 'bug', checked: true },
        { name: 'feature (green)', value: 'feature', checked: true },
        { name: 'documentation (blue)', value: 'documentation', checked: true },
        { name: 'chore (gray)', value: 'chore', checked: true },
        {
          name: 'good first issue (purple)',
          value: 'good-first-issue',
          checked: false,
        },
        { name: 'help wanted (yellow)', value: 'help-wanted', checked: false },
      ],
    })

    const labelMap: Record<
      string,
      { name: string; color: string; description: string }
    > = {
      bug: {
        name: 'bug',
        color: 'd73a4a',
        description: "Something isn't working",
      },
      feature: {
        name: 'feature',
        color: '0e8a16',
        description: 'New feature request',
      },
      documentation: {
        name: 'documentation',
        color: '0075ca',
        description: 'Documentation improvements',
      },
      chore: {
        name: 'chore',
        color: '6c757d',
        description: 'Maintenance tasks',
      },
      'good-first-issue': {
        name: 'good first issue',
        color: '7057ff',
        description: 'Good for newcomers',
      },
      'help-wanted': {
        name: 'help wanted',
        color: 'fbca04',
        description: 'Extra attention is needed',
      },
    }

    config.labels = {
      replace_default: replaceDefault,
      items: labelPresets.map((preset) => labelMap[preset]),
    }
  }

  // Branch protection
  const configureBranchProtection = await confirm({
    message: 'Configure branch protection?',
    default: true,
  })

  if (configureBranchProtection) {
    const branchName = await input({
      message: 'Branch to protect:',
      default: 'main',
    })

    const requiredReviews = await select({
      message: 'Required approving reviews:',
      choices: [
        { name: '0 (no reviews required)', value: 0 },
        { name: '1', value: 1 },
        { name: '2', value: 2 },
        { name: '3', value: 3 },
      ],
      default: 1,
    })

    const dismissStaleReviews = await confirm({
      message: 'Dismiss stale reviews on new commits?',
      default: true,
    })

    const requireStatusChecks = await confirm({
      message: 'Require status checks?',
      default: false,
    })

    let statusChecks: string[] = []
    if (requireStatusChecks) {
      const checksInput = await input({
        message: 'Status check names (comma-separated):',
        default: 'ci',
      })
      statusChecks = checksInput
        .split(',')
        .map((c) => c.trim())
        .filter(Boolean)
    }

    const enforceAdmins = await confirm({
      message: 'Enforce rules for administrators?',
      default: false,
    })

    config.branch_protection = {
      [branchName]: {
        required_reviews: requiredReviews,
        dismiss_stale_reviews: dismissStaleReviews,
        ...(requireStatusChecks && {
          require_status_checks: true,
          status_checks: statusChecks,
        }),
        enforce_admins: enforceAdmins,
      },
    }
  }

  // Secrets
  const configureSecrets = await confirm({
    message: 'Add required secrets check?',
    default: false,
  })

  if (configureSecrets) {
    const secretsInput = await input({
      message: 'Required secret names (comma-separated):',
      default: 'API_TOKEN',
    })
    config.secrets = {
      required: secretsInput
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean),
    }
  }

  // Environment variables
  const configureEnv = await confirm({
    message: 'Add required environment variables check?',
    default: false,
  })

  if (configureEnv) {
    const envInput = await input({
      message: 'Required variable names (comma-separated):',
      default: 'DATABASE_URL',
    })
    config.env = {
      required: envInput
        .split(',')
        .map((e) => e.trim())
        .filter(Boolean),
    }
  }

  // Generate YAML
  const yaml = stringify(config)

  // Write file
  const parentDir = outputPath.split('/').slice(0, -1).join('/')
  if (parentDir && !existsSync(parentDir)) {
    const { mkdirSync } = await import('~/infra/fs')
    mkdirSync(parentDir, { recursive: true })
  }

  writeFileSync(outputPath, yaml)

  logger.log('')
  logger.success(`Created ${outputPath}`)
  logger.log('')
  logger.info('Next steps:')
  logger.log(`  1. Review the generated config: ${outputPath}`)
  logger.log('  2. Preview changes: gh repo-settings plan')
  logger.log('  3. Apply changes: gh repo-settings apply')
}
