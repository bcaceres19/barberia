---
prompt_id: "PROMPT-CHORE-135-SHELL-PRIVADO-NAVA-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "135"
issue_url: "https://github.com/bcaceres19/barberia/issues/135"
suggested_issue_title: "feat(web): shell privado NAVA (wordmark, header, dock, IA de navegación)"
branch: "feat/135-shell-privado-nava"
pr: 136
pr_url: "https://github.com/bcaceres19/barberia/pull/136"
depends_on:
  - "Fase 1 (issue #132, PR #133) integrada en main"
rules:
  - "RN-DIS-02"
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

# Fase 2 · Shell privado — wordmark, header, dock, IA de navegación

## Instrucción para el agente

Implementa únicamente la Fase 2 (Shell privado) del prompt de orquestación `PROMPT-ORCH-NAVA-FRONTEND-v1`: recomponer `AppHeader`/`AppNav`/`PrivateShell` a la identidad NAVA, sin tocar ninguna pantalla de negocio más allá de lo estrictamente necesario para la propia composición del shell.

## Objetivo

El cascarón privado (`PrivateShell.vue`) sigue el patrón de `estandar-diseno-visual.md` §5.3: wordmark NAVA + nombre de barbería en el header, "Nuevo turno" como acción primaria global, dock de navegación fijo al pie con icono + texto y los 5 destinos P0 de `especificacion-frontend-nava.md` §5.1.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar Fase 1 (PR #133).
2. Auditoría de `AppHeader.vue`/`AppNav.vue`/`PrivateShell.vue` existentes (HU-012) y de los 6 `index.ts` de módulo que aportan `NavItem[]` al router.
3. Issue real creado (`#135`) antes de escribir código; rama `feat/135-shell-privado-nava` desde `main` actualizada.

## Alcance incluido

- `NavaWordmark` (nuevo, `shared/ui`): texto accesible, variantes tinta/invertida.
- `AppHeader.vue`: wordmark + barbería, "Nuevo turno" como `RouterLink` con estilo de acción primaria, cierre de sesión conservado (misma lógica y pruebas, sin convertirlo en ítem de dock).
- `AppNav.vue`: dock fijo al pie (64px, `env(safe-area-inset-bottom)`, icono SVG decorativo + texto), `aria-current` nativo de `RouterLink`.
- IA de navegación: "Panel" → "Agenda"; "Barbería" → "Configuración" (solo el rótulo del dock; ruta y título de pantalla sin cambios); "Nuevo turno" deja de ser destino de navegación.
- `PrivateShell.vue`: orden header → main → dock; `padding-bottom` en el contenido para no quedar oculto bajo el dock fijo.

## Fuera de alcance

- Cualquier pantalla de negocio (agenda, catálogo, staff, schedules, settings) más allá de los archivos de composición del shell.
- Self-hosting de fuentes NAVA.
- Reserva pública, P1/P2.

## Desviación registrada durante la implementación

El plan original de este issue proponía mover "Servicios por barbero" (destino que no forma parte de los 5 P0 canónicos) a un enlace textual dentro de `CatalogPage.vue`, para poder retirarlo del dock sin perder su alcanzabilidad. Al implementarlo, rompió la prueba `never shows placeholders for barber assignment, availability, or appointments` de `CatalogPage.test.ts` (HU-022), que verifica expresamente que esa pantalla nunca mencione "barbero" fuera de su una-sola-preocupación.

Revertido en la misma sesión (ver comentario en el issue #135). Resolución final: `barberServicesNavItems` se conserva sin vaciar y "Servicios por barbero" queda como **sexto ítem temporal del dock**, documentado como tal en `apps/web/src/modules/barberServices/index.ts`. Su fusión real dentro de "Servicios" (un único destino, sin ítem aparte) queda para la Fase 5, cuando ambas pantallas se rediseñen juntas.

## Estado existente que se conservó

- Lógica y pruebas de cierre de sesión de `AppHeader.vue` (CA-012-07: doble envío bloqueado, limpia estado local, navega a acceso incluso ante fallo de red ambiguo) sin tocar el script, solo el template/estilo.
- Rutas, guardas (`requireSession`) y permisos de todo el cascarón privado, sin cambios.
- Mecanismo de composición `extraNavItems` (app → modules → shared) entre `auth` y los módulos hermanos, sin alterar su forma (`NavItem[]`).

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format` / `format:write`.
- `pnpm --filter @system-barbershop/web lint`: 0 errores (27 warnings, 26 preexistentes + 1 `vue/no-v-html` en `AppNav.vue`, mismo patrón ya presente en `BaseAlert`/`BaseBadge` para iconos SVG inline decorativos).
- `pnpm --filter @system-barbershop/web typecheck`: limpio.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 591 pruebas en verde, incluye `vitest-axe` sobre `AppHeader`/`AppNav`/`PrivateShell` en estado autenticado y pruebas nuevas: wordmark + CTA "Nuevo turno" en `AppHeader`, icono decorativo + `aria-current` + composición de `extraItems` en `AppNav`.
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Verificación visual en navegador real: **no realizada esta sesión** (sin backend/PostgreSQL local corriendo, sin credenciales disponibles); CI sí levanta el stack completo (`Frontend (lint, tipos, pruebas, build)`, `Atlas + suites SQL`, `Go`).

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice.
- `docs/00-control/historial-cambios.md` y `matriz-trazabilidad.md`: entrada de esta entrega tras confirmar CI verde y merge real.

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
| Wordmark + header NAVA | Cumplido | `AppHeader.vue`, `AppHeader.test.ts` |
| Dock fijo, icono + texto, `aria-current`, safe area | Cumplido | `AppNav.vue`, `AppNav.test.ts` |
| 5 destinos P0 con rótulos correctos | Cumplido | `router/index.ts`, `agenda`/`settings`/`barberServices`/`index.ts` |
| "Nuevo turno" como acción global, no destino | Cumplido | `AppHeader.vue` (CTA), `agenda/index.ts` (nav vacío) |
| Cierre de sesión conservado sin regresión | Cumplido | `AppHeader.test.ts` (4 pruebas ya existentes sin cambio de aserciones) |
| Ninguna pantalla de negocio modificada fuera del shell | Cumplido | Diff de PR #136 limitado a `auth`, `shared/ui`, `app/router`, e `index.ts` de módulo |
| "Servicios por barbero" fusionado a "Servicios" | Diferido a Fase 5 (documentado) | Comentario en issue #135; comentario en `barberServices/index.ts` |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final, salida capturada en esta sesión |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 136` (4/4 verde); `gh pr view 136 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-02T03:36:39Z`, commit `7eb928e` |

## Git y PR

- Commit: `feat(web): rediseña el shell privado a la identidad NAVA (Fase 2)`.
- PR: [#136](https://github.com/bcaceres19/barberia/pull/136), `Closes #135`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
