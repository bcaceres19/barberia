# Registro de cambios del contrato OpenAPI

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Ver [`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
sección 18 para qué cuenta como cambio compatible o incompatible.

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
