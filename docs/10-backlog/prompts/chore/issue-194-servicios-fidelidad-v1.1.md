---
prompt_id: "PROMPT-CHORE-194-SERVICIOS-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-022"
related_hu: ["HU-024", "HU-009", "HU-012"]
issue: 194
issue_url: "https://github.com/bcaceres19/barberia/issues/194"
branch: "chore/194-servicios-fidelidad"
pr: null
pr_url: null
depends_on: ["Issue #193 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-SER-01", "RN-SER-02", "RN-SER-03", "RN-SER-04", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-067", "DEC-069", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-022-01", "CA-022-02", "CA-022-03", "CA-022-04", "CA-022-05", "CA-022-06", "CA-022-07", "CA-024-01", "CA-024-02", "CA-024-04", "CA-024-05"]
created_at: "2026-09-08"
updated_at: "2026-09-08"
supersedes: null
superseded_by: null
---

# Fidelidad de `/panel/servicios` al atlas NAVA

Implementa el issue [#194](https://github.com/bcaceres19/barberia/issues/194)
desde la rama indicada, usando la mitad izquierda de
`05-servicios-y-asignacion.png` como referencia exacta.

## Alcance ejecutado

- Recompone encabezado editorial, CTA, tabla compacta de catálogo y reflow
  móvil de `/panel/servicios`.
- Conserva alta, edición, paginación, precio decimal COP, duración, estados
  Activo/Inactivo, desactivación, reactivación, idempotencia e impacto real.
- No agrega asignaciones, disponibilidad, citas, borrado físico, API ni
  cambios de base de datos.

## Evidencia y puerta de calidad

- `e2e/servicios-fidelidad-mock.spec.ts` prueba tabla en 1280 y 360, foco y
  reflow en 320/360/768/1280/zoom 200 con datos ficticios.
- La diferencia heredada de shell (dock horizontal frente a sidebar del
  atlas) se conserva aislada de esta ruta, como en #192 y #193.
- Antes de integrar: componente, tipos, lint, E2E, build, diff y checks CI.
