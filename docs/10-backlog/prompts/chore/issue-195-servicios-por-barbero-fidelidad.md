---
prompt_id: "PROMPT-CHORE-195-SERVICIOS-POR-BARBERO-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-023"
related_hu: ["HU-009", "HU-021", "HU-022"]
issue: 195
issue_url: "https://github.com/bcaceres19/barberia/issues/195"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #194 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-SER-02", "RN-TEN-01", "RN-IDE-01"]
decisions: ["DEC-068", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-023-01", "CA-023-02", "CA-023-03", "CA-023-04", "CA-023-05", "CA-023-06", "CA-023-07", "CA-023-08"]
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
  - "docs/10-backlog/prompts/hu/hu-023-asignacion-servicios-barberos.md"
  - "apps/web/src/modules/barberServices"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Servicios por barbero al atlas NAVA

Rediseña `/panel/servicios-por-barbero` según el panel de asignación de `05-servicios-y-asignacion.png`, issue [#195](https://github.com/bcaceres19/barberia/issues/195). Conserva selección, asignación, retiro y rechazo de la última asignación de un servicio activo (`DEC-068`).

Espera #194; luego crea `chore/195-servicios-por-barbero-fidelidad` desde `main` actualizada.

Mide referencia y app; cubre carga, vacío, selección, asignado/no asignado, guardado, error y rechazo de última asignación. Reproduce composición, jerarquía, controles, áreas táctiles, alertas y densidad. No uses color como única señal. Entrega referencia/baseline/final/lado a lado/overlay o diff, 320/360/768/1280, zoom 200 %, teclado, foco, axe, consola y Network.

No cambies backend, OpenAPI, cliente, migraciones, reglas ni agregues orden, porcentaje, precio por barbero o eliminación. Ejecuta formato, lint, tipos, pruebas de componente, E2E afectada y build. Commit/PR: `chore(web): reproduce asignaciones según el atlas NAVA`; no mezcles #196.
