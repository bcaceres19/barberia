// Modelo de dominio del reto telefónico de HU-007 (DEC-062). Misma
// separación que `loginOutcome.ts`: la página/componente nunca ven `Problem`
// ni `status` HTTP crudo, solo estas uniones discriminadas.

/** Resultado de `POST /auth/challenge`. Siempre `accepted` cuando el
 * servidor respondió 202 (DEC-062: no enumeración — `accepted` no implica
 * que se haya enviado un mensaje real, solo que la solicitud se procesó). */
export type ChallengeRequestOutcome =
  | { kind: 'accepted' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }

/** Resultado de `POST /auth/challenge/verify`. `invalid-code` cubre código
 * incorrecto, vencido, agotado o cuenta inexistente: el servidor no
 * distingue el motivo (DEC-062), así que este tipo tampoco lo hace. */
export type ChallengeVerifyOutcome =
  | { kind: 'verified' }
  | { kind: 'invalid-code' }
  | { kind: 'validation-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
