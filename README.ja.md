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

## インストール

```bash
gh extension install myzkey/gh-repo-settings
```

## クイックスタート

```bash
# 現在のリポジトリ設定を YAML にエクスポート
gh repo-settings export -s repo-settings.yaml

# 変更をプレビュー（terraform plan のように）
gh repo-settings plan -c repo-settings.yaml

# 変更を適用
gh repo-settings apply -c repo-settings.yaml
```

## コマンド

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

```bash
# すべての変更をプレビュー
gh repo-settings plan -c .github/repo-settings.yaml

# ディレクトリ設定でプレビュー
gh repo-settings plan -d .github/repo-settings/

# スキーマ検証のみ（API 呼び出しなし）
gh repo-settings plan --schema-only

# シークレットの存在のみチェック
gh repo-settings plan --secrets

# 環境変数のみチェック
gh repo-settings plan --env
```

### `apply` - 変更の適用

YAML 設定を GitHub リポジトリに適用します。

```bash
# 変更を適用
gh repo-settings apply -c .github/repo-settings.yaml

# ドライラン（plan と同じ）
gh repo-settings apply --dry-run

# ディレクトリから適用
gh repo-settings apply -d .github/repo-settings/
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
└── env.yaml
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
    required_reviews: 1          # 必要な承認数
    dismiss_stale_reviews: true  # 新しいコミットで承認を却下
    require_code_owner: false    # CODEOWNERS のレビューを必須
    require_status_checks: true  # ステータスチェックを必須
    status_checks:               # 必須のステータスチェック名
      - ci/test
    strict_status_checks: false  # 最新ブランチを必須
    enforce_admins: false        # 管理者にも適用
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

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install gh-repo-settings
        run: gh extension install myzkey/gh-repo-settings
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: スキーマ検証
        run: gh repo-settings plan --schema-only
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: 設定差分チェック
        run: gh repo-settings plan
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## グローバルオプション

| オプション | 説明 |
|-----------|------|
| `-v, --verbose` | デバッグ出力を表示 |
| `-q, --quiet` | エラーのみ表示 |
| `-r, --repo <owner/name>` | 対象リポジトリ（デフォルト: 現在のリポジトリ） |

## 開発

```bash
# 依存関係のインストール
pnpm install

# ビルド
pnpm build

# テスト実行
pnpm test

# リント
pnpm lint

# 型チェック
pnpm typecheck
```

## ライセンス

MIT
