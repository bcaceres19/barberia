// Modelo de dominio de la recuperación de acceso (HU-008, DEC-063–DEC-066).
// Mismo patrón que `loginOutcome.ts`/`challengeOutcome.ts`: cada función de
// `api/recoveryApi.ts` traduce la respuesta real a una de estas uniones
// discriminadas; ningún componente ve `Problem`, `status` HTTP crudo ni
// cabeceras.

/** Resultado de `POST /recovery/request`. Siempre `accepted` ante un 202
 * (DEC-065: no enumeración — el servidor responde igual exista o no la
 * cuenta). `network-error`/`unexpected-error` cubren un fallo real de
 * transporte (la solicitud nunca llegó o el servidor no la procesó), que no
 * revela nada sobre la cuenta y por eso sí se distingue de `accepted`. */
export type RecoveryRequestOutcome =
  | { kind: 'accepted' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }

/** Resultado de `POST /recovery/verify`. `invalid-code` cubre código
 * incorrecto, vencido, agotado o cuenta inexistente: el servidor no
 * distingue el motivo (DEC-064/DEC-065), así que este tipo tampoco lo hace
 * (mismo patrón que `ChallengeVerifyOutcome` de HU-007). Un `verified`
 * exitoso trae el token de reinicio de un solo uso y el destino
 * enmascarado (CA-008-06): esta es la única vez que el destino aparece. */
export type RecoveryVerifyOutcome =
  | { kind: 'verified'; resetToken: string; maskedPhone: string; maskedEmail: string }
  | { kind: 'invalid-code' }
  | { kind: 'validation-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }

/** Resultado de `POST /recovery/reset-password`. `invalid-token` cubre
 * token de reinicio desconocido, vencido, consumido o de otra cuenta (401,
 * mismo error uniforme). `policy-violation` cubre un 422: el backend
 * mapea todo incumplimiento de `DEC-063` al mismo `code` genérico de
 * validación (`apperr.KindValidation`), así que el cliente no puede
 * distinguir cuál regla exacta fue la que falló solo por la respuesta;
 * por eso la política se explica ANTES de escribir (CA-011-06) y la
 * validación de forma en cliente cubre lo que sí puede anticiparse
 * (longitud, igual al correo). */
export type RecoveryResetPasswordOutcome =
  | { kind: 'success' }
  | { kind: 'invalid-token' }
  | { kind: 'policy-violation' }
  | { kind: 'validation-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
