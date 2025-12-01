# gh-repo-settings

[English](../README.md) | [日本語](./README.ja.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

通过 YAML 配置管理 GitHub 仓库设置的 GitHub CLI 扩展。灵感来自 Terraform 的工作流程——用代码定义期望状态，预览变更，然后应用。

## 特性

- **基础设施即代码**: 用 YAML 文件定义仓库设置
- **Terraform 风格工作流**: `plan` 预览，`apply` 执行
- **导出现有设置**: 从当前仓库配置生成 YAML
- **Schema 验证**: 应用前验证配置
- **多种配置格式**: 单文件或目录结构配置
- **Secrets/Env 检查**: 验证必需的密钥和环境变量是否存在
- **Actions 权限设置**: 配置 GitHub Actions 权限和工作流设置

## 安装

### 通过 GitHub CLI（推荐）

```bash
gh extension install myzkey/gh-repo-settings
```

### 升级

```bash
gh extension upgrade myzkey/gh-repo-settings
```

### 手动安装

从 [Releases](https://github.com/myzkey/gh-repo-settings/releases) 下载最新的二进制文件并添加到 PATH。

## 快速开始

```bash
# 交互式创建配置文件
gh repo-settings init

# 预览变更（类似 terraform plan）
gh repo-settings plan

# 应用变更
gh repo-settings apply
```

默认配置文件路径（按优先级）：
1. `.github/repo-settings/`（目录）
2. `.github/repo-settings.yaml`（单文件）

## 命令

### `init` - 初始化配置文件

交互式创建配置文件。

```bash
# 交互式创建 .github/repo-settings.yaml
gh repo-settings init

# 指定输出路径
gh repo-settings init -o config.yaml

# 覆盖现有文件
gh repo-settings init -f
```

### `export` - 导出仓库设置

将当前 GitHub 仓库设置导出为 YAML 格式。

```bash
# 导出到标准输出
gh repo-settings export

# 导出到单个文件
gh repo-settings export -s .github/repo-settings.yaml

# 导出到目录（多个文件）
gh repo-settings export -d .github/repo-settings/

# 包含密钥名称
gh repo-settings export -s settings.yaml --include-secrets

# 从指定仓库导出
gh repo-settings export -r owner/repo -s settings.yaml
```

### `plan` - 预览变更

验证配置并显示计划的变更，不实际应用。

```bash
# 预览所有变更（使用默认配置路径）
gh repo-settings plan

# 指定配置文件
gh repo-settings plan -c custom-config.yaml

# 使用目录配置预览
gh repo-settings plan -d .github/repo-settings/

# 显示当前 GitHub 设置（用于调试）
gh repo-settings plan --show-current

# 检查密钥
gh repo-settings plan --secrets

# 检查环境变量
gh repo-settings plan --env

# 显示配置中没有的变量/密钥的删除计划
gh repo-settings plan --env --secrets --sync
```

`--show-current` 选项显示当前 GitHub 仓库设置，适用于：
- 调试配置问题
- 查找 GitHub 上存在但配置文件中没有的设置
- 验证仓库的实际配置

**Status Check 验证**: 运行 `plan` 时，工具会自动验证分支保护规则中的 `status_checks` 名称是否与 `.github/workflows/` 文件中的作业名称匹配。如果发现不匹配，将显示警告：

```
⚠ status check lint not found in workflows
⚠ status check test not found in workflows
  Available checks: build, golangci-lint, Run tests
```

### `apply` - 应用变更

将 YAML 配置应用到 GitHub 仓库。

```bash
# 应用变更（使用默认配置路径）
gh repo-settings apply

# 无需确认自动批准
gh repo-settings apply -y

# 指定配置文件
gh repo-settings apply -c custom-config.yaml

# 从目录应用
gh repo-settings apply -d .github/repo-settings/

# 应用变量和密钥
gh repo-settings apply --env --secrets

# 同步模式：删除配置中没有的变量/密钥
gh repo-settings apply --env --secrets --sync
```

## 配置

### 单文件

创建 `.github/repo-settings.yaml`：

```yaml
repo:
  description: "我的优秀项目"
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
      description: 出了问题
    - name: feature
      color: 0e8a16
      description: 新功能请求

branch_protection:
  main:
    required_reviews: 1
    dismiss_stale_reviews: true
    require_status_checks: true
    status_checks:
      - ci/test
      - ci/lint
    enforce_admins: false

env:
  variables:
    NODE_ENV: production
    API_URL: https://api.example.com
  secrets:
    - API_TOKEN
    - DEPLOY_KEY

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

### 目录结构

也可以将配置拆分为多个文件：

```
.github/repo-settings/
├── repo.yaml
├── topics.yaml
├── labels.yaml
├── branch-protection.yaml
├── env.yaml
└── actions.yaml
```

## 配置参考

### `repo` - 仓库设置

| 字段 | 类型 | 描述 |
|-----|------|------|
| `description` | string | 仓库描述 |
| `homepage` | string | 主页 URL |
| `visibility` | `public` \| `private` \| `internal` | 仓库可见性 |
| `allow_merge_commit` | boolean | 允许合并提交 |
| `allow_rebase_merge` | boolean | 允许变基合并 |
| `allow_squash_merge` | boolean | 允许压缩合并 |
| `delete_branch_on_merge` | boolean | 合并后自动删除分支 |
| `allow_update_branch` | boolean | 允许更新 PR 分支 |

### `topics` - 仓库主题

主题字符串数组：

```yaml
topics:
  - javascript
  - nodejs
  - cli
```

### `labels` - Issue 标签

| 字段 | 类型 | 描述 |
|-----|------|------|
| `replace_default` | boolean | 删除配置中没有的标签 |
| `items` | array | 标签定义列表 |
| `items[].name` | string | 标签名称 |
| `items[].color` | string | 十六进制颜色（不带 `#`） |
| `items[].description` | string | 标签描述 |

### `branch_protection` - 分支保护规则

```yaml
branch_protection:
  <分支名>:
    # 拉取请求审查
    required_reviews: 1          # 所需审批数
    dismiss_stale_reviews: true  # 新提交时撤销审批
    require_code_owner: false    # 需要 CODEOWNERS 审查

    # 状态检查
    require_status_checks: true  # 需要状态检查
    status_checks:               # 必需的状态检查名称
      - ci/test
    strict_status_checks: false  # 需要最新分支

    # 部署
    required_deployments:        # 必需的部署环境
      - production

    # 提交要求
    require_signed_commits: false # 需要签名提交
    require_linear_history: false # 禁止合并提交

    # 推送/合并限制
    enforce_admins: false        # 对管理员也适用
    restrict_creations: false    # 限制分支创建
    restrict_pushes: false       # 限制推送
    allow_force_pushes: false    # 允许强制推送
    allow_deletions: false       # 允许删除分支
```

### `env` - 环境变量和密钥

管理仓库的变量和密钥：

```yaml
env:
  # 带默认值的变量（可通过 .env 文件覆盖）
  variables:
    NODE_ENV: production
    API_URL: https://api.example.com
  # 密钥名称（值来自 .env 文件或交互式输入）
  secrets:
    - API_TOKEN
    - DEPLOY_KEY
```

| 字段 | 类型 | 描述 |
|-----|------|------|
| `variables` | map | 仓库变量的键值对 |
| `secrets` | array | 要管理的密钥名称列表 |

#### 使用 `.env` 文件

创建 `.github/.env` 文件（建议添加到 gitignore）来存储实际值：

```bash
# .github/.env
NODE_ENV=staging
API_URL=https://staging-api.example.com
API_TOKEN=your-secret-token
DEPLOY_KEY=your-deploy-key
```

**优先级**：`.env` 文件的值会覆盖 YAML 中的默认值。

#### 命令

```bash
# 预览变量/密钥的变更
gh repo-settings plan --env --secrets

# 应用变量和密钥
gh repo-settings apply --env --secrets

# 删除配置中没有的变量/密钥（同步模式）
gh repo-settings apply --env --secrets --sync
```

如果密钥的值在 `.env` 中不存在，`apply` 时会提示交互式输入。

### `actions` - GitHub Actions 权限设置

配置仓库的 GitHub Actions 权限：

```yaml
actions:
  # 启用/禁用 GitHub Actions
  enabled: true

  # 允许哪些 actions: "all", "local_only", "selected"
  allowed_actions: selected

  # 当 allowed_actions 为 "selected" 时
  selected_actions:
    github_owned_allowed: true    # 允许 GitHub 官方 actions
    verified_allowed: true        # 允许已验证创建者的 actions
    patterns_allowed:             # 允许的 action 模式
      - "actions/*"
      - "github/codeql-action/*"

  # 默认 GITHUB_TOKEN 权限: "read" 或 "write"
  default_workflow_permissions: read

  # 允许 GitHub Actions 创建/批准拉取请求
  can_approve_pull_request_reviews: false
```

| 字段 | 类型 | 描述 |
|-----|------|------|
| `enabled` | boolean | 启用 GitHub Actions |
| `allowed_actions` | `all` \| `local_only` \| `selected` | 允许的 actions |
| `selected_actions.github_owned_allowed` | boolean | 允许 GitHub 官方 actions |
| `selected_actions.verified_allowed` | boolean | 允许已验证创建者 |
| `selected_actions.patterns_allowed` | array | 允许的 action 模式 |
| `default_workflow_permissions` | `read` \| `write` | GITHUB_TOKEN 默认权限 |
| `can_approve_pull_request_reviews` | boolean | 允许 Actions 批准 PR |

### `pages` - GitHub Pages 配置

配置仓库的 GitHub Pages：

```yaml
pages:
  # 构建类型: "workflow" (GitHub Actions) 或 "legacy" (基于分支)
  build_type: workflow

  # 源配置（仅用于 legacy 构建类型）
  source:
    branch: main
    path: /docs  # "/" 或 "/docs"
```

| 字段 | 类型 | 描述 |
|-----|------|------|
| `build_type` | `workflow` \| `legacy` | Pages 构建方式 |
| `source.branch` | string | legacy 构建的分支 |
| `source.path` | `/` \| `/docs` | 分支内的路径 |

## 编辑器集成 (VSCode)

本项目提供 JSON Schema 用于 VSCode 中的 YAML 验证和自动补全。

### 设置

1. 安装 [YAML 扩展](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

2. 添加到 `.vscode/settings.json`:

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

### 功能

- 所有字段的自动补全
- 悬停显示文档
- 枚举值建议（`public`/`private`/`internal`、`read`/`write` 等）
- 未知字段检测
- 类型验证

## CI/CD 集成

### GitHub Actions 工作流

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
        run: gh repo-settings plan
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 全局选项

| 选项 | 描述 |
|-----|------|
| `-v, --verbose` | 显示调试输出 |
| `-q, --quiet` | 仅显示错误 |
| `-r, --repo <owner/name>` | 目标仓库（默认：当前仓库） |

## 开发

```bash
# 构建
make build

# 运行测试
make test

# 代码检查（需要 golangci-lint）
make lint

# 全平台构建
make build-all

# 清理构建产物
make clean
```

## 许可证

MIT
