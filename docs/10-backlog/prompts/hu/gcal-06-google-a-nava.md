---
prompt_id: "PROMPT-FEAT-GCAL-06-GOOGLE-A-NAVA-v1"
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
suggested_issue_title: "feat(integraciones): recibir cambios de Google Calendar por webhook y sincronizar incrementalmente"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FEAT-GCAL-02-v1 integrado"
  - "PROMPT-FEAT-GCAL-03-v1 integrado"
  - "PROMPT-FEAT-GCAL-05-v1 integrado"
rules:
  - "RN-CON-03"
  - "RN-TEN-01"
decisions:
  - "DEC-073"
  - "DEC-076"
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
  - "apps/api/internal/modules/googlecalendar"
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/modules/schedule"
  - "apps/api/cmd/worker"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Sincronización Google Calendar → NAVA con webhook y syncToken

## Instrucción para el agente

Implementa únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones. Es una entrega de la orquestación [PROMPT-ORCH-GCAL-BARBERO-v1](../orchestration/google-calendar-barbero.md); léela para conocer el orden y el principio de autoridad (NAVA manda, Google es una vista).

## Objetivo

Los cambios que el barbero hace en Google Calendar llegan a NAVA por notificaciones push y sincronización incremental: eventos personales bloquean la agenda y los eventos de cita se reconcilian a través de `booking`, sin duplicados ni bucles.

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos; trabaja en un worktree si hay cambios sin relación.
2. Actualiza `main` mediante fast-forward y verifica que cada dependencia de los metadatos esté integrada.
3. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`, y las `DEC-*` enlazadas.
4. Comprueba `dudas-pendientes.md` y `contradicciones.md`. Si falta una decisión, registra el bloqueo antes de cambiar código.
5. Este prompt nace con `issue: pending` y en `draft`. Antes de ejecutarlo crea el issue real con criterios verificables (título sugerido en los metadatos), actualiza este archivo y el índice, y solo entonces pásalo a `ready`.
6. Si esta entrega introduce o toca una historia, redáctala en `docs/02-requisitos/historias-usuario.md` con su número real y criterios `CA-*` antes de codificar; no inventes números.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- OpenAPI del webhook público, del «Sincronizar ahora» (seguro ante múltiples clics) y de sus errores.
- Canal `watch` (`channelId`, `resourceId`, expiración, secreto por canal con hash), renovación por el worker antes de expirar y parada al desconectar.
- Sincronización inicial y con `syncToken` (incluida la invalidación `410`), sin descargar todo en cada notificación.
- Evento externo → `time_block` `google_calendar` (creación, modificación del mismo bloque, eliminación lógica); día completo por zona de la barbería; recurrencias por ocurrencias dentro de la ventana.
- Evento de cita vinculado: mover → reprogramación y eliminar → cancelación por el barbero mediante los casos de uso de `PROMPT-FEAT-GCAL-03-v1`; rechazo y reconciliación al estado canónico con el conflicto registrado.
- Prevención de eco y de bucle (huella, `etag`, propiedades extendidas privadas) e idempotencia por notificación duplicada (`X-Goog-Message-Number`).
- Desconexión que retira lógicamente los bloques de la conexión y detiene el canal.
- Documentación del webhook por ambiente (HTTPS, dominio, redirect URI) y del túnel para desarrollo local.

## Fuera de alcance

- UI (`PROMPT-FEAT-GCAL-07-v1`).
- Crear citas, clientes o servicios desde eventos de Google (prohibido por `DEC-099`).
- Ampliar el modelo de recurrencias de NAVA.

## Estado existente que debe conservarse

- `booking` y `schedule` siguen siendo los únicos que escriben citas y bloqueos.
- La disponibilidad pública no consulta Google en tiempo real.
- Aislamiento por tenant en la resolución del webhook mediante una función acotada, sin `BYPASSRLS`.

## Trabajo requerido

1. Verifica contra la documentación vigente de Google la combinación de `singleEvents`, `syncToken` y la ventana; registra el resultado en el PR.
2. Escribe las pruebas 6 a 12, 14, 18, 19 y 20 antes de implementar.
3. Implementa webhook, canal, sincronización, reconciliación y limpieza.
4. Registra los eventos de log `google_calendar.webhook.*`, `incremental_sync.*`, `watch.*` y `conflict.detected` sin secretos.

## Pruebas y evidencia

- Servidor Google falso: notificación duplicada (un solo cambio), mover cita válida (historial con actor sistema), mover a un horario ocupado (rechazo y evento restablecido), eliminar cita, evento personal crear/modificar/eliminar, día completo, recurrente, evento de cita terminal.
- Eco de una escritura de NAVA: sin segunda reprogramación ni historial.
- Webhook con token de canal inválido o desconocido; disponibilidad pública sin la franja del evento personal.
- PostgreSQL real con dos tenants y ejecución concurrente de la misma sincronización.

## Documentación y trazabilidad

- Actualiza matriz de trazabilidad, historial, contrato OpenAPI, diccionario y diagrama de datos que realmente resulten afectados.
- Documenta la operación del canal, la renovación y el diagnóstico sin secretos.
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
