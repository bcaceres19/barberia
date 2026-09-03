---
prompt_id: "PROMPT-CHORE-186-FUNDACIONES-TAILORED-GRID-v1"
version: "1.0"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "186"
issue_url: "https://github.com/bcaceres19/barberia/issues/186"
suggested_issue_title: "chore(web): fundaciones visuales compartidas Tailored Grid"
branch: "chore/186-fundaciones-tailored-grid"
pr: 199
pr_url: "https://github.com/bcaceres19/barberia/pull/199"
depends_on:
  - "Atlas integral integrado en main (issue #183/PR #185)"
rules:
  - "RN-DIS-02"
  - "RN-DIS-04"
decisions:
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v2.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 1 · Fundaciones de composición Tailored Grid

## Instrucción para el agente

Ejecuta únicamente la Fase 1 (Fundaciones) de `PROMPT-ORCH-NAVA-ATLAS-v2`: añadir las primitivas de composición que faltan en `shared/ui` para que las Fases 2-12 puedan traducir el atlas a carriles, divisores y densidad Tailored Grid, sin tocar ninguna pantalla de negocio ni el shell.

## Objetivo

`tokens.css`/`base.css` y las primitivas existentes (`BaseButton`, `BaseInput`, `BaseAlert`, `BaseBadge`, `BaseDialog`) ya cumplían la firma cromática NAVA desde la orquestación anterior (#168/PR #133): auditoría completa sin hallazgos, ningún hexadecimal o tamaño suelto fuera de tokens. El gap real detectado en la línea base visual (capturas Chrome DevTools previas a esta fase, cuenta QA `bacg20044@gmail.com`) no era de color sino de composición: las 11 rutas repiten un encabezado ad-hoc por pantalla y listan sus registros en tarjetas blancas con borde, sin el carácter de carril/divisor fino del atlas.

## Preflight ejecutado

1. Atlas (#183/PR #185) integrado en `main`; `main` actualizado por fast-forward.
2. Router real inventariado contra los 11 archivos `routes.ts` de cada módulo: coincide exactamente con las 11 rutas del atlas.
3. Línea base visual capturada con Chrome DevTools (cuenta QA `bacg20044@gmail.com`, stack `barberia-qa-local`) en 1280 px para las 11 rutas: confirma composición plana, sin carriles, con tarjetas blancas repetidas.
4. Lectura completa de `tokens.css`, `base.css`, `BaseButton.vue`, `BaseInput.vue`: ya usan solo tokens semánticos, sin hexadecimal o tamaño suelto.
5. Issue real creado (`#186`) bajo el maestro `#184` antes de escribir código; rama `chore/186-fundaciones-tailored-grid` desde `main` actualizada.

## Alcance incluido

- `PageHeader.vue` (`shared/ui`): cabecera de página repetible — título editorial único `h1`, contexto opcional, enlace de retorno opcional y región de acciones que se apila en móvil (`max-width: 480px`). Sustituye el encabezado que cada pantalla reinventaba.
- `RecordRow.vue` (`shared/ui`): renglón de carril para listas de registros — divisor fino entre filas (`border-bottom`, sin borde en el último), regiones `leading`/principal/`trailing`, `tag` configurable (`li` por defecto para `<ul>`, `div` cuando no aplica semántica de lista). No es interactivo por sí mismo: el consumidor coloca el enlace/botón dentro de una región, igual que ya hace la agenda con el turno listado.
- Export público desde `shared/ui/index.ts`.
- `estandar-diseno-visual.md` §6.6 nuevo, documentando el patrón (4.1 → 4.2).
- Pruebas de componente (renderizado, slots condicionales, `tag`, enlace de retorno con `vue-router`) y `vitest-axe` para ambos componentes.

## Fuera de alcance

- Cualquier pantalla de negocio: ninguna adopta `PageHeader`/`RecordRow` todavía; eso ocurre en las Fases 3-12 del prompt de orquestación, ruta por ruta.
- Shell privado (`PrivateShell`/`AppNav`), dock móvil: Fase 2 del prompt de orquestación.
- Self-hosting de fuentes, tokens de color nuevos: no había hallazgo que lo requiriera.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`: sin cambios de contenido en `tokens.css`/`base.css` (normalización de línea preexistente, sin diff real).
- `pnpm --filter @system-barbershop/web lint`: 0 errores, 27 warnings preexistentes (ninguno en `PageHeader.vue`/`RecordRow.vue`).
- `pnpm --filter @system-barbershop/web typecheck`: limpio.
- `pnpm --filter @system-barbershop/web test:unit`: 46 archivos, 627 pruebas en verde (incluye los 13 casos nuevos de `PageHeader`/`RecordRow`, con `vitest-axe`).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Verificación manual con Chrome DevTools (cuenta QA `bacg20044@gmail.com`, stack `barberia-qa-local`, 1280 px): línea base capturada para las 11 rutas antes de este cambio; ninguna pantalla se modificó, por lo que no hay regresión visual que verificar en esta fase.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice para este prompt.
- `docs/00-control/historial-cambios.md`: entrada de esta entrega con el PR real.

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
| Primitivas de composición nuevas sin tocar pantallas de negocio | Cumplido | Diff de PR #199 limitado a `shared/ui/` y al estándar |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final, salida capturada en esta sesión |
| CI real (GitHub Actions) en verde | Cumplido | `gh pr checks 199` (4/4 verde); `gh pr view 199 --json state,mergedAt,mergeCommit`: `MERGED`, `2026-09-03T02:16:38Z`, commit `2dc5a0b` |

## Git y PR

- Commit: `chore(web): fundaciones de composicion Tailored Grid (#186)`.
- PR: [#199](https://github.com/bcaceres19/barberia/pull/199), `Refs #184`, `Closes #186`.
- Sin push directo ni force-push; squash-merge tras CI verde, rama eliminada.
