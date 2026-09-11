---
prompt_id: "PROMPT-ORCH-B4-RESERVA-PUBLICA-v1"
version: "1.0"
kind: "orchestration"
status: "draft"
target_agents: ["codex", "claude", "human"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: ["HU-090", "HU-091", "HU-092", "HU-093", "HU-094", "HU-095", "HU-096", "HU-097", "HU-098", "HU-099"]
issue: "pending"
issue_url: null
suggested_issue_title: "chore(orchestration): coordinar bloque B4 de reserva pública"
branch: null
pr: null
pr_url: null
depends_on: ["B3 con criterio de salida completo", "DP-PUB-01 a DP-PUB-06 resueltas", "CT-011 resuelta", "Un issue real por cada HU"]
rules: ["RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-04", "RN-DIS-05", "RN-DIS-06", "RN-DIS-07", "RN-CON-01", "RN-CON-02", "RN-CON-03", "RN-CON-04", "RN-CON-05", "RN-CON-06", "RN-CAN-01", "RN-CAN-02", "RN-CAN-04", "RN-CNF-01", "RN-CNF-02", "RN-DAT-01", "RN-DAT-02", "RN-DAT-03", "RN-IDE-01", "RN-TEN-01"]
decisions: ["DEC-005", "DEC-006", "DEC-007", "DEC-008", "DEC-009", "DEC-010", "DEC-011", "DEC-012", "DEC-013", "DEC-014", "DEC-016", "DEC-017", "DEC-018", "DEC-019", "DEC-020", "DEC-022", "DEC-024", "DEC-041", "DEC-043", "DEC-045", "DEC-046", "DEC-049", "DEC-067", "DEC-068", "DEC-069", "DEC-070", "DEC-073", "DEC-076", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-090-01", "CA-090-02", "CA-090-03", "CA-090-04", "CA-090-05", "CA-091-01", "CA-091-02", "CA-091-03", "CA-091-04", "CA-091-05", "CA-092-01", "CA-092-02", "CA-092-03", "CA-092-04", "CA-092-05", "CA-093-01", "CA-093-02", "CA-093-03", "CA-093-04", "CA-093-05", "CA-094-01", "CA-094-02", "CA-094-03", "CA-094-04", "CA-094-05", "CA-094-06", "CA-095-01", "CA-095-02", "CA-095-03", "CA-095-04", "CA-095-05", "CA-096-01", "CA-096-02", "CA-096-03", "CA-096-04", "CA-096-05", "CA-096-06", "CA-097-01", "CA-097-02", "CA-097-03", "CA-097-04", "CA-097-05", "CA-097-06", "CA-097-07", "CA-098-01", "CA-098-02", "CA-098-03", "CA-098-04", "CA-098-05", "CA-098-06", "CA-099-01", "CA-099-02", "CA-099-03", "CA-099-04", "CA-099-05", "CA-099-06"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/00-control/contradicciones.md", "docs/00-control/matriz-trazabilidad.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/02-requisitos/estados-citas.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/03-desarrollo/flujo-git-github.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "docs/10-backlog/plan-bloques.md"]
created_at: "2026-09-10"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Orquestar B4: reserva pública y disponibilidad

## Instrucción para el agente

Coordina las diez entregas sin implementarlas en una sola rama. Este prompt permanece `draft` hasta cerrar B3, resolver `DP-PUB-01`–`DP-PUB-06`/`CT-011` y crear un issue verificable por HU.

## Objetivo

Entregar B4 en diez PR independientes, integrados en el orden aprobado, con el criterio de salida del bloque demostrado de punta a punta.

## Preflight obligatorio

1. Verifica que B3 cumple su criterio de salida con evidencia, incluida `HU-068`.
2. Confirma que cada duda/contradicción del lote tiene `DEC-*` propagada.
3. Crea o enlaza un issue por HU y cambia a `ready` solo el prompt cuya dependencia inmediata esté satisfecha.
4. Relee las fuentes antes de cada entrega; una decisión material nueva versiona el prompt afectado.

## Secuencia y gates

1. `HU-090` entrada pública.
2. `HU-091` catálogo público.
3. `HU-092` selección de barbero.
4. `HU-093` configuración de políticas.
5. `HU-094` motor de disponibilidad.
6. `HU-095` fechas y horarios.
7. `HU-096` datos del cliente/persona atendida.
8. `HU-097` confirmación concurrente.
9. `HU-098` acceso del cliente.
10. `HU-099` cancelación pública.

No abras la rama siguiente hasta que el PR anterior esté integrado en `main` con checks verdes y el prompt/índice refleje el estado real.

## Alcance incluido

- Verificar dependencias, issue, estado de prompt, CI, integración y trazabilidad entre entregas.
- Ejecutar al final el recorrido público completo, concurrencia de N clientes, cancelación y reoferta.

## Fuera de alcance

- Combinar HU en un issue/PR, adelantar B5, resolver dudas por inferencia o saltar checks.
- Hacer merge de un PR fallido o declarar B4 cerrado con evidencia parcial.

## Pruebas y evidencia de cierre

- Todos los CA de `HU-090`–`HU-099` enlazados a pruebas/evidencia.
- N confirmaciones concurrentes: una cita; perdedores con alternativas y datos conservados.
- PostgreSQL real con dos tenants; E2E completo y evidencia 320/360/768/1280, teclado, zoom 200 % y axe-core.
- Enlace customer consulta y cancela según política; la franja futura vuelve a disponibilidad.

## Documentación y trazabilidad

Tras cada PR actualiza historia, plan, matriz, historial y catálogo. Al cierre de B4 registra SHA/PR y no cambies B5 a iniciado hasta demostrar el criterio de salida.

## Verificación final

Ejecuta el conjunto completo vigente de OpenAPI, Go `-race`, PostgreSQL/Atlas, Vue, E2E, `git diff --check` y Graphify. Entrega una matriz `HU | Issue | PR | CA | Checks | Evidencia | Estado`.

## Git y PR

- Diez ramas `feat/<issue>-<descripcion>` y diez PR; nunca una rama de orquestación con código de negocio.
- Conventional Commits, squash, sin push directo ni force push a `main`.
