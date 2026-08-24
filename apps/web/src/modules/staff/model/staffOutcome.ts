// Resultados discriminados de las operaciones de HU-021, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6),
// mismo criterio que settingsOutcome.ts. Un 401 real de estas rutas ya lo
// intercepta la coordinación única de `installSessionHandling`; todos
// igual declaran `unexpected-error` como frontera defensiva por si esa
// solicitud concreta llegara a resolverse antes de que la redirección
// navegue.
import type { Barber, BarberPage } from './barber'

export type FetchBarbersOutcome =
  { kind: 'success'; page: BarberPage } | { kind: 'network-error' } | { kind: 'unexpected-error' }

export type CreateBarberOutcome =
  | { kind: 'success'; barber: Barber }
  | { kind: 'validation-error' }
  // RN-IDE-01: la misma clave de idempotencia ya se usó con un contenido
  // distinto, o la operación con esa clave sigue en curso (DEC-043). No es
  // un error de validación de campo: el formulario no cambió, es un
  // reintento el que debe generar una clave nueva.
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type RenameBarberOutcome =
  | { kind: 'success'; barber: Barber }
  | { kind: 'validation-error' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
