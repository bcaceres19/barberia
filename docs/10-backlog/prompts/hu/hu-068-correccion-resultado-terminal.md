---
prompt_id: "PROMPT-HU-068-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-068"
related_hu: ["HU-004", "HU-060", "HU-064", "HU-066", "HU-067"]
issue: 227
issue_url: "https://github.com/bcaceres19/barberia/issues/227"
suggested_issue_title: "feat(booking): implementar HU-068 corrección auditada de resultado"
branch: "feat/227-hu068-correccion-resultado"
pr: 239
pr_url: "https://github.com/bcaceres19/barberia/pull/239"
depends_on:
  - "HU-066 y HU-067 integradas en main con CI verde"
  - "Issue #227 con CA-068-01 a CA-068-08"
rules: ["RN-CIT-01", "RN-CIT-03", "RN-CIT-04", "RN-HIS-01", "RN-HIS-02", "RN-CON-01", "RN-CON-03", "RN-TEN-01", "RN-IDE-01"]
decisions: ["DEC-011", "DEC-012", "DEC-014", "DEC-016", "DEC-017", "DEC-024", "DEC-035", "DEC-036", "DEC-037", "DEC-038", "DEC-041", "DEC-043", "DEC-077", "DEC-078", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-068-01", "CA-068-02", "CA-068-03", "CA-068-04", "CA-068-05", "CA-068-06", "CA-068-07", "CA-068-08"]
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/detalle-turno-eventos/README.md"
  - "api/openapi/paths/private-appointments.yaml"
  - "apps/api/internal/modules/booking"
  - "apps/web/src/modules/agenda"
created_at: "2026-09-08"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Implementar HU-068: corrección auditada de un resultado terminal

## Instrucción para el agente

Implementa únicamente T8. HU-066 (PR #235) y HU-067 (PR #237) ya están integradas en `main` con CI verde (verificado 2026-09-10), así que el prompt pasa a `ready`. No conviertas la corrección en edición/borrado del historial ni en reapertura a `confirmed`.

## Objetivo

Que el barbero corrija una clasificación terminal errónea hacia otra terminal con motivo obligatorio, rastro completo y protección contra la reocupación conflictiva de una franja.

## Preflight obligatorio

1. Detente si HU-066/HU-067 no están integradas en `main` con CI verde.
2. Comprueba árbol limpio, actualiza `main`, lee todos los `source_docs` y revisa dudas/contradicciones.
3. Confirma issue #227, actualiza el prompt y crea `feat/227-hu068-correccion-resultado`.
4. Ejecuta Graphify sobre máquina de estados, exclusión, historial, idempotencia y detalle Vue.

## Alcance incluido

- Comando T8 contract-first con destino terminal, motivo, idempotencia y versión.
- Corrección entre cuatro terminales; `confirmed` nunca es origen ni destino.
- Estado + `appointment_status_corrected` + cambios/motivo en una transacción.
- Reaplicación de tiempo/exclusión al volver a un estado que ocupa agenda.
- Liberación atómica al corregir hacia cancelado.
- UI, historial completo, pruebas concurrentes y evidencia accesible.

## Fuera de alcance

- Reabrir/recrear automáticamente, T3, cierre automático o cambio de otros datos.
- Editar/borrar/ocultar eventos históricos.
- Efectos de notificación o recordatorios.

## Estado existente que debe conservarse

- `appointment_history` y sus cambios son append-only por permisos y pruebas.
- La exclusión cubre `confirmed`, `completed` y `no_show`.
- HU-066/HU-067 aportan los terminales y el patrón de comando condicionado.
- El detalle pagina eventos en orden estable y no expone IDs internos.

## Trabajo requerido

1. Define un request cerrado: destino terminal permitido y motivo normalizado/no vacío; deriva actor/tenant.
2. Valida origen/destino, frontera temporal para `completed`/`no_show` y versión.
3. Bloquea la cita y aplica estado/evento/cambios en una transacción corta.
4. De cancelado a estado ocupante, deja que PostgreSQL sea la última defensa y traduce `23P01` a `409`.
5. De estado ocupante a cancelado, confirma que la exclusión se libera al commit.
6. Conserva repetición exacta sin duplicados y resuelve carreras sin `sleep`.
7. En Vue exige destino/motivo, resume consecuencias y recarga detalle/historial tras éxito.

## Pruebas y evidencia

- Matriz de 12 cambios terminales, valores inválidos, motivo, tiempo, repetición y versión.
- HTTP/contrato con todos los códigos y body cerrado.
- PostgreSQL real con dos tenants: append-only, rollback, liberar/reocupar y carrera con barreras.
- Componente y E2E `completed` → `no_show`, historial completo y rechazo de reapertura.
- 320, 360, 768, 1280 px, teclado, foco, zoom 200 % y axe-core.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG/cliente, README, matriz, historial, plan y prompt.
- Registra una duda y bloquea antes de elegir una semántica no definida.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && pnpm run db:atlas:validate && git diff --check && graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-068-01`–`CA-068-08`.

## Git y PR

- Rama: `feat/227-hu068-correccion-resultado`.
- Commit/PR: `feat(booking): implementa HU-068 corrección auditada`.
- `Closes #227` solo si se cubre todo; si no, `Refs #227`.
