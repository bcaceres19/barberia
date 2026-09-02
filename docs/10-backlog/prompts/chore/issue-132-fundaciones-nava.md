---
prompt_id: "PROMPT-CHORE-132-FUNDACIONES-NAVA-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "132"
issue_url: "https://github.com/bcaceres19/barberia/issues/132"
suggested_issue_title: "feat(web): migra fundaciones NAVA (tokens, tipografía, primitivas base)"
branch: "feat/132-fundaciones-nava"
pr: 133
pr_url: "https://github.com/bcaceres19/barberia/pull/133"
depends_on:
  - "DEC-077 registrada y especificación NAVA integrada en main (PR #131)"
rules:
  - "RN-CIT-02"
  - "RN-DIS-02"
decisions:
  - "DEC-039"
  - "DEC-077"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/prompts/orchestration/adopcion-frontend-nava.md"
created_at: "2026-09-01"
updated_at: "2026-09-01"
supersedes: null
superseded_by: null
---

# Fase 1 · Fundaciones NAVA — tokens, tipografía y primitivas base

## Instrucción para el agente

Implementa únicamente la Fase 1 (Fundaciones) del prompt de orquestación `PROMPT-ORCH-NAVA-FRONTEND-v1`: migrar `tokens.css`/`base.css` a los valores NAVA 2.0, sin tocar pantallas de negocio ni self-hostear fuentes.

## Objetivo

`apps/web/src/styles/tokens.css` y `base.css` reflejan la identidad NAVA / Tailored Grid (`DEC-077`, `estandar-diseno-visual.md` v2.0): color, tipografía, espaciado, radios, sombras y overlays. Las primitivas ya construidas (`BaseButton`, `BaseInput`, `BaseDialog`, `BaseAlert`, `BaseBadge`) heredan los nuevos valores sin edición propia.

## Preflight ejecutado

1. Árbol limpio verificado; `main` actualizado por fast-forward tras integrar `PR #131` (docs NAVA).
2. Auditoría de consumidores: `--font-family-base` y `--radius-*` se usan en primitivas y en pantallas de negocio (`DailyAgendaPage`, `AppointmentDetailPage`, `SchedulesPage`, `CatalogPage`, `StaffPage`, `BarberServicesPage`, `AppNav`, `LoginForm`, `RecoveryPage`); ninguna tiene un hexadecimal o tamaño local, todas consumen solo tokens semánticos.
3. Issue real creado (`#132`) antes de escribir código; rama `feat/132-fundaciones-nava` desde `main` actualizada.

## Alcance incluido

- Reescritura completa de `tokens.css`: paleta NAVA (tinta, marfil, grafito, latón, salvia, piedra), escala tipográfica (`display` responsivo 36/40 móvil → 48/52 desde 1024px, `h1` 32/38, `h2` 24/30, `h3` 19/26), radios (4/6/8px), sombras y overlays derivados de `nava-ink`.
- `--font-display` / `--font-sans` nuevos, con fallback canónico; `--font-family-base` se conserva como alias de `--font-sans` para no editar las pantallas de negocio que ya lo consumen.
- Comentario normativo de `base.css` actualizado a `DEC-077` (sin cambio funcional).

## Fuera de alcance

- Self-hosting de `Instrument Serif`/`Instrument Sans` como WOFF2: requiere issue propio con licencia, peso de archivos, `font-display` y medición (`estandar-diseno-visual.md` §5.1). Hasta entonces el fallback declarado es el comportamiento válido.
- Cualquier edición de pantallas de negocio: heredan los nuevos tokens sin que se les toque un archivo.
- Shell privado, wordmark, dock (Fase 2 del prompt de orquestación).

## Estado existente que se conservó

- Las cinco primitivas de `shared/ui` seguían el patrón "solo tokens semánticos" antes de este cambio; se confirmó por lectura completa de cada archivo, sin encontrar un hexadecimal o `px` de color/tipografía suelto.
- Ningún test de componente asocia un valor de token con un resultado esperado (`grep` sin coincidencias de `--font-weight-*`/tamaños en `__tests__`), por lo que el cambio de valores no arriesga una prueba existente.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format` / `format:write` (solo `tokens.css`/`base.css` reformateados).
- `pnpm --filter @system-barbershop/web lint`: 0 errores (26 warnings preexistentes, no relacionados).
- `pnpm --filter @system-barbershop/web typecheck`: limpio.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 588 pruebas en verde (incluye `vitest-axe`).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Verificación visual con `pnpm dev` y Chrome MCP en `/acceso`: lienzo marfil `#F4F0E7`, botón primario en tinta `#101B2B`, texto grafito — confirma que la cascada de tokens llega a una pantalla existente sin editarla.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice para este prompt.
- `docs/00-control/historial-cambios.md` y `matriz-trazabilidad.md`: entrada de esta entrega tras confirmar CI verde y merge real (no antes).

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
| `tokens.css`/`base.css` migrados a NAVA 2.0 | Cumplido | Diff completo en PR #133 |
| Primitivas heredan sin edición propia | Cumplido | Lectura de las 5 primitivas: solo tokens semánticos, sin diff propio |
| Pantallas de negocio sin tocar | Cumplido | `git status`/diff de PR #133 limitado a `styles/` |
| Fuentes NAVA self-hosted | No incluido (fuera de alcance, issue propio pendiente) | — |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final, salida capturada en esta sesión |
| CI real (GitHub Actions) en verde | Pendiente de confirmar | `gh pr checks 133` |

## Git y PR

- Commit: `feat(web): migra tokens.css y base.css al sistema NAVA 2.0`.
- PR: [#133](https://github.com/bcaceres19/barberia/pull/133), `Closes #132`.
- Sin push directo ni force-push; sin merge de `main` fuera del squash-merge autorizado tras CI verde.
