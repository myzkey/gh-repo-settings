# gh-repo-settings

[English](./README.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md) | [Español](./README.es.md)

GitHub リポジトリの設定を YAML で管理する GitHub CLI 拡張機能。Terraform のワークフローにインスパイアされており、望む状態をコードで定義し、変更をプレビューしてから適用できます。

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

### 手動インストール

[Releases](https://github.com/myzkey/gh-repo-settings/releases) から最新のバイナリをダウンロードし、PATH に追加してください。

## クイックスタート

```bash
# 対話形式で設定ファイルを作成
gh rset init

# 変更をプレビュー（terraform plan のように）
gh rset plan

# 変更を適用
gh rset apply
```

デフォルトの設定ファイルパス（優先順）:
1. `.github/repo-settings/`（ディレクトリ）
2. `.github/repo-settings.yaml`（単一ファイル）

## コマンド

### `init` - 設定ファイルの初期化

対話形式で設定ファイルを作成します。

```bash
# .github/repo-settings.yaml を対話形式で作成
gh rset init

# 出力先を指定
gh rset init -o config.yaml

# 既存ファイルを上書き
gh rset init -f
```

### `export` - リポジトリ設定のエクスポート

現在の GitHub リポジトリ設定を YAML 形式でエクスポートします。

```bash
# 標準出力にエクスポート
gh rset export

# 単一ファイルにエクスポート
gh rset export -s .github/repo-settings.yaml

# ディレクトリにエクスポート（複数ファイル）
gh rset export -d .github/repo-settings/

# シークレット名を含める
gh rset export -s settings.yaml --include-secrets

# 特定のリポジトリからエクスポート
gh rset export -r owner/repo -s settings.yaml
```

### `plan` - 変更のプレビュー

設定を検証し、適用せずに計画された変更を表示します。

```bash
# すべての変更をプレビュー（デフォルトパスを使用）
gh rset plan

# 設定ファイルを指定
gh rset plan -c custom-config.yaml

# ディレクトリ設定でプレビュー
gh rset plan -d .github/repo-settings/

# スキーマ検証のみ（API 呼び出しなし）
gh rset plan --schema-only

# シークレットの存在のみチェック
gh rset plan --secrets

# 環境変数のみチェック
gh rset plan --env
```

### `apply` - 変更の適用

YAML 設定を GitHub リポジトリに適用します。

```bash
# 変更を適用（デフォルトパスを使用）
gh rset apply

# ドライラン（plan と同じ）
gh rset apply --dry-run

# 設定ファイルを指定
gh rset apply -c custom-config.yaml

# ディレクトリから適用
gh rset apply -d .github/repo-settings/
```

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

### ディレクトリ構造

設定を複数ファイルに分割することもできます:

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

### `secrets` - 必須シークレット

必要なリポジトリシークレットの存在をチェック（値は管理しません）:

```yaml
secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY
```

### `env` - 必須環境変数

必要なリポジトリ変数の存在をチェック:

```yaml
env:
  required:
    - DATABASE_URL
    - SENTRY_DSN
```

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
        run: gh rset plan
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## グローバルオプション

| オプション | 説明 |
|-----------|------|
| `-v, --verbose` | デバッグ出力を表示 |
| `-q, --quiet` | エラーのみ表示 |
| `-r, --repo <owner/name>` | 対象リポジトリ（デフォルト: 現在のリポジトリ） |

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
