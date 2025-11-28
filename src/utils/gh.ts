import { execSync, spawnSync } from "node:child_process";
import type { RepoInfo } from "~/types";

export function ghApi<T>(
  method: string,
  endpoint: string,
  data?: Record<string, unknown>
): T {
  const args = ["api", "-X", method, endpoint];

  if (data) {
    args.push("-H", "Accept: application/vnd.github+json");
    args.push("--input", "-");
  }

  const result = spawnSync("gh", args, {
    input: data ? JSON.stringify(data) : undefined,
    encoding: "utf-8",
    maxBuffer: 10 * 1024 * 1024,
  });

  if (result.error) {
    throw new Error(`Failed to execute gh: ${result.error.message}`);
  }

  if (result.status !== 0) {
    const errorMsg = result.stderr || result.stdout || "Unknown error";
    throw new Error(`gh api failed: ${errorMsg}`);
  }

  if (!result.stdout || result.stdout.trim() === "") {
    return {} as T;
  }

  try {
    return JSON.parse(result.stdout) as T;
  } catch {
    return result.stdout as unknown as T;
  }
}

export function ghApiGet<T>(endpoint: string): T {
  return ghApi<T>("GET", endpoint);
}

export function ghApiPost<T>(
  endpoint: string,
  data?: Record<string, unknown>
): T {
  return ghApi<T>("POST", endpoint, data);
}

export function ghApiPatch<T>(
  endpoint: string,
  data: Record<string, unknown>
): T {
  return ghApi<T>("PATCH", endpoint, data);
}

export function ghApiPut<T>(
  endpoint: string,
  data?: Record<string, unknown>
): T {
  return ghApi<T>("PUT", endpoint, data);
}

export function ghApiDelete(endpoint: string): void {
  ghApi<void>("DELETE", endpoint);
}

export function getCurrentRepo(): RepoInfo | null {
  try {
    const result = execSync(
      "gh repo view --json owner,name --jq '.owner.login + \"/\" + .name'",
      { encoding: "utf-8" }
    ).trim();

    if (!result) return null;

    const [owner, name] = result.split("/");
    if (!owner || !name) return null;

    return { owner, name };
  } catch {
    return null;
  }
}

export function parseRepoArg(repo: string): RepoInfo {
  const parts = repo.split("/");
  if (parts.length !== 2) {
    throw new Error(`Invalid repo format: ${repo}. Expected: owner/name`);
  }
  return { owner: parts[0], name: parts[1] };
}

export function getRepoInfo(repoArg?: string): RepoInfo {
  if (repoArg) {
    return parseRepoArg(repoArg);
  }

  const current = getCurrentRepo();
  if (!current) {
    throw new Error(
      "Could not determine repository. Use --repo owner/name or run from a git repository."
    );
  }

  return current;
}
