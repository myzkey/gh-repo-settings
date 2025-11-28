import type { IGitHubClient } from '../ports'

export interface DiffContext {
  client: IGitHubClient
  owner: string
  name: string
}
