---
prompt_id: "PROMPT-FEAT-GCAL-03-BOOKING-ACCION-INTEGRACION-v1"
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
suggested_issue_title: "feat(citas): reprogramar y cancelar citas por una integración con actor sistema"
branch: null
pr: null
pr_url: null
depends_on:
  - "PR del issue #284 integrado (DEC-099 a DEC-102)"
rules:
  - "RN-CON-03"
  - "RN-HIS-01"
decisions:
  - "DEC-073"
  - "DEC-076"
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
  - "docs/02-requisitos/historias-usuario.md (HU-062, HU-066)"
  - "docs/02-requisitos/estados-citas.md"
  - "apps/api/internal/modules/booking"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Casos de uso de `booking` para cambios originados por una integración

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

`booking` ofrece reprogramación y cancelación por el barbero solicitadas por una integración, con actor `system` y origen visible en el historial, reutilizando exactamente las mismas invariantes que el flujo del barbero.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Análisis de `RescheduleService` y `CancelAppointmentByBarberService`: extraer el núcleo común y añadir un caso de uso de integración en lugar de aceptar un `ActorStaffUserID` falso.
- Origen «sincronización Google Calendar» en `appointment_history` (columna o dato de la entrada, según el estándar; sin `jsonb` sin justificación) y su exposición en el detalle/historial de OpenAPI.
- Comportamiento definido ante estado terminal, versión desactualizada, cruce con cita, bloqueo (`DEC-073`, `DEC-076`) y fuera de jornada: error de dominio tipificado que la integración pueda distinguir del error interno.

## Fuera de alcance

- Cualquier código de Google o de la integración.
- Cambiar la semántica de los estados o de `appointment`.
- UI del historial (la etiqueta visible va en `PROMPT-FEAT-GCAL-07-v1`).

## Estado existente que debe conservarse

- `DEC-076` (reprogramación voluntaria bloquea si cruza), `RN-HIS-01` e idempotencia vigentes.
- El actor `system` ya existe en `appointment_history_actor_shape_ck`; no se debilita la auditoría.
- `booking` no importa `google.golang.org/api` ni el módulo de integración.

## Trabajo requerido

1. Escribe primero las pruebas de las reglas con actor de integración.
2. Refactoriza el núcleo común sin cambiar el comportamiento de los flujos existentes (las pruebas actuales siguen verdes).
3. Añade los casos de uso y el origen en el historial (migración Atlas si hace falta).
4. Actualiza OpenAPI, handler y cliente solo donde el origen se expone.

## Pruebas y evidencia

- Unitarias: reprogramación válida, cruce, bloqueo, estado terminal, versión, idempotencia; cancelación válida y sobre estado terminal.
- PostgreSQL real con dos tenants: exclusión de cruces bajo concurrencia y `FOR UPDATE`.
- Historial con `actor_type = system` y un solo registro ante repetición.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Actualiza `estados-citas.md` o el diccionario si el origen se expone.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
cd apps/api && go vet ./... && go test ./... && go test -race ./... && govulncheck ./...
pnpm run openapi:lint (y el bundle/cliente tipado definidos por el repositorio)
atlas migrate validate y pruebas contra PostgreSQL real con al menos dos tenants
tools/ai/validate-agent-system.sh --strict
```

Entrega una tabla `Criterio | Estado | Prueba o evidencia`. No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR con Conventional Commits, sin secretos ni datos personales.
- `Closes #<issue>` solo si se cubre el issue completo; de lo contrario `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`; integra por PR y squash tras CI verde.
