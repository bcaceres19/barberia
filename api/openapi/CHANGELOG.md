# Registro de cambios del contrato OpenAPI

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Ver [`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
sección 18 para qué cuenta como cambio compatible o incompatible.

## [0.19.0] - 2026-09-11

### Agregado

- `GET /public/barbershops/{slug}` (`operationId: resolvePublicBarbershop`,
  `CA-090-01` a `CA-090-04`, tag `PublicBooking`): entrada pública de
  reservas (`HU-090`). Resuelve `slug` -el identificador del enlace
  público generado automáticamente desde el nombre de la barbería
  (`DEC-082`, resuelve `DP-PUB-01`)- a la barbería habilitada
  correspondiente, sin sesión y sin que el cliente pueda fijar
  `barbershopId` (`CA-090-03`), mediante la función `SECURITY DEFINER`
  `public_resolve_barbershop_by_slug`
  (`20260911045044_add_barbershop_public_slug.sql`). Un `slug` mal
  formado, desconocido o de una barbería no publicable produce
  EXACTAMENTE la misma respuesta `404` uniforme (`CA-090-02`, `RN-TEN-01`).
  `200`, `404`, `500`.

## [0.18.0] - 2026-09-10

### Agregado

- `POST /private/appointments/{appointmentId}/complete` (`operationId:
  completeAppointment`) y `POST /private/appointments/{appointmentId}/no-show`
  (`operationId: markAppointmentNoShow`) (`CA-067-01` a `CA-067-08`): `T4`
  manual y `T7` (`HU-067`), cierran una cita `confirmed` cuyo `starts_at` ya
  pasó según el reloj del servidor como `completed` o `no_show`
  respectivamente, insertando un único evento (`appointment_completed` o
  `appointment_no_show`) dentro de la misma transacción que la
  actualización de estado. Antes de `starts_at` responden `422` sin
  persistir nada, ni siquiera una reclamación de idempotencia
  (`CA-067-03`). Protegidas con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`)
  y con la precondición `If-Match` (el `versionToken` de `HU-064`). Sin
  cuerpo de solicitud. Repetir el mismo resultado es un no-op exitoso sin
  duplicar el evento; el resultado contrario o cualquier otro estado
  terminal responde `409` orientando a la futura corrección `T8`
  (`CA-067-05`). `200`, `400`, `401`, `404`, `409`
  (versión/estado/idempotencia, códigos distinguibles), `422`, `500`.

## [0.17.0] - 2026-09-10

### Agregado

- `POST /private/appointments/{appointmentId}/cancel` (`operationId:
  cancelAppointmentByBarber`, `CA-066-01` a `CA-066-08`): T6 (`HU-066`),
  cancela una cita `confirmed` en cualquier momento (sin ventana temporal),
  cambiándola a `cancelled_by_barber` e insertando un único evento
  `appointment_cancelled_by_barber`, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`) y con la precondición `If-Match` (el
  `versionToken` de `HU-064`). Sin cuerpo de solicitud. Repetir la
  cancelación sobre una cita ya `cancelled_by_barber` es un no-op exitoso
  sin duplicar el evento (`CA-066-04`); cualquier otro estado terminal
  responde `409`. `200`, `400`, `401`, `404`, `409`
  (versión/estado/idempotencia, códigos distinguibles), `500`.

## [0.16.0] - 2026-09-01

### Agregado

- `POST /private/appointments/{appointmentId}/reschedule` (`operationId:
  rescheduleAppointment`, `CA-065-01` a `CA-065-08`): T2 (`HU-065`), mueve
  el intervalo de una cita `confirmed` a un nuevo inicio futuro, protegida
  con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`) y con la precondición
  `If-Match` (el `versionToken` de `HU-064`). Un bloqueo vigente en el
  nuevo intervalo rechaza con el mismo `409` que un cruce de citas
  (`DEC-076`); un intervalo idéntico es un no-op exitoso. `200`, `400`,
  `401`, `404`, `409` (agenda/versión/estado/idempotencia, códigos
  distinguibles), `422`, `500`.

## [0.15.0] - 2026-09-01

### Agregado

- `GET /private/appointments/{appointmentId}` (`operationId:
  getAppointmentDetail`, `CA-064-01` a `CA-064-04`): detalle completo de un
  turno (`HU-064`), incluidos contacto y nota (solo visibles aquí, nunca en
  la agenda diaria). Nunca incluye `customerId`, ningún identificador de
  actor, `barbershopId` ni `updatedAt` crudo; `versionToken` es el único
  dato de concurrencia expuesto, opaco. `200`, `401`, `404`, `500`.
- `GET /private/appointments/{appointmentId}/history` (`operationId:
  listAppointmentHistory`, `CA-064-05`): historial inmutable paginado por
  cursor (`HU-060`, `HU-064`), orden estable por instante y luego por
  identificador. Cada entrada trae `actorLabel`, un nombre visible seguro
  (nunca correo ni identificador de actor). `200`, `400`, `401`, `404`,
  `500`.

## [0.14.0] - 2026-08-29

### Agregado

- `GET /private/barbers/{barberId}/appointments/daily-agenda`
  (`operationId: listDailyAgenda`, `CA-062-01` a `CA-062-07`): agenda
  diaria de lectura de un único barbero (`HU-062`), calculada en la zona
  IANA de la barbería. La ruta exige `barberId` explícito, sin vista
  consolidada de varios barberos (`DEC-074`); un turno que cruza
  medianoche aparece en cada agenda diaria cuyo rango interseca su
  intervalo (`DEC-075`). Respuesta mínima sin teléfono, correo, nota ni
  `customerId` (`CA-062-05`). `200`, `400`, `401`, `404`, `500`.

## [0.13.0] - 2026-08-27

### Agregado

- `POST /private/appointments` (`operationId: createManualAppointment`,
  `CA-061-01` a `CA-061-08`): alta de un turno manual `confirmed` a
  partir de un turno recibido por teléfono, WhatsApp o en persona
  (`HU-061`), protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`).
  Sin anticipación mínima ni ventana máxima (`RN-DIS-04`, `DEC-005`,
  `DEC-018`); el servicio debe estar activo y asignado al barbero elegido
  (`DEC-072`); un cruce con otra cita (`RN-CON-01`) o con un bloqueo
  vigente (`DEC-073`) responde `409`. El cliente se reconcilia por
  teléfono o correo (`DEC-045`, `DEC-046`, `DEC-071`). `201`, `400`,
  `401`, `404`, `409` (idempotencia o conflicto), `422`, `500`.

## [0.12.0] - 2026-08-25

### Agregado

- `GET`/`PATCH /private/barbers/{barberId}/holiday-calendar`
  (`operationId: getHolidayCalendar`/`updateHolidayCalendar`,
  `CA-041-01`, `CA-041-02`): interruptor de calendario colombiano de
  festivos por barbero. `200`, `400`, `401`, `404`, `500`.
- `GET /private/barbers/{barberId}/schedule-exceptions`
  (`operationId: listScheduleExceptions`, `CA-041-04`, `CA-041-05`): lista
  paginada por cursor de excepciones de jornada, ordenada por fecha
  efectiva. `200`, `400`, `401`, `404`, `500`.
- `POST /private/barbers/{barberId}/schedule-exceptions`
  (`operationId: createScheduleException`, `CA-041-04`, `CA-041-05`): alta
  de una excepción (día cerrado o abierto con tramos), protegida con
  `Idempotency-Key` (`RN-IDE-01`, `DEC-043`). Una excepción ya existente
  para esa fecha, o tramos que se solapan entre sí, responden `409`
  (`code: conflict`). `201`, `400`, `401`, `404`, `409` (idempotencia o
  conflicto), `422`, `500`.
- `GET /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: getScheduleException`, `CA-041-06`). `200`, `401`,
  `404`, `500`.
- `PATCH /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: updateScheduleException`, `CA-041-04`, `CA-041-05`,
  `CA-041-06`): reemplaza la excepción completa. `200`, `400`, `401`,
  `404`, `409`, `422`, `500`.
- `DELETE /private/barbers/{barberId}/schedule-exceptions/{exceptionId}`
  (`operationId: deleteScheduleException`, `CA-041-06`): retiro físico;
  reintentar tras un `404` es seguro. `204`, `401`, `404`, `500`.
- `GET /private/schedule/colombian-holidays`
  (`operationId: listColombianHolidays`): festivos colombianos de un año
  calendario, calculados de forma determinista (Ley 51 de 1983, "Ley
  Emiliani"), sin tabla ni dependencia nueva. `200`, `400`, `401`, `500`.

No mezcla con horario semanal (`working-hours`, `HU-040`) ni con bloqueos
(`HU-042`). La precedencia de resolución (excepción manual >
festivo automático > horario semanal) vive en un puerto interno de Go, sin
endpoint propio: la disponibilidad (B4) lo consumirá cuando exista.

## [0.11.0] - 2026-08-25

### Agregado

- Tag `Schedule` (HU-040): horario laboral recurrente de cada barbero.
- `GET /private/barbers/{barberId}/working-hours`
  (`operationId: listWorkingHours`, `CA-040-01`): lista paginada por cursor
  de los tramos del barbero de la ruta, ordenada por día ISO y hora de
  inicio. `200` (`WorkingHourListResponse`), `400`, `401`, `404`, `500`.
- `POST /private/barbers/{barberId}/working-hours`
  (`operationId: createWorkingHour`, `CA-040-02`, `CA-040-03`,
  `CA-040-04`): alta de un tramo, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`). Un tramo cuyo `startsTime` + `durationMinutes`
  cruza medianoche es válido (`DEC-020`). Un solape o una hora de inicio
  repetida con otro tramo del mismo barbero y día responde `409`
  (`code: conflict`). `201` (`WorkingHourResponse`), `400`, `401`, `404`,
  `409` (idempotencia o solape), `422`, `500`.
- `GET /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: getWorkingHour`, `CA-040-05`): lectura individual; un
  tramo inexistente, de otro barbero o de otra barbería responde `404`.
  `200`, `401`, `404`, `500`.
- `PATCH /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: updateWorkingHour`, `CA-040-04`, `CA-040-05`): reemplaza
  el intervalo completo del tramo (`isoWeekday`, `startsTime`,
  `durationMinutes` obligatorios). `200`, `400`, `401`, `404`, `409`,
  `422`, `500`.
- `DELETE /private/barbers/{barberId}/working-hours/{workingHourId}`
  (`operationId: deleteWorkingHour`, `CA-040-05`): retiro físico
  (`working_hour` no tiene eliminación lógica); reintentar tras un `404` es
  seguro. `204`, `401`, `404`, `500`.

No mezcla con excepciones por fecha/festivos (`HU-041`) ni con bloqueos
(`HU-042`): sus operaciones llegan con sus propias historias. `CT-008` se
resolvió como `DEC-070` antes de esta migración: las FK del modelo físico
de B2 usan `ON DELETE RESTRICT`, no `CASCADE`.

## [0.10.0] - 2026-08-25

### Agregado

- `GET /private/services/{serviceId}/deactivation-impact` (HU-024,
  `operationId: getServiceDeactivationImpact`, `CA-024-01`): previsualiza el
  impacto real de desactivar el servicio, calculado en el momento de la
  solicitud. Siempre `0` en B1 (`DEC-069`): `appointment` no existe todavía
  en la cadena migrada. `200` (`ServiceDeactivationImpactResponse`), `401`,
  `404`, `500`.
- `POST /private/services/{serviceId}/deactivate`
  (`operationId: deactivateService`, `CA-024-02`, `CA-024-03`, `CA-024-04`):
  transición activo → inactivo, protegida con `Idempotency-Key`
  (`RN-IDE-01`, `DEC-043`). Vuelve a consultar el impacto real dentro de la
  misma operación, sin bloqueo optimista (`DEC-069`). Un servicio ya
  inactivo con una clave nueva responde `409` (`code: conflict`,
  transición inválida). `200` (`ServiceDeactivationResponse`), `400`, `401`,
  `404`, `409` (idempotencia u transición inválida), `500`.
- `POST /private/services/{serviceId}/reactivate`
  (`operationId: reactivateService`, `CA-024-05`): transición inactivo →
  activo, misma protección de idempotencia; nunca crea otra fila ni altera
  duración, precio o asignaciones. Un servicio ya activo con una clave nueva
  responde `409`. `200` (`ServiceResponse`), `400`, `401`, `404`, `409`,
  `500`.
- `ServiceResponse` gana `isActive`/`deactivatedAt`, de solo lectura
  (`CA-024-02`, `CA-024-05`): ningún endpoint de HU-022 (`create`/`update`)
  los acepta como entrada; solo cambian mediante las dos operaciones
  anteriores.
- Esquemas `ServiceDeactivationImpactResponse`, `ServiceDeactivationResponse`.
- Responses `ServiceDeactivationImpact`, `ServiceDeactivated`,
  `ServiceReactivated`.
- Tag `Catalog` ampliado para cubrir también el ciclo de vida de servicios.

## [0.9.0] - 2026-08-24

### Agregado

- `GET /private/barbers/{barberId}/services` (HU-023,
  `operationId: listBarberServices`): lista paginada por cursor de los
  servicios asignados a un barbero de la barbería activa. Un `barberId`
  inexistente o de otra barbería responde `404` uniforme (`RN-TEN-01`).
  `200` (`AssignmentListResponse`), `400`, `401`, `404`, `500`.
- `PUT /private/barbers/{barberId}/services/{serviceId}`
  (`operationId: assignServiceToBarber`): asigna un servicio a un barbero.
  Semántica HTTP naturalmente repetible (sin `Idempotency-Key`): repetir la
  misma operación no crea una segunda fila (`CA-023-02`), responde `200` en
  vez de `201` conservando el `createdAt` original. Un mismo servicio puede
  asignarse a varios barberos (`CA-023-03`). Un `barberId`/`serviceId`
  ajeno o inexistente responde el mismo `404` uniforme (`CA-023-04`).
  `200`/`201` (`AssignmentResponse` + `Location`), `401`, `404`, `500`.
- `DELETE /private/barbers/{barberId}/services/{serviceId}`
  (`operationId: unassignServiceFromBarber`): retira la asignación.
  Retirar la última asignación ACTIVA de un servicio ACTIVO se rechaza con
  `409` (`DEC-068`, `code: conflict`); nunca borra al barbero, al servicio
  ni una cita. `204`, `401`, `404`, `409`, `500`.
- Esquemas `AssignmentResponse`, `AssignmentListResponse`: exactamente
  `barberId`, `serviceId`, `createdAt` (`CA-023-07`) — nunca nombre,
  duración, precio ni estado del barbero o del servicio.
- Responses `AssignmentListSuccess`, `AssignmentAssigned`,
  `AssignmentUnassigned`, `LastActiveAssignmentConflictProblem` (nuevo uso
  de `code: conflict` sobre esta relación, `DEC-068`).
- Tag `Catalog` ampliado para cubrir también la asignación de servicios a
  barberos.

## [0.8.0] - 2026-08-24

### Agregado

- `GET /private/services` (HU-022, `operationId: listServices`): lista
  paginada por cursor de los servicios del catálogo de la barbería activa.
  `200` (`ServiceListResponse`), `400` (cursor/limit inválido), `401`,
  `500`.
- `POST /private/services` (`operationId: createService`): alta de un
  servicio, protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`).
  `201` (`ServiceResponse` + `Location`), `400`, `401`, `409` (conflicto u
  operación en curso de idempotencia, `code: idempotency-conflict`/
  `idempotency-locked`; o nombre ya usado por otro servicio activo de la
  misma barbería, `code: conflict`, `DEC-067`), `422` (nombre/descripción/
  duración/precio inválidos), `500`.
- `GET /private/services/{serviceId}` (`operationId: getService`): lectura
  individual; un identificador inexistente o de otra barbería responde el
  mismo `404` (`CA-022-06`, `RN-TEN-01`).
- `PATCH /private/services/{serviceId}` (`operationId: updateService`):
  edición parcial de `name`/`description`/`durationMinutes`/`price`;
  rechaza objetos vacíos y campos desconocidos (nunca `isActive`,
  `currency`, asignaciones a barberos ni alcance de propagación hacia una
  cita, `RN-SER-04`). `200`, `400`, `401`, `404`, `409` (nombre duplicado),
  `422`, `500`.
- Esquemas `ServiceResponse`, `ServiceListResponse`, `CreateServiceRequest`,
  `UpdateServiceRequest`. `price` viaja como string decimal exacto;
  `currency` siempre `"COP"` (`DEC-067`), informativo, nunca aceptado como
  entrada.
- Responses `ServiceListSuccess`, `ServiceSuccess`, `ServiceCreated`,
  `ServiceUpdated`, `ServiceValidationProblem`, `ServiceConflictProblem`
  (nuevo problem type `conflict`, distinto de `idempotency-conflict`).
- Tag `Catalog`.

## [0.7.0] - 2026-08-23

### Agregado

- `GET /private/barbers` (HU-021, `operationId: listBarbers`): lista
  paginada por cursor de los barberos de la barbería activa. `200`
  (`BarberListResponse`), `400` (cursor/limit inválido), `401`, `500`.
- `POST /private/barbers` (`operationId: createBarber`): alta de un barbero,
  protegida con `Idempotency-Key` (`RN-IDE-01`, `DEC-043`). `201`
  (`BarberResponse` + `Location`), `400`, `401`, `409` (conflicto u
  operación en curso de idempotencia), `422` (nombre vacío/solo
  espacios/mayor de 120 caracteres), `500`. No impone unicidad de
  `fullName`.
- `GET /private/barbers/{barberId}` (`operationId: getBarber`): lectura
  individual; un identificador inexistente o de otra barbería responde el
  mismo `404` (`CA-021-05`, `RN-TEN-01`).
- `PATCH /private/barbers/{barberId}` (`operationId: renameBarber`):
  renombrado limitado a `fullName`; rechaza objetos vacíos y campos
  desconocidos (nunca `active`, `deletedAt`, `sortOrder`, `staffUserId`,
  servicios ni horarios, `DEC-047`). `200`, `400`, `401`, `404`, `422`,
  `500`.
- Esquemas `BarberResponse`, `BarberListResponse`, `CreateBarberRequest`,
  `UpdateBarberRequest`.
- Responses `BarberListSuccess`, `BarberSuccess`, `BarberCreated`,
  `BarberUpdated`, `BarberValidationProblem`.
- Header `Location` (componente compartido, primer uso de un `201` en este
  contrato).
- Tag `Staff`.

## [0.6.0] - 2026-08-23

### Agregado

- `GET /private/settings/barbershop` (HU-020, `operationId:
  getBarbershopSettings`): lectura autenticada de la configuración básica
  de la barbería activa (`name`, `timezone`, `contactEmail`,
  `contactPhone`). `200` (`BarbershopSettingsResponse`), `401`, `404`
  defensivo, `500`.
- `PATCH /private/settings/barbershop` (`operationId:
  updateBarbershopSettings`): actualiza los mismos cuatro campos.
  `barbershopId` nunca es un campo aceptado (`CA-020-05`). `200`
  (representación canónica guardada), `400` (JSON/campo desconocido),
  `401`, `404` defensivo, `422` (validación de campo, incluida zona IANA no
  reconocida, `CA-020-03`), `500`.
- Esquemas `BarbershopSettingsResponse`, `UpdateBarbershopSettingsRequest`.
- Responses `BarbershopSettingsSuccess`, `BarbershopSettingsUpdated`,
  `BarbershopSettingsValidationProblem`.
- Tag `Settings`.

## [0.5.0] - 2026-08-17

### Agregado

- `POST /public/auth/recovery/request` (`operationId: requestRecovery`,
  `DEC-064`/`DEC-065`/`DEC-066`): solicita el código de recuperación de
  acceso. Siempre `202` (`RecoveryRequestAcceptedResponse`), sin destino,
  exista o no la cuenta (no enumeración, `CA-008-01`).
- `POST /public/auth/recovery/verify` (`operationId: verifyRecovery`,
  `DEC-064`/`DEC-065`): verifica el código de 6 dígitos. `200` en éxito
  (`RecoveryVerifyResponse`: token de reinicio de un solo uso + destino
  enmascarado, `CA-008-06`); `401` uniforme para código incorrecto, vencido,
  agotado o cuenta inexistente.
- `POST /public/auth/recovery/reset-password`
  (`operationId: resetPasswordWithRecoveryToken`, `DEC-063`/`DEC-064`):
  establece la contraseña nueva con el token de reinicio. `204` en éxito
  (revoca todas las sesiones activas, `CA-008-05`); `401` uniforme para
  token inválido/vencido/reutilizado; `422` si la contraseña incumple la
  política de `DEC-063`.

## [0.4.0] - 2026-08-17

### Agregado

- `429` en `POST /public/auth/login` (HU-007, `DEC-061`/`DEC-062`): la sexta
  solicitud desde la misma IP dentro de la ventana de 15 minutos, y
  cualquier otra mientras el escalamiento siga vigente, responde
  `ChallengeRequiredProblem` (`code: challenge-required`, cabecera
  `Retry-After`) sin evaluar la contraseña.
- `POST /public/auth/challenge` (`operationId: requestPhoneChallenge`,
  `DEC-062`): solicita el código del reto telefónico. Siempre `202`
  (`ChallengeAcceptedResponse`), exista o no la cuenta (no enumeración).
- `POST /public/auth/challenge/verify` (`operationId: verifyPhoneChallenge`,
  `DEC-062`): verifica el código de 6 dígitos. `204` en éxito (limpia el
  escalamiento de la IP); `401` uniforme (`UnauthorizedProblem`) para código
  incorrecto, vencido, agotado o cuenta inexistente.
- Esquemas `ChallengeRequest`, `ChallengeVerifyRequest`,
  `ChallengeAcceptedResponse`. Responses `ChallengeAccepted`,
  `ChallengeVerifySuccess`, `ChallengeRequiredProblem`. Header `Retry-After`.

## [0.3.0] - 2026-08-17

### Agregado

- `GET /private/auth/session` (HU-012, `operationId: getSessionContext`,
  `DEC-060`): lectura no destructiva del contexto de sesión vigente, con
  seguridad `SessionCookie`. Respuesta `200` (`SessionContextResponse`):
  `barbershop.id`/`barbershop.name` y `expiresAt`, sin `Set-Cookie` (no
  emite ni modifica la cookie, a diferencia de login/logout). Errores `401`
  (sesión ausente/inválida/vencida/revocada, mismo `UnauthorizedProblem`
  uniforme que logout), `500`.
- Esquema `SessionContextResponse`.
- Response `SessionContextSuccess`.

### Nota

- Este registro no documenta la adición de `POST /private/auth/logout`
  (HU-006) como una versión propia: quedó dentro de la entrada `0.2.0` de
  abajo aunque `openapi.yaml` ya reflejaba `version: 0.2.0` desde antes de
  esa operación. Se detectó al preparar esta entrada; se deja constancia
  aquí en vez de reescribir el historial de una versión ya publicada.

## [0.2.0] - 2026-08-13

### Agregado

- `POST /public/auth/login` (HU-005, `operationId: loginWithPassword`):
  inicio de sesión con correo y contraseña, sin seguridad (`security: []`,
  `DEC-055`). Respuesta `200` sin secretos (`LoginResponse`, solo
  `expiresAt`); la sesión se fija en `Set-Cookie`, documentada sin su valor
  funcional. Errores `400` (JSON/campo desconocido), `401` (credenciales
  inválidas, unificado para correo inexistente/contraseña
  incorrecta/usuario inactivo), `422` (validación de campo), `500`.
- Esquemas `LoginRequest`, `LoginResponse`.
- Responses `LoginSuccess`, `UnauthorizedProblem`, `ValidationProblem`.
- Header `Set-Cookie` (`SetCookieSession.yaml`).
- Esquema de seguridad `SessionCookie` (`components/security-schemes/`), sin
  registrar todavía en `openapi.yaml` porque ninguna operación lo referencia
  aún (HU-006 lo activa junto con la primera operación privada real).
- Tag `Auth`.

## [0.1.0] - 2026-08-07

`openapi.yaml` existe como documento de entrada sin operaciones (HU-003).
