---
prompt_id: "PROMPT-FEAT-GCAL-02-SCHEDULE-FUENTE-GOOGLE-CALENDAR-v1"
version: "1.0"
kind: "hu"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "feat(agenda): admitir bloqueos de origen google_calendar en schedule"
branch: null
pr: null
pr_url: null
depends_on:
  - "PR del issue #284 integrado (DEC-099 a DEC-102)"
rules:
  - "RN-CON-03"
  - "RN-DIS-06"
decisions:
  - "DEC-020"
  - "DEC-073"
  - "DEC-099"
  - "DEC-101"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/02-requisitos/historias-usuario.md (HU-042)"
  - "apps/api/internal/modules/schedule"
  - "apps/api/internal/modules/availability"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Fuente `google_calendar` para los bloqueos de agenda

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

`schedule` puede crear, actualizar y retirar lógicamente bloqueos con `source = google_calendar` por medio de un puerto propio, y la disponibilidad pública y la agenda los tratan como cualquier bloqueo vigente.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Nueva migración que amplía `time_block_source_ck` a `manual`, `holiday_calendar`, `google_calendar` (sin tocar la migración aplicada).
- Constante de dominio de fuente y puerto de integración en `schedule` (crear, actualizar el intervalo, retirar con `deleted_at` y `deleted_by` nulo para el sistema).
- Verificación de que `availability`, la creación manual (`DEC-073`), la reprogramación (`DEC-076`) y la agenda diaria ya respetan cualquier bloqueo vigente; ajustes solo si no lo hacen.
- Exposición del origen en OpenAPI solo si la agenda ya devuelve la fuente.

## Fuera de alcance

- Cualquier llamada a Google, OAuth o sincronización.
- Mapear bloqueos a eventos (vive en `PROMPT-FEAT-GCAL-05-v1`).
- Cambiar el modelo de series o de festivos.

## Estado existente que debe conservarse

- Bloqueos manuales y de festivo, sus servicios y su retiro lógico.
- `time_block_deleted_by_ck` permite retiro sin `deleted_by`; verifica que el sistema pueda retirar sin usuario.
- El dominio de `schedule` no importa Google ni Chi ni pgx.

## Trabajo requerido

1. Documenta la decisión en el diccionario de datos y comprueba si el `CHECK` es la única restricción de fuente.
2. Escribe primero las pruebas de la regla de solape entre bloqueos y entre bloqueo y cita para el nuevo origen.
3. Implementa migración, dominio, puerto y repositorio.
4. Verifica que las consultas de disponibilidad no filtran por fuente.

## Pruebas y evidencia

- Migración con PostgreSQL real y dos tenants: inserción de `google_calendar`, rechazo de fuentes inválidas, RLS.
- Servicio: crear, actualizar el mismo bloque (sin duplicados), retirar; idempotencia ante repetición.
- Disponibilidad: un bloqueo `google_calendar` deja de ofrecer la franja.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Actualiza `database/modelo-fisico-referencia.sql`, diccionario y diagrama de datos.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/api && go vet ./... && go test ./... && go test -race ./... && govulncheck ./...
atlas migrate validate y pruebas contra PostgreSQL real con al menos dos tenants
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
