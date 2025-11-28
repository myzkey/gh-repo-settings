import { spawnSync } from 'node:child_process'

export function ghApi<T>(
  method: string,
  endpoint: string,
  data?: Record<string, unknown>,
): T {
  const args = ['api', '-X', method, endpoint]

  if (data) {
    args.push('-H', 'Accept: application/vnd.github+json')
    args.push('--input', '-')
  }

  const result = spawnSync('gh', args, {
    input: data ? JSON.stringify(data) : undefined,
    encoding: 'utf-8',
    maxBuffer: 10 * 1024 * 1024,
  })

  if (result.error) {
    throw new Error(`Failed to execute gh: ${result.error.message}`)
  }

  if (result.status !== 0) {
    const errorMsg = result.stderr || result.stdout || 'Unknown error'
    throw new Error(`gh api failed: ${errorMsg}`)
  }

  if (!result.stdout || result.stdout.trim() === '') {
    return {} as T
  }

  try {
    return JSON.parse(result.stdout) as T
  } catch {
    return result.stdout as unknown as T
  }
}
