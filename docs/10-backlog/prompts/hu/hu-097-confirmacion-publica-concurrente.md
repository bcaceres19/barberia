---
prompt_id: "PROMPT-HU-097-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-097"
related_hu: ["HU-004", "HU-060", "HU-061", "HU-090", "HU-091", "HU-092", "HU-093", "HU-094", "HU-095", "HU-096", "HU-098"]
issue: 268
issue_url: "https://github.com/bcaceres19/barberia/issues/268"
suggested_issue_title: "feat(public-booking): implementar HU-097 confirmación concurrente"
branch: "feat/268-hu-097-confirmacion-publica-concurrente"
pr: 270
pr_url: "https://github.com/bcaceres19/barberia/pull/270"
depends_on: ["HU-090 a HU-096 integradas (satisfecho)", "DP-PUB-05 resuelta mediante DEC-089 (satisfecho)", "DP-PUB-06 resuelta mediante DEC-090 (satisfecho)", "CT-011 resuelta mediante DEC-091 (satisfecho)", "Issue real con CA-097-01 a CA-097-07 (satisfecho, issue #268)"]
rules: ["RN-RES-01", "RN-RES-02", "RN-RES-03", "RN-DIS-03", "RN-DIS-04", "RN-DIS-05", "RN-DIS-06", "RN-DIS-07", "RN-CON-01", "RN-CON-02", "RN-CON-03", "RN-CON-04", "RN-CON-05", "RN-CON-06", "RN-CNF-01", "RN-HIS-01", "RN-IDE-01", "RN-TEN-01"]
decisions: ["DEC-005", "DEC-006", "DEC-007", "DEC-008", "DEC-012", "DEC-013", "DEC-016", "DEC-019", "DEC-020", "DEC-022", "DEC-024", "DEC-041", "DEC-043", "DEC-045", "DEC-046", "DEC-073", "DEC-085", "DEC-089", "DEC-090", "DEC-091"]
acceptance_criteria: ["CA-097-01", "CA-097-02", "CA-097-03", "CA-097-04", "CA-097-05", "CA-097-06", "CA-097-07"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/00-control/contradicciones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/02-requisitos/estados-citas.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/05-backend/migraciones-atlas.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/booking", "apps/api/internal/modules/publicbooking", "apps/api/internal/modules/schedule", "apps/api/internal/platform/idempotency"]
created_at: "2026-09-10"
updated_at: "2026-09-15"
supersedes: null
superseded_by: null
---

# Implementar HU-097: confirmación pública concurrente

## Instrucción para el agente

Implementa solo T1 pública y su recuperación de conflicto. `DP-PUB-05` (`DEC-089`), `DP-PUB-06` (`DEC-090`) y `CT-011` (`DEC-091`) ya están resueltas: aplícalas literalmente, sin reinterpretarlas.

## Objetivo

Crear cero o una cita pública completa por intención, con exactamente un ganador concurrente y alternativas seguras para perdedores.

## Preflight obligatorio

1. Verifica todas las HU previas integradas, decisiones propagadas, issue real y rama.
2. Ejecuta Graphify sobre booking, schedule, catalog, idempotencia, token e historial.
3. Revisa la migración aplicada de HU-060: no la edites; todo cambio es roll-forward.

## Alcance incluido

- Comando público contract-first con `Idempotency-Key` y body cerrado.
- Revalidación completa en servidor y persistencia atómica aprobada.
- Exclusión PostgreSQL, carrera con bloqueo, alternativas (hasta 3, mismo barbero, `DEC-090`) y conservación de formulario.
- Resumen/confirmación UI y estados de conflicto/red.
- Emisión del token de acceso (`appointment_access_token`, 256 bits, hash SHA-256, vigencia 90 días, `DEC-089`) y envío síncrono simple del correo de confirmación con el proveedor ya integrado (Resend, `DEC-091`), dentro de o inmediatamente después de la transacción.

## Fuera de alcance

- Hold, pago, aprobación manual, lista de espera o la maquinaria completa de notificaciones de B5 (cola, reintentos, múltiples proveedores, recordatorios) que `DEC-091` reserva para después.

## Estado existente que debe conservarse

- Núcleo booking, historial append-only, snapshots, RLS, exclusión e idempotencia de B3.

## Trabajo requerido

1. Diseña OpenAPI y transacción según decisiones de token/confirmación.
2. Deriva tenant, snapshots, fin y estado; revalida asignación/jornada/bloqueo/límites.
3. Traduce `23P01` a conflicto público y calcula alternativas según `DEC-090`.
4. Coordina la carrera bloqueo/reserva sin borrar la cita; registra afectación al barbero.
5. Implementa UI contra códigos estables, nunca texto de `detail`.

## Pruebas y evidencia

- PostgreSQL real con dos tenants, N conexiones y barreras sin `sleep`: exactamente un éxito.
- Rollback de cada fallo, idempotencia, bloqueo en ambos órdenes, contrato, componente y E2E.
- Evidencia 320/360/768/1280, teclado, zoom 200 % y axe-core.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, Atlas/datos si aplica, matriz, estados, módulos, HU y prompt/índice.

## Verificación final

```text
pnpm run openapi:check-config && pnpm run openapi:lint && pnpm run openapi:bundle && pnpm run db:atlas:validate
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-097-01`–`CA-097-07` y evidencia de concurrencia.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-097 confirmación pública`.
