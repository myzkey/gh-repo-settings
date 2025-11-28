import {
  existsSync,
  join,
  readFileSync,
  readdirSync,
} from '~/infra/fs'
import { parse, stringify } from '~/infra/yaml'
import type { Config } from '../types'

const DEFAULT_DIR = '.github/repo-settings'
const DEFAULT_SINGLE_FILE = '.github/repo-settings.yaml'

export interface LoadConfigOptions {
  dir?: string
  config?: string
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
  const config = parse<Config>(content)

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
    const parsed = parse<Record<string, unknown>>(content)

    if (!parsed || typeof parsed !== 'object') {
      continue
    }

    // Merge based on filename or content
    const baseName = file.replace(/\.(yaml|yml)$/, '')

    switch (baseName) {
      case 'repo':
        config.repo = (parsed.repo ?? parsed) as Config['repo']
        break
      case 'topics':
        if (Array.isArray(parsed.topics)) {
          config.topics = parsed.topics
        } else if (Array.isArray(parsed)) {
          config.topics = parsed
        }
        break
      case 'labels':
        config.labels = (parsed.labels ?? parsed) as Config['labels']
        break
      case 'branch-protection':
      case 'branch_protection':
        config.branch_protection = (parsed.branch_protection ??
          parsed) as Config['branch_protection']
        break
      case 'secrets':
        config.secrets = (parsed.secrets ?? parsed) as Config['secrets']
        break
      case 'env':
        config.env = (parsed.env ?? parsed) as Config['env']
        break
      default:
        // For any other file, merge at top level
        Object.assign(config, parsed)
    }
  }

  return config
}

export function configToYaml(config: Config): string {
  return stringify(config)
}
