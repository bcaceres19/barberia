---
prompt_id: "PROMPT-HU-098-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-098"
related_hu: ["HU-097", "HU-099"]
issue: "pending"
issue_url: null
suggested_issue_title: "feat(customer-access): implementar HU-098 acceso al turno"
branch: null
pr: null
pr_url: null
depends_on: ["HU-097 integrada", "DP-PUB-05 resuelta", "CT-011 resuelta", "Issue real con CA-098-01 a CA-098-06"]
rules: ["RN-CNF-01", "RN-CNF-02", "RN-DAT-01", "RN-DAT-02", "RN-DAT-03", "RN-TEN-01"]
decisions: ["DEC-016", "DEC-022", "DEC-024", "DEC-025", "DEC-049", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-098-01", "CA-098-02", "CA-098-03", "CA-098-04", "CA-098-05", "CA-098-06"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/00-control/contradicciones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/05-backend/migraciones-atlas.md", "docs/06-api/estandar-openapi.md", "api/openapi/paths/customer-appointments.yaml", "apps/api/internal/platform/httpserver"]
created_at: "2026-09-10"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Implementar HU-098: acceso del cliente a su turno

## Instrucción para el agente

Implementa solo confirmación visual y lectura por token. No agregues portal, búsqueda por contacto ni edición.

## Objetivo

Permitir que el cliente consulte una única cita mediante una credencial larga, segura, revocable y nunca registrada en claro.

## Preflight obligatorio

1. Confirma HU-097, decisiones de token/correo, issue/rama y fuentes.
2. Ejecuta Graphify sobre audiencia customer, logging, tokens, booking y anonimización.
3. Realiza threat model corto de fuga por URL, logs, referer, caché e historial.

## Alcance incluido

- Pantalla de éxito y operación customer de lectura mínima.
- Hash, vigencia/rotación/revocación según `DP-PUB-05`; respuesta uniforme.
- Redacción del token en observabilidad y aislamiento de una cita.

## Fuera de alcance

- Cuenta, portal, lista de citas, historial, edición, reprogramación o cancelación (HU-099).

## Estado existente que debe conservarse

- Prefijo `/customer`, logger que normaliza rutas y anonimización/revocación de referencia.

## Trabajo requerido

1. Diseña contrato y esquema de seguridad customer explícito.
2. Implementa validación constante/segura, consulta mínima y políticas de caché/referrer.
3. Construye UI de confirmación/consulta y estados de credencial inválida.
4. Integra entrega por correo únicamente según resolución de `CT-011`.

## Pruebas y evidencia

- Token válido/inválido/vencido/revocado, dos tenants y enumeración uniforme.
- Inspección de logs/headers; contrato, componente y E2E creación→consulta.
- Responsive/accesibilidad en cuatro anchos, teclado, zoom y axe-core.

## Documentación y trazabilidad

Actualiza OpenAPI/cliente, threat model/privacidad, matriz, HU y catálogo.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle && pnpm run db:atlas:validate
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-098-01`–`CA-098-06`.

## Git y PR

- Commit/PR: `feat(customer-access): implementa HU-098 acceso al turno`.
