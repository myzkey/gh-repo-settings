# gh-repo-settings

[English](../README.md) | [日本語](./README.ja.md) | [简体中文](./README.zh-CN.md) | [한국어](./README.ko.md)

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

### Actualizar

```bash
gh extension upgrade myzkey/gh-repo-settings
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

# Mostrar configuración actual de GitHub (para depuración)
gh repo-settings plan --show-current

# Verificar secrets
gh repo-settings plan --secrets

# Verificar variables de entorno
gh repo-settings plan --env

# Mostrar variables/secrets a eliminar (no en config)
gh repo-settings plan --env --secrets --sync
```

La opción `--show-current` muestra la configuración actual del repositorio de GitHub, útil para:
- Depurar problemas de configuración
- Encontrar configuraciones que existen en GitHub pero no en tu archivo de configuración
- Verificar qué está realmente configurado en el repositorio

**Validación de Status Checks**: Al ejecutar `plan`, la herramienta valida automáticamente que los nombres de `status_checks` en tus reglas de protección de rama coincidan con los nombres de jobs definidos en tus archivos `.github/workflows/`. Si se encuentra una discrepancia, verás una advertencia:

```
⚠ status check lint not found in workflows
⚠ status check test not found in workflows
  Available checks: build, golangci-lint, Run tests
```

### `apply` - Aplicar cambios

Aplica la configuración YAML al repositorio de GitHub.

```bash
# Aplicar cambios (usa ruta por defecto)
gh repo-settings apply

# Aprobar automáticamente sin confirmación
gh repo-settings apply -y

# Especificar archivo de configuración
gh repo-settings apply -c custom-config.yaml

# Aplicar desde directorio
gh repo-settings apply -d .github/repo-settings/

# Aplicar variables y secrets
gh repo-settings apply --env --secrets

# Modo sincronización: eliminar variables/secrets no en config
gh repo-settings apply --env --secrets --sync
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

### Estructura de Directorios

También puedes dividir la configuración en múltiples archivos:

```
.github/repo-settings/
├── repo.yaml
├── topics.yaml
├── labels.yaml
├── branch-protection.yaml
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

### `env` - Variables de Entorno y Secrets

Gestiona las variables y secrets del repositorio:

```yaml
env:
  # Variables con valores por defecto (pueden sobrescribirse con archivo .env)
  variables:
    NODE_ENV: production
    API_URL: https://api.example.com
  # Nombres de secrets (valores provienen del archivo .env o entrada interactiva)
  secrets:
    - API_TOKEN
    - DEPLOY_KEY
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `variables` | map | Pares clave-valor para variables del repositorio |
| `secrets` | array | Lista de nombres de secrets a gestionar |

#### Usando archivo `.env`

Crea un archivo `.github/.env` (agregar a gitignore) para almacenar valores reales:

```bash
# .github/.env
NODE_ENV=staging
API_URL=https://staging-api.example.com
API_TOKEN=your-secret-token
DEPLOY_KEY=your-deploy-key
```

**Prioridad**: Los valores del archivo `.env` sobrescriben los valores por defecto del YAML.

#### Comandos

```bash
# Previsualizar cambios de variables/secrets
gh repo-settings plan --env --secrets

# Aplicar variables y secrets
gh repo-settings apply --env --secrets

# Eliminar variables/secrets no en config (modo sincronización)
gh repo-settings apply --env --secrets --sync
```

Si el valor de un secret no se encuentra en `.env`, se solicitará entrada interactiva durante `apply`.

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

### `pages` - Configuración de GitHub Pages

Configura GitHub Pages para el repositorio:

```yaml
pages:
  # Tipo de build: "workflow" (GitHub Actions) o "legacy" (basado en rama)
  build_type: workflow

  # Configuración de origen (solo para tipo legacy)
  source:
    branch: main
    path: /docs  # "/" o "/docs"
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `build_type` | `workflow` \| `legacy` | Cómo se construye Pages |
| `source.branch` | string | Rama para builds legacy |
| `source.path` | `/` \| `/docs` | Ruta dentro de la rama |

## Integración con Editor (VSCode)

Este proyecto proporciona un JSON Schema para validación y autocompletado de YAML en VSCode.

### Configuración

1. Instala la [extensión YAML](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

2. Agrega a tu `.vscode/settings.json`:

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

### Características

- Autocompletado para todos los campos
- Documentación al pasar el cursor
- Sugerencias de enum (`public`/`private`/`internal`, `read`/`write`, etc.)
- Detección de campos desconocidos
- Validación de tipos

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
