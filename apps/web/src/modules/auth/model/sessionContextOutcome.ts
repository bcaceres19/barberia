import type { SessionContext } from './sessionContext'

/**
 * Resultado discriminado de consultar GET /private/auth/session
 * (DEC-060). `network-error` y `unexpected-error` se distinguen del mismo
 * modo que `LoginOutcome` (api/loginApi.ts): un `fetch` que ni siquiera
 * obtuvo respuesta HTTP frente a un estado inesperado con respuesta real.
 */
export type SessionContextOutcome =
  | ({ kind: 'authenticated' } & SessionContext)
  | { kind: 'unauthenticated' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
