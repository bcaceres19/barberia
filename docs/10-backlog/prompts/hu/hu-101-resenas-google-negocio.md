---
prompt_id: "PROMPT-HU-101-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-101"
related_hu: ["HU-090", "HU-025"]
issue: 358
issue_url: "https://github.com/bcaceres19/barberia/issues/358"
suggested_issue_title: "feat(public-booking): implementar HU-101 reseñas y estrellas de Google del negocio"
branch: "feat/358-hu101-resenas-google"
pr: null
pr_url: null
depends_on: ["Credencial de servidor de Google Places disponible en el ambiente de despliegue"]
rules: ["RN-TEN-01"]
decisions: ["DEC-082", "DEC-110", "DEC-129"]
acceptance_criteria: ["CA-101-01", "CA-101-02", "CA-101-03", "CA-101-04", "CA-101-05", "CA-101-06"]
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "apps/api/internal/modules/shops"
  - "apps/api/internal/modules/publicbooking"
  - "apps/api/internal/modules/googlecalendar"
created_at: "2026-10-26"
updated_at: "2026-10-29"
supersedes: null
superseded_by: null
---

# Implementar HU-101: reseñas y estrellas de Google del negocio

## Instrucción para el agente

Implementa únicamente la integración de solo lectura con Google Places descrita en `HU-101`/`DEC-129`. No implementes reseñas internas (`HU-100`), ni ubicación/domicilio (`HU-102`/`HU-103`), aunque comparta credencial con ellas.

## Objetivo

Que la reserva pública muestre la calificación real de Google y hasta 5 de sus reseñas cuando la barbería configuró su `googlePlaceId`, con atribución a Google, y que la sección quede oculta sin error en cualquier otro caso (sin ficha, sin credencial, o ante un fallo de la consulta).

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward.
2. Lee completamente `AGENTS.md`, `CLAUDE.md`, `DEC-110` y `DEC-129`.
3. Revisa cómo `googlecalendar` aísla su dependencia oficial y documenta credenciales en `apps/api/README.md` (`DEC-099`/`DEC-102`); replica ese criterio para Google Places.
4. Si el metadato `issue` sigue `pending`, créalo contra este prompt antes de escribir código y actualiza este archivo con el número real.
5. Crea la rama `feat/<issue>-hu101-resenas-google` desde `main` actualizada.

## Alcance incluido

- Columna `barbershop.google_place_id` (`text` `NULL`), editable junto a los demás ajustes de marca (`DEC-110`).
- Adaptador de servidor hacia la Google Places API (Place Details: `rating`, `userRatingCount`, `reviews`), con timeout corto y sin reintento síncrono (mismo criterio que `DEC-066`).
- Caché de 6 horas sin persistencia permanente del texto de las reseñas.
- `GET /public/barbershops/{slug}/reviews` (sin sesión, resuelve el slug con el mismo patrón sin tenant de `DEC-082`).
- Credencial de servidor nueva (`GOOGLE_PLACES_API_KEY` o equivalente), validada al arrancar, con la integración desactivada por completo si falta.

## Fuera de alcance

- Responder reseñas, importarlas al sistema interno de `HU-100`, otros proveedores, ficha de Google por barbero individual.
- Ubicación de la barbería y domicilio (`HU-102`, `HU-103`), aunque reutilicen la misma familia de credenciales de Google Maps Platform.

## Estado existente que debe conservarse

- El recurso de marca/configuración de `DEC-110` y su patrón de edición.
- El patrón de resolución pública sin tenant de `DEC-082`/`HU-090`.
- El aislamiento de dependencias oficiales detrás de un adaptador, igual que `googlecalendar` (`DEC-099`).

## Trabajo requerido

1. Migración que añade `barbershop.google_place_id`.
2. Adaptador `googlereviews` (o paquete equivalente) que consulta Place Details con la clave de servidor y aplica la caché de 6 horas (en memoria o en una tabla pequeña con `expires_at`).
3. Endpoint `GET /public/barbershops/{slug}/reviews` contract-first; sin `googlePlaceId`, sin credencial o ante fallo, responde una estructura vacía explícita (nunca un error que el cliente vea).
4. Edición de `googlePlaceId` en Configuración (backend y frontend).
5. Frontend: sección de reseñas de Google en la reserva pública, con atribución visible y enlace a la ficha real; oculta cuando la respuesta viene vacía.
6. Guía de credenciales en `apps/api/README.md`, mismo formato que la de Google Calendar.

## Pruebas y evidencia

- Adaptador con servidor de Google simulado: éxito, fallo de red, cuota agotada, `googlePlaceId` inválido.
- Verificación de que una segunda consulta dentro de la ventana de caché no repite la llamada externa (conteo de invocaciones en la prueba).
- Contrato HTTP; componente Vue con `vitest-axe`; E2E con interceptor de red que simula la respuesta de Google.

## Documentación y trazabilidad

- Actualiza OpenAPI (`CHANGELOG.md`), cliente generado, `database/README.md`, `apps/api/README.md`.
- Actualiza `matriz-trazabilidad.md`, `historial-cambios.md`, este prompt y el índice de `prompts/README.md` con issue, rama, PR y estado reales.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
```

Entrega tabla `CA-101-01`–`CA-101-06` | Estado | Prueba o evidencia.

## Git y PR

- Commit/PR: `feat(public-booking): implementar HU-101 reseñas y estrellas de Google del negocio`.
- `Closes #<issue>` solo si cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
