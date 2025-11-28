import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import yaml from 'js-yaml'
import type { Config, ValidationResult } from '~/types'
import { colors } from '~/utils/colors'
import { validateConfig } from '~/utils/schema'

const DEFAULT_DIR = '.github/repo-settings'
const DEFAULT_SINGLE_FILE = '.github/repo-settings.yaml'

export interface LoadConfigOptions {
  dir?: string
  config?: string
  validate?: boolean
}

export function loadConfig(options: LoadConfigOptions): Config {
  // Priority: --dir > --config > default dir > default single file
  if (options.dir) {
    return loadFromDirectory(options.dir)
  }

  if (options.config) {
    return loadSingleFile(options.config)
  }

  // Check default directory
  if (existsSync(DEFAULT_DIR)) {
    return loadFromDirectory(DEFAULT_DIR)
  }

  // Check default single file
  if (existsSync(DEFAULT_SINGLE_FILE)) {
    return loadSingleFile(DEFAULT_SINGLE_FILE)
  }

  throw new Error(
    `No config found. Create ${DEFAULT_DIR}/ or ${DEFAULT_SINGLE_FILE}`,
  )
}

function loadSingleFile(filePath: string): Config {
  if (!existsSync(filePath)) {
    throw new Error(`Config file not found: ${filePath}`)
  }

  const content = readFileSync(filePath, 'utf-8')
  const config = yaml.load(content) as Config

  if (!config || typeof config !== 'object') {
    throw new Error(`Invalid config in ${filePath}`)
  }

  return config
}

function loadFromDirectory(dirPath: string): Config {
  if (!existsSync(dirPath)) {
    throw new Error(`Config directory not found: ${dirPath}`)
  }

  const files = readdirSync(dirPath).filter(
    (f) => f.endsWith('.yaml') || f.endsWith('.yml'),
  )

  if (files.length === 0) {
    throw new Error(`No YAML files found in ${dirPath}`)
  }

  const config: Config = {}

  for (const file of files) {
    const filePath = join(dirPath, file)
    const content = readFileSync(filePath, 'utf-8')
    const parsed = yaml.load(content) as Record<string, unknown>

    if (!parsed || typeof parsed !== 'object') {
      continue
    }

    // Merge based on filename or content
    const baseName = file.replace(/\.(yaml|yml)$/, '')

    switch (baseName) {
      case 'repo':
        if (parsed.repo) {
          config.repo = parsed.repo as Config['repo']
        } else {
          config.repo = parsed as Config['repo']
        }
        break
      case 'topics':
        if (Array.isArray(parsed.topics)) {
          config.topics = parsed.topics
        } else if (Array.isArray(parsed)) {
          config.topics = parsed
        }
        break
      case 'labels':
        if (parsed.labels) {
          config.labels = parsed.labels as Config['labels']
        } else if (parsed.items) {
          config.labels = parsed as Config['labels']
        }
        break
      case 'branch-protection':
      case 'branch_protection':
        if (parsed.branch_protection) {
          config.branch_protection =
            parsed.branch_protection as Config['branch_protection']
        } else {
          config.branch_protection = parsed as Config['branch_protection']
        }
        break
      case 'secrets':
        if (parsed.secrets) {
          config.secrets = parsed.secrets as Config['secrets']
        } else {
          config.secrets = parsed as Config['secrets']
        }
        break
      case 'env':
        if (parsed.env) {
          config.env = parsed.env as Config['env']
        } else {
          config.env = parsed as Config['env']
        }
        break
      default:
        // For any other file, merge at top level
        Object.assign(config, parsed)
    }
  }

  return config
}

export function configToYaml(config: Config): string {
  return yaml.dump(config, {
    indent: 2,
    lineWidth: -1,
    noRefs: true,
    sortKeys: false,
  })
}

export function loadAndValidateConfig(options: LoadConfigOptions): Config {
  const config = loadConfig(options)
  const result = validateConfig(config)

  if (!result.valid) {
    printValidationErrors(result)
    throw new Error('Config validation failed')
  }

  return config
}

export function printValidationErrors(result: ValidationResult): void {
  console.error(colors.red('\nConfig validation failed:\n'))
  for (const error of result.errors) {
    const path = error.path || '(root)'
    console.error(colors.red(`  - ${path}: ${error.message}`))
  }
  console.error()
}
