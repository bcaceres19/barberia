---
prompt_id: "PROMPT-TEST-301-UI-PUBLIC-BARBER-v1"
version: "1.0"
kind: "test"
status: "ready"
target_agents: ["codex"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: ["HU-092"]
issue: 301
issue_url: "https://github.com/bcaceres19/barberia/issues/301"
branch: null
pr: null
pr_url: null
depends_on: ["#300"]
rules: ["RN-TEN-01", "RN-DAT-02", "RN-IDE-01"]
decisions: ["DEC-104", "DEC-111", "DEC-120"]
acceptance_criteria: ["CA-092-01"]
source_docs:
  - "AGENTS.md"
  - "docs/07-calidad/pruebas-ui/02-protocolo-agentes.md"
  - "docs/07-calidad/pruebas-ui/04-transversal.md"
  - "docs/07-calidad/pruebas-ui/06-informe.md"
  - "docs/07-calidad/pruebas-ui/pantallas/public-barber.md"
created_at: "2026-10-07"
updated_at: "2026-10-07"
supersedes: null
superseded_by: null
---

# Explorar Selección pública de profesional

## Instrucción y preflight

GPT-6 Luna / medium. Ejecutar exclusivamente scope **public-barber**, ruta /reservar/:slug/servicios/:serviceId/barbero, una partición de campaña #301. Leer source_docs y los fragmentos RN/DEC/HU de la ficha. Coordinador proporciona campaña/SHA/URL/salida; cuenta scope=public-barber se lee privadamente de apps/web/.auth/nava-qa/current.json, sin imprimirla.

Confirmar issue #301 vigente, prompt in_progress asignado por el coordinador y dependencia #300 preparada. Falta entorno/cuenta/navegador = BLOCKED. Conservar trabajo ajeno; no cambiar rama ni arrancar otro servidor. Coordinador actualiza estado/índice antes de delegar. No Graphify para navegar.

## Alcance y exclusiones

Todos los UI-PUBLIC-BARBER de la ficha, UI-T aplicables, cuatro viewports, teclado, zoom y estados. Datos propios; éxito/guardado contra API/PostgreSQL reales. Fuera: fixes, cambios de producto/pruebas/migraciones, terceros reales, recursos de otras tareas.

## Criterios y entrega

Cada caso tiene estado y evidencia; recuperación del guardado tras F5. FAIL con pasos/esperado/observado. Apoyo temporal/SQL faltante BLOCKED; no simular y afirmar regla real. Tanda de 12 casos/20 min; restantes NOT_RUN se reencolan.

Escribir solo apps/web/test-results/ui-qa/<campaña>/public-barber/report.md y results.json según 06-informe. Capturas/trazas privadas y versión sanitizada al entregar. Resumen <=400 palabras con conteos/pendientes/enlaces. No afirmar ausencia de bugs.

## Git y relevo

Sin commits/branch/PR de producto del subagente: informe es partición de #301. Corrección futura es otra entrega con issue/branch/PR. Solo coordinador actualiza metadatos/índice. Cambio material ejecutado crea versión/supersedes.
