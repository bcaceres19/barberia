---
prompt_id: "PROMPT-HU-102-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents: ["codex", "claude"]
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-102"
related_hu: ["HU-090", "HU-025", "HU-101", "HU-103"]
issue: 359
issue_url: "https://github.com/bcaceres19/barberia/issues/359"
suggested_issue_title: "feat(public-booking): implementar HU-102 ubicación de la barbería"
branch: "feat/359-hu102-ubicacion-barberia"
pr: null
pr_url: null
depends_on: ["Credencial de Google Geocoding/Maps disponible en el ambiente de despliegue"]
rules: ["RN-TEN-01"]
decisions: ["DEC-082", "DEC-110", "DEC-115", "DEC-130"]
acceptance_criteria: ["CA-102-01", "CA-102-02", "CA-102-03", "CA-102-04", "CA-102-05", "CA-102-06"]
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
created_at: "2026-10-26"
updated_at: "2026-10-29"
supersedes: null
superseded_by: null
---

# Implementar HU-102: ubicación de la barbería en la reserva pública

## Instrucción para el agente

Implementa únicamente la ubicación configurable descrita en `HU-102`/`DEC-130`. No implementes domicilio (`HU-103`) ni reseñas de Google (`HU-101`), aunque comparte credencial de Google Maps Platform con ambas.

## Objetivo

Que la reserva pública muestre un mapa real con marcador y la dirección en texto únicamente cuando la barbería activó `has_physical_location` y tiene coordenadas geocodificadas guardadas; en cualquier otro caso, la sección no aparece, sin error visible.

## Preflight obligatorio

1. Comprueba árbol limpio, actualiza `main` por fast-forward.
2. Lee completamente `AGENTS.md`, `CLAUDE.md`, `DEC-110`, `DEC-115` y `DEC-130`.
3. Si `HU-101` ya introdujo la credencial compartida de Google Maps Platform, reutilízala; si no, documenta la credencial igual que `HU-101` lo haría.
4. Si el metadato `issue` sigue `pending`, créalo contra este prompt antes de escribir código y actualiza este archivo con el número real.
5. Crea la rama `feat/<issue>-hu102-ubicacion-barberia` desde `main` actualizada.

## Alcance incluido

- Columnas nuevas en `barbershop`: `has_physical_location boolean NOT NULL DEFAULT true`, `address text NULL`, `latitude numeric NULL`, `longitude numeric NULL`.
- Edición en Configuración junto a los demás ajustes de marca (`DEC-110`); al guardar una dirección nueva, geocodificación en el servidor (Google Geocoding) con `422` si Google no puede resolverla.
- Campos nuevos necesarios en `GET /public/barbershops/{slug}` para mostrar el mapa cuando corresponda.
- Mapa interactivo de Google con marcador y dirección en texto en la reserva pública, visible solo con `has_physical_location = true` y coordenadas presentes.

## Fuera de alcance

- Varias sedes, horario distinto por sede, instrucciones de llegada.
- Domicilio con dirección del cliente (`HU-103`).
- Geocodificación en el navegador del cliente: siempre se geocodifica una sola vez en el servidor al guardar.

## Estado existente que debe conservarse

- El recurso de marca/configuración de `DEC-110` y su patrón de edición.
- El patrón de resolución pública sin tenant de `DEC-082`/`HU-090`.
- El criterio de ajuste explícito sobre inferencia implícita ya usado por `panel_profile` (`DEC-115`): `has_physical_location` nunca se deriva de si `address` está vacío.

## Trabajo requerido

1. Migración que añade `has_physical_location`/`address`/`latitude`/`longitude` a `barbershop`.
2. Adaptador de Geocoding (puede vivir junto al de `HU-101` si ya existe) que resuelve una dirección a coordenadas y rechaza la que Google no puede geocodificar.
3. Endpoint de edición en Configuración (backend y frontend) que geocodifica al guardar y conserva `address`/coordenadas al apagar `has_physical_location`.
4. Campos nuevos en `GET /public/barbershops/{slug}` (contract-first en OpenAPI).
5. Frontend: sección de ubicación en la reserva pública con mapa interactivo (iframe o SDK de Google Maps) y dirección en texto; oculta según la regla de visibilidad.

## Pruebas y evidencia

- Adaptador de Geocoding simulado: dirección válida, ambigua, no encontrada.
- PostgreSQL real con dos tenants: persistencia de coordenadas, conservación del dato al apagar el interruptor.
- Contrato HTTP; componente Vue con `vitest-axe`; E2E con y sin ubicación visible.

## Documentación y trazabilidad

- Actualiza OpenAPI (`CHANGELOG.md`), cliente generado, `database/README.md`, `modelo-fisico-referencia.sql`.
- Actualiza `matriz-trazabilidad.md`, `historial-cambios.md`, este prompt y el índice de `prompts/README.md` con issue, rama, PR y estado reales.

## Verificación final

```text
pnpm run openapi:lint && pnpm run openapi:bundle
cd apps/api && gofmt -l . && go vet ./... && go test -race ./... && go build ./...
cd ../web && pnpm run format && pnpm run lint && pnpm run typecheck && pnpm run test:unit && pnpm run build && pnpm run test:e2e
```

Entrega tabla `CA-102-01`–`CA-102-06` | Estado | Prueba o evidencia.

## Git y PR

- Commit/PR: `feat(public-booking): implementar HU-102 ubicación de la barbería`.
- `Closes #<issue>` solo si cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
