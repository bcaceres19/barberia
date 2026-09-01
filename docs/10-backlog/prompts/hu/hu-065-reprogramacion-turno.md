---
prompt_id: "PROMPT-HU-065-v1"
version: "1.2"
kind: "hu"
status: "in_progress"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-065"
related_hu:
  - "HU-004"
  - "HU-040"
  - "HU-041"
  - "HU-042"
  - "HU-060"
  - "HU-064"
issue: 123
issue_url: "https://github.com/bcaceres19/barberia/issues/123"
suggested_issue_title: "feat(booking): implementar HU-065 reprogramación auditada"
branch: "feat/123-hu065-reprogramacion-turno"
pr: null
pr_url: null
depends_on:
  - "HU-064 integrada en main con token opaco de versión mediante PR #121 (cumplido)"
  - "DP-CIT-06 resuelta y propagada mediante DEC-076 (cumplido)"
  - "Issue real propio con CA-065-01 a CA-065-08 (cumplido: #123)"
rules:
  - "RN-CIT-01"
  - "RN-CIT-03"
  - "RN-CON-01"
  - "RN-CON-03"
  - "RN-DIS-05"
  - "RN-DIS-07"
  - "RN-HIS-01"
  - "RN-HIS-02"
  - "RN-TEN-01"
  - "RN-IDE-01"
decisions:
  - "DEC-002"
  - "DEC-004"
  - "DEC-007"
  - "DEC-014"
  - "DEC-016"
  - "DEC-020"
  - "DEC-024"
  - "DEC-035"
  - "DEC-036"
  - "DEC-037"
  - "DEC-038"
  - "DEC-039"
  - "DEC-043"
  - "DEC-074"
  - "DEC-075"
  - "DEC-076"
acceptance_criteria:
  - "CA-065-01"
  - "CA-065-02"
  - "CA-065-03"
  - "CA-065-04"
  - "CA-065-05"
  - "CA-065-06"
  - "CA-065-07"
  - "CA-065-08"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/00-control/glosario.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/estados-citas.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/04-arquitectura/backend-go.md"
  - "docs/04-arquitectura/frontend.md"
  - "docs/05-backend/base-datos.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/plan-bloques.md"
  - "database/migrations/20260827110000_create_appointment_core.sql"
  - "api/openapi/paths/private-appointments.yaml"
  - "apps/api/internal/modules/booking"
  - "apps/api/internal/modules/schedule"
  - "apps/api/internal/platform/idempotency"
  - "apps/web/src/modules/agenda"
created_at: "2026-08-31"
updated_at: "2026-09-01T06:30:00Z"
supersedes: null
superseded_by: null
---

# Implementar HU-065: reprogramación auditada de un turno

## Instrucción para el agente

Implementa únicamente la transición `T2`: mover el intervalo de una cita `confirmed` sin cambiar su estado ni ningún otro dato. El cambio debe ser tenant-aware, idempotente, condicional a la versión leída, protegido por la exclusión PostgreSQL y auditado en la misma transacción.

El issue real [#123](https://github.com/bcaceres19/barberia/issues/123) ya está enlazado y este prompt pasó a `ready`: `HU-064` está integrada en `main` mediante [PR #121](https://github.com/bcaceres19/barberia/pull/121) con token opaco de versión probado, y `DP-CIT-06` quedó resuelta como `DEC-076` (bloqueo duro frente a un bloqueo vigente, mismo tratamiento que un cruce de citas). Si aparece otra contradicción entre contrato, esquema o UX, regístrala antes de codificar.

## Objetivo

Que el barbero reprograme un turno confirmado a otra fecha/hora futura válida, conserve el intervalo anterior en historial y reciba un conflicto recuperable si la agenda o la representación cambiaron, sin perder ni sobrescribir silenciosamente trabajo concurrente.

## Preflight obligatorio

1. Comprueba árbol limpio y `main` actualizada mediante fast-forward; preserva cambios ajenos.
2. Confirma `HU-064` integrada con detalle, historial y token opaco de versión probado.
3. Confirma que `DP-CIT-06` tiene un nuevo `DEC-*` y que este fue propagado a reglas, historia, matriz y prompt. `DEC-073` por sí sola no autoriza extrapolar la creación manual a T2.
4. Localiza el issue real. Mientras falte issue o dependencia, detente antes de cambiar el repositorio.
5. Actualiza este prompt y crea `feat/<issue>-hu065-reprogramacion-turno` desde `main`.
6. Ejecuta Graphify sobre `booking`, exclusión GiST, idempotencia, `schedule`, detalle/agenda Vue, contrato y utilidades de tiempo.
7. Lee completamente los `source_docs`; revisa T2, transiciones prohibidas, efectos secundarios y comportamiento repetido de `estados-citas.md`.
8. Verifica que ninguna migración aplicada vaya a editarse y que B5 siga fuera de alcance.

## Alcance incluido

- Comando HTTP explícito para reprogramar una cita, con sesión, `Idempotency-Key` y precondición basada en el token opaco de `HU-064`.
- Solo `starts_at`/`ends_at`; estado, barbero, servicio, duración, precio, cliente, persona, origen y nota permanecen iguales.
- Inicio estrictamente futuro en la zona de la barbería; fin derivado de la duración snapshot.
- Jornada efectiva consultada mediante puertos de `schedule`; un bloqueo vigente en el nuevo intervalo **rechaza** la reprogramación con el mismo `409` uniforme que un cruce de citas (`DEC-076`), sin importar implementación de `schedule` desde `booking`.
- Exclusión PostgreSQL y contigüidad `[inicio, fin)`; conflicto uniforme y seguro.
- Actualización más evento `appointment_rescheduled` más cambios anterior/nuevo dentro de una sola transacción.
- No-op del mismo intervalo, idempotencia concurrente y conflicto por token obsoleto.
- Formulario/confirmación desde detalle, retorno coherente, pruebas y evidencia responsive/accesible.

## Fuera de alcance

- `T3`: cambiar servicio, duración o precio; editar persona/contacto/nota; mover a otro barbero.
- Cancelar, completar, marcar `no_show`, corregir estado o reabrir una terminal.
- Reprogramación masiva, arrastrar, calendario semanal, sugerencias o disponibilidad pública.
- Enviar notificaciones o crear/regenerar recordatorios. B5 integrará esos efectos; no registres un envío ficticio.
- Relajar la exclusión, confiar solo en un `SELECT` previo o usar un mutex en memoria.
- Modificar `20260827110000_create_appointment_core.sql`; cualquier evolución usa migración Atlas nueva y solo si es necesaria.
- Resolver `DP-CIT-06` dentro del código o extender `DEC-073` por analogía sin decisión del propietario.

## Estado existente que debe conservarse

- `appointment_barber_interval_excl` impide cruces para estados que ocupan agenda y permite contigüidad.
- `appointment_history` admite `appointment_rescheduled`; `appointment_history_change` conserva pares anterior/nuevo y ambas tablas son append-only.
- `ManualBookingService` ya demuestra cómo colaborar con `schedule`, resolver zona y traducir conflictos sin acoplar módulos; su bloqueo duro aplica a creación manual por `DEC-073` y no se copia a T2 hasta resolver `DP-CIT-06`.
- El protocolo de `HU-004`/`DEC-043` maneja repetición exacta, contenido distinto y operación concurrente.
- `HU-064` entrega el estado y token de versión; `HU-063` conserva el contexto fecha/barbero.
- La ausencia actual de B5 no autoriza a afirmar que hubo aviso o recordatorio regenerado.

## Trabajo requerido

### 1. Contrato y precondición

1. Modela T2 como comando explícito, no como `PATCH` genérico que permita otros campos.
2. Exige `Idempotency-Key` y una precondición estándar basada en el token opaco del detalle; documenta `409` por agenda, estado, versión e idempotencia con códigos distinguibles.
3. Request cerrado: nuevo inicio civil y solo los campos estrictamente necesarios. Tenant, actor, fin y todos los snapshots se derivan.
4. Define repetición: mismo intervalo o misma intención devuelve éxito sin evento extra; intención distinta con la misma clave entra en conflicto.
5. Mantén compatibilidad del contrato v1 y actualiza cliente generado.

### 2. Dominio, aplicación y persistencia

1. Crea un caso de uso T2 en `booking` independiente de Chi/pgx, con reloj, zona, `schedule` y repositorio por puertos.
2. Valida cita propia `confirmed`, token vigente, inicio futuro y jornada efectiva; un bloqueo vigente en el nuevo intervalo rechaza con el mismo conflicto uniforme que un cruce de citas (`DEC-076`).
3. Deriva `ends_at` con `duration_minutes_snapshot`; no consulta catálogo ni cambia snapshots.
4. En una transacción tenant-aware, bloquea/condiciona la fila, vuelve a verificar estado/versión, actualiza intervalo e inserta historial/cambios.
5. Traduce exclusión `23P01` y resolución concurrente equivalente al mismo conflicto seguro; no deshabilites la restricción.
6. El no-op no actualiza `updated_at`, no inserta historial ni crea efectos secundarios.
7. No hagas llamadas de red en la transacción y no registres nombres, contacto, intervalos sensibles ni SQL con valores.

### 3. Frontend Vue

1. Añade “Reprogramar turno” solo para `confirmed` en el detalle; estados terminales no muestran la acción.
2. Formulario corto con fecha/hora, zona, intervalo actual, nuevo intervalo calculado y consecuencia antes de confirmar.
3. Envía el token de la representación y una clave estable por intento; bloquea doble toque sin sustituir la idempotencia.
4. Ante conflicto de versión, conserva la intención, explica que el turno cambió y ofrece recargar; ante conflicto de agenda, conserva datos y permite elegir otra hora.
5. Tras éxito actualiza detalle/historial y permite volver a la fecha nueva; no deja la fecha vieja presentada como vigente.
6. Cumple 44 px, teclado, foco, zoom 200 %, reducción de movimiento y anchos aprobados.

## Pruebas y evidencia

- Dominio: futuro, mismo intervalo, zona/DST, fin derivado, terminal, jornada, ramas aprobadas para bloqueo, token obsoleto y cancelación de contexto.
- HTTP/contrato: éxito/repetición, `400`, `401`, `404`, variantes `409`, `422`, campos desconocidos, idempotencia, precondición y RFC 9457.
- PostgreSQL real: dos tenants, atomicidad, rollback, historial/cambios exactos, contigüidad, cruce total/parcial/un minuto y no-op.
- Concurrencia real con barreras: dos destinos incompatibles o carrera con otra cita; nunca `sleep`, exactamente un estado final válido y ningún evento huérfano/duplicado.
- Componente: acción por estado, resumen, confirmación, doble toque, conflicto de agenda/versión, conservación de datos, éxito, foco y axe-core.
- E2E: reprogramar a otro día, comprobar agenda anterior/nueva y evento de historial; repetir solicitud y conservar un solo evento.
- Evidencia: 320, 360, 768 y 1280 px, zoom 200 %, teclado, contraste y foco.

## Documentación y trazabilidad

- Actualiza OpenAPI/CHANGELOG, cliente generado, README API/web/database, matriz, historial, plan y este prompt.
- Documenta atomicidad, precondición, traducción de conflictos, pruebas concurrentes y cualquier migración/índice real.
- Mantén T3, cancelación, estados y B5 pendientes; registra nuevas dudas antes de escoger defaults.

## Verificación final

```text
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle

cd apps/api
gofmt -l .
go vet ./...
go test -race ./...
go build ./...

cd ../web
pnpm run format
pnpm run lint
pnpm run typecheck
pnpm run test:unit
pnpm run build
pnpm run test:e2e

cd ../..
pnpm run db:atlas:hash
pnpm run db:atlas:validate
git diff --check
graphify update .
```

Ejecuta PostgreSQL 14 real con dos tenants y la carrera coordinada. Los comandos Atlas solo son obligatorios si existe migración; si no, documenta por qué no se necesitó. Entrega `Criterio | Estado | Prueba o evidencia` para `CA-065-01`–`CA-065-08`.

## Git y PR

- Rama sugerida: `feat/<issue>-hu065-reprogramacion-turno`.
- Commit/título: `feat(booking): implementa HU-065 reprogramación auditada`.
- Usa `Closes #<issue>` solo con ocho criterios y evidencia completos; si no, `Refs #<issue>`.
- No hagas push directo, force push, merge de `main`, edites migraciones aplicadas ni mezcles T3/cancelación/estados/B5.
