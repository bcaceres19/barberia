// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend
// (CA-042 real vive en el servicio Go). No valida solape ni conflicto de
// negocio: RN-BLQ-03 exige que un bloqueo puntual nunca falle por eso, y el
// resto depende del estado ya persistido, que solo el servidor conoce con
// certeza. Reutiliza MIN_DURATION_MINUTES/MAX_DURATION_MINUTES y
// validateISOWeekday/validateStartsTime de scheduleValidation.ts: mismo
// límite [1,1440] y misma forma HH:MM que working_hour/time_block_series.
import { validateISOWeekday, validateStartsTime } from './scheduleValidation'

export { validateISOWeekday, validateStartsTime }

const MAX_REASON_LENGTH = 200

export function validateReason(reason: string): string | undefined {
  if (reason.length > MAX_REASON_LENGTH) {
    return `El motivo no puede exceder ${MAX_REASON_LENGTH} caracteres.`
  }
  return undefined
}

// validateInstant confirma que raw es una marca de tiempo ISO 8601 válida
// (el input datetime-local del navegador ya produce esta forma sin offset;
// la pantalla la reinterpreta en la zona IANA de la barbería antes de
// enviarla, CA-042).
export function validateInstant(raw: string): string | undefined {
  if (!raw || Number.isNaN(Date.parse(raw))) {
    return 'Escribe una fecha y hora válidas.'
  }
  return undefined
}

// validateBlockInterval exige que endsAt sea estrictamente posterior a
// startsAt (intervalo semiabierto [starts, ends), RN-DIS-05). Ambos ya deben
// haber pasado validateInstant.
export function validateBlockInterval(startsAt: string, endsAt: string): string | undefined {
  if (Date.parse(endsAt) <= Date.parse(startsAt)) {
    return 'La hora de fin debe ser posterior a la hora de inicio.'
  }
  return undefined
}

// validateEffectiveFrom confirma la forma civil YYYY-MM-DD (el input date
// del navegador ya produce esta forma).
export function validateEffectiveFrom(raw: string): string | undefined {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(raw) || Number.isNaN(Date.parse(raw))) {
    return 'Escribe una fecha válida.'
  }
  return undefined
}
