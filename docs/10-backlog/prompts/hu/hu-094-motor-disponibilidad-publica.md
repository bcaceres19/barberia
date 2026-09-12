---
prompt_id: "PROMPT-HU-094-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-094"
related_hu: ["HU-040", "HU-041", "HU-042", "HU-060", "HU-093"]
issue: 251
issue_url: "https://github.com/bcaceres19/barberia/issues/251"
suggested_issue_title: "feat(availability): implementar HU-094 disponibilidad pública real"
branch: "feat/251-hu094-disponibilidad-publica"
pr: null
pr_url: null
depends_on: ["HU-091 a HU-093 integradas (satisfecho)", "DP-PUB-03 resuelta mediante DEC-084 (satisfecho, 2026-09-11)", "Issue real con CA-094-01 a CA-094-06 (satisfecho, issue #251)"]
rules: ["RN-DIS-01", "RN-DIS-02", "RN-DIS-03", "RN-DIS-04", "RN-DIS-05", "RN-DIS-06", "RN-DIS-07", "RN-BLQ-01", "RN-BLQ-02", "RN-BLQ-03", "RN-BLQ-04", "RN-CON-01", "RN-CON-03", "RN-CAN-04", "RN-TEN-01"]
decisions: ["DEC-002", "DEC-005", "DEC-006", "DEC-007", "DEC-008", "DEC-009", "DEC-012", "DEC-018", "DEC-019", "DEC-020", "DEC-024", "DEC-070", "DEC-073", "DEC-076", "DEC-084"]
acceptance_criteria: ["CA-094-01", "CA-094-02", "CA-094-03", "CA-094-04", "CA-094-05", "CA-094-06"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/02-requisitos/estados-citas.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/schedule", "apps/api/internal/modules/booking"]
created_at: "2026-09-10"
updated_at: "2026-09-11"
supersedes: null
superseded_by: null
---

# Implementar HU-094: motor de disponibilidad pública

## Instrucción para el agente

Implementa el cálculo de `HU-094`; no construyas UI de calendario ni creación de citas. `DP-PUB-03` debe estar resuelta.

## Objetivo

Proyectar inicios válidos donde el servicio completo cabe al aplicar todos los factores confirmados.

## Preflight obligatorio

1. Confirma dependencias, decisión, issue y rama `feat/<issue>-hu094-disponibilidad-publica`.
2. Ejecuta Graphify sobre schedule, booking, shops, catalog y rangos PostgreSQL.
3. Lee fuentes y contratos reales; registra cualquier semántica faltante.

## Alcance incluido

- Dominio puro de intervalos/fechas/zona y adaptadores de jornada, bloqueos y citas.
- Operación pública de lectura sin efectos; límites de costo/rango documentados.
- Rejilla, duración, anticipación, ventana, `[inicio, fin)` y cruces de medianoche.

## Fuera de alcance

- Hold, caché distribuida, UI, confirmación, alternativas o reglas predictivas.

## Estado existente que debe conservarse

- B2 proyecta jornada/bloqueos y B3 protege cruces; no dupliques sus reglas ni relajes exclusión.
- Dominio/servicios Go permanecen independientes de Chi y PostgreSQL.

## Trabajo requerido

1. Define una API de dominio determinista con reloj/zoneinfo inyectables.
2. Une restricciones solapadas y genera slots según la decisión de rejilla.
3. Consulta datos tenant-aware con índices justificados y cancellation de contexto.
4. Expón contrato público mínimo y orden estable.

## Pruebas y evidencia

- Tabla exhaustiva y PostgreSQL real con dos tenants, once factores, nocturnidad y fronteras.
- `EXPLAIN (ANALYZE, BUFFERS)` con dataset reproducible; pruebas sin `sleep` ni hora real.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, módulos, matriz, HU, diccionario si cambia y prompt/índice.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle && pnpm run db:atlas:validate
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-094-01`–`CA-094-06`, incluido plan de consulta.

## Git y PR

- Commit/PR: `feat(availability): implementa HU-094 disponibilidad pública`.
