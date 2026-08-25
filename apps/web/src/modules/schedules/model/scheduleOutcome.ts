// Resultados discriminados de las operaciones de HU-040, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6),
// mismo criterio que staffOutcome.ts/barberServicesOutcome.ts. Un 401 real
// de estas rutas ya lo intercepta la coordinación única de
// `installSessionHandling`; todos igual declaran `unexpected-error` como
// frontera defensiva.
import type { BarberSummary, WorkingHour, WorkingHourPage } from './workingHour'

export type FetchBarberSummariesOutcome =
  | { kind: 'success'; items: BarberSummary[] }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

// CA-040-06: la pantalla siempre indica la zona IANA de la barbería, nunca
// la del dispositivo; un fallo al obtenerla no bloquea la pantalla (la hora
// civil ya viaja correcta de todas formas, sin conversión alguna), así que
// esta consulta no tiene un desenlace de error propio que la página deba
// tratar como bloqueante.
export type FetchBarbershopTimezoneOutcome =
  { kind: 'success'; timezone: string } | { kind: 'unavailable' }

export type FetchWorkingHoursOutcome =
  | { kind: 'success'; page: WorkingHourPage }
  // El barbero seleccionado ya no existe/es de otra barbería (RN-TEN-01):
  // la pantalla lo trata igual que una carga fallida.
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type CreateWorkingHourOutcome =
  | { kind: 'success'; workingHour: WorkingHour }
  | { kind: 'validation-error' }
  // CA-040-04: el intervalo se solapa con otro tramo existente del mismo
  // barbero y día, o repite exactamente su hora de inicio.
  | { kind: 'overlap-conflict' }
  | { kind: 'not-found' }
  // RN-IDE-01: la misma clave de idempotencia ya se usó con un contenido
  // distinto, o la operación con esa clave sigue en curso (DEC-043).
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type UpdateWorkingHourOutcome =
  | { kind: 'success'; workingHour: WorkingHour }
  | { kind: 'validation-error' }
  | { kind: 'overlap-conflict' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type DeleteWorkingHourOutcome =
  | { kind: 'success' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
