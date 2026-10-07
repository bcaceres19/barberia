---
prompt_id: "PROMPT-ORCH-301-UI-LUNA-v1"
version: "1.0"
kind: "orchestration"
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
rules: ["RN-TEN-01", "RN-DAT-02", "RN-IDE-01"]
decisions: ["DEC-050", "DEC-106", "DEC-116"]
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "docs/07-calidad/pruebas-ui/README.md"
  - "docs/07-calidad/pruebas-ui/01-entorno-cuentas.md"
  - "docs/07-calidad/pruebas-ui/02-protocolo-agentes.md"
  - "docs/07-calidad/pruebas-ui/03-inventario.md"
created_at: "2026-10-07"
updated_at: "2026-10-07"
supersedes: null
superseded_by: null
---

# Probar toda la interfaz de NAVA

## Instrucción y preflight

Coordinar una única campaña #301, agentes GPT-6 Luna / medium, máximo dos simultáneos. Activar ui-app-testing. No fixes ni terceros reales. Confirmar issue vigente/#300 listo, SHA, Atlas, API/frontend/Chromium y smoke autenticado. Preservar trabajo ajeno, sin main ni Graphify para navegar.

## Trabajo y criterios

Crear registro de campaña y salida privada. Serializar access/recovery; cola por cada ficha; finalizar journeys/aislamiento. Cuenta/tenant/contexto/salida exclusivos. Mensaje mínimo por agente: prompt guardado y parámetros no secretos. Son particiones de este único informe; no implementaciones independientes.

Coordinador actualiza metadatos/índice a in_progress antes de cada encargo. Campaña posterior tiene issue y versiones nuevas; no reescribir cuerpos ejecutados ni dejar ready con issue cerrado.

Todos los casos/rutas actuales reciben estado; UI-T/UI-J, responsive/teclado/zoom, sesión/persistencia reales y evidencia. BLOCKED/NOT_RUN no pasan. No cambiar código/esquema/proveedor.

## Encargos persistentes

- [PROMPT-TEST-301-UI-ACCESS-v1](../test/issue-301-ui-access.md)
- [PROMPT-TEST-301-UI-RECOVERY-v1](../test/issue-301-ui-recovery.md)
- [PROMPT-TEST-301-UI-SHOP-v1](../test/issue-301-ui-shop.md)
- [PROMPT-TEST-301-UI-POLICY-v1](../test/issue-301-ui-policy.md)
- [PROMPT-TEST-301-UI-STAFF-v1](../test/issue-301-ui-staff.md)
- [PROMPT-TEST-301-UI-CATALOG-v1](../test/issue-301-ui-catalog.md)
- [PROMPT-TEST-301-UI-ASSIGNMENTS-v1](../test/issue-301-ui-assignments.md)
- [PROMPT-TEST-301-UI-SCHEDULES-v1](../test/issue-301-ui-schedules.md)
- [PROMPT-TEST-301-UI-AGENDA-v1](../test/issue-301-ui-agenda.md)
- [PROMPT-TEST-301-UI-MANUAL-v1](../test/issue-301-ui-manual.md)
- [PROMPT-TEST-301-UI-DETAIL-v1](../test/issue-301-ui-detail.md)
- [PROMPT-TEST-301-UI-PUBLIC-ENTRY-v1](../test/issue-301-ui-public-entry.md)
- [PROMPT-TEST-301-UI-PUBLIC-CATALOG-v1](../test/issue-301-ui-public-catalog.md)
- [PROMPT-TEST-301-UI-PUBLIC-BARBER-v1](../test/issue-301-ui-public-barber.md)
- [PROMPT-TEST-301-UI-PUBLIC-SLOTS-v1](../test/issue-301-ui-public-slots.md)
- [PROMPT-TEST-301-UI-PUBLIC-CONFIRM-v1](../test/issue-301-ui-public-confirm.md)
- [PROMPT-TEST-301-UI-CUSTOMER-v1](../test/issue-301-ui-customer.md)
- [PROMPT-TEST-301-UI-JOURNEYS-v1](../test/issue-301-ui-journeys.md)

## Entrega y Git

Consolidar 06-informe: modelo/esfuerzo reales, SHA, conteos, evidencia sanitizada y pendientes. Crudos ignorados, copia pública revisada. Informe de campaña puede versionarse en test/301-pruebas-ui desde main actualizada; fixes usan otros issues/ramas/PR. No push directo ni merge. executed acredita ejecución informada, no ausencia de bugs. Cerrar #301 solo con sus criterios cumplidos.
