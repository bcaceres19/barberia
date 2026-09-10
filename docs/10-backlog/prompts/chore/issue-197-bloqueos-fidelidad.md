---
prompt_id: "PROMPT-CHORE-197-BLOQUEOS-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-042"
related_hu: ["HU-009", "HU-040", "HU-041"]
issue: 197
issue_url: "https://github.com/bcaceres19/barberia/issues/197"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #196 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-BLQ-01", "RN-BLQ-02", "RN-BLQ-03", "RN-BLQ-04", "RN-TEN-01"]
decisions: ["DEC-008", "DEC-009", "DEC-020", "DEC-070", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-042-01", "CA-042-02", "CA-042-03", "CA-042-04", "CA-042-05", "CA-042-06", "CA-042-07", "CA-042-08"]
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
  - "docs/10-backlog/prompts/hu/hu-042-bloqueos-agenda.md"
  - "apps/web/src/modules/schedules"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Bloqueos al atlas NAVA

Rediseña `/panel/bloqueos` según `08-bloqueos-puntuales-y-series.png`, issue [#197](https://github.com/bcaceres19/barberia/issues/197). Conserva bloqueos puntuales y series realmente conectados. No presentes edición, impacto sobre citas o acciones que el frontend/contrato vigente no soporte.

Espera #196; luego crea `chore/197-bloqueos-fidelidad` desde `main` actualizada.

Lee fuentes y estado de #98/#100; mide referencia y captura tipos, puntual, semanal, lista explícita/excepciones/edición solo si ya están implementadas, retiro lógico, carga, vacío, validación y errores. Reproduce ficha, jerarquía, líneas, acciones y estados. Entrega referencia/baseline/final/lado a lado/overlay o diff y 320/360/768/1280, zoom, teclado, foco, axe, consola y Network.

No cambies backend, OpenAPI, migraciones, reglas de impacto ni implementes pendientes funcionales de #100 dentro de este rediseño. Ejecuta formato, lint, tipos, componentes, E2E afectadas y build. Commit/PR: `chore(web): reproduce Bloqueos según el atlas NAVA`; no mezcles #198.
