---
prompt_id: "PROMPT-CHORE-191-DETALLE-HISTORIAL-REPROGRAMACION-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-064"
related_hu: ["HU-009", "HU-063", "HU-065"]
issue: 191
issue_url: "https://github.com/bcaceres19/barberia/issues/191"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
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
  - "docs/10-backlog/prompts/hu/hu-064-detalle-historial-turno.md"
  - "docs/10-backlog/prompts/hu/hu-065-reprogramacion-turno.md"
  - "apps/web/src/modules/agenda"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de detalle, historial y reprogramación al atlas NAVA

Rediseña únicamente `/panel/turnos/:appointmentId` y su diálogo de reprogramación según `04-detalle-historial-reprogramacion.png`, dentro del issue [#191](https://github.com/bcaceres19/barberia/issues/191). Conserva snapshots, privacidad, historial paginado, token opaco, estados y comando T2 existentes.

No ejecutes antes de integrar #190. Después crea `chore/191-detalle-historial-reprogramacion-fidelidad` desde `main` actualizada y registra estado/rama.

## Trabajo

1. Lee todos los `source_docs`; abre el PNG original, mide paneles, columnas, hechos, cronología, acciones y diálogo.
2. Captura baseline y estados carga/error/historial vacío o paginado, `confirmed`, terminales, diálogo, conflicto de agenda, versión obsoleta y éxito.
3. Reproduce geometría, jerarquía, densidad, tipografía, color, alineaciones y estados, manteniendo el regreso a fecha/barbero.
4. No muestres correo, teléfono, IDs internos ni valores técnicos fuera de las superficies autorizadas.
5. Compara referencia/baseline/final/lado a lado/overlay o diff y verifica 320/360/768/1280, zoom 200 %, teclado, foco, axe, consola y Network.

## Fuera de alcance

- Añadir cancelar, completar, `no_show`, corrección de estado, cambio de servicio/duración/persona o notificaciones.
- Cambiar backend, OpenAPI, migraciones, paginación, concurrencia o reglas de reprogramación.
- Sustituir snapshots por catálogo actual o crear botones decorativos.

Ejecuta formato, lint, tipos, unitarias/componentes, E2E afectadas y build. Entrega matrices de criterios y diferencias. Commit/PR: `chore(web): reproduce detalle de turno según el atlas NAVA`; `Closes #191` solo con evidencia completa y sin mezclar #192.
