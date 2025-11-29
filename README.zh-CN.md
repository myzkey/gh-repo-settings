# gh-repo-settings

[English](./README.md) | [日本語](./README.ja.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

通过 YAML 配置管理 GitHub 仓库设置的 GitHub CLI 扩展。灵感来自 Terraform 的工作流程——用代码定义期望状态，预览变更，然后应用。

## 特性

- **基础设施即代码**: 用 YAML 文件定义仓库设置
- **Terraform 风格工作流**: `plan` 预览，`apply` 执行
- **导出现有设置**: 从当前仓库配置生成 YAML
- **Schema 验证**: 应用前验证配置
- **多种配置格式**: 单文件或目录结构配置
- **Secrets/Env 检查**: 验证必需的密钥和环境变量是否存在

## 安装

```bash
gh extension install myzkey/gh-repo-settings
```

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

# 仅验证 Schema（不调用 API）
gh repo-settings plan --schema-only

# 仅检查密钥是否存在
gh repo-settings plan --secrets

# 仅检查环境变量
gh repo-settings plan --env
```

### `apply` - 应用变更

将 YAML 配置应用到 GitHub 仓库。

```bash
# 应用变更（使用默认配置路径）
gh repo-settings apply

# 演练模式（与 plan 相同）
gh repo-settings apply --dry-run

# 指定配置文件
gh repo-settings apply -c custom-config.yaml

# 从目录应用
gh repo-settings apply -d .github/repo-settings/
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

secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY

env:
  required:
    - DATABASE_URL
```

### 目录结构

也可以将配置拆分为多个文件：

```
.github/repo-settings/
├── repo.yaml
├── topics.yaml
├── labels.yaml
├── branch-protection.yaml
├── secrets.yaml
└── env.yaml
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
    required_reviews: 1          # 所需审批数
    dismiss_stale_reviews: true  # 新提交时撤销审批
    require_code_owner: false    # 需要 CODEOWNERS 审查
    require_status_checks: true  # 需要状态检查
    status_checks:               # 必需的状态检查名称
      - ci/test
    strict_status_checks: false  # 需要最新分支
    enforce_admins: false        # 对管理员也适用
    restrict_pushes: false       # 限制推送
    allow_force_pushes: false    # 允许强制推送
    allow_deletions: false       # 允许删除分支
```

### `secrets` - 必需密钥

检查必需的仓库密钥是否存在（不管理值）：

```yaml
secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY
```

### `env` - 必需环境变量

检查必需的仓库变量是否存在：

```yaml
env:
  required:
    - DATABASE_URL
    - SENTRY_DSN
```

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

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install gh-repo-settings
        run: gh extension install myzkey/gh-repo-settings
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: 验证 Schema
        run: gh repo-settings plan --schema-only
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: 检查设置差异
        run: gh repo-settings plan
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 全局选项

| 选项 | 描述 |
|-----|------|
| `-v, --verbose` | 显示调试输出 |
| `-q, --quiet` | 仅显示错误 |
| `-r, --repo <owner/name>` | 目标仓库（默认：当前仓库） |

## 开发

```bash
# 安装依赖
pnpm install

# 构建
pnpm build

# 运行测试
pnpm test

# 代码检查
pnpm lint

# 类型检查
pnpm typecheck
```

## 许可证

MIT
