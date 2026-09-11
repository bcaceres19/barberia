---
prompt_id: "PROMPT-HU-099-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-099"
related_hu: ["HU-066", "HU-093", "HU-094", "HU-098"]
issue: "pending"
issue_url: null
suggested_issue_title: "feat(customer-access): implementar HU-099 cancelación pública"
branch: null
pr: null
pr_url: null
depends_on: ["HU-093, HU-094 y HU-098 integradas", "DP-PUB-02 y DP-PUB-05 resueltas", "Issue real con CA-099-01 a CA-099-06"]
rules: ["RN-CAN-01", "RN-CAN-02", "RN-CAN-03", "RN-CAN-04", "RN-CIT-03", "RN-HIS-01", "RN-HIS-02", "RN-DIS-04", "RN-IDE-01", "RN-TEN-01"]
decisions: ["DEC-010", "DEC-011", "DEC-012", "DEC-014", "DEC-016", "DEC-017", "DEC-018", "DEC-022", "DEC-024", "DEC-041", "DEC-043"]
acceptance_criteria: ["CA-099-01", "CA-099-02", "CA-099-03", "CA-099-04", "CA-099-05", "CA-099-06"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/02-requisitos/estados-citas.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/booking", "api/openapi/paths/customer-appointments.yaml"]
created_at: "2026-09-10"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Implementar HU-099: cancelación pública

## Instrucción para el agente

Implementa únicamente T5 customer. No mezcles T6, reprogramación, pagos ni efectos de B5.

## Objetivo

Cancelar una cita confirmada solo cuando la política vigente lo autoriza, con historial atómico y liberación inmediata.

## Preflight obligatorio

1. Confirma dependencias, decisiones, issue real y rama.
2. Ejecuta Graphify sobre T5/T6, políticas, token, exclusión, historial e idempotencia.
3. Relee máquina de estados y no agregues un `PATCH status` genérico.

## Alcance incluido

- Comando customer con token, `Idempotency-Key`, `If-Match` y reloj servidor.
- Política dentro/exacto/fuera del plazo y motivo condicional.
- Estado/evento/motivo atómicos, no-op exacto y liberación de exclusión.
- UI con consecuencia, rechazo/contacto y estados accesibles.

## Fuera de alcance

- T6, reapertura, reprogramación, cancelación masiva, penalidad/pago o notificaciones de B5.

## Estado existente que debe conservarse

- Núcleo booking, historial append-only, exclusión, idempotencia y patrones de comando T6/T8.

## Trabajo requerido

1. Diseña operación customer contract-first con errores distinguibles por `code`.
2. Evalúa política con reloj inyectado y versión vigente.
3. Persiste T5 e historial en una transacción tenant-aware; verifica disponibilidad posterior.
4. Implementa UI sin comparar `detail` y sin exponer token en logs.

## Pruebas y evidencia

- Tabla antes/exacto/después × política × motivo; estados terminales y repetición.
- PostgreSQL real con dos tenants, rollback, liberación y carreras coordinadas.
- Contrato, componente/E2E cancelar→reconsultar y evidencia responsive/accesible.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, estados, módulos, matriz, HU y prompt/índice.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle && pnpm run db:atlas:validate
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-099-01`–`CA-099-06`.

## Git y PR

- Commit/PR: `feat(customer-access): implementa HU-099 cancelación pública`.
