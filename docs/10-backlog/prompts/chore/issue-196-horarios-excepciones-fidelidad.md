---
prompt_id: "PROMPT-CHORE-196-HORARIOS-EXCEPCIONES-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-040"
related_hu: ["HU-009", "HU-041"]
issue: 196
issue_url: "https://github.com/bcaceres19/barberia/issues/196"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #195 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-05", "RN-DIS-07", "RN-TEN-01"]
decisions: ["DEC-007", "DEC-020", "DEC-070", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-040-01", "CA-040-02", "CA-040-03", "CA-040-04", "CA-040-05", "CA-040-06", "CA-040-07", "CA-040-08", "CA-041-08"]
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
  - "docs/10-backlog/prompts/hu/hu-040-horario-laboral.md"
  - "docs/10-backlog/prompts/hu/hu-041-excepciones-festivos.md"
  - "apps/web/src/modules/schedules"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Horarios y excepciones al atlas NAVA

Rediseña `/panel/horarios` según `07-horarios-y-excepciones.png`, issue [#196](https://github.com/bcaceres19/barberia/issues/196). Conserva jornada semanal, segmentos, días cerrados, excepciones y festivos existentes, zonas IANA e intervalos `[inicio, fin)`.

Espera #195; luego crea `chore/196-horarios-excepciones-fidelidad` desde `main` actualizada.

Lee fuentes, mide referencia y captura selector de barbero, semana, alta/edición/retiro, solape, excepción, festivo, carga, vacío y errores. Reproduce jerarquía de filas, cifras tabulares, controles, diálogos, alertas y densidad. Entrega comparación visual más 320/360/768/1280, zoom, teclado, foco, axe, consola y Network.

No cambies backend, contrato, migraciones, recurrencia, zona, cruces de medianoche ni agregues plantillas/calendarios no existentes. Ejecuta formato, lint, tipos, componentes, E2E afectadas y build. Commit/PR: `chore(web): reproduce Horarios según el atlas NAVA`; no mezcles #197.
