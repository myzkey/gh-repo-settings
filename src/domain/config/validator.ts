import { logger } from '~/utils/logger'
import type { Config, ValidationResult } from '../types'
import { type LoadConfigOptions, loadConfig } from './loader'
import { validateConfig } from './schema/config'

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
  logger.error('\nConfig validation failed:\n')
  for (const error of result.errors) {
    const path = error.path || '(root)'
    logger.error(`  - ${path}: ${error.message}`)
  }
  logger.log('')
}
