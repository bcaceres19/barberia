---
prompt_id: "PROMPT-FIX-146-DETALLE-TURNO-TOKENS-v1"
version: "1.0"
kind: "fix"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-064"
  - "HU-065"
issue: "146"
issue_url: "https://github.com/bcaceres19/barberia/issues/146"
suggested_issue_title: "fix(web): corrige tokens de tipografía rotos en Detalle de turno"
branch: "fix/146-detalle-turno-tokens"
pr: 147
pr_url: "https://github.com/bcaceres19/barberia/pull/147"
depends_on:
  - "Fase 4b (issue #144, PR #145) integrada en main"
rules: []
decisions:
  - "DEC-077"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 4c · Corrige tokens de tipografía rotos en Detalle de turno

## Instrucción para el agente

Corrige únicamente el defecto de tokens de tipografía inexistentes en `AppointmentDetailPage.vue`, encontrado durante la auditoría de la Fase 4 (mismo defecto ya corregido en `DailyAgendaPage.vue`, Fase 4a, issue #141).

## Objetivo

`.appointment-detail-page__title` y `.appointment-detail-page__history-title` usan tokens de tipografía reales (`--font-size-h1`/`--font-size-h2`) en vez de `var(--font-size-heading-lg)`/`var(--font-size-heading-md)`, que nunca existieron en `tokens.css`.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar Fase 4b (PR #145).
2. Auditoría de `AppointmentDetailPage.vue` durante la planificación de la Fase 4c (detalle/historial/reprogramación): el diálogo de reprogramación y la sección de historial ya usaban tokens reales; el único hallazgo fueron los dos tokens de encabezado rotos.
3. Issue real creado (`#146`) antes de escribir código; rama `fix/146-detalle-turno-tokens` desde `main` actualizada.

## Alcance incluido

- `.appointment-detail-page__title` → `--font-size-h1`/`--font-size-h1-line`/`--font-weight-h1`.
- `.appointment-detail-page__history-title` → `--font-size-h2`/`--font-size-h2-line`/`--font-weight-h2`.

## Fuera de alcance

- Cualquier cambio de contrato, lógica de reprogramación o historial.
- El mismo defecto en `StaffPage`/`BarberServicesPage`/`SettingsPage`/`SchedulesPage`/`CatalogPage`/`BlocksPage`: corregido por separado en el issue #149.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 604 pruebas en verde (sin cambio: fix de solo estilo, 20 pruebas preexistentes de `AppointmentDetailPage.test.ts` sin cambio de aserciones).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice.
- `docs/00-control/historial-cambios.md`: entrada de esta entrega con confirmación real de CI/merge.

## Verificación final

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
git diff --check
```

| Criterio | Estado | Prueba o evidencia |
| --- | --- | --- |
| Título del turno con jerarquía real | Cumplido | Revisión de `AppointmentDetailPage.vue` §style |
| Título "Historial" con jerarquía real | Cumplido | Revisión de `AppointmentDetailPage.vue` §style |
| Sin regresión en HU-064/HU-065 | Cumplido | 20 pruebas preexistentes sin cambio de aserciones |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 147` (4/4 verde tras un rerun del job Go, ver nota); `gh pr view 147 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T04:18:29Z`, commit `473b163` |

## Nota de CI

El primer intento de CI de este PR falló en `go test -race` (`TestWorkingHours_HTTP_CrossTenantBarber_Returns404`: esperaba `404` y recibió `401`, una carrera de expiración de sesión en la suite de integración) en un PR que solo toca `AppointmentDetailPage.vue` — un fallo pre-existente y no relacionado con este cambio. Se confirmó investigando el log (`gh run view --job --log-failed`) antes de decidir, y se resolvió con `gh run rerun --failed`, que pasó 4/4 en verde.

## Git y PR

- Commit: `fix(web): corrige tokens de tipografia rotos en Detalle de turno`.
- PR: [#147](https://github.com/bcaceres19/barberia/pull/147), `Closes #146`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
