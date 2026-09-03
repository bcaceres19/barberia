---
prompt_id: "PROMPT-CHORE-171-CATALOGO-SERVICIOS-NAVA-v1"
version: "1.1"
kind: "chore"
status: "executed"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-022"
  - "HU-024"
issue: 171
issue_url: "https://github.com/bcaceres19/barberia/issues/171"
suggested_issue_title: "feat(web): rediseñar catálogo de servicios NAVA"
branch: "feat/171-catalogo-servicios-nava"
pr: 180
pr_url: "https://github.com/bcaceres19/barberia/pull/180"
depends_on:
  - "Issue maestro #168 (orquestación NAVA); DEC-077-DEC-079 integradas en main"
rules: []
decisions:
  - "DEC-079"
  - "DEC-067"
acceptance_criteria:
  - "La pantalla /panel/servicios respeta la firma cromática de DEC-079 en la lista, el alta/edición y la desactivación."
  - "vitest-axe sin violaciones en la lista, los diálogos de alta/desactivación y el error de carga."
  - "Evidencia responsiva real capturada y commiteada bajo apps/web/e2e/evidence/servicios/."
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

# Fase 9 · Catálogo de servicios NAVA

## Instrucción para el agente

Verifica `/panel/servicios` (`apps/web/src/modules/catalog/pages/CatalogPage.vue`) contra el contrato visual mínimo de `DEC-079`.

## Objetivo

Confirmar con evidencia real que la lista, el alta/edición y la desactivación pertenecen a la familia NAVA.

## Preflight ejecutado

1. Árbol limpio verificado por preocupación; `main` actualizado por fast-forward tras integrar #166–#173.
2. Auditoría visual real (usuario QA aislado `qa.visual@ejemplo.test`): lista con tarjeta, insignia Activo/Inactivo (nunca solo color), botones Editar/Desactivar, diálogo "Agregar servicio" con label→control→ayuda, y diálogo "Desactivar servicio" con consecuencia real ("El servicio y su historial se conservan; nunca se borra", "No hay citas futuras que se vean afectadas ahora mismo") ya son NAVA. El precio se muestra como `"25000.00 COP"` (sin separador de miles ni símbolo `"$"`) de forma DELIBERADA: `estandar-diseno-visual.md` §10.3 prohíbe concatenar `"$"` o asumir pesos desde el componente (issue #155 ya resolvió este mismo punto); no se toca.
3. Esta pantalla ya tenía la cobertura `vitest-axe` más completa de las cuatro fases pendientes (lista, diálogo de alta, diálogo de desactivación); solo faltaba el estado de error de carga.

## Alcance incluido

- Cobertura `vitest-axe` adicional con el estado de error de carga visible (antes solo se verificaban la lista y los dos diálogos).
- Evidencia responsiva real (320/360/768/1280 px + zoom 200%, normal y foco) capturada contra el API local real con `axe-core` en vivo, reemplazando las capturas de 2026-08-24 (previas a NAVA).

## Fuera de alcance

- Cualquier rediseño estructural: la composición, la anatomía de campos y los diálogos ya cumplían.
- El formato del precio (`"25000.00 COP"`): es la aplicación correcta y ya documentada de `estandar-diseno-visual.md` §10.3, no un defecto.
- Asignación a barberos (`HU-023`, pantalla separada, Fase 10/#172).
- Cambios de contrato, moneda configurable, disponibilidad o cálculo de impacto de desactivación en el cliente.

## Estado existente que se conservó

- Nombre único activo, duración planificada, precio COP mayor que cero (`DEC-067`), impacto real de desactivación consultado al servidor (nunca calculado en Vue), reactivación sin campos de confirmación adicionales (`CA-024-05`): sin cambios.
- Las 27 pruebas unitarias previas de `CatalogPage.test.ts`: sin cambio de aserciones.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismas advertencias preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 614 pruebas en verde (incluye la prueba `vitest-axe` nueva).
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Evidencia real con datos sintéticos (usuario QA aislado) en `apps/web/e2e/evidence/servicios/{320,360,768,1280,1280-zoom200}/{normal,foco}.png`, verificada además con `axe-core` en vivo sobre el DOM real (0 violaciones) en los cinco anchos.

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
| Firma cromática y tipografía NAVA reales en lista/diálogos | Cumplido | Evidencia real capturada 2026-09-02 |
| `vitest-axe` sin violaciones en lista, diálogos y error de carga | Cumplido | 4 pruebas `vitest-axe` en `CatalogPage.test.ts` |
| 320/360/768/1280 px y zoom 200% sin scroll horizontal, sin violaciones axe reales | Cumplido | Evidencia real; `axe-core` en vivo, 0 violaciones en los 5 anchos |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final (614/614 pruebas) |
| Sin dependencias nuevas | Cumplido | `git diff` no toca `package.json`/`pnpm-lock.yaml` |

## Git y PR

- Rama: `feat/171-catalogo-servicios-nava`.
- PR: [#180](https://github.com/bcaceres19/barberia/pull/180), `Closes #171`, integrada en `main` con CI 4/4 verde (commit `a908bef`).
- Sin push directo ni force-push; squash-merge solo tras CI verde.
