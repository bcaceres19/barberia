---
prompt_id: "PROMPT-TEST-301-UI-JOURNEYS-v1"
version: "1.0"
kind: "test"
status: "ready"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: 301
issue_url: "https://github.com/bcaceres19/barberia/issues/301"
branch: null
pr: null
pr_url: null
depends_on: ["#300"]
rules: ["RN-CON-01", "RN-TEN-01", "RN-HIS-01", "RN-IDE-01"]
decisions: ["DEC-050", "DEC-114", "DEC-116"]
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "docs/07-calidad/pruebas-ui/02-protocolo-agentes.md"
  - "docs/07-calidad/pruebas-ui/04-transversal.md"
  - "docs/07-calidad/pruebas-ui/06-informe.md"
  - "docs/07-calidad/pruebas-ui/05-recorridos.md"
  - "docs/07-calidad/pruebas-ui/08-cobertura.md"
created_at: "2026-10-07"
updated_at: "2026-10-07"
supersedes: null
superseded_by: null
---

# Recorridos y aislamiento

GPT-6 Luna / medium. Al terminar pantallas, ejecutar en serie UI-J-01 a UI-J-12 y UI-T-16. Usar solo scopes journeys/isolation-a/isolation-b privados. Crear precondiciones por UI; no datos mutados por otros.

Confirmar #301/SHA/entorno; sin Graphify ni cambio de rama. UI-S se reutiliza solo mismo SHA/campaña, si no repetir con coordinador y scope access sin concurrencia.

Fuera: producto, terceros reales, cobertura inventada de worker/retención/Google Calendar. Cada caso tiene estado/evidencia, FAIL reproducible. BACKEND_ONLY/NOT_RUN explícitos. No imprimir cookies/tokens/credenciales.

Solo apps/web/test-results/ui-qa/<campaña>/journeys/report.md y results.json. Tanda 12 casos/20 min y reencolar pendientes. Coordinador actualiza estado/índice; executed no significa todos PASS. Sin rama/PR independiente de producto; fix posterior usa otro issue/rama/PR.
