---
prompt_id: "PROMPT-HU-096-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-096"
related_hu: ["HU-061", "HU-095", "HU-097"]
issue: "pending"
issue_url: null
suggested_issue_title: "feat(public-booking): implementar HU-096 datos del cliente"
branch: null
pr: null
pr_url: null
depends_on: ["HU-095 integrada", "DP-PUB-04 resuelta mediante DEC-*", "Issue real con CA-096-01 a CA-096-06"]
rules: ["RN-RES-01", "RN-RES-02", "RN-RES-03", "RN-DAT-01", "RN-DAT-02", "RN-TEN-01"]
decisions: ["DEC-016", "DEC-022", "DEC-024", "DEC-045", "DEC-046", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-096-01", "CA-096-02", "CA-096-03", "CA-096-04", "CA-096-05", "CA-096-06"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/booking", "apps/web/src/modules/agenda"]
created_at: "2026-09-10"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Implementar HU-096: datos del cliente y persona atendida

## Instrucción para el agente

No ejecutes hasta resolver `DP-PUB-04`. Implementa solo captura, validación y semántica de identidad; HU-097 persiste la cita.

## Objetivo

Capturar los datos mínimos aprobados y resolver correctamente quién reserva y quién será atendido.

## Preflight obligatorio

1. Confirma decisión propagada, HU-095, issue/rama y fuentes completas.
2. Ejecuta Graphify sobre booking/customer, validadores y formulario manual reutilizable.
3. Revisa privacidad y evita conservar PII en logs, URL o fixtures.

## Alcance incluido

- Request/modelo cerrado: nombre, teléfono, correo, nota opcional y persona atendida condicional.
- Normalización/validación server-side y UI equivalente.
- Puerto/política de identidad según `DP-PUB-04`, sin ejecutar todavía la cita si la división vigente así lo exige.

## Fuera de alcance

- Cuenta/portal, datos sensibles, contacto del atendido, cita o notificación.

## Estado existente que debe conservarse

- Reglas de customer de HU-061, unicidad tenant-aware y snapshots de booking.

## Trabajo requerido

1. Define schemas direccionales y límites aprobados; rechaza campos desconocidos.
2. Implementa normalización y política de coincidencias sin unión ambigua.
3. Construye formulario condicional accesible, autofill y conservación de datos ante error.

## Pruebas y evidencia

- Tabla para sí/tercero, campos, normalización y conflictos.
- PostgreSQL real con dos tenants; contrato, componente y E2E; inspección de logs.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, privacidad/matriz, módulos, HU y prompt/índice.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-096-01`–`CA-096-06`.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-096 datos del cliente`.
