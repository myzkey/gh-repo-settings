import { Command } from 'commander'
import { applyCommand } from '~/commands/apply'
import { exportCommand } from '~/commands/export'
import { initCommand } from '~/commands/init'
import { planCommand } from '~/commands/plan'
import { logger, setLogLevel } from '~/utils/logger'

const program = new Command()

program
  .name('gh-repo-settings')
  .description('Manage GitHub repository settings via YAML configuration')
  .version('0.1.0')
  .option('-v, --verbose', 'Show debug output')
  .option('-q, --quiet', 'Only show errors')
  .hook('preAction', () => {
    const opts = program.opts()
    if (opts.verbose) {
      setLogLevel('verbose')
    } else if (opts.quiet) {
      setLogLevel('quiet')
    }
  })

program
  .command('export')
  .description('Export current GitHub repository settings to YAML')
  .option(
    '-r, --repo <owner/name>',
    'Target repository (default: current repo)',
  )
  .option('-d, --dir <path>', 'Export to directory (multiple YAML files)')
  .option('-s, --single <path>', 'Export to single YAML file')
  .option('--include-secrets', 'Include secret names in export')
  .action(async (options) => {
    try {
      await exportCommand({
        repo: options.repo,
        dir: options.dir,
        single: options.single,
        includeSecrets: options.includeSecrets,
      })
    } catch (error) {
      logger.error(
        `Error: ${error instanceof Error ? error.message : String(error)}`,
      )
      process.exit(1)
    }
  })

program
  .command('apply')
  .description('Apply YAML configuration to GitHub repository')
  .option(
    '-r, --repo <owner/name>',
    'Target repository (default: current repo)',
  )
  .option('-c, --config <path>', 'Path to single YAML config file')
  .option('-d, --dir <path>', 'Path to config directory')
  .option('--dry-run', 'Show planned changes without applying')
  .action(async (options) => {
    try {
      await applyCommand({
        repo: options.repo,
        config: options.config,
        dir: options.dir,
        dryRun: options.dryRun,
      })
    } catch (error) {
      logger.error(
        `Error: ${error instanceof Error ? error.message : String(error)}`,
      )
      process.exit(1)
    }
  })

program
  .command('plan')
  .description('Validate config and show planned changes')
  .option(
    '-r, --repo <owner/name>',
    'Target repository (default: current repo)',
  )
  .option('-c, --config <path>', 'Path to single YAML config file')
  .option('-d, --dir <path>', 'Path to config directory')
  .option('--secrets', 'Check only secrets existence')
  .option('--env', 'Check only environment variables existence')
  .option('--schema-only', 'Only validate YAML schema (no GitHub API calls)')
  .action(async (options) => {
    try {
      await planCommand({
        repo: options.repo,
        config: options.config,
        dir: options.dir,
        secrets: options.secrets,
        env: options.env,
        schemaOnly: options.schemaOnly,
      })
    } catch (error) {
      logger.error(
        `Error: ${error instanceof Error ? error.message : String(error)}`,
      )
      process.exit(1)
    }
  })

program
  .command('init')
  .description('Initialize a new configuration file interactively')
  .option(
    '-o, --output <path>',
    'Output file path',
    '.github/repo-settings.yaml',
  )
  .option('-f, --force', 'Overwrite existing file')
  .action(async (options) => {
    try {
      await initCommand({
        output: options.output,
        force: options.force,
      })
    } catch (error) {
      if ((error as { name?: string }).name === 'ExitPromptError') {
        logger.info('\nInit cancelled.')
        process.exit(0)
      }
      logger.error(
        `Error: ${error instanceof Error ? error.message : String(error)}`,
      )
      process.exit(1)
    }
  })

program.parse()
