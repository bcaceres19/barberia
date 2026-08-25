// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend
// (CA-040-04 real vive en el servicio Go). No valida solape: esa
// comprobación depende del estado ya persistido de otros tramos y solo el
// servidor lo conoce con certeza (mismo criterio que DEC-067/DEC-068 en
// otros módulos).
export const MIN_DURATION_MINUTES = 1
export const MAX_DURATION_MINUTES = 1440

const STARTS_TIME_PATTERN = /^([01]\d|2[0-3]):[0-5]\d$/

export function validateISOWeekday(isoWeekday: number): string | undefined {
  if (!Number.isInteger(isoWeekday) || isoWeekday < 1 || isoWeekday > 7) {
    return 'Elige un día de la semana.'
  }
  return undefined
}

export function validateStartsTime(startsTime: string): string | undefined {
  if (!STARTS_TIME_PATTERN.test(startsTime)) {
    return 'Escribe una hora válida (HH:MM, 24 horas).'
  }
  return undefined
}

export function validateDurationMinutes(durationMinutes: number): string | undefined {
  if (
    !Number.isInteger(durationMinutes) ||
    durationMinutes < MIN_DURATION_MINUTES ||
    durationMinutes > MAX_DURATION_MINUTES
  ) {
    return `La duración debe ser mayor que cero y no exceder ${MAX_DURATION_MINUTES} minutos.`
  }
  return undefined
}
