# gh-repo-settings

[日本語](./README.ja.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

A GitHub CLI extension to manage repository settings via YAML configuration. Inspired by Terraform's workflow - define your desired state in code, see planned changes, and apply them.

## Features

- **Infrastructure as Code**: Define repository settings in YAML files
- **Terraform-like workflow**: `plan` to preview, `apply` to execute
- **Export existing settings**: Generate YAML from current repository configuration
- **Schema validation**: Validate configuration before applying
- **Multiple config formats**: Single file or directory-based configuration
- **Secrets/Env check**: Verify required secrets and environment variables exist

## Installation

```bash
gh extension install myzkey/gh-repo-settings
```

## Quick Start

```bash
# Initialize config interactively
gh repo-settings init

# Preview changes (like terraform plan)
gh repo-settings plan

# Apply changes
gh repo-settings apply
```

Default config paths (in priority order):
1. `.github/repo-settings/` (directory)
2. `.github/repo-settings.yaml` (single file)

## Commands

### `init` - Initialize configuration

Create a configuration file interactively.

```bash
# Create .github/repo-settings.yaml interactively
gh repo-settings init

# Specify output path
gh repo-settings init -o config.yaml

# Overwrite existing file
gh repo-settings init -f
```

### `export` - Export repository settings

Export current GitHub repository settings to YAML format.

```bash
# Export to stdout
gh repo-settings export

# Export to single file
gh repo-settings export -s .github/repo-settings.yaml

# Export to directory (multiple files)
gh repo-settings export -d .github/repo-settings/

# Include secret names
gh repo-settings export -s settings.yaml --include-secrets

# Export from specific repository
gh repo-settings export -r owner/repo -s settings.yaml
```

### `plan` - Preview changes

Validate configuration and show planned changes without applying them.

```bash
# Preview all changes (uses default config path)
gh repo-settings plan

# Specify config file
gh repo-settings plan -c custom-config.yaml

# Specify config directory
gh repo-settings plan -d .github/repo-settings/

# Schema validation only (no API calls)
gh repo-settings plan --schema-only

# Check only secrets existence
gh repo-settings plan --secrets

# Check only environment variables
gh repo-settings plan --env
```

### `apply` - Apply changes

Apply YAML configuration to the GitHub repository.

```bash
# Apply changes (uses default config path)
gh repo-settings apply

# Dry run (same as plan)
gh repo-settings apply --dry-run

# Specify config file
gh repo-settings apply -c custom-config.yaml

# Apply from directory
gh repo-settings apply -d .github/repo-settings/
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
└── env.yaml
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

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install gh-repo-settings
        run: gh extension install myzkey/gh-repo-settings
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Validate schema
        run: gh repo-settings plan --schema-only
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Check drift
        run: gh repo-settings plan
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Global Options

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Show debug output |
| `-q, --quiet` | Only show errors |
| `-r, --repo <owner/name>` | Target repository (default: current) |

## Development

```bash
# Install dependencies
pnpm install

# Build
pnpm build

# Run tests
pnpm test

# Lint
pnpm lint

# Type check
pnpm typecheck
```

## License

MIT
