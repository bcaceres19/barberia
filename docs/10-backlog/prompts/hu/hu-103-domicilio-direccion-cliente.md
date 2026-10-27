---
prompt_id: "PROMPT-HU-103-v1"
version: "1.0"
kind: "hu"
status: "ready"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-103"
related_hu: ["HU-096", "HU-097", "HU-102"]
issue: 360
issue_url: "https://github.com/bcaceres19/barberia/issues/360"
suggested_issue_title: "feat(public-booking): implementar HU-103 domicilio con dirección del cliente"
branch: null
pr: null
pr_url: null
depends_on: ["HU-096 integrada (satisfecho)", "HU-097 integrada (satisfecho, PR #270)", "Credencial de Google Places Autocomplete disponible (comparte familia con HU-102)"]
rules: ["RN-TEN-01", "RN-DAT-02", "RN-CON-01"]
decisions: ["DEC-085", "DEC-089", "DEC-090", "DEC-091", "DEC-131"]
acceptance_criteria: ["CA-103-01", "CA-103-02", "CA-103-03", "CA-103-04", "CA-103-05", "CA-103-06", "CA-103-07"]
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
  - "apps/api/internal/modules/shops"
  - "apps/api/internal/modules/publicbooking"
  - "apps/api/internal/modules/booking"
created_at: "2026-10-26"
updated_at: "2026-10-26"
supersedes: null
superseded_by: null
---

# Implementar HU-103: domicilio con dirección del cliente y mapa

## Instrucción para el agente

Implementa únicamente el interruptor general de domicilio y la captura de dirección descritos en `HU-103`/`DEC-131`. No conviertas el domicilio en un servicio propio del catálogo ni lo hagas configurable por servicio individual: la decisión del propietario fue un interruptor general por barbería.

## Objetivo

Que, cuando la barbería activa `offers_home_service`, el cliente pueda elegir «a domicilio» al reservar cualquier servicio e indicar su dirección con autocompletado y mapa de Google, y que ningún turno quede con datos de dirección parciales o inconsistentes con su modalidad.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward.
2. Verifica que `HU-096` y `HU-097` están integradas en `main` (confirmación pública real, con `Idempotency-Key` y revalidación).
3. Lee completamente `AGENTS.md`, `CLAUDE.md`, `DEC-085`, `DEC-089`–`DEC-091` y `DEC-131`.
4. Si `HU-102` ya introdujo la credencial de Google Places Autocomplete/Maps, reutilízala.
5. Si el metadato `issue` sigue `pending`, créalo contra este prompt antes de escribir código y actualiza este archivo con el número real.
6. Crea la rama `feat/<issue>-hu103-domicilio-direccion` desde `main` actualizada.

## Alcance incluido

- Columna `barbershop.offers_home_service boolean NOT NULL DEFAULT false`, editable en Configuración.
- Paso «¿en el local o a domicilio?» en el flujo público entre `HU-096` y la confirmación de `HU-097`, presente solo cuando `offers_home_service = true`.
- Campo de dirección con autocompletado de Google Places y mapa interactivo de confirmación; exige una selección válida con coordenadas.
- Columnas nuevas en `appointment`: `is_home_service boolean NOT NULL DEFAULT false`, `service_address text NULL`, `service_latitude numeric NULL`, `service_longitude numeric NULL`, con un `CHECK` que exige los tres datos de dirección juntos cuando `is_home_service = true` y los tres nulos cuando es `false`.
- `POST .../appointments` (`HU-097`) amplía su cuerpo para aceptar la modalidad y, si es a domicilio, la dirección elegida; revalida `offers_home_service` al confirmar y rechaza con `422` si cambió a `false`.
- Visibilidad de la dirección a domicilio en el panel privado (agenda y detalle del turno), nunca en la reserva pública ni en las reseñas de `HU-100`/`HU-101`.

## Fuera de alcance

- Cobro o recargo distinto por domicilio, radio de cobertura, ruta u optimización de desplazamiento.
- Domicilio como servicio propio del catálogo con su propio precio/duración.
- Domicilio configurable por servicio individual en vez de por barbería completa.

## Estado existente que debe conservarse

- El comando público de confirmación de `HU-097` (`Idempotency-Key`, revalidación, alternativas ante conflicto) se amplía, no se reescribe.
- La reconciliación de `customer` de `DEC-085` no cambia.
- `RN-DAT-02`: la dirección del cliente nunca se registra en logs.

## Trabajo requerido

1. Migraciones: `barbershop.offers_home_service`; `appointment.is_home_service`/`service_address`/`service_latitude`/`service_longitude` con su `CHECK`.
2. Edición de `offers_home_service` en Configuración (backend y frontend).
3. Paso de modalidad en el flujo público, con el campo de autocompletado (Google Places) y el mapa de confirmación.
4. Ampliación del contrato y del `ConfirmationService`/`booking.CreatePublic` para aceptar, revalidar y persistir la modalidad y la dirección en la misma transacción.
5. Exposición de la dirección a domicilio en la agenda/detalle privados de esa barbería sobre ese turno.

## Pruebas y evidencia

- PostgreSQL real con dos tenants sobre el `CHECK` de consistencia (domicilio completo o ausente, nunca parcial).
- Contrato HTTP: confirmación con y sin domicilio, y con `offers_home_service` revalidado (apagado entre elección y envío → `422` sin crear el turno).
- Componente Vue del selector de modalidad y el autocompletado (simulado); `vitest-axe`.
- E2E completo de principio a fin, con domicilio y sin él.

## Documentación y trazabilidad

- Actualiza OpenAPI (`CHANGELOG.md`), cliente generado, `database/README.md`, `modelo-fisico-referencia.sql`.
- Actualiza `matriz-trazabilidad.md`, `historial-cambios.md`, este prompt y el índice de `prompts/README.md` con issue, rama, PR y estado reales.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
```

Entrega tabla `CA-103-01`–`CA-103-07` | Estado | Prueba o evidencia.

## Git y PR

- Commit/PR: `feat(public-booking): implementar HU-103 domicilio con dirección del cliente`.
- `Closes #<issue>` solo si cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
