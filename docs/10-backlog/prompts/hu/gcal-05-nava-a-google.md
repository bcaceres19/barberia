---
prompt_id: "PROMPT-FEAT-GCAL-05-NAVA-A-GOOGLE-v1"
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
suggested_issue_title: "feat(integraciones): sincronizar citas y bloqueos de NAVA hacia Google Calendar"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-04-v1 integrado (conexión OAuth)"
rules:
  - "RN-CON-03"
  - "RN-TEN-01"
decisions:
  - "DEC-099"
  - "DEC-101"
  - "DEC-102"
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
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/modules/publicbooking"
  - "apps/api/internal/modules/schedule"
  - "apps/api/cmd/worker"
  - "apps/api/internal/platform/idempotency"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Sincronización NAVA → Google Calendar con cola propia

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

Toda cita confirmada, reprogramada o cancelada y todo bloqueo compatible del barbero conectado se refleja en su Google Calendar mediante una cola confiable; una caída de Google nunca impide crear la cita.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Migración Atlas y RLS de `google_calendar_event_link` (`resource_type` `appointment`/`time_block`, ids de Google, `etag`, `updated`, huella del estado escrito) y de `google_calendar_sync_job` con claim con lease y CAS por `claim_token`.
- Puertos definidos por `booking` y `schedule` (sin importar Google) que encolan el trabajo en la misma transacción del cambio; cubrir reserva pública y creación manual.
- Worker: reclamo, ejecución fuera de transacción, backoff exponencial con tope y máximo de intentos; error permanente cambia el estado de la conexión.
- Crear, actualizar el mismo evento y eliminar por cancelación; `completed` y `no_show` conservan el evento; título y descripción según `DEC-101`; propiedades extendidas privadas; zona de la barbería.
- Empuje inicial de las citas y bloqueos futuros dentro de la ventana de 6 meses al conectar, sin vincular por heurística.

## Fuera de alcance

- Webhook, `watch`, `syncToken` y Google → NAVA (`PROMPT-FEAT-GCAL-06-v1`).
- UI.
- Series de bloqueo de NAVA (`DP-INT-02`).

## Estado existente que debe conservarse

- Las transacciones de reserva no esperan a ningún servicio externo.
- Idempotencia de la reserva y de la reprogramación existentes.
- Retención y datos personales: nada de teléfono, correo, notas ni tokens en el evento, los logs o el trabajo.

## Trabajo requerido

1. Diseña y documenta el mecanismo de encolado transaccional elegido (puerto en el consumidor).
2. Escribe las pruebas 1 a 5, 13, 15, 16, 17 y 20 de la orquestación antes de implementar.
3. Implementa migraciones, adaptador Google de eventos, encolado y ejecución.
4. Registra los eventos de log `google_calendar.event.*` con conexión, recurso y operación, sin secretos.

## Pruebas y evidencia

- Servidor Google falso: creación única (1 cita, 1 evento, 1 vínculo), reprogramación sobre el mismo `googleEventId`, cancelaciones eliminan, terminales conservan.
- Google 503/429/timeout: la cita existe, el trabajo reintenta con backoff y no hay bucle infinito; token revocado: `reauth_required`.
- Concurrencia y varios workers con `SKIP LOCKED`; PostgreSQL real y dos tenants.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Actualiza diccionario y diagrama de datos y el documento de operación del worker.
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
