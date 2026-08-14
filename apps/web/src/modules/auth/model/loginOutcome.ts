// Modelo de dominio del intento de acceso. `LoginCredentials` es lo único
// que el formulario conoce; `LoginOutcome` es lo único que la página
// coordinadora recibe del cliente API. Ninguno de los dos expone la forma
// RFC 9457 (`Problem`) del contrato: eso queda dentro de `api/loginApi.ts`
// (docs/03-desarrollo/estandar-frontend-vue.md §6, "la vista no conoce la
// forma interna de errores de cada proveedor").

/** Credenciales tal como las escribe el barbero. Nunca se serializan a
 * `sessionStorage`, `localStorage` ni a la URL (CA-010-07). */
export interface LoginCredentials {
  email: string
  password: string
}

/** Resultado discriminado de un intento de acceso contra el API real.
 * Cada variante corresponde a un estado observable exigido por HU-010:
 * `success` (CA-010-01), `invalid-credentials` (CA-010-02),
 * `network-error` (CA-010-03), `rate-limited` (429 documentado, ver
 * trabajo requerido §8) y `unexpected-error`/`validation-error` como
 * fronteras defensivas del contrato. */
export type LoginOutcome =
  | { kind: 'success'; expiresAt: string }
  | { kind: 'invalid-credentials' }
  | { kind: 'validation-error' }
  | { kind: 'rate-limited'; retryAfterSeconds?: number }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
