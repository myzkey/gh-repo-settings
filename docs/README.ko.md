# gh-repo-settings

[English](../README.md) | [日本語](./README.ja.md) | [简体中文](./README.zh-CN.md) | [Español](./README.es.md)

YAML 설정을 통해 GitHub 저장소 설정을 관리하는 GitHub CLI 확장 프로그램입니다. Terraform의 워크플로우에서 영감을 받아 원하는 상태를 코드로 정의하고, 변경 사항을 미리 보고, 적용할 수 있습니다.

## 특징

- **Infrastructure as Code**: YAML 파일로 저장소 설정 정의
- **Terraform 스타일 워크플로우**: `plan`으로 미리보기, `apply`로 실행
- **기존 설정 내보내기**: 현재 저장소 설정에서 YAML 생성
- **스키마 검증**: 적용 전 설정 검증
- **다양한 설정 형식**: 단일 파일 또는 디렉토리 기반 설정
- **Secrets/Env 확인**: 필수 시크릿 및 환경 변수 존재 확인
- **Actions 권한 설정**: GitHub Actions 권한 및 워크플로우 설정 관리

## 설치

### GitHub CLI를 통해 (권장)

```bash
gh extension install myzkey/gh-repo-settings
```

### 업그레이드

```bash
gh extension upgrade myzkey/gh-repo-settings
```

### 수동 설치

[Releases](https://github.com/myzkey/gh-repo-settings/releases)에서 최신 바이너리를 다운로드하고 PATH에 추가하세요.

## 빠른 시작

```bash
# 대화형으로 설정 파일 생성
gh repo-settings init

# 변경 사항 미리보기 (terraform plan처럼)
gh repo-settings plan

# 변경 사항 적용
gh repo-settings apply
```

기본 설정 파일 경로 (우선순위 순):
1. `.github/repo-settings/` (디렉토리)
2. `.github/repo-settings.yaml` (단일 파일)

## 명령어

### `init` - 설정 파일 초기화

대화형으로 설정 파일을 생성합니다.

```bash
# .github/repo-settings.yaml을 대화형으로 생성
gh repo-settings init

# 출력 경로 지정
gh repo-settings init -o config.yaml

# 기존 파일 덮어쓰기
gh repo-settings init -f
```

### `export` - 저장소 설정 내보내기

현재 GitHub 저장소 설정을 YAML 형식으로 내보냅니다.

```bash
# 표준 출력으로 내보내기
gh repo-settings export

# 단일 파일로 내보내기
gh repo-settings export -s .github/repo-settings.yaml

# 디렉토리로 내보내기 (여러 파일)
gh repo-settings export -d .github/repo-settings/

# 시크릿 이름 포함
gh repo-settings export -s settings.yaml --include-secrets

# 특정 저장소에서 내보내기
gh repo-settings export -r owner/repo -s settings.yaml
```

### `plan` - 변경 사항 미리보기

설정을 검증하고 적용하지 않고 계획된 변경 사항을 표시합니다.

```bash
# 모든 변경 사항 미리보기 (기본 경로 사용)
gh repo-settings plan

# 설정 파일 지정
gh repo-settings plan -c custom-config.yaml

# 디렉토리 설정으로 미리보기
gh repo-settings plan -d .github/repo-settings/

# 현재 GitHub 설정 표시 (디버깅용)
gh repo-settings plan --show-current

# 시크릿 존재 여부만 확인
gh repo-settings plan --secrets

# 환경 변수만 확인
gh repo-settings plan --env
```

`--show-current` 옵션은 현재 GitHub 저장소 설정을 표시합니다. 다음 경우에 유용합니다:
- 설정 문제 디버깅
- GitHub에 존재하지만 설정 파일에 없는 설정 찾기
- 저장소의 실제 설정 확인

**Status Check 검증**: `plan` 실행 시 브랜치 보호 규칙의 `status_checks` 이름이 `.github/workflows/` 파일의 작업 이름과 일치하는지 자동으로 검증합니다. 불일치가 발견되면 경고가 표시됩니다:

```
⚠ status check lint not found in workflows
⚠ status check test not found in workflows
  Available checks: build, golangci-lint, Run tests
```

### `apply` - 변경 사항 적용

YAML 설정을 GitHub 저장소에 적용합니다.

```bash
# 변경 사항 적용 (기본 경로 사용)
gh repo-settings apply

# 드라이 런 (plan과 동일)
gh repo-settings apply --dry-run

# 설정 파일 지정
gh repo-settings apply -c custom-config.yaml

# 디렉토리에서 적용
gh repo-settings apply -d .github/repo-settings/
```

## 설정

### 단일 파일

`.github/repo-settings.yaml` 생성:

```yaml
repo:
  description: "멋진 프로젝트"
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
      description: 버그 리포트
    - name: feature
      color: 0e8a16
      description: 새 기능 요청

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

### 디렉토리 구조

설정을 여러 파일로 분리할 수도 있습니다:

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

## 설정 레퍼런스

### `repo` - 저장소 설정

| 필드 | 타입 | 설명 |
|-----|------|------|
| `description` | string | 저장소 설명 |
| `homepage` | string | 홈페이지 URL |
| `visibility` | `public` \| `private` \| `internal` | 저장소 공개 범위 |
| `allow_merge_commit` | boolean | 머지 커밋 허용 |
| `allow_rebase_merge` | boolean | 리베이스 머지 허용 |
| `allow_squash_merge` | boolean | 스쿼시 머지 허용 |
| `delete_branch_on_merge` | boolean | 머지 후 브랜치 자동 삭제 |
| `allow_update_branch` | boolean | PR 브랜치 업데이트 허용 |

### `topics` - 저장소 토픽

토픽 문자열 배열:

```yaml
topics:
  - javascript
  - nodejs
  - cli
```

### `labels` - 이슈 라벨

| 필드 | 타입 | 설명 |
|-----|------|------|
| `replace_default` | boolean | 설정에 없는 라벨 삭제 |
| `items` | array | 라벨 정의 목록 |
| `items[].name` | string | 라벨 이름 |
| `items[].color` | string | 16진수 색상 (`#` 제외) |
| `items[].description` | string | 라벨 설명 |

### `branch_protection` - 브랜치 보호 규칙

```yaml
branch_protection:
  <브랜치명>:
    # 풀 리퀘스트 리뷰
    required_reviews: 1          # 필요한 승인 수
    dismiss_stale_reviews: true  # 새 커밋 시 승인 취소
    require_code_owner: false    # CODEOWNERS 리뷰 필수

    # 상태 체크
    require_status_checks: true  # 상태 체크 필수
    status_checks:               # 필수 상태 체크 이름
      - ci/test
    strict_status_checks: false  # 최신 브랜치 필수

    # 배포
    required_deployments:        # 필수 배포 환경
      - production

    # 커밋 요구사항
    require_signed_commits: false # 서명된 커밋 필수
    require_linear_history: false # 머지 커밋 금지

    # 푸시/머지 제한
    enforce_admins: false        # 관리자에게도 적용
    restrict_creations: false    # 브랜치 생성 제한
    restrict_pushes: false       # 푸시 제한
    allow_force_pushes: false    # 강제 푸시 허용
    allow_deletions: false       # 브랜치 삭제 허용
```

### `secrets` - 필수 시크릿

필수 저장소 시크릿 존재 확인 (값은 관리하지 않음):

```yaml
secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY
```

### `env` - 필수 환경 변수

필수 저장소 변수 존재 확인:

```yaml
env:
  required:
    - DATABASE_URL
    - SENTRY_DSN
```

### `actions` - GitHub Actions 권한 설정

저장소의 GitHub Actions 권한을 설정합니다:

```yaml
actions:
  # GitHub Actions 활성화/비활성화
  enabled: true

  # 허용되는 actions: "all", "local_only", "selected"
  allowed_actions: selected

  # allowed_actions가 "selected"인 경우
  selected_actions:
    github_owned_allowed: true    # GitHub 공식 actions 허용
    verified_allowed: true        # 인증된 제작자의 actions 허용
    patterns_allowed:             # 허용되는 action 패턴
      - "actions/*"
      - "github/codeql-action/*"

  # 기본 GITHUB_TOKEN 권한: "read" 또는 "write"
  default_workflow_permissions: read

  # GitHub Actions의 PR 생성/승인 허용
  can_approve_pull_request_reviews: false
```

| 필드 | 타입 | 설명 |
|-----|------|------|
| `enabled` | boolean | GitHub Actions 활성화 |
| `allowed_actions` | `all` \| `local_only` \| `selected` | 허용되는 actions |
| `selected_actions.github_owned_allowed` | boolean | GitHub 공식 actions 허용 |
| `selected_actions.verified_allowed` | boolean | 인증된 제작자 허용 |
| `selected_actions.patterns_allowed` | array | 허용되는 action 패턴 |
| `default_workflow_permissions` | `read` \| `write` | GITHUB_TOKEN 기본 권한 |
| `can_approve_pull_request_reviews` | boolean | Actions의 PR 승인 허용 |

### `pages` - GitHub Pages 설정

리포지토리의 GitHub Pages 설정:

```yaml
pages:
  # 빌드 타입: "workflow" (GitHub Actions) 또는 "legacy" (브랜치 기반)
  build_type: workflow

  # 소스 설정 (legacy 빌드 타입 전용)
  source:
    branch: main
    path: /docs  # "/" 또는 "/docs"
```

| 필드 | 타입 | 설명 |
|-----|------|------|
| `build_type` | `workflow` \| `legacy` | Pages 빌드 방식 |
| `source.branch` | string | legacy 빌드 브랜치 |
| `source.path` | `/` \| `/docs` | 브랜치 내 경로 |

## 에디터 연동 (VSCode)

VSCode에서 YAML 검증 및 자동 완성을 위한 JSON Schema를 제공합니다.

### 설정

1. [YAML 확장 프로그램](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) 설치

2. `.vscode/settings.json`에 추가:

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

### 기능

- 모든 필드 자동 완성
- 호버 시 문서 표시
- enum 값 제안 (`public`/`private`/`internal`, `read`/`write` 등)
- 알 수 없는 필드 감지
- 타입 검증

## CI/CD 통합

### GitHub Actions 워크플로우

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

## 전역 옵션

| 옵션 | 설명 |
|-----|------|
| `-v, --verbose` | 디버그 출력 표시 |
| `-q, --quiet` | 오류만 표시 |
| `-r, --repo <owner/name>` | 대상 저장소 (기본: 현재 저장소) |

## 개발

```bash
# 빌드
make build

# 테스트 실행
make test

# 린트 (golangci-lint 필요)
make lint

# 전체 플랫폼 빌드
make build-all

# 빌드 아티팩트 정리
make clean
```

## 라이선스

MIT
