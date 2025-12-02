# gh-repo-settings

[English](../README.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

GitHub リポジトリの設定を YAML で管理する GitHub CLI 拡張機能。Terraform のワークフローにインスパイアされており、望む状態をコードで定義し、変更をプレビューしてから適用できます。

## なぜ gh-repo-settings？

GitHub リポジトリの設定を一貫して管理するのは難しいです：

- Settings UI をクリックして回るのはスケールしない
- リポジトリ管理者が望ましい設定からドリフトしがち
- Terraform の GitHub Provider は強力だが：
  - 別途バックエンド（状態管理）が必要
  - GitHub Provider の認証設定が必要
  - HCL ファイルをリポジトリとは別に管理する必要がある

**gh-repo-settings** は以下を提供します：

- **バックエンド不要** - 状態ファイルの管理が不要
- **外部依存なし** - GitHub CLI だけで動作
- **YAML 設定** - `.github/` にコードと一緒に配置
- **Terraform 風ワークフロー** - お馴染みの `plan` / `apply` コマンド
- **ワークフロー検証** - `status_checks` と実際のワークフロージョブ名の不一致を検出

## 特徴

- **Infrastructure as Code**: リポジトリ設定を YAML ファイルで定義
- **Terraform 風ワークフロー**: `plan` でプレビュー、`apply` で実行
- **既存設定のエクスポート**: 現在のリポジトリ設定から YAML を生成
- **スキーマ検証**: 適用前に設定を検証
- **複数の設定形式**: 単一ファイルまたはディレクトリベースの設定
- **Secrets/Env チェック**: 必要なシークレットと環境変数の存在確認
- **Actions 権限設定**: GitHub Actions の権限とワークフロー設定を管理

## インストール

### GitHub CLI 経由（推奨）

```bash
gh extension install myzkey/gh-repo-settings
```

### アップグレード

```bash
gh extension upgrade myzkey/gh-repo-settings
```

### 手動インストール

[Releases](https://github.com/myzkey/gh-repo-settings/releases) から最新のバイナリをダウンロードし、PATH に追加してください。

## クイックスタート

```bash
# 対話形式で設定ファイルを作成
gh repo-settings init

# 変更をプレビュー（terraform plan のように）
gh repo-settings plan

# 変更を適用
gh repo-settings apply
```

デフォルトの設定ファイルパス（優先順）:
1. `.github/repo-settings/`（ディレクトリ）
2. `.github/repo-settings.yaml`（単一ファイル）

## コマンド

### `init` - 設定ファイルの初期化

対話形式で設定ファイルを作成します。

```bash
# .github/repo-settings.yaml を対話形式で作成
gh repo-settings init

# 出力先を指定
gh repo-settings init -o config.yaml

# 既存ファイルを上書き
gh repo-settings init -f
```

### `export` - リポジトリ設定のエクスポート

現在の GitHub リポジトリ設定を YAML 形式でエクスポートします。

```bash
# 標準出力にエクスポート
gh repo-settings export

# 単一ファイルにエクスポート
gh repo-settings export -s .github/repo-settings.yaml

# ディレクトリにエクスポート（複数ファイル）
gh repo-settings export -d .github/repo-settings/

# シークレット名を含める
gh repo-settings export -s settings.yaml --include-secrets

# 特定のリポジトリからエクスポート
gh repo-settings export -r owner/repo -s settings.yaml
```

### `plan` - 変更のプレビュー

設定を検証し、適用せずに計画された変更を表示します。

#### 出力例

```diff
Repository: owner/my-repo

repo:
  ~ description: "古い説明" → "新しい説明"

labels:
  + feature (color: 0e8a16)
  ~ bug: color ff0000 → d73a4a
  - old-label

branch_protection (main):
  ~ required_reviews: 1 → 2

Plan: 2 to add, 2 to change, 1 to delete
```

```bash
# すべての変更をプレビュー（デフォルトパスを使用）
gh repo-settings plan

# 設定ファイルを指定
gh repo-settings plan -c custom-config.yaml

# ディレクトリ設定でプレビュー
gh repo-settings plan -d .github/repo-settings/

# 現在のGitHub設定を表示（デバッグ用）
gh repo-settings plan --show-current

# シークレットをチェック
gh repo-settings plan --secrets

# 環境変数をチェック
gh repo-settings plan --env

# 設定にない変数/シークレットの削除を表示
gh repo-settings plan --env --secrets --sync
```

`--show-current` オプションは現在のGitHub設定を表示します。これは以下の場合に便利です：
- 設定の問題をデバッグする
- GitHubに存在するが設定ファイルにない設定を見つける
- リポジトリの実際の設定を確認する

**Status Check 検証**: `plan` 実行時、ブランチ保護ルールの `status_checks` 名が `.github/workflows/` ファイルのジョブ名と一致するか自動で検証します。不一致がある場合、警告が表示されます：

```
⚠ status check lint not found in workflows
⚠ status check test not found in workflows
  Available checks: build, golangci-lint, Run tests
```

### `apply` - 変更の適用

YAML 設定を GitHub リポジトリに適用します。

```bash
# 変更を適用（デフォルトパスを使用）
gh repo-settings apply

# 確認なしで自動承認
gh repo-settings apply -y

# 設定ファイルを指定
gh repo-settings apply -c custom-config.yaml

# ディレクトリから適用
gh repo-settings apply -d .github/repo-settings/

# 変数とシークレットを適用
gh repo-settings apply --env --secrets

# 同期モード: 設定にない変数/シークレットを削除
gh repo-settings apply --env --secrets --sync
```

### ⚠️ 同期モードの注意

`--sync` フラグは**破壊的な操作**を有効にします：

- 設定に定義されていないラベルを削除（`labels.replace_default: true` の場合）
- 設定に定義されていない変数を削除
- 設定に定義されていないシークレットを削除

**適用前に必ず `plan --sync` を実行**して、削除される内容を確認してください：

```bash
# 適用前に削除内容をプレビュー
gh repo-settings plan --env --secrets --sync

# プランが正しければ適用
gh repo-settings apply --env --secrets --sync
```

> **ヒント**: CI で `--sync` を使う場合は、人間によるレビューなしで実行しないでください。

## 設定

### 単一ファイル

`.github/repo-settings.yaml` を作成:

```yaml
repo:
  description: "素晴らしいプロジェクト"
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
      description: 不具合報告
    - name: feature
      color: 0e8a16
      description: 新機能リクエスト

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

### ディレクトリ構造

設定を複数ファイルに分割することもできます:

```
.github/repo-settings/
├── repo.yaml
├── topics.yaml
├── labels.yaml
├── branch-protection.yaml
├── env.yaml
└── actions.yaml
```

## 設定リファレンス

### `repo` - リポジトリ設定

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `description` | string | リポジトリの説明 |
| `homepage` | string | ホームページ URL |
| `visibility` | `public` \| `private` \| `internal` | リポジトリの可視性 |
| `allow_merge_commit` | boolean | マージコミットを許可 |
| `allow_rebase_merge` | boolean | リベースマージを許可 |
| `allow_squash_merge` | boolean | スカッシュマージを許可 |
| `delete_branch_on_merge` | boolean | マージ後にブランチを自動削除 |
| `allow_update_branch` | boolean | PR ブランチの更新を許可 |

### `topics` - リポジトリトピック

トピック文字列の配列:

```yaml
topics:
  - javascript
  - nodejs
  - cli
```

### `labels` - Issue ラベル

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `replace_default` | boolean | 設定にないラベルを削除 |
| `items` | array | ラベル定義のリスト |
| `items[].name` | string | ラベル名 |
| `items[].color` | string | 16進数カラー（`#` なし） |
| `items[].description` | string | ラベルの説明 |

### `branch_protection` - ブランチ保護ルール

```yaml
branch_protection:
  <ブランチ名>:
    # プルリクエストレビュー
    required_reviews: 1          # 必要な承認数
    dismiss_stale_reviews: true  # 新しいコミットで承認を却下
    require_code_owner: false    # CODEOWNERS のレビューを必須

    # ステータスチェック
    require_status_checks: true  # ステータスチェックを必須
    status_checks:               # 必須のステータスチェック名
      - ci/test
    strict_status_checks: false  # 最新ブランチを必須

    # デプロイメント
    required_deployments:        # 必須のデプロイ環境
      - production

    # コミット要件
    require_signed_commits: false # 署名付きコミットを必須
    require_linear_history: false # マージコミットを禁止

    # プッシュ/マージ制限
    enforce_admins: false        # 管理者にも適用
    restrict_creations: false    # ブランチ作成を制限
    restrict_pushes: false       # プッシュを制限
    allow_force_pushes: false    # 強制プッシュを許可
    allow_deletions: false       # ブランチ削除を許可
```

### `env` - 環境変数とシークレット

リポジトリの変数とシークレットを管理:

```yaml
env:
  # デフォルト値付きの変数（.env ファイルで上書き可能）
  variables:
    NODE_ENV: production
    API_URL: https://api.example.com
  # シークレット名（値は .env ファイルまたは対話入力から）
  secrets:
    - API_TOKEN
    - DEPLOY_KEY
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `variables` | map | リポジトリ変数のキーと値 |
| `secrets` | array | 管理するシークレット名のリスト |

#### `.env` ファイルの使用

実際の値を格納する `.github/.env` ファイル（gitignore 推奨）を作成:

```bash
# .github/.env
NODE_ENV=staging
API_URL=https://staging-api.example.com
API_TOKEN=your-secret-token
DEPLOY_KEY=your-deploy-key
```

**優先順位**: `.env` ファイルの値は YAML のデフォルト値を上書きします。

#### コマンド

```bash
# 変数/シークレットの変更をプレビュー
gh repo-settings plan --env --secrets

# 変数とシークレットを適用
gh repo-settings apply --env --secrets

# 設定にない変数/シークレットを削除（同期モード）
gh repo-settings apply --env --secrets --sync
```

シークレットの値が `.env` にない場合、`apply` 時に対話形式で入力を求められます。

### `actions` - GitHub Actions 権限設定

リポジトリの GitHub Actions 権限を設定:

```yaml
actions:
  # GitHub Actions の有効/無効
  enabled: true

  # 使用可能なアクション: "all", "local_only", "selected"
  allowed_actions: selected

  # allowed_actions が "selected" の場合
  selected_actions:
    github_owned_allowed: true    # GitHub 製アクションを許可
    verified_allowed: true        # 認証済み作成者のアクションを許可
    patterns_allowed:             # 許可するアクションのパターン
      - "actions/*"
      - "github/codeql-action/*"

  # デフォルトの GITHUB_TOKEN 権限: "read" または "write"
  default_workflow_permissions: read

  # GitHub Actions による PR 作成/承認を許可
  can_approve_pull_request_reviews: false
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `enabled` | boolean | GitHub Actions を有効にする |
| `allowed_actions` | `all` \| `local_only` \| `selected` | 許可するアクション |
| `selected_actions.github_owned_allowed` | boolean | GitHub 製アクションを許可 |
| `selected_actions.verified_allowed` | boolean | 認証済み作成者を許可 |
| `selected_actions.patterns_allowed` | array | 許可するアクションのパターン |
| `default_workflow_permissions` | `read` \| `write` | GITHUB_TOKEN のデフォルト権限 |
| `can_approve_pull_request_reviews` | boolean | Actions による PR 承認を許可 |

### `pages` - GitHub Pages 設定

リポジトリの GitHub Pages を設定:

```yaml
pages:
  # ビルドタイプ: "workflow" (GitHub Actions) または "legacy" (ブランチベース)
  build_type: workflow

  # ソース設定（legacy ビルドタイプの場合のみ）
  source:
    branch: main
    path: /docs  # "/" または "/docs"
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `build_type` | `workflow` \| `legacy` | Pages のビルド方法 |
| `source.branch` | string | legacy ビルドのブランチ |
| `source.path` | `/` \| `/docs` | ブランチ内のパス |

## エディタ連携 (VSCode)

VSCode での YAML 検証と自動補完のための JSON Schema を提供しています。

### セットアップ

1. [YAML 拡張機能](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)をインストール

2. `.vscode/settings.json` に追加:

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

### 機能

- すべてのフィールドの自動補完
- ホバーでドキュメント表示
- enum の候補表示（`public`/`private`/`internal`、`read`/`write` など）
- 未知のフィールドの検出
- 型の検証

## CI/CD 連携

### GitHub Actions ワークフロー

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

### 複数リポジトリの管理

このツールは**1回の実行で1つのリポジトリ**に設定を適用します。複数のリポジトリを同じ設定で管理するには、GitHub Actions の matrix 戦略を使用します：

```yaml
name: Sync Settings Across Repos

on:
  workflow_dispatch:
  push:
    branches: [main]
    paths:
      - ".github/repo-settings.yaml"

jobs:
  sync:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        repo:
          - myorg/service-a
          - myorg/service-b
          - myorg/service-c
      fail-fast: false

    steps:
      - uses: actions/checkout@v4

      - name: Install gh-repo-settings
        run: gh extension install myzkey/gh-repo-settings
        env:
          GH_TOKEN: ${{ secrets.ADMIN_TOKEN }}

      - name: Apply settings to ${{ matrix.repo }}
        run: gh repo-settings apply -y -r ${{ matrix.repo }}
        env:
          GH_TOKEN: ${{ secrets.ADMIN_TOKEN }}
```

> **注意**: 対象の全リポジトリへの管理者アクセス権を持つ PAT（`ADMIN_TOKEN`）が必要です。

## グローバルオプション

| オプション | 説明 |
|-----------|------|
| `-v, --verbose` | デバッグ出力を表示 |
| `-q, --quiet` | エラーのみ表示 |
| `-r, --repo <owner/name>` | 対象リポジトリ（デフォルト: 現在のリポジトリ） |

## FAQ

### 複数リポジトリに一括で設定を適用できますか？

直接はできません。このツールは**1回の実行で1つのリポジトリ**を管理します。

複数リポジトリに同じ設定を適用するには：
1. 各リポジトリに同じ YAML 設定を配置する、または
2. GitHub Actions の matrix 戦略を使用する（[複数リポジトリの管理](#複数リポジトリの管理)を参照）

### 複数環境（dev/staging/prod）に対応していますか？

ネイティブには対応していません。`env` ブロックは1リポジトリにつき1セットの変数/シークレットを管理します。

環境固有の値については：
- 異なる `.env` ファイル（`.env.dev`、`.env.staging`、`.env.prod`）を CI で切り替える
- GitHub Environments を使用する（このツールではまだ未対応）

### `plan` を実行せずに `apply` を実行するとどうなりますか？

ツールは計画された変更を表示し、適用前に確認を求めます。`-y` フラグで確認をスキップできますが、初回使用時は推奨しません。

### 組織レベルの設定を管理できますか？

いいえ。このツールはリポジトリレベルの設定のみを管理します。組織設定は異なる API 権限が必要であり、対象外です。

## 開発

```bash
# ビルド
make build

# テスト実行
make test

# リント（golangci-lint が必要）
make lint

# 全プラットフォーム向けビルド
make build-all

# ビルド成果物の削除
make clean
```

## ライセンス

MIT
