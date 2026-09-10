---
prompt_id: "PROMPT-CHORE-196-HORARIOS-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-040"
related_hu: ["HU-041", "HU-009", "HU-012"]
issue: 196
issue_url: "https://github.com/bcaceres19/barberia/issues/196"
branch: "chore/196-horarios-fidelidad"
pr: 234
pr_url: "https://github.com/bcaceres19/barberia/pull/234"
depends_on: ["Issue #195 integrado en main con squash 0b39cda", "Issue maestro #184"]
rules: ["RN-DIS-06", "RN-DIS-07", "RN-BLQ-01", "RN-BLQ-02", "RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-006", "DEC-007", "DEC-020", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-040-01", "CA-040-02", "CA-040-03", "CA-040-04", "CA-040-05", "CA-040-06", "CA-041-01", "CA-041-02", "CA-041-04", "CA-041-05"]
created_at: "2026-09-09"
updated_at: "2026-09-09"
supersedes: null
superseded_by: null
---

# Fidelidad de `/panel/horarios` al atlas NAVA

Implementa [#196](https://github.com/bcaceres19/barberia/issues/196) desde la
rama indicada, usando `07-horarios-y-excepciones.png` como referencia visual
exacta sin ampliar las operaciones reales de jornada, calendario y excepciones.

## Alcance

- Recompone la superficie editorial de jornada semanal, selector de barbero,
  tramos, días cerrados, festivos y excepciones.
- Conserva la zona IANA de la barbería, las siete jornadas ISO, creación,
  edición, retiro, idempotencia y conflictos reales existentes.
- No agrega calendario mensual, disponibilidad pública, asignaciones, API,
  migraciones ni cambios de reglas de horario.

## Calidad requerida

- Prueba de componente y E2E mock con evidencia a 320/360/768/1280 y zoom 200,
  incluidos foco, conflicto y reflow sin scroll horizontal.
- Verificar teclado, axe, tipo, lint, build, `git diff --check` y CI antes de
  abrir PR; la diferencia heredada entre dock móvil y sidebar del atlas queda
  aislada en el shell.

## Pull request

- [PR #234](https://github.com/bcaceres19/barberia/pull/234), pendiente de
  checks antes del squash merge autorizado.
