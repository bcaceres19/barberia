# Registro de cambios del contrato OpenAPI

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Ver [`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
sección 18 para qué cuenta como cambio compatible o incompatible.

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
