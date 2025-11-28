import { colors, logger } from '~/utils/logger'
import type { DiffItem } from '../types'

export function printDiff(diffs: DiffItem[]): void {
  if (diffs.length === 0) {
    logger.success('No changes detected.')
    return
  }

  logger.heading('\nPlanned changes:\n')

  const grouped = groupBy(diffs, (d) => d.type)

  for (const [type, items] of Object.entries(grouped)) {
    logger.section(`[${type}]`)

    for (const item of items) {
      const icon = getActionIcon(item.action)
      const color = getActionColor(item.action)
      logger.log(`  ${icon} ${color(item.details)}`)
      if (item.apiCall) {
        logger.debug(`    API: ${item.apiCall}`)
      }
    }
    logger.log('')
  }
}

function getActionIcon(action: DiffItem['action']): string {
  switch (action) {
    case 'create':
      return colors.green('+')
    case 'update':
      return colors.yellow('~')
    case 'delete':
      return colors.red('-')
    case 'check':
      return colors.blue('?')
  }
}

function getActionColor(action: DiffItem['action']): (text: string) => string {
  switch (action) {
    case 'create':
      return colors.green
    case 'update':
      return colors.yellow
    case 'delete':
      return colors.red
    case 'check':
      return colors.blue
  }
}

function groupBy<T>(arr: T[], fn: (item: T) => string): Record<string, T[]> {
  return arr.reduce(
    (acc, item) => {
      const key = fn(item)
      if (!acc[key]) {
        acc[key] = []
      }
      acc[key].push(item)
      return acc
    },
    {} as Record<string, T[]>,
  )
}
