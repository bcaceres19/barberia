---
prompt_id: "PROMPT-CHORE-144-NUEVO-TURNO-NAVA-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-061"
issue: "144"
issue_url: "https://github.com/bcaceres19/barberia/issues/144"
suggested_issue_title: "feat(web): NAVA en Nuevo turno — resumen antes del CTA y tokens"
branch: "feat/144-nuevo-turno-nava"
pr: 145
pr_url: "https://github.com/bcaceres19/barberia/pull/145"
depends_on:
  - "Fase 4a (issue #141, PR #142) integrada en main"
rules: []
decisions:
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

# Fase 4b · Nuevo turno — resumen antes del CTA y tokens

## Instrucción para el agente

Implementa únicamente la Fase 4b (segunda entrega de "Operación de agenda") del prompt de orquestación `PROMPT-ORCH-NAVA-FRONTEND-v1`: el resumen antes del CTA en `NewAppointmentPage.vue` y la corrección de referencias CSS rotas encontradas en la auditoría, sin cambiar contrato ni reordenar los campos existentes.

## Objetivo

`NewAppointmentPage.vue` muestra, antes del botón "Registrar turno", un resumen en vivo de lo ya seleccionado (barbero, servicio, persona atendida, fecha/hora legible), como exige `especificacion-frontend-nava.md` §7.3. Las referencias `var(--color-border, #ccc)`/`var(--color-danger, #b00020)`, que nunca correspondieron a tokens reales, quedan corregidas a `--color-border-control`/`--color-danger-text`.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar Fase 4a (PR #142).
2. Auditoría completa de `NewAppointmentPage.vue`: se encontró que el archivo usaba `var(--color-border, #ccc)`, `var(--color-danger, #b00020)` y valores `rem`/`px` sueltos sin ningún token real detrás — `tokens.css` nunca definió `--color-border` ni `--color-danger` (los nombres reales son `--color-border-control`/`--color-danger-text`), así que el navegador venía renderizando directamente los hexadecimales de reserva. Además, `.new-appointment-page`, `__header`, `__title`, `__timezone`, `__state` y `__empty` no tenían ninguna regla CSS propia (dependían por completo de estilos por defecto del navegador).
3. Issue real creado (`#144`) antes de escribir código; rama `feat/144-nuevo-turno-nava` desde `main` actualizada.

## Corrección de proceso durante esta entrega

La implementación se hizo por error sobre la rama de otro PR ya abierto (`docs/confirma-fase4a-agenda-timeline`, con PR #143 pendiente de CI). Se detectó antes de commitear: los cambios (sin commitear) se guardaron con `git stash`, se creó el issue real `#144`, se abrió la rama correcta `feat/144-nuevo-turno-nava` desde `main` actualizada, y se aplicó `git stash pop` ahí. El PR #143 nunca se vio afectado.

## Alcance incluido

- Resumen en vivo antes del CTA: `dl` con barbero, servicio (solo nombre, sin inventar duración/precio que `ServiceSummary` no expone), persona atendida y fecha/hora (`formatCivilDateFull` + hora tal cual se escribió).
- Corrección de tokens rotos: `--color-border-control`, `--color-danger-text`, escala real de espaciado (`--space-*`) y radio (`--radius-sm`).
- Estilos nuevos de página que faltaban por completo (header, título, timezone, estados de carga/vacío).

## Fuera de alcance

- Cambios de contrato (`ServiceSummary` no gana duración/precio en esta entrega).
- Reordenar los campos existentes del formulario (bajo valor frente al riesgo de tocar un formulario ya probado con HU-061).
- Cualquier otra pantalla de negocio.

## Estado existente que se conservó

- Toda la lógica de `onSubmit`/`validateAll`/`watch(selectedBarberId)` (HU-061, DEC-072/073): sin tocar el `<script>` salvo los nuevos computeds de resumen, que solo leen refs ya existentes.
- Los 13 casos de prueba preexistentes (carga, validación, conflicto de agenda/bloqueo, conflicto de idempotencia, error de red, doble envío bloqueado, accesibilidad) sin cambio de aserciones.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismos 27 warnings preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 604 pruebas en verde (601 previas + 3 nuevas: sin resumen en formulario vacío, resumen parcial con solo barbero elegido, resumen completo con fecha legible).
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
| Resumen en vivo antes del CTA | Cumplido | 3 pruebas nuevas en `NewAppointmentPage.test.ts` |
| Sin datos inventados (duración/precio) | Cumplido | Revisión de `ServiceSummary` (solo `id`/`name`) |
| Tokens CSS rotos corregidos | Cumplido | Revisión de `NewAppointmentPage.vue` §style |
| Sin regresión en HU-061 | Cumplido | 10 pruebas preexistentes sin cambio de aserciones |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 145` (4/4 verde); `gh pr view 145 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T04:14:57Z`, commit `d943d1d` |

## Git y PR

- Commit: `feat(web): Nuevo turno NAVA - resumen antes del CTA y tokens (Fase 4b)`.
- PR: [#145](https://github.com/bcaceres19/barberia/pull/145), `Closes #144`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
