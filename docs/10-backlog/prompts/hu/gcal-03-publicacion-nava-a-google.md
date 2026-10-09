---
prompt_id: "PROMPT-FEAT-GCAL-03-PUBLICACION-NAVA-A-GOOGLE-v1"
version: "1.0"
kind: "hu"
status: "ready"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "324"
issue_url: "https://github.com/bcaceres19/barberia/issues/324"
suggested_issue_title: "feat(integraciones): publicar citas y bloqueos de NAVA en Google Calendar"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-02-v1 integrado (conexión OAuth)"
rules:
  - "RN-CON-03"
  - "RN-TEN-01"
decisions:
  - "DEC-099"
  - "DEC-122"
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
updated_at: "2026-10-09"
supersedes: null
superseded_by: null
---

# Publicación de citas y bloqueos de NAVA en Google Calendar con cola propia

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

Toda cita confirmada, reprogramada o cancelada y todo bloqueo compatible del barbero conectado se refleja en su Google Calendar mediante una cola confiable; una caída de Google nunca impide crear la cita, y nada de lo que ocurra en Google altera NAVA.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- Migración Atlas y RLS de `google_calendar_event_link` (`resource_type` `appointment`/`time_block`, id del evento, `etag`) y de `google_calendar_sync_job` con claim con lease y CAS por `claim_token`.
- Puertos definidos por `booking` y `schedule` (sin importar Google) que encolan el trabajo en la misma transacción del cambio; cubrir reserva pública y creación manual.
- Worker: reclamo, ejecución fuera de transacción, backoff exponencial con tope y máximo de intentos; un error permanente cambia el estado de la conexión.
- Crear, actualizar el mismo evento y eliminar por cancelación; `completed` y `no_show` conservan el evento; título, descripción, propiedades extendidas privadas, recordatorio según `reminder_minutes` (un `override` emergente, o `useDefault = true` si es nulo) y zona de la barbería según `DEC-101`.
- Trabajo periódico del worker (cadencia de pocos minutos, a fijar y documentar) que lista solo los eventos publicados por NAVA (propiedad extendida privada `navaConnectionId`, sin `syncToken`) y recrea los que falten de citas confirmadas futuras y bloqueos vigentes futuros; nunca restaura citas canceladas, terminales ni pasadas, y nunca modifica el dominio de NAVA.
- Cancelación desde la app: elimina el evento, cierra el vínculo para que no se recree y trata `404`/`410` como éxito; un evento modificado en Google se restablece en la siguiente actualización de NAVA.
- Actualización de los eventos futuros cuando el barbero cambia `reminder_minutes`.
- Invitación al cliente (`DEC-122`): si la cita tiene correo y la conexión está `connected`, el evento lo incluye como asistente con `sendUpdates=all`, `guestsCanModify=false`, `guestsCanInviteOthers=false` y `guestsCanSeeOtherGuests=false`; el correo solo viaja como asistente, nunca en título, descripción ni propiedades extendidas. La respuesta del cliente no modifica NAVA.
- Publicación inicial de las citas y bloqueos futuros dentro de la ventana de 6 meses al conectar, y acción «Sincronizar ahora» que solo procesa la cola pendiente de esa conexión, segura ante varios clics.

## Fuera de alcance

- Importar cambios de Google, webhook, `watch`, `syncToken`, eventos externos como bloqueos o reprogramar/cancelar citas desde Google (`DEC-099`).
- UI.
- Series de bloqueo de NAVA (`DP-INT-02`, fuera de la primera fase).

## Estado existente que debe conservarse

- Las transacciones de reserva no esperan a ningún servicio externo.
- Idempotencia de la reserva y de la reprogramación existentes; sin cambios en `appointment`, `time_block` ni sus reglas.
- Datos personales: nada de teléfono, notas ni tokens en el evento, los logs o el trabajo; el correo del cliente solo como asistente (`DEC-122`) y nunca en logs.

## Trabajo requerido

1. Diseña y documenta el mecanismo de encolado transaccional elegido (puerto en el consumidor).
2. Escribe primero las pruebas 1 a 5, 13, 15, 16, 17 y 18 de la orquestación.
3. Implementa migraciones, adaptador Google de eventos, encolado y ejecución.
4. Registra los eventos de log `google_calendar.event.*` con conexión, recurso y operación, sin secretos.

## Pruebas y evidencia

- Servidor Google falso: creación única (1 cita, 1 evento, 1 vínculo), reprogramación sobre el mismo evento, cancelaciones eliminan, terminales conservan, bloque NAVA → evento.
- Google 503/429/timeout: la cita existe, el trabajo reintenta con backoff y no hay bucle infinito; token revocado: `reauth_required`; token expirado: refresh automático.
- Concurrencia y varios workers con `SKIP LOCKED`; PostgreSQL real y dos tenants.
- Un cambio o borrado hecho en Google no modifica ninguna cita ni bloqueo de NAVA.
- Evento borrado en Google: se recrea en el siguiente ciclo, una sola vez por ciclo y sin duplicados; una cita cancelada en la app no se recrea; una cita pasada o terminal no se recrea.
- `reminder_minutes`: con valor, el evento lleva ese único recordatorio; nulo, `useDefault`; cambiarlo actualiza los eventos futuros.

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
