---
prompt_id: "PROMPT-HU-090-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-090"
related_hu: ["HU-091"]
issue: 243
issue_url: "https://github.com/bcaceres19/barberia/issues/243"
suggested_issue_title: "feat(public-booking): implementar HU-090 entrada pública"
branch: "feat/243-hu090-entrada-publica"
pr: null
pr_url: null
depends_on: ["B3 con criterio de salida cumplido (cerrado)", "DP-PUB-01 resuelta mediante DEC-082 (resuelta)", "Issue real #243 con CA-090-01 a CA-090-05"]
rules: ["RN-TEN-01", "RN-DAT-01", "RN-DAT-02"]
decisions: ["DEC-016", "DEC-019", "DEC-022", "DEC-024", "DEC-033", "DEC-034", "DEC-035", "DEC-036", "DEC-037", "DEC-038", "DEC-039", "DEC-077", "DEC-078", "DEC-079", "DEC-082"]
acceptance_criteria: ["CA-090-01", "CA-090-02", "CA-090-03", "CA-090-04", "CA-090-05"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/00-control/contradicciones.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/05-backend/migraciones-atlas.md", "docs/06-api/estandar-openapi.md", "api/openapi/paths/public-booking.yaml", "apps/web/src/modules/public-booking"]
created_at: "2026-09-10"
updated_at: "2026-09-10"
supersedes: null
superseded_by: null
---

# Implementar HU-090: entrada pública de reservas

## Instrucción para el agente

Implementa solo `HU-090`. B3 cerró su criterio de salida y `DP-PUB-01` quedó resuelta como `DEC-082`; el issue real es [#243](https://github.com/bcaceres19/barberia/issues/243).

## Objetivo

Abrir sin cuenta el contexto público correcto de una barbería mediante un identificador seguro, con respuesta uniforme y sin fuga tenant.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada, issue real y rama `feat/<issue>-hu090-entrada-publica`.
2. Lee todos los `source_docs`; detente ante duda o contradicción abierta.
3. Ejecuta Graphify sobre shops, router público, RLS y shell frontend.
4. Confirma que la decisión de `DP-PUB-01` está reflejada en HU, contrato y modelo.

## Alcance incluido

- Resolución pública de barbería habilitada sin confiar en `barbershopId` del cliente.
- Contrato, caso de uso, persistencia estrictamente necesaria y pantalla pública NAVA.
- Carga, enlace inválido/no disponible, error y reintento accesibles.

## Fuera de alcance

- Servicios, barberos, disponibilidad, formulario, cita o cancelación.
- Portal/cuenta de cliente o semántica no aprobada del identificador.

## Estado existente que debe conservarse

- Audiencias `/public`, `/customer`, `/private`, RLS y módulo `shops` existentes; `public-booking.yaml` y `modules/public-booking` son placeholders sin implementación.
- Shell privado y datos de configuración no se convierten en API pública por reutilización accidental.

## Trabajo requerido

1. Diseña OpenAPI primero y un request sin tenant confiado.
2. Resuelve el contexto según `DP-PUB-01`, con errores uniformes y logs redactados.
3. Añade puertos/caso de uso independientes de Chi/PostgreSQL y adaptadores tenant-aware.
4. Construye la ruta Vue pública, aislada del shell privado y con estados completos.

## Pruebas y evidencia

- Unidad/HTTP y PostgreSQL real con dos tenants; inválido, desconocido y no publicable.
- Componente/E2E en 320, 360, 768 y 1280 px, teclado, foco, zoom 200 % y axe-core.

## Documentación y trazabilidad

Actualiza OpenAPI/CHANGELOG/cliente, matriz, historia, plan, README de módulos y este prompt/índice con datos reales.

## Verificación final

```text
pnpm run openapi:check-config && pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && pnpm run db:atlas:validate && git diff --check && graphify update .
```

Entrega `Criterio | Estado | Prueba o evidencia` para `CA-090-01`–`CA-090-05`.

## Git y PR

- Commit/PR: `feat(public-booking): implementa HU-090 entrada pública`.
- `Closes #<issue>` solo si cubre todo; no push directo ni merge sin checks.
