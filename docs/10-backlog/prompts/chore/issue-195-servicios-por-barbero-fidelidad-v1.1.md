---
prompt_id: "PROMPT-CHORE-195-SERVICIOS-POR-BARBERO-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-023"
related_hu: ["HU-021", "HU-022", "HU-024", "HU-009", "HU-012"]
issue: 195
issue_url: "https://github.com/bcaceres19/barberia/issues/195"
branch: "chore/195-servicios-por-barbero-fidelidad"
pr: 233
pr_url: "https://github.com/bcaceres19/barberia/pull/233"
depends_on: ["Issue #194 integrado en main con squash d2c0f06", "Issue maestro #184"]
rules: ["RN-SER-03", "RN-SER-04", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-068", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-023-01", "CA-023-02", "CA-023-03", "CA-023-04", "CA-023-05", "CA-023-06", "CA-023-07", "CA-023-08"]
created_at: "2026-09-08"
updated_at: "2026-09-08"
supersedes: null
superseded_by: null
---

# Fidelidad de `/panel/servicios-por-barbero` al atlas NAVA

Implementa el issue [#195](https://github.com/bcaceres19/barberia/issues/195)
desde la rama indicada, usando la mitad derecha de
`05-servicios-y-asignacion.png` como referencia exacta.

## Alcance

- Recompone el encabezado editorial, selector compacto, carril de asignaciones,
  chips de estado, guardado y alerta de última asignación de
  `/panel/servicios-por-barbero`.
- Conserva selector de barbero, escritura real por casilla, prevención de doble
  envío, reversión al estado confirmado y rechazo visible de `DEC-068`.
- No expone duración o precio en esta ruta: `CA-023-07` prohíbe que el contrato
  de asignación incluya datos del catálogo por barbero. El mockup ilustra esos
  valores, pero no autoriza ampliar el contrato ni fabricar datos.
- No modifica API, OpenAPI, migraciones, RLS, datos de prueba ni la regla de
  negocio de última asignación.

## Evidencia y calidad

- `e2e/servicios-por-barbero-fidelidad-mock.spec.ts` usa datos sintéticos y
  captura escritorio, móvil, conflicto, foco y reflow en 320/360/768/1280/zoom
  200. Sus artefactos quedan bajo `apps/web/e2e/evidence/servicios-por-barbero/fidelidad-195/`.
- Se mantienen las pruebas de componente, incluida `vitest-axe`, y el E2E real
  responsivo existente de HU-023.
- La navegación global horizontal en móvil difiere de la sidebar del atlas por
  ser una decisión heredada del shell (#187), fuera de esta ruta.
- Antes de integrar: formato dirigido, tipo, lint, componente, E2E mock, build,
  `git diff --check` y checks CI verdes.

## Pull request

- [PR #233](https://github.com/bcaceres19/barberia/pull/233), pendiente de
  checks antes del squash merge autorizado.
