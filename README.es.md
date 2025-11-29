# gh-repo-settings

[English](./README.md) | [日本語](./README.ja.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md)

Una extensión de GitHub CLI para gestionar la configuración de repositorios mediante YAML. Inspirado en el flujo de trabajo de Terraform: define el estado deseado en código, previsualiza los cambios y aplícalos.

## Características

- **Infraestructura como Código**: Define la configuración del repositorio en archivos YAML
- **Flujo de trabajo estilo Terraform**: `plan` para previsualizar, `apply` para ejecutar
- **Exportar configuración existente**: Genera YAML desde la configuración actual del repositorio
- **Validación de esquema**: Valida la configuración antes de aplicar
- **Múltiples formatos de configuración**: Archivo único o configuración basada en directorios
- **Verificación de Secrets/Env**: Verifica que existan los secrets y variables de entorno requeridos
- **Permisos de Actions**: Configura permisos de GitHub Actions y ajustes de flujos de trabajo

## Instalación

### Via GitHub CLI (Recomendado)

```bash
gh extension install myzkey/gh-repo-settings
```

### Instalación Manual

Descarga el binario más reciente desde [Releases](https://github.com/myzkey/gh-repo-settings/releases) y agrégalo a tu PATH.

## Inicio Rápido

```bash
# Crear archivo de configuración interactivamente
gh repo-settings init

# Previsualizar cambios (como terraform plan)
gh repo-settings plan

# Aplicar cambios
gh repo-settings apply
```

Rutas de configuración por defecto (en orden de prioridad):
1. `.github/repo-settings/` (directorio)
2. `.github/repo-settings.yaml` (archivo único)

## Comandos

### `init` - Inicializar configuración

Crea un archivo de configuración interactivamente.

```bash
# Crear .github/repo-settings.yaml interactivamente
gh repo-settings init

# Especificar ruta de salida
gh repo-settings init -o config.yaml

# Sobrescribir archivo existente
gh repo-settings init -f
```

### `export` - Exportar configuración del repositorio

Exporta la configuración actual del repositorio de GitHub en formato YAML.

```bash
# Exportar a stdout
gh repo-settings export

# Exportar a archivo único
gh repo-settings export -s .github/repo-settings.yaml

# Exportar a directorio (múltiples archivos)
gh repo-settings export -d .github/repo-settings/

# Incluir nombres de secrets
gh repo-settings export -s settings.yaml --include-secrets

# Exportar desde repositorio específico
gh repo-settings export -r owner/repo -s settings.yaml
```

### `plan` - Previsualizar cambios

Valida la configuración y muestra los cambios planificados sin aplicarlos.

```bash
# Previsualizar todos los cambios (usa ruta por defecto)
gh repo-settings plan

# Especificar archivo de configuración
gh repo-settings plan -c custom-config.yaml

# Previsualizar con configuración de directorio
gh repo-settings plan -d .github/repo-settings/

# Solo validación de esquema (sin llamadas API)
gh repo-settings plan --schema-only

# Verificar solo existencia de secrets
gh repo-settings plan --secrets

# Verificar solo variables de entorno
gh repo-settings plan --env
```

### `apply` - Aplicar cambios

Aplica la configuración YAML al repositorio de GitHub.

```bash
# Aplicar cambios (usa ruta por defecto)
gh repo-settings apply

# Ejecución en seco (igual que plan)
gh repo-settings apply --dry-run

# Especificar archivo de configuración
gh repo-settings apply -c custom-config.yaml

# Aplicar desde directorio
gh repo-settings apply -d .github/repo-settings/
```

## Configuración

### Archivo Único

Crear `.github/repo-settings.yaml`:

```yaml
repo:
  description: "Mi proyecto increíble"
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
      description: Algo no funciona
    - name: feature
      color: 0e8a16
      description: Solicitud de nueva funcionalidad

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

### Estructura de Directorios

También puedes dividir la configuración en múltiples archivos:

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

## Referencia de Configuración

### `repo` - Configuración del Repositorio

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `description` | string | Descripción del repositorio |
| `homepage` | string | URL de la página principal |
| `visibility` | `public` \| `private` \| `internal` | Visibilidad del repositorio |
| `allow_merge_commit` | boolean | Permitir merge commits |
| `allow_rebase_merge` | boolean | Permitir rebase merge |
| `allow_squash_merge` | boolean | Permitir squash merge |
| `delete_branch_on_merge` | boolean | Eliminar rama automáticamente después del merge |
| `allow_update_branch` | boolean | Permitir actualizar rama del PR |

### `topics` - Temas del Repositorio

Array de strings de temas:

```yaml
topics:
  - javascript
  - nodejs
  - cli
```

### `labels` - Etiquetas de Issues

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `replace_default` | boolean | Eliminar etiquetas que no están en la configuración |
| `items` | array | Lista de definiciones de etiquetas |
| `items[].name` | string | Nombre de la etiqueta |
| `items[].color` | string | Color hexadecimal (sin `#`) |
| `items[].description` | string | Descripción de la etiqueta |

### `branch_protection` - Reglas de Protección de Rama

```yaml
branch_protection:
  <nombre_rama>:
    # Revisiones de pull request
    required_reviews: 1          # Número de aprobaciones requeridas
    dismiss_stale_reviews: true  # Descartar aprobaciones en nuevos commits
    require_code_owner: false    # Requerir revisión de CODEOWNERS

    # Checks de estado
    require_status_checks: true  # Requerir checks de estado
    status_checks:               # Nombres de checks requeridos
      - ci/test
    strict_status_checks: false  # Requerir rama actualizada

    # Despliegues
    required_deployments:        # Entornos de despliegue requeridos
      - production

    # Requisitos de commits
    require_signed_commits: false # Requerir commits firmados
    require_linear_history: false # Prohibir merge commits

    # Restricciones de push/merge
    enforce_admins: false        # Incluir administradores
    restrict_creations: false    # Restringir creación de ramas
    restrict_pushes: false       # Restringir quién puede hacer push
    allow_force_pushes: false    # Permitir force push
    allow_deletions: false       # Permitir eliminación de rama
```

### `secrets` - Secrets Requeridos

Verifica que existan los secrets requeridos del repositorio (no gestiona valores):

```yaml
secrets:
  required:
    - API_TOKEN
    - DEPLOY_KEY
```

### `env` - Variables de Entorno Requeridas

Verifica que existan las variables requeridas del repositorio:

```yaml
env:
  required:
    - DATABASE_URL
    - SENTRY_DSN
```

### `actions` - Permisos de GitHub Actions

Configura los permisos de GitHub Actions para el repositorio:

```yaml
actions:
  # Habilitar/deshabilitar GitHub Actions
  enabled: true

  # Qué actions están permitidas: "all", "local_only", "selected"
  allowed_actions: selected

  # Cuando allowed_actions es "selected"
  selected_actions:
    github_owned_allowed: true    # Permitir actions de GitHub
    verified_allowed: true        # Permitir actions de creadores verificados
    patterns_allowed:             # Patrones de actions permitidas
      - "actions/*"
      - "github/codeql-action/*"

  # Permisos predeterminados de GITHUB_TOKEN: "read" o "write"
  default_workflow_permissions: read

  # Permitir que GitHub Actions cree/apruebe pull requests
  can_approve_pull_request_reviews: false
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `enabled` | boolean | Habilitar GitHub Actions |
| `allowed_actions` | `all` \| `local_only` \| `selected` | Actions permitidas |
| `selected_actions.github_owned_allowed` | boolean | Permitir actions de GitHub |
| `selected_actions.verified_allowed` | boolean | Permitir creadores verificados |
| `selected_actions.patterns_allowed` | array | Patrones de actions permitidas |
| `default_workflow_permissions` | `read` \| `write` | Permisos predeterminados de GITHUB_TOKEN |
| `can_approve_pull_request_reviews` | boolean | Permitir que Actions apruebe PRs |

## Integración CI/CD

### Flujo de Trabajo de GitHub Actions

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

## Opciones Globales

| Opción | Descripción |
|--------|-------------|
| `-v, --verbose` | Mostrar salida de depuración |
| `-q, --quiet` | Mostrar solo errores |
| `-r, --repo <owner/name>` | Repositorio destino (predeterminado: actual) |

## Desarrollo

```bash
# Compilar
make build

# Ejecutar tests
make test

# Lint (requiere golangci-lint)
make lint

# Compilar para todas las plataformas
make build-all

# Limpiar artefactos de compilación
make clean
```

## Licencia

MIT
