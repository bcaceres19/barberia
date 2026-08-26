// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6) para HU-041: ayuda a corregir antes de enviar, nunca sustituye al
// backend (CA-041-04/05 reales viven en el servicio Go). No valida fecha
// duplicada ni solape entre excepciones: esas comprobaciones dependen del
// estado ya persistido y solo el servidor lo conoce con certeza (mismo
// criterio que scheduleValidation.ts, HU-040).
import { validateDurationMinutes, validateStartsTime } from './scheduleValidation'

export const MAX_REASON_LENGTH = 200

const EFFECTIVE_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/

export function validateEffectiveDate(effectiveDate: string): string | undefined {
  if (!EFFECTIVE_DATE_PATTERN.test(effectiveDate)) {
    return 'Elige una fecha.'
  }
  // Confirma que la fecha existe de verdad (por ejemplo, rechaza
  // 2027-02-30): el mismo criterio de reconstrucción que
  // schedule.ValidateEffectiveDate en el backend.
  const [year, month, day] = effectiveDate.split('-').map(Number) as [number, number, number]
  const parsed = new Date(Date.UTC(year, month - 1, day))
  const reconstructed =
    parsed.getUTCFullYear() === year &&
    parsed.getUTCMonth() === month - 1 &&
    parsed.getUTCDate() === day
  if (!reconstructed) {
    return 'Esa fecha no existe.'
  }
  return undefined
}

export function validateReason(reason: string): string | undefined {
  if (reason.length > MAX_REASON_LENGTH) {
    return `El motivo no puede superar ${MAX_REASON_LENGTH} caracteres.`
  }
  return undefined
}

export interface ExceptionSegmentDraft {
  startsTime: string
  durationMinutes: number
}

// validateExceptionShape aplica CA-041-04: cerrada nunca admite tramos;
// abierta exige al menos un tramo válido, sin solapes entre sí (semántica
// semiabierta [inicio, fin), permite cruzar medianoche, mismo criterio que
// schedule.ValidateExceptionShape).
export function validateExceptionShape(
  isClosed: boolean,
  segments: ExceptionSegmentDraft[],
): string | undefined {
  if (isClosed) {
    return undefined
  }
  if (segments.length === 0) {
    return 'Agrega al menos un tramo, o marca el día como cerrado.'
  }
  for (const seg of segments) {
    if (validateStartsTime(seg.startsTime) || validateDurationMinutes(seg.durationMinutes)) {
      return 'Revisa la hora y la duración de cada tramo.'
    }
  }
  const intervals = segments
    .map((seg) => ({ start: toMinutes(seg.startsTime), duration: seg.durationMinutes }))
    .sort((a, b) => a.start - b.start)
  for (let i = 0; i < intervals.length - 1; i++) {
    const current = intervals[i]!
    const next = intervals[i + 1]!
    if (current.start + current.duration > next.start) {
      return 'Dos tramos se solapan entre sí.'
    }
  }
  return undefined
}

function toMinutes(startsTime: string): number {
  const [hours, minutes] = startsTime.split(':').map(Number) as [number, number]
  return hours * 60 + minutes
}
