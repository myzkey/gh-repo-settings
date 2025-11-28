import { execSync } from 'node:child_process'
import type { RepoIdentifier } from '~/domain'
import { GitHubClient } from './client'

export function getCurrentRepo(): RepoIdentifier | null {
  try {
    const result = execSync(
      'gh repo view --json owner,name --jq \'.owner.login + "/" + .name\'',
      { encoding: 'utf-8' },
    ).trim()

    if (!result) return null

    const [owner, name] = result.split('/')
    if (!owner || !name) return null

    return { owner, name }
  } catch {
    return null
  }
}

export function parseRepoArg(repo: string): RepoIdentifier {
  const parts = repo.split('/')
  if (parts.length !== 2) {
    throw new Error(`Invalid repo format: ${repo}. Expected: owner/name`)
  }
  return { owner: parts[0], name: parts[1] }
}

export function getRepoInfo(repoArg?: string): RepoIdentifier {
  if (repoArg) {
    return parseRepoArg(repoArg)
  }

  const current = getCurrentRepo()
  if (!current) {
    throw new Error(
      'Could not determine repository. Use --repo owner/name or run from a git repository.',
    )
  }

  return current
}

export function createClient(repoArg?: string): GitHubClient {
  const repo = getRepoInfo(repoArg)
  return new GitHubClient(repo)
}
