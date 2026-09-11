---
prompt_id: "PROMPT-HU-093-v1"
version: "1.0"
kind: "hu"
status: "in_progress"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-093"
related_hu: ["HU-020", "HU-094", "HU-099"]
issue: 249
issue_url: "https://github.com/bcaceres19/barberia/issues/249"
suggested_issue_title: "feat(settings): implementar HU-093 políticas de reserva pública"
branch: "feat/249-hu093-configuracion-reserva-cancelacion"
pr: null
pr_url: null
depends_on: ["B3 cerrado (satisfecho)", "DP-PUB-02 resuelta mediante DEC-083 (satisfecho, 2026-09-11)", "Issue real con CA-093-01 a CA-093-05 (satisfecho, issue #249)"]
rules: ["RN-DIS-04", "RN-DIS-06", "RN-CAN-01", "RN-CAN-02", "RN-TEN-01"]
decisions: ["DEC-005", "DEC-006", "DEC-010", "DEC-018", "DEC-024", "DEC-077", "DEC-078", "DEC-079"]
acceptance_criteria: ["CA-093-01", "CA-093-02", "CA-093-03", "CA-093-04", "CA-093-05"]
source_docs: ["AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/00-control/registro-decisiones.md", "docs/00-control/dudas-pendientes.md", "docs/01-producto/alcance-mvp.md", "docs/01-producto/reglas-negocio.md", "docs/02-requisitos/historias-usuario.md", "docs/03-desarrollo/estandar-backend-go.md", "docs/03-desarrollo/estandar-frontend-vue.md", "docs/03-desarrollo/estandar-diseno-visual.md", "docs/03-desarrollo/estrategia-pruebas.md", "docs/05-backend/estandar-base-datos.md", "docs/05-backend/migraciones-atlas.md", "docs/06-api/estandar-openapi.md", "apps/api/internal/modules/shops", "apps/web/src/modules/settings"]
created_at: "2026-09-10"
updated_at: "2026-09-11"
supersedes: null
superseded_by: null
---

# Implementar HU-093: configuración de reserva y cancelación

## Instrucción para el agente

`DP-PUB-02` quedó resuelta por `DEC-083` (2026-09-11): rangos de anticipación `[0,1440]` min, ventana `[1,90]` días, rejilla `{5,10,15,20,30,60}` min, plazo de cancelación `[0,10080]` min, y default `permite_cliente=true`/`motivo_obligatorio=true`. Implementa solo configuración; no cambies citas.

## Objetivo

Persistir y editar por barbería anticipación, ventana, rejilla y política de cancelación con valores y rangos aprobados.

## Preflight obligatorio

1. Confirma decisión propagada, issue real, rama y B3 cerrado.
2. Revisa modelo de referencia sin copiarlo como norma.
3. Ejecuta Graphify sobre shops/settings, migraciones y consumidores.

## Alcance incluido

- Modelo normalizado o columnas justificadas, Atlas roll-forward, RLS y configuración versionada.
- OpenAPI privado, caso de uso y formulario accesible con unidades/consecuencias.
- Defaults exactos de `DEC-018` y política aprobada por `DP-PUB-02`.

## Fuera de alcance

- Disponibilidad, cancelación, notificaciones o mutación retroactiva.
- Rangos/defaults tomados solo del SQL de referencia.

## Estado existente que debe conservarse

- Configuración básica de `HU-020`, migraciones aplicadas inmutables y cliente generado.

## Trabajo requerido

1. Diseña contrato y migración Atlas con análisis de datos, bloqueo y recuperación.
2. Implementa lectura/actualización tenant-aware, request cerrado y `If-Match`.
3. Añade UI sin estado global y actualiza consumidores solo si es imprescindible para compilar.

## Pruebas y evidencia

- Fronteras, combinaciones y defaults; PostgreSQL real con dos tenants y migración vacía/con datos.
- Contrato, componente/E2E y evidencia accesible/responsive.

## Documentación y trazabilidad

Actualiza OpenAPI, Atlas sum, diccionario/diagrama si aplica, matriz, HU y catálogo.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle && pnpm run db:atlas:validate
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
cd ../.. && git diff --check && graphify update .
```

Entrega tabla para `CA-093-01`–`CA-093-05`.

## Git y PR

- Commit/PR: `feat(settings): implementa HU-093 políticas públicas`.
