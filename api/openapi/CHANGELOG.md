# Registro de cambios del contrato OpenAPI

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Ver [`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
sección 18 para qué cuenta como cambio compatible o incompatible.

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
