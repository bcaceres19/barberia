---
prompt_id: "PROMPT-HU-100-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-100"
related_hu: ["HU-092", "HU-098"]
issue: 357
issue_url: "https://github.com/bcaceres19/barberia/issues/357"
suggested_issue_title: "feat(public-booking): implementar HU-100 reseñas internas por barbero"
branch: "feat/357-hu100-resenas-internas"
pr: 371
pr_url: "https://github.com/bcaceres19/barberia/pull/371"
depends_on: ["HU-098 integrada (satisfecho, PR #272)"]
rules: ["RN-TEN-01", "RN-DAT-02"]
decisions: ["DEC-089", "DEC-104", "DEC-128"]
acceptance_criteria: ["CA-100-01", "CA-100-02", "CA-100-03", "CA-100-04", "CA-100-05", "CA-100-06"]
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "apps/api/internal/modules/customeraccess"
  - "apps/api/internal/modules/staff"
  - "apps/api/internal/modules/publicbooking"
created_at: "2026-10-26"
updated_at: "2026-10-29"
supersedes: null
superseded_by: null
---

# Implementar HU-100: reseñas y calificación interna por barbero

## Instrucción para el agente

Implementa únicamente el sistema de reseñas internas por barbero descrito en `HU-100` y `DEC-128`. No construyas reseñas de la barbería/local (`HU-101`, datos de Google), ni cola de moderación previa, ni respuesta pública del barbero.

## Objetivo

Que un cliente con un turno `completed` pueda calificar una sola vez a su barbero desde la misma página de acceso al turno (`HU-098`), y que esa calificación se refleje de inmediato en el perfil público del barbero, con ocultamiento reversible desde el panel.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward.
2. Verifica que `HU-098` (acceso por token) está integrada en `main`.
3. Lee completamente `AGENTS.md`, `CLAUDE.md`, `DEC-089`, `DEC-104` y `DEC-128`.
4. Si el metadato `issue` sigue `pending`, créalo contra este prompt antes de escribir código y actualiza este archivo con el número real.
5. Crea la rama `feat/<issue>-hu100-resenas-internas` desde `main` actualizada.

## Alcance incluido

- Tabla nueva `barber_review` (tenant-aware, RLS, `ON DELETE RESTRICT` hacia `barber` y `appointment`, `UNIQUE (appointment_id)`).
- Acción de calificación accesible solo desde `GET /customer/appointments/{token}` cuando el turno es `completed` y no tiene reseña.
- `rating` (1–5, obligatorio) y `comment` (≤500 caracteres, recorte de espacios, vacío → `NULL`).
- `rating`/`ratingCount` del barbero (promedio/total sobre reseñas no ocultas) en `BarberResponse`/`PublicBarberResponse`.
- Acción de ocultar/mostrar una reseña desde el panel del barbero o del propietario (`hidden_at`, reversible, sin borrado físico).

## Fuera de alcance

- Reseñas del negocio/local (`HU-101`).
- Respuesta pública del barbero, edición de una reseña ya enviada, aprobación previa a publicar, notificación al barbero.
- Cualquier cambio a la matriz 0/1/N de `HU-092` o a su proyección pública ya fijada (nombre, foto y descripción).

## Estado existente que debe conservarse

- El token opaco de acceso al turno (`DEC-089`, `HU-098`) es el único mecanismo de verificación; no se inventa un token ni enlace nuevo.
- El patrón de ocultamiento reversible sin borrado físico ya usado por la fotografía del barbero (`DEC-104`).
- Los estados de cita de `estados-citas.md`; `completed` es la única puerta de entrada.

## Trabajo requerido

1. Migración `barber_review` (columnas `id`, `barber_id`, `appointment_id` único, `rating smallint CHECK (rating BETWEEN 1 AND 5)`, `comment text NULL`, `hidden_at timestamptz NULL`, `created_at timestamptz NOT NULL DEFAULT now()`), prueba SQL con dos tenants.
2. Servicio de dominio que valida/normaliza `rating`/`comment`, resuelve si el turno es calificable (mismo criterio uniforme de `CA-098-02`) y aplica la unicidad por turno.
3. Endpoint(s) bajo `/customer/appointments/{token}/review` (contract-first en OpenAPI) para crear la reseña; endpoint(s) privados para ocultar/mostrar.
4. Cálculo de `rating`/`ratingCount` en la consulta pública existente de `staff`/`publicbooking`, sin columna derivada.
5. Frontend: acción «Calificar» en `CustomerAppointmentPage.vue`; reseñas visibles en la selección pública de barbero (`PublicBarberSelectionPage.vue` o equivalente); acción de ocultar/mostrar en el panel de Barberos.

## Pruebas y evidencia

- PostgreSQL real con dos tenants: unicidad por turno, promedio que excluye ocultas, `ON DELETE RESTRICT`.
- Contrato HTTP: turno no completado/cancelado/ajeno/ya calificado → mismo `404`/`409`.
- Componente Vue con `vitest-axe`.
- E2E: calificar tras turno completado → aparece en selección pública → ocultar → desaparece.

## Documentación y trazabilidad

- Actualiza OpenAPI (`CHANGELOG.md` con el bump de versión), cliente generado, `database/README.md` y `modelo-fisico-referencia.sql`.
- Actualiza `matriz-trazabilidad.md`, `historial-cambios.md`, este prompt y el índice de `prompts/README.md` con issue, rama, PR y estado reales.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
```

Entrega tabla `CA-100-01`–`CA-100-06` | Estado | Prueba o evidencia.

## Git y PR

- Commit/PR: `feat(public-booking): implementar HU-100 reseñas internas por barbero`.
- `Closes #<issue>` solo si cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
