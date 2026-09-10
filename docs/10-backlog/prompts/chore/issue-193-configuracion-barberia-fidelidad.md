---
prompt_id: "PROMPT-CHORE-193-CONFIGURACION-BARBERIA-FIDELIDAD-v1"
version: "1.0"
kind: "chore"
status: "blocked"
target_agents: ["claude", "codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-020"
related_hu: ["HU-009", "HU-012"]
issue: 193
issue_url: "https://github.com/bcaceres19/barberia/issues/193"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on: ["Issue #192 integrado en main con CI verde", "Issue maestro #184"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-007", "DEC-077", "DEC-079", "DEC-080"]
acceptance_criteria: ["CA-020-01", "CA-020-02", "CA-020-03", "CA-020-04", "CA-020-05", "CA-020-06", "CA-020-07", "CA-020-08"]
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
  - "docs/10-backlog/prompts/hu/hu-020-configuracion-barberia.md"
  - "apps/web/src/modules/settings"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Fidelidad de Configuración de barbería al atlas NAVA

Rediseña `/panel/barberia` según la mitad de configuración de `06-barberos-y-barberia.png`, issue [#193](https://github.com/bcaceres19/barberia/issues/193). Conserva únicamente nombre, zona IANA y contactos del contrato de `HU-020`, junto con guardado, validación y privacidad reales.

Espera #192; luego crea `chore/193-configuracion-barberia-fidelidad` desde `main` actualizada.

## Trabajo

- Lee todos los `source_docs`; mide referencia y baseline de carga, éxito, edición, guardado, validación, error y reintento.
- Reproduce panel, jerarquía editorial, agrupación de campos, estado de guardado, tipografía, color y espaciado.
- Conserva datos escritos ante fallos y evita incluir contactos en logs, URL o evidencia.
- Compara referencia/baseline/final/lado a lado/overlay o diff; verifica 320/360/768/1280, zoom 200 %, teclado, foco, axe, consola y Network.

No agregues logo, tema, canales OTP/notificación, políticas, horarios ni otra configuración futura. No cambies API, backend, migraciones o cliente generado. Ejecuta formato, lint, tipos, componentes, E2E y build. Commit/PR: `chore(web): reproduce configuración según el atlas NAVA`; no mezcles #194.
