// Resultados discriminados de las dos operaciones de HU-093, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6). Un
// 401 real de estas rutas ya lo intercepta la coordinación única de
// `installSessionHandling` (redirige a acceso); ambos tipos igual declaran
// `unexpected-error` como frontera defensiva. `version-conflict` es
// EXCLUSIVO de la escritura (CA-093-02): distinto de `validation-error`
// porque la recuperación correcta es recargar, no corregir un campo.
import type { BookingPolicy } from './bookingPolicy'

export type FetchBookingPolicyOutcome =
  | { kind: 'success'; policy: BookingPolicy }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SaveBookingPolicyOutcome =
  | { kind: 'success'; policy: BookingPolicy }
  | { kind: 'validation-error' }
  | { kind: 'version-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
