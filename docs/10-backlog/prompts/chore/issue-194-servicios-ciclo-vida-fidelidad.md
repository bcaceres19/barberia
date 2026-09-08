---
prompt_id: "PROMPT-CHORE-194-SERVICIOS-CICLO-VIDA-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-022"
related_hu: ["HU-009", "HU-024"]
issue: 194
issue_url: "https://github.com/bcaceres19/barberia/issues/194"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #193 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-SER-01", "RN-SER-02", "RN-SER-03", "RN-SER-04", "RN-TEN-01"]
decisions: ["DEC-003", "DEC-004", "DEC-067", "DEC-069", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-022-01", "CA-022-02", "CA-022-03", "CA-022-04", "CA-022-05", "CA-022-06", "CA-022-07", "CA-022-08", "CA-024-08"]
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md"
  - "docs/10-backlog/prompts/hu/hu-022-catalogo-servicios.md"
  - "docs/10-backlog/prompts/hu/hu-024-ciclo-vida-servicios.md"
  - "apps/web/src/modules/catalog"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Servicios y ciclo de vida al atlas NAVA

Rediseña `/panel/servicios` según el panel de catálogo de `05-servicios-y-asignacion.png`, issue [#194](https://github.com/bcaceres19/barberia/issues/194). Conserva COP, precio mayor que cero, duración, unicidad, alta/edición y desactivación/reactivación reales.

Espera #193; luego crea `chore/194-servicios-ciclo-vida-fidelidad` desde `main` actualizada.

## Trabajo

- Lee fuentes completas; mide la referencia y captura carga, vacío, lista, alta, edición, validación, conflicto, impacto de desactivación, confirmación y reactivación.
- Reproduce composición, filas, cifras tabulares, jerarquía, diálogos, alertas, acciones, color, tipografía y densidad sin iconografía cliché que sustituya nombres.
- Conserva `$`/COP sin concatenaciones erróneas y no prometa cancelación futura inexistente.
- Entrega comparación visual y evidencia 320/360/768/1280, zoom, teclado, foco, axe, consola y Network.

No cambies backend, contrato, migraciones, reglas de impacto o asignaciones. Ejecuta formato, lint, tipos, componentes, `servicios-ciclo-vida` E2E y build. Commit/PR: `chore(web): reproduce Servicios según el atlas NAVA`; no mezcles #195.
