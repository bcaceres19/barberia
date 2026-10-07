---
prompt_id: "PROMPT-REVIEW-300-UI-SKILL-v1"
version: "1.0"
kind: "review"
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
decisions: ["DEC-116"]
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - ".agents/skills/ui-app-testing/SKILL.md"
  - "docs/07-calidad/pruebas-ui/README.md"
created_at: "2026-10-07"
updated_at: "2026-10-07"
supersedes: null
superseded_by: null
---

# Evaluación independiente del flujo de pruebas

GPT-6 Luna / medium, contexto mínimo. Solicitud realista: «Prueba toda la app». Aplicar ui-app-testing en modo de evaluación sin iniciar campaña #301 ni modificar producto, datos o estado de prompts.

Usar skill, catálogo y herramientas del kit para demostrar en salida privada el preflight y plan de dos primeras tareas: documentos, cuentas por scope, modelo/esfuerzo, aislamiento, prompt guardado y evidencia. No credenciales. Comprobar un acceso sintético por navegador si está listo; escribir solo apps/web/test-results/ui-qa/skill-evaluation/report.md.

Revisar decisiones ante runtime sin Luna, issue cerrado, sesión vencida y prueba con mock de éxito. No considerar el ejercicio campaña completa. Reportar problemas concretos con ruta/impacto y pasos reproducibles, o ninguna observación con límites. Sin Graphify, fixes, PR ni acciones externas.
