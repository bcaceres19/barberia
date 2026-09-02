---
prompt_id: "PROMPT-CHORE-172-SERVICIOS-POR-BARBERO-NAVA-v1"
version: "1.1"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-023"
issue: 172
issue_url: "https://github.com/bcaceres19/barberia/issues/172"
suggested_issue_title: "feat(web): rediseñar asignación de servicios por barbero NAVA"
branch: "feat/172-servicios-por-barbero-nava"
pr: 178
pr_url: "https://github.com/bcaceres19/barberia/pull/178"
depends_on:
  - "Issue maestro #168 (orquestación NAVA); DEC-077-DEC-079 integradas en main"
rules: []
decisions:
  - "DEC-079"
acceptance_criteria:
  - "La pantalla /panel/servicios-por-barbero respeta la firma cromática de DEC-079, incluida la casilla de asignación."
  - "vitest-axe sin violaciones en la lista lista, el conflicto de última asignación y el error de carga."
  - "Evidencia responsiva real capturada y commiteada bajo apps/web/e2e/evidence/servicios-por-barbero/."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/10-backlog/prompts/orchestration/adopcion-frontend-nava.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Fase 10 · Servicios por barbero NAVA

## Instrucción para el agente

Verifica `/panel/servicios-por-barbero` (`apps/web/src/modules/barberServices/pages/BarberServicesPage.vue`) contra el contrato visual mínimo de `DEC-079` y corrige el defecto real encontrado.

## Objetivo

Confirmar con evidencia real que la pantalla pertenece a la familia NAVA, y corregir la única desviación cromática real detectada.

## Preflight ejecutado

1. Árbol limpio verificado por preocupación; `main` actualizado por fast-forward tras integrar #166–#170.
2. Auditoría visual real (usuario QA aislado `qa.visual@ejemplo.test`): la composición, el selector de barbero, la lista de servicios con objetivo táctil de 44 px (fix #157) y los tokens de superficie/borde ya son NAVA. Hallazgo real: la casilla de asignación (`<input type="checkbox">` nativo) se renderizaba con el azul por defecto del navegador — ningún `accent-color` la ligaba a la tinta NAVA. Ningún otro control nativo similar existe todavía en el repositorio (sin precedente de `accent-color` en `apps/web/src`).

## Alcance incluido

- `accent-color: var(--color-action-primary)` en `.barber-services-page__checkbox` (una línea, sin tocar tamaño, objetivo táctil ni comportamiento).
- Cobertura `vitest-axe` adicional con la alerta de conflicto de última asignación (`DEC-068`) y el estado de error de carga visibles (antes solo se verificaba la vista lista).
- Evidencia responsiva real (320/360/768/1280 px + zoom 200%, normal y foco) capturada contra el API local real, con `axe-core` ejecutado en vivo sobre cada breakpoint (mismo patrón que `servicios-por-barbero-evidencia-responsiva.spec.ts`).

## Fuera de alcance

- Cualquier rediseño estructural: selector, fieldset, lista y estados ya cumplían.
- Fusionar esta ruta dentro de Servicios (la especificación lo permite pero no lo exige; ningún issue lo pide).
- Cambios de contrato, idempotencia de asignación o la regla de última asignación activa.

## Estado existente que se conservó

- Aislamiento entre barberos, guardia de doble envío por casilla, reversión exacta al estado real confirmado por el servidor ante cualquier error, y objetivo táctil de 44 px (issue #157): sin cambios.
- Las 16 pruebas unitarias previas: sin cambio de aserciones.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismas advertencias preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 612 pruebas en verde (incluye las 2 pruebas `vitest-axe` nuevas).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Evidencia real con datos sintéticos (usuario QA aislado) en `apps/web/e2e/evidence/servicios-por-barbero/{320,360,768,1280,1280-zoom200}/{normal,foco}.png`, verificada además con `axe-core` en vivo sobre el DOM real (sin violaciones) en los cinco anchos.

## Documentación y trazabilidad

- `docs/10-backlog/prompts/README.md`: nueva fila de índice.
- `docs/00-control/matriz-trazabilidad.md` / `historial-cambios.md`: actualizados al confirmar el merge real.

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
| Casilla de asignación con tinta NAVA (no azul del navegador) | Cumplido | `accent-color` añadido; evidencia real `1280/normal.png` |
| Estados normal/vacío/error/conflicto/pendiente verificados | Cumplido | Pruebas unitarias existentes + 2 `vitest-axe` nuevas |
| 320/360/768/1280 px y zoom 200% sin scroll horizontal, sin violaciones axe reales | Cumplido | Evidencia real; `axe-core` ejecutado en vivo, 0 violaciones en los 5 anchos |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final (612/612 pruebas) |
| Sin dependencias nuevas | Cumplido | `git diff` no toca `package.json`/`pnpm-lock.yaml` |

## Git y PR

- Rama: `feat/172-servicios-por-barbero-nava`.
- PR: [#178](https://github.com/bcaceres19/barberia/pull/178), `Closes #172`, integrada en `main` con CI 4/4 verde (commit `19be1ca`).
- Sin push directo ni force-push; squash-merge solo tras CI verde.
