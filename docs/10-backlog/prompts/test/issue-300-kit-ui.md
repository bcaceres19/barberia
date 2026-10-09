---
prompt_id: "PROMPT-TEST-300-KIT-UI-v1"
version: "1.0"
kind: "test"
status: "executed"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: 300
issue_url: "https://github.com/bcaceres19/barberia/issues/300"
branch: "test/300-exploracion-ui-luna"
pr: null
pr_url: null
depends_on: []
rules: ["RN-TEN-01", "RN-DAT-02", "RN-IDE-01"]
decisions: ["DEC-050", "DEC-106", "DEC-120"]
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "docs/07-calidad/pruebas-ui/README.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
created_at: "2026-10-07"
updated_at: "2026-10-07"
supersedes: null
superseded_by: null
---

# Preparar el kit de pruebas UI

Solicitud del propietario: documentación, usuarios ficticios y skill automático para probar por pantalla con Luna medium.

Alcance #300: skill/adaptador, fichas y prompts, aprovisionamiento privado y smoke de preparación. Fuera: ejecutar toda la campaña #301 y fixes de producto.

Criterios: rutas implementadas catalogadas, fuentes reales, usuarios/tenants creados, login UI y dos tenants verificados, guardado/recarga, rutas privadas/públicas y sesión del smoke; guards de provisión, validación estricta y revisión independiente con Luna. No credenciales/códigos/enlaces en Git.

Validar node --test tools/qa/*.test.mjs, bash -n tools/qa/start-local.sh, quick_validate.py, tools/ai/validate-agent-system.sh --strict y node tools/qa/verify-ui-ready.mjs con PostgreSQL real. Smoke no acredita toda la campaña.

Rama test/300-exploracion-ui-luna, cambios ajenos conservados, Conventional Commits. PR si se solicita publicación. Actualizar estado/índice executed solo tras verificar preparación. Campaña #301 permanece ready pendiente. Versión nueva/supersedes para cuerpo ya ejecutado modificado.
