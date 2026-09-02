---
prompt_id: "PROMPT-CHORE-173-DETALLE-HISTORIAL-TURNO-NAVA-v1"
version: "1.0"
kind: "chore"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-064"
  - "HU-065"
issue: 173
issue_url: "https://github.com/bcaceres19/barberia/issues/173"
suggested_issue_title: "feat(web): rediseñar detalle e historial del turno NAVA"
branch: "feat/173-detalle-historial-reprogramacion-nava"
pr: null
pr_url: null
depends_on:
  - "Issue maestro #168 (orquestación NAVA); DEC-077-DEC-079 integradas en main"
rules: []
decisions:
  - "DEC-079"
acceptance_criteria:
  - "La pantalla /panel/turnos/:appointmentId respeta la firma cromática de DEC-079 en detalle, historial y el diálogo de reprogramación."
  - "vitest-axe sin violaciones en la vista lista, el diálogo de reprogramación y el conflicto de reprogramación visibles."
  - "Evidencia responsiva real capturada y commiteada bajo apps/web/e2e/evidence/agenda-detalle-historial/ (carpeta nueva, no existía)."
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

# Fase 6 · Detalle, historial y reprogramación NAVA

## Instrucción para el agente

Verifica `/panel/turnos/:appointmentId` (`apps/web/src/modules/agenda/pages/AppointmentDetailPage.vue`) contra el contrato visual mínimo de `DEC-079`.

## Objetivo

Confirmar con evidencia real (creando un turno real con datos sintéticos, no solo leyendo el código) que el detalle, el historial y el diálogo de reprogramación pertenecen a la familia NAVA.

## Preflight ejecutado

1. Árbol limpio verificado por preocupación; `main` actualizado por fast-forward tras integrar #166–#172.
2. Auditoría visual real: se creó un turno real con datos sintéticos (usuario QA aislado `qa.visual@ejemplo.test`, cliente "Cliente Evidencia QA") y se visitó su detalle. Composición, badge de estado, jerarquía de datos, sección "Historial" y el diálogo "Reprogramar turno" (horario actual, nueva fecha/hora, vista previa del nuevo horario) ya son NAVA: sin defecto cromático real encontrado, a diferencia de la Fase 10 (#172).
3. No existía evidencia responsiva previa para esta pantalla (`apps/web/e2e/evidence/agenda-detalle-historial/` no existía); se crea desde cero.

## Alcance incluido

- Cobertura `vitest-axe` adicional con el conflicto de reprogramación (`409`) visible (antes solo se verificaban la vista lista y el diálogo vacío).
- Evidencia responsiva real nueva (320/360/768/1280 px + zoom 200%, normal/foco/diálogo de reprogramación) capturada contra el API local real.

## Fuera de alcance

- Cualquier rediseño estructural: la composición ya cumple.
- Edición T3, cancelación, completar, no-asistencia o corrección terminal (sin HU/issue propio).
- Cambios de contrato, token de versión/concurrencia o reglas de reprogramación.

## Estado existente que se conservó

- Vocabulario "turno", snapshots históricos, permisos, conflicto por cruce/bloqueo (`DEC-076`), preservación de fecha/barbero de origen (`HU-063`) al volver: sin cambios.
- Las 20 pruebas unitarias previas de `AppointmentDetailPage.test.ts`: sin cambio de aserciones.

## Pruebas y evidencia

- `pnpm --filter @system-barbershop/web format`/`lint` (0 errores, mismas advertencias preexistentes)/`typecheck`: limpios.
- `pnpm --filter @system-barbershop/web test:unit`: 44 archivos, 613 pruebas en verde (incluye la prueba `vitest-axe` nueva). Dos reintentos locales mostraron fallos de timeout en pruebas `vitest-axe` no relacionadas (`SchedulesPage`, y una vez la propia `AppointmentDetailPage`) mientras el servidor `vite` corría en paralelo; con el servidor detenido, la suite completa pasó limpia dos veces seguidas — ruido de carga del entorno local, no una regresión (mismo patrón ya documentado en PR #147: "fallo flaky no relacionado").
- `pnpm --filter @system-barbershop/web build`: build de producción exitoso.
- `git diff --check`: sin errores de espacio en blanco.
- Evidencia real con datos sintéticos (usuario QA aislado, turno real creado y luego reprogramado) en `apps/web/e2e/evidence/agenda-detalle-historial/{320,360,768,1280,1280-zoom200}/{normal,foco,reprogramar}.png`.

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
| Firma cromática y tipografía NAVA reales en detalle/historial/reprogramación | Cumplido | Evidencia real capturada 2026-09-02 con un turno real |
| `vitest-axe` sin violaciones en lista, diálogo y conflicto de reprogramación | Cumplido | 3 pruebas `vitest-axe` en `AppointmentDetailPage.test.ts` |
| 320/360/768/1280 px y zoom 200% sin scroll horizontal | Cumplido | Evidencia real, `hasHorizontalScroll` verificado en cada ancho |
| Formato/lint/tipos/tests/build en verde | Cumplido | Comandos de verificación final (613/613 pruebas) |
| Sin dependencias nuevas | Cumplido | `git diff` no toca `package.json`/`pnpm-lock.yaml` |

## Git y PR

- Rama: `feat/173-detalle-historial-reprogramacion-nava`.
- PR: pendiente de abrir, `Closes #173`.
- Sin push directo ni force-push; squash-merge solo tras CI verde.
