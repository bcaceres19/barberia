// Resultados discriminados de las operaciones de HU-041, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6),
// mismo criterio que scheduleOutcome.ts (HU-040).
import type {
  ColombianHoliday,
  ScheduleException,
  ScheduleExceptionPage,
} from './scheduleException'

export type FetchHolidayCalendarOutcome =
  | { kind: 'success'; enabled: boolean }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type UpdateHolidayCalendarOutcome =
  | { kind: 'success'; enabled: boolean }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchScheduleExceptionsOutcome =
  | { kind: 'success'; page: ScheduleExceptionPage }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type CreateScheduleExceptionOutcome =
  | { kind: 'success'; exception: ScheduleException }
  | { kind: 'validation-error' }
  // CA-041-05: ya existe otra excepción del mismo barbero para esa fecha.
  | { kind: 'date-conflict' }
  | { kind: 'not-found' }
  // RN-IDE-01: la misma clave de idempotencia ya se usó con un contenido
  // distinto, o la operación con esa clave sigue en curso (DEC-043).
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type UpdateScheduleExceptionOutcome =
  | { kind: 'success'; exception: ScheduleException }
  | { kind: 'validation-error' }
  | { kind: 'date-conflict' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type DeleteScheduleExceptionOutcome =
  | { kind: 'success' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

// CA-041-01/02: un fallo al listar festivos colombianos de referencia no
// bloquea la pantalla (es solo un atajo para prellenar la fecha), así que
// no distingue causas de error.
export type FetchColombianHolidaysOutcome =
  { kind: 'success'; items: ColombianHoliday[] } | { kind: 'unavailable' }
