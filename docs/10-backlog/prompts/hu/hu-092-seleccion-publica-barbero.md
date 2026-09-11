---
prompt_id: "PROMPT-HU-092-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-092"
related_hu: ["HU-021", "HU-023", "HU-091"]
issue: 247
issue_url: "https://github.com/bcaceres19/barberia/issues/247"
suggested_issue_title: "feat(public-booking): implementar HU-092 selección de barbero"
branch: "feat/247-hu092-seleccion-barbero"
pr: null
pr_url: null
depends_on: ["HU-091 integrada (satisfecho, PR #246)", "Issue real con CA-092-01 a CA-092-05 (satisfecho, issue #247)"]
rules: ["RN-TEN-01", "RN-CON-01", "RN-DAT-02"]
decisions: ["DEC-019", "DEC-024", "DEC-047", "DEC-068", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-092-01", "CA-092-02", "CA-092-03", "CA-092-04", "CA-092-05"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/staff", "apps/api/internal/modules/catalog"]
created_at: "2026-09-10"
updated_at: "2026-09-11"
supersedes: null
superseded_by: null
---

# Implementar HU-092: selección pública de barbero

## Instrucción para el agente

Implementa solo la selección descrita; no agregues “cualquiera”, ranking, perfiles ni asignación automática.

## Objetivo

Resolver exactamente un barbero elegible para el servicio: preselección con uno y elección explícita con varios.

## Preflight obligatorio

1. Verifica HU-091 integrada, issue/rama reales y fuentes completas.
2. Ejecuta Graphify sobre `staff`, `catalog` y el flujo público.
3. Detente si una fuente requiere una opción no decidida.

## Alcance incluido

- Lectura pública de barberos del tenant asignados al servicio activo.
- Matriz 0/1/N y limpieza/revalidación al cambiar servicio.
- Selector accesible y tipado.

## Fuera de alcance

- Disponibilidad, cita, baja de barbero, foto, biografía, preferencia o “primero libre”.

## Estado existente que debe conservarse

- Ciclo de `barber` recortado por `DEC-047` y asignación por `HU-023`.
- Ningún vínculo automático `staff_user`→`barber` se inventa.

## Trabajo requerido

1. Define contrato público sin exponer tenant ni campos privados.
2. Consulta relación vigente con RLS y orden estable.
3. Implementa estados 0/1/N y evita respuestas tardías sobre un servicio nuevo.

## Pruebas y evidencia

- Unidad/contrato, PostgreSQL real con dos tenants y asignaciones cruzadas.
- Componente/E2E, cuatro anchos, teclado, zoom 200 % y axe-core.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, README, matriz, historia y prompt/índice.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-092-01`–`CA-092-05`.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-092 selección de barbero`.
