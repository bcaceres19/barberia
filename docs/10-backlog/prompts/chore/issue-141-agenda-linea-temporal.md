---
prompt_id: "PROMPT-CHORE-141-AGENDA-LINEA-TEMPORAL-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-062"
  - "HU-063"
issue: "141"
issue_url: "https://github.com/bcaceres19/barberia/issues/141"
suggested_issue_title: "feat(web): agenda diaria NAVA — línea temporal de escritorio y ficha de turno"
branch: "feat/141-agenda-linea-temporal"
pr: 142
pr_url: "https://github.com/bcaceres19/barberia/pull/142"
depends_on:
  - "Fase 3 (issue #138, PR #139) integrada en main"
rules:
  - "RN-DIS-05"
  - "RN-DIS-07"
decisions:
  - "DEC-074"
  - "DEC-075"
  - "DEC-077"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/prompts/orchestration/adopcion-frontend-nava.md"
created_at: "2026-09-01"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 4a · Agenda diaria — línea temporal de escritorio y ficha de turno

## Instrucción para el agente

Implementa únicamente la Fase 4a (primera entrega de "Operación de agenda") del prompt de orquestación `PROMPT-ORCH-NAVA-FRONTEND-v1`: la línea temporal horizontal opcional de escritorio y el ajuste de jerarquía visual de la ficha de turno en `DailyAgendaPage.vue`, sin tocar ninguna acción sobre un turno existente.

## Objetivo

`DailyAgendaPage.vue` ofrece, desde 1024px, una representación temporal horizontal del día del barbero seleccionado (eje con marcas de hora, fichas posicionadas por tiempo real, marcador "Ahora" solo si se ve hoy), sin sustituir la lista cronológica accesible. La ficha de la lista sube su jerarquía visual (persona atendida como texto principal, hora en cifras tabulares, 64px mínimo).

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar Fase 3 (PR #139).
2. Auditoría de `DailyAgendaPage.vue`, `shared/time/civilDate.ts` y `shared/time/formatInstant.ts`: sin utilidad existente para "minutos dentro de un día civil dado un instante arbitrario" (`getCivilDateInTimezone` solo resuelve el día civil de "ahora").
3. Issue real creado (`#141`) antes de escribir código; rama `feat/141-agenda-linea-temporal` desde `main` actualizada.

## Corrección de proceso durante esta entrega

El primer commit de este trabajo se hizo por error directamente sobre `main` (sin issue ni rama previos), incumpliendo el preflight que este mismo prompt exige. Se detectó de inmediato: el commit se movió a una rama nueva (`git branch` + `git reset --hard origin/main`), se creó el issue real `#141`, se renombró la rama a la convención `<tipo>/<issue>-<descripcion>` y se enmendó el commit (todavía no publicado en ningún remoto en ese momento) para referenciar el issue real antes de cualquier `git push`. `main` nunca quedó con el commit fuera de proceso.

## Alcance incluido

- `shared/time/civilDate.ts`: `minutesIntoCivilDate(isoInstant, civilDate, timezone)`, con 4 pruebas nuevas (hora exacta, turno nocturno recortado a 0, turno que cruza a 1440, límites de medianoche).
- `DailyAgendaPage.vue`: línea temporal horizontal (`aria-hidden`, enlaces con `tabindex="-1"`, misma información que la lista), visible solo desde 1024px; título "Agenda"; fichas de lista a 64px mínimo con persona atendida en peso 600 y hora en cifras tabulares; corrección de `--font-size-heading-lg` (token inexistente) a `--font-size-h1`/`--font-size-h1-line`/`--font-weight-h1`.

## Fuera de alcance

- Cualquier acción sobre un turno existente (editar/cancelar/reprogramar/completar): Fase 4b/4c y HUs con contrato propio.
- Selector de barbero, navegador de fecha: sin cambio de comportamiento.
- Mosaico multi-barbero (`P2`, reservado mientras `DEC-074` gobierne P0).

## Estado existente que se conservó

- Toda la lógica de `syncFromRoute`/`loadAgenda` (HU-062/HU-063): sin tocar el `<script>` salvo los nuevos computeds de la línea temporal, que solo leen `entries`/`selectedDate`/`barbershopTimezone` ya existentes.
- La lista cronológica (`<ul class="daily-agenda-page__list">`) sigue siendo la única fuente que un lector de pantalla anuncia; la línea temporal es un artefacto puramente visual añadido, no un reemplazo.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismos 27 warnings preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 601 pruebas en verde (594 previas + 4 `minutesIntoCivilDate` + 3 de la línea temporal: fichas ocultas/no-focalizables junto a la lista accesible, ausencia de línea temporal sin turnos, ausencia de "Ahora" en una fecha que no es hoy).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Verificación visual en navegador real: no realizada esta sesión (sin backend/PostgreSQL local); CI levanta el stack completo.

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
| Línea temporal horizontal desde 1024px | Cumplido | `DailyAgendaPage.vue`, 3 pruebas nuevas |
| Marcador "Ahora" solo si se ve hoy | Cumplido | Prueba "does not show 'Ahora' for a date other than today" |
| Turno nocturno recortado sin reinterpretar el instante | Cumplido | `minutesIntoCivilDate`, 4 pruebas nuevas |
| Lista accesible sin cambio de comportamiento | Cumplido | 18 pruebas preexistentes de `DailyAgendaPage.test.ts` sin cambio de aserciones |
| Ficha 64px, persona como texto principal, hora tabular | Cumplido | Revisión de `DailyAgendaPage.vue` §style |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 142` (4/4 verde); `gh pr view 142 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T04:04:14Z`, commit `f7d6be1` |

## Git y PR

- Commit: `feat(web): agenda diaria NAVA - linea temporal de escritorio (Fase 4a)`.
- PR: [#142](https://github.com/bcaceres19/barberia/pull/142), `Closes #141`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
