---
prompt_id: "PROMPT-FIX-149-TOKENS-CONFIGURACION-v1"
version: "1.0"
kind: "fix"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-020"
  - "HU-021"
  - "HU-022"
  - "HU-023"
  - "HU-040"
  - "HU-041"
  - "HU-042"
issue: "149"
issue_url: "https://github.com/bcaceres19/barberia/issues/149"
suggested_issue_title: "fix(web): corrige tokens de tipografía y color rotos en pantallas de configuración"
branch: "fix/149-tokens-pantallas-configuracion"
pr: 150
pr_url: "https://github.com/bcaceres19/barberia/pull/150"
depends_on:
  - "Fase 4c (issue #146, PR #147) integrada en main"
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

# Fix · Tokens rotos en las seis pantallas de configuración

## Instrucción para el agente

Corrige el mismo defecto de tokens de tipografía/color inexistentes (ya corregido en `DailyAgendaPage.vue` issue #141 y `AppointmentDetailPage.vue` issue #146) en las seis pantallas de configuración restantes, sin ningún rediseño estructural.

## Objetivo

`StaffPage.vue`, `CatalogPage.vue`, `SchedulesPage.vue`, `BarberServicesPage.vue`, `SettingsPage.vue` y `BlocksPage.vue` usan tokens de tipografía y color reales en sus títulos y bordes.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar el fix de Fase 4c (PR #147).
2. Auditoría con `grep` de `--font-size-heading-lg`/`--font-size-heading-md`/`var(--color-border,`/`var(--color-danger,` en `apps/web/src/modules`: confirmó exactamente estos 6 archivos, ninguno más.
3. Issue real creado (`#149`) antes de escribir código; rama `fix/149-tokens-pantallas-configuracion` desde `main` actualizada.

## Alcance incluido

- `StaffPage.vue` ("Barberos"): título → `--font-size-h1`; fila de la lista gana `min-height` 64px y el nombre pasa a peso 600.
- `CatalogPage.vue` ("Servicios"): título → `--font-size-h1`.
- `SchedulesPage.vue` ("Horarios"): título → `--font-size-h1`.
- `BarberServicesPage.vue` ("Servicios por barbero"): título → `--font-size-h1`.
- `SettingsPage.vue` ("Barbería"): título → `--font-size-h1`.
- `BlocksPage.vue` ("Horarios y bloqueos"): título sin ninguna regla CSS propia (se agrega); `.blocks-page__select`/`__item` usaban `var(--color-border, #ccc)` (nunca existió) → `--color-border-control`/`--color-border-subtle` y radio real.

## Fuera de alcance

- Cualquier rediseño estructural (secciones, resumen, líneas temporales, fichas nuevas).
- Cambios de contrato o de lógica de negocio en ninguna de las seis pantallas.

## Estado existente que se conservó

- Los 89 casos de prueba de los 5 archivos con cobertura unitaria (`StaffPage`, `CatalogPage`, `SchedulesPage`, `BarberServicesPage`, `SettingsPage`) sin cambio de aserciones. `BlocksPage.vue` no tiene pruebas unitarias propias (gap preexistente, issue #100).

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismos 27 warnings preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 604 pruebas en verde (sin cambio: fix de solo estilo).
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
| 5 títulos con jerarquía real (`--font-size-h1`) | Cumplido | Revisión de cada archivo §style |
| `BlocksPage` con título estilado y bordes reales | Cumplido | Revisión de `BlocksPage.vue` §style |
| Sin defecto restante en `apps/web/src/modules` | Cumplido | `grep` de los 4 patrones tras el cambio: sin coincidencias |
| Sin regresión en las 5 pantallas con pruebas unitarias | Cumplido | 89 pruebas sin cambio de aserciones |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 150` (4/4 verde); `gh pr view 150 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T04:27:26Z`, commit `21abdc7` |

## Git y PR

- Commit: `fix(web): corrige tokens de tipografia y color rotos en pantallas de configuracion`.
- PR: [#150](https://github.com/bcaceres19/barberia/pull/150), `Closes #149`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
