---
prompt_id: "PROMPT-HU-091-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-091"
related_hu: ["HU-022", "HU-023", "HU-024", "HU-090"]
issue: 245
issue_url: "https://github.com/bcaceres19/barberia/issues/245"
suggested_issue_title: "feat(public-booking): implementar HU-091 catálogo público"
branch: "feat/245-hu091-catalogo-publico"
pr: 246
pr_url: "https://github.com/bcaceres19/barberia/pull/246"
depends_on: ["HU-090 integrada", "HU-022 a HU-024 integradas", "Issue real con CA-091-01 a CA-091-05"]
rules: ["RN-SER-01", "RN-SER-02", "RN-SER-03", "RN-SER-04", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-002", "DEC-003", "DEC-004", "DEC-016", "DEC-019", "DEC-024", "DEC-067", "DEC-068", "DEC-069", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-091-01", "CA-091-02", "CA-091-03", "CA-091-04", "CA-091-05"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/catalog", "apps/web/src/modules/catalog"]
created_at: "2026-09-10"
updated_at: "2026-09-11"
supersedes: null
superseded_by: null
---

# Implementar HU-091: catálogo público de servicios

## Instrucción para el agente

Implementa únicamente la selección pública de servicios de `HU-091`; no reutilices DTO privados ni abras operaciones administrativas.

## Objetivo

Que un visitante vea y elija solo servicios activos, asignados y vigentes de la barbería pública.

## Preflight obligatorio

1. Confirma HU-090 integrada, issue real, rama corta y fuentes vigentes.
2. Ejecuta Graphify sobre catálogo, asignaciones, RLS y contrato público.
3. Comprueba que no hay duda/contradicción abierta aplicable.

## Alcance incluido

- Proyección pública mínima: nombre, descripción, duración, precio y COP.
- Filtro por servicio activo y al menos una asignación vigente del mismo tenant.
- Selección y estados carga/vacío/error/reintento accesibles.

## Fuera de alcance

- CRUD, inactivos, auditoría, disponibilidad, barbero o creación de cita.

## Estado existente que debe conservarse

- `catalog` gobierna servicios/asignaciones; snapshots históricos nunca se sustituyen por catálogo vigente.
- Cliente TypeScript generado y dirección de dependencias `app → modules → shared`.

## Trabajo requerido

1. Define operación pública contract-first y proyección separada de la privada.
2. Implementa caso de uso/consulta tenant-aware con orden estable y costo acotado.
3. Integra selector Vue sin estado global y revalida cambios de actividad/asignación.

## Pruebas y evidencia

- PostgreSQL real con dos tenants, activos/inactivos, 0/1/N asignaciones.
- Contrato, componente, E2E y evidencia 320/360/768/1280, teclado, zoom y axe-core.

## Documentación y trazabilidad

Actualiza OpenAPI, CHANGELOG, cliente, módulos, matriz, HU y catálogo de prompts.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-091-01`–`CA-091-05`.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-091 catálogo público`.
- `Closes #<issue>` solo con todos los criterios probados.
