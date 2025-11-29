# gh-repo-settings

[![CI](https://github.com/myzkey/gh-repo-settings/actions/workflows/lint.yml/badge.svg)](https://github.com/myzkey/gh-repo-settings/actions/workflows/lint.yml)
[![Test](https://github.com/myzkey/gh-repo-settings/actions/workflows/test.yml/badge.svg)](https://github.com/myzkey/gh-repo-settings/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/myzkey/gh-repo-settings)](https://github.com/myzkey/gh-repo-settings)
[![Release](https://img.shields.io/github/v/release/myzkey/gh-repo-settings)](https://github.com/myzkey/gh-repo-settings/releases)

[日本語](./README.ja.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

**Keep your repo settings, labels and branch protections in sync across dozens of repositories via YAML + GitHub CLI.**

A GitHub CLI extension to manage repository settings via YAML configuration. Inspired by Terraform's workflow - define your desired state in code, see planned changes, and apply them.

## Features

- **Infrastructure as Code**: Define repository settings in YAML files
- **Terraform-like workflow**: `plan` to preview, `apply` to execute
- **Export existing settings**: Generate YAML from current repository configuration
- **Schema validation**: Validate configuration before applying
- **Multiple config formats**: Single file or directory-based configuration
- **Secrets/Env check**: Verify required secrets and environment variables exist
- **Actions permissions**: Configure GitHub Actions permissions and workflow settings

## Installation

### Via GitHub CLI (Recommended)

```bash
gh extension install myzkey/gh-repo-settings
```

### Manual Installation

Download the latest binary from [Releases](https://github.com/myzkey/gh-repo-settings/releases) and add it to your PATH.

## Quick Start

```bash
# Initialize config interactively
gh rset init

# Preview changes (like terraform plan)
gh rset plan

# Apply changes
gh rset apply
```

Default config paths (in priority order):
1. `.github/repo-settings/` (directory)
2. `.github/repo-settings.yaml` (single file)

## Commands

### `init` - Initialize configuration

Create a configuration file interactively.

```bash
# Create .github/repo-settings.yaml interactively
gh rset init

# Specify output path
gh rset init -o config.yaml

# Overwrite existing file
gh rset init -f
```

### `export` - Export repository settings

Export current GitHub repository settings to YAML format.

```bash
# Export to stdout
gh rset export

# Export to single file
gh rset export -s .github/repo-settings.yaml

# Export to directory (multiple files)
gh rset export -d .github/repo-settings/

# Include secret names
gh rset export -s settings.yaml --include-secrets

# Export from specific repository
gh rset export -r owner/repo -s settings.yaml
```

### `plan` - Preview changes

Validate configuration and show planned changes without applying them.

```bash
# Preview all changes (uses default config path)
gh rset plan

# Specify config file
gh rset plan -c custom-config.yaml

# Specify config directory
gh rset plan -d .github/repo-settings/

# Schema validation only (no API calls)
gh rset plan --schema-only

# Check only secrets existence
gh rset plan --secrets

# Check only environment variables
gh rset plan --env
```

### `apply` - Apply changes

Apply YAML configuration to the GitHub repository.

```bash
# Apply changes (uses default config path)
gh rset apply

# Dry run (same as plan)
gh rset apply --dry-run

# Specify config file
gh rset apply -c custom-config.yaml

# Apply from directory
gh rset apply -d .github/repo-settings/
```

## Configuration

### Single File

Create `.github/repo-settings.yaml`:

```yaml
repo:
  description: "My awesome project"
  homepage: "https://example.com"
  visibility: public
  allow_merge_commit: false
  allow_rebase_merge: true
  allow_squash_merge: true
  delete_branch_on_merge: true

topics:
  - typescript
  - cli
  - github

labels:
  replace_default: true
  items:
    - name: bug
      color: ff0000
      description: Something isn't working
    - name: feature
      color: 0e8a16
      description: New feature request

branch_protection:
  main:
    required_reviews: 1
    dismiss_stale_reviews: true
    require_status_checks: true
    status_checks:
      - ci/test
      - ci/lint
    enforce_admins: false

secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY

env:
  required:
    - DATABASE_URL

actions:
  enabled: true
  allowed_actions: selected
  selected_actions:
    github_owned_allowed: true
    verified_allowed: true
    patterns_allowed:
      - "actions/*"
  default_workflow_permissions: read
  can_approve_pull_request_reviews: false
```

### Directory Structure

Alternatively, split configuration into multiple files:

```
.github/repo-settings/
├── repo.yaml
├── topics.yaml
├── labels.yaml
├── branch-protection.yaml
├── secrets.yaml
├── env.yaml
└── actions.yaml
```

## Configuration Reference

### `repo` - Repository Settings

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Repository description |
| `homepage` | string | Homepage URL |
| `visibility` | `public` \| `private` \| `internal` | Repository visibility |
| `allow_merge_commit` | boolean | Allow merge commits |
| `allow_rebase_merge` | boolean | Allow rebase merging |
| `allow_squash_merge` | boolean | Allow squash merging |
| `delete_branch_on_merge` | boolean | Auto-delete head branches |
| `allow_update_branch` | boolean | Allow updating PR branches |

### `topics` - Repository Topics

Array of topic strings:

```yaml
topics:
  - javascript
  - nodejs
  - cli
```

### `labels` - Issue Labels

| Field | Type | Description |
|-------|------|-------------|
| `replace_default` | boolean | Delete labels not in config |
| `items` | array | List of label definitions |
| `items[].name` | string | Label name |
| `items[].color` | string | Hex color (without `#`) |
| `items[].description` | string | Label description |

### `branch_protection` - Branch Protection Rules

```yaml
branch_protection:
  <branch_name>:
    # Pull request reviews
    required_reviews: 1          # Number of required approvals
    dismiss_stale_reviews: true  # Dismiss approvals on new commits
    require_code_owner: false    # Require CODEOWNERS review

    # Status checks
    require_status_checks: true  # Require status checks
    status_checks:               # Required status check names
      - ci/test
    strict_status_checks: false  # Require up-to-date branches

    # Deployments
    required_deployments:        # Required deployment environments
      - production

    # Commit requirements
    require_signed_commits: false # Require signed commits
    require_linear_history: false # Prevent merge commits

    # Push/merge restrictions
    enforce_admins: false        # Include administrators
    restrict_creations: false    # Restrict branch creation
    restrict_pushes: false       # Restrict who can push
    allow_force_pushes: false    # Allow force pushes
    allow_deletions: false       # Allow branch deletion
```

### `secrets` - Required Secrets

Check that required repository secrets exist (values are not managed):

```yaml
secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY
```

### `env` - Required Environment Variables

Check that required repository variables exist:

```yaml
env:
  required:
    - DATABASE_URL
    - SENTRY_DSN
```

### `actions` - GitHub Actions Permissions

Configure GitHub Actions permissions for the repository:

```yaml
actions:
  # Enable/disable GitHub Actions
  enabled: true

  # Which actions can be used: "all", "local_only", "selected"
  allowed_actions: selected

  # When allowed_actions is "selected"
  selected_actions:
    github_owned_allowed: true    # Allow actions from GitHub
    verified_allowed: true        # Allow actions from verified creators
    patterns_allowed:             # Allow specific action patterns
      - "actions/*"
      - "github/codeql-action/*"

  # Default GITHUB_TOKEN permissions: "read" or "write"
  default_workflow_permissions: read

  # Allow GitHub Actions to create/approve pull requests
  can_approve_pull_request_reviews: false
```

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | boolean | Enable GitHub Actions for this repository |
| `allowed_actions` | `all` \| `local_only` \| `selected` | Which actions are allowed |
| `selected_actions.github_owned_allowed` | boolean | Allow actions created by GitHub |
| `selected_actions.verified_allowed` | boolean | Allow actions from verified creators |
| `selected_actions.patterns_allowed` | array | Patterns for allowed actions |
| `default_workflow_permissions` | `read` \| `write` | Default GITHUB_TOKEN permissions |
| `can_approve_pull_request_reviews` | boolean | Allow Actions to approve PRs |

## Editor Integration (VSCode)

This project provides a JSON Schema for YAML validation and auto-completion in VSCode.

### Setup

1. Install the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

2. Add to your `.vscode/settings.json`:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/myzkey/gh-repo-settings/main/schema.json": [
      ".github/repo-settings.yaml",
      ".github/repo-settings/*.yaml"
    ]
  }
}
```

### Features

- Auto-completion for all fields
- Hover documentation
- Enum suggestions (`public`/`private`/`internal`, `read`/`write`, etc.)
- Unknown field detection
- Type validation

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Repo Settings Check

on:
  pull_request:
    paths:
      - ".github/repo-settings.yaml"
      - ".github/repo-settings/**"

jobs:
  check:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      actions: read

    steps:
      - uses: actions/checkout@v4

      - name: Install gh-repo-settings
        run: gh extension install myzkey/gh-repo-settings
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Check drift
        run: gh rset plan
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Global Options

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Show debug output |
| `-q, --quiet` | Only show errors |
| `-r, --repo <owner/name>` | Target repository (default: current) |

## Authentication & Permissions

This extension uses the GitHub CLI (`gh`) for authentication. Make sure you're logged in:

```bash
gh auth login
```

### Required Token Permissions

| Feature | Required Scopes |
|---------|-----------------|
| Repository settings | `repo` |
| Branch protection | `repo` (admin access to repository) |
| Secrets check | `repo`, `admin:repo_hook` |
| Environment variables | `repo` |
| Actions permissions | `repo`, `admin:repo_hook` |

### Token Types

- **`GITHUB_TOKEN` (GitHub Actions)**: Works for most operations within the same repository. However, branch protection rules require admin access, which `GITHUB_TOKEN` may not have by default.
- **Personal Access Token (PAT)**: Required for cross-repository operations or when `GITHUB_TOKEN` lacks sufficient permissions. Use a fine-grained PAT with `Repository administration` permission for full functionality.

## Development

```bash
# Build
make build

# Run tests
make test

# Lint (requires golangci-lint)
make lint

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## License

MIT
