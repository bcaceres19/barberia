---
prompt_id: "PROMPT-CHORE-192-BARBEROS-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-021"
related_hu: ["HU-009", "HU-012"]
issue: 192
issue_url: "https://github.com/bcaceres19/barberia/issues/192"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #191 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-047", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-021-01", "CA-021-02", "CA-021-03", "CA-021-04", "CA-021-05", "CA-021-06", "CA-021-07", "CA-021-08"]
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
  - "docs/10-backlog/prompts/hu/hu-021-registro-listado-barberos.md"
  - "apps/web/src/modules/staff"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Barberos al atlas NAVA

Rediseña `/panel/barberos` según la mitad de barberos de `06-barberos-y-barberia.png`, issue [#192](https://github.com/bcaceres19/barberia/issues/192). Conserva lista, alta y renombrado de `HU-021`; no inventes fotografía, rol, orden, vínculo automático, desactivación ni eliminación (`DEC-047`).

Espera la integración de #191. Luego crea `chore/192-barberos-fidelidad` desde `main` actualizada.

## Trabajo y puertas

- Lee todos los `source_docs`, mide la referencia original y captura baseline de carga, vacío, lista, alta, edición, validación y errores.
- Reproduce composición, jerarquía de filas, monograma permitido, tipografía, color, espaciado, acciones y diálogos con componentes compartidos.
- Mantén nombres accesibles, teclado, foco, objetivos táctiles y aislamiento visible sin exponer IDs.
- Entrega referencia/baseline/final/lado a lado/overlay o diff y 320/360/768/1280, zoom 200 %, axe, consola y Network.

No cambies backend, OpenAPI, cliente tipado, migraciones, permisos ni operaciones. Ejecuta formato, lint, tipos, pruebas de componente, E2E afectadas y build. Commit/PR: `chore(web): reproduce Barberos según el atlas NAVA`; no mezcles #193 y usa `Closes #192` solo con todas las puertas.
