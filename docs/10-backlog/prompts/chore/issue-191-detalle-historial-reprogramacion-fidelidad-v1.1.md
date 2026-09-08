---
prompt_id: "PROMPT-CHORE-191-DETALLE-HISTORIAL-REPROGRAMACION-FIDELIDAD-v1.1"
version: "1.1"
kind: "chore"
status: "in_progress"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-064"
related_hu: ["HU-009", "HU-063", "HU-065"]
issue: 191
issue_url: "https://github.com/bcaceres19/barberia/issues/191"
branch: "chore/191-detalle-historial-reprogramacion-fidelidad"
pr: 229
pr_url: "https://github.com/bcaceres19/barberia/pull/229"
depends_on: ["Issue #190 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-CIT-03", "RN-HIS-01", "RN-HIS-02", "RN-TEN-01"]
decisions: ["DEC-004", "DEC-014", "DEC-041", "DEC-076", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-064-01", "CA-064-02", "CA-064-03", "CA-064-04", "CA-064-05", "CA-064-06", "CA-064-07", "CA-064-08", "CA-065-08"]
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
  - "apps/web/src/modules/agenda"
created_at: "2026-09-04"
updated_at: "2026-09-08"
supersedes: "PROMPT-CHORE-191-DETALLE-HISTORIAL-REPROGRAMACION-FIDELIDAD-v1"
superseded_by: null
---

# Fidelidad de detalle, historial y reprogramación al atlas NAVA

Rediseña únicamente `/panel/turnos/:appointmentId` y su diálogo de reprogramación según `04-detalle-historial-reprogramacion.png`, dentro del issue [#191](https://github.com/bcaceres19/barberia/issues/191). Conserva privacidad, historial paginado, token opaco, estados y comando T2 existentes.

## Trabajo en curso

- Rama creada desde `main` con #190 integrado: `chore/191-detalle-historial-reprogramacion-fidelidad`.
- La pantalla reproduce la composición de identidad, hechos reglados, cronología y diálogo ancho/inferior. La foto del atlas se sustituye por monograma porque el contrato no entrega un activo fotográfico.
- La evidencia determinista usa datos ficticios solo dentro de Playwright; cubre detalle listo, conflicto de versión, 320/360/768/1280, aproximación de zoom 200 %, teclado, foco y movimiento reducido.

## Fuera de alcance

- Añadir cancelar, completar, `no_show`, corrección de estado, cambio de servicio/duración/persona o notificaciones.
- Cambiar backend, OpenAPI, migraciones, paginación, concurrencia o reglas de reprogramación.
- Sustituir snapshots por catálogo actual o crear botones decorativos.

El [PR #229](https://github.com/bcaceres19/barberia/pull/229) contiene formato, lint, tipos, 54 pruebas de componente, seis E2E visuales en Chromium, build y `git diff --check`; no mezcla #192.
