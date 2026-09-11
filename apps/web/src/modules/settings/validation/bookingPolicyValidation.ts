// Validación de forma en cliente (docs/03-desarrollo/estandar-frontend-vue.md
// §7.6): ayuda a corregir antes de enviar, nunca sustituye al backend. Los
// rangos y el conjunto discreto son EXACTAMENTE los de DEC-083
// (barbershop_*_ck en la base): el cliente no inventa un límite más
// estricto ni más laxo que el servidor.
import { ALLOWED_SLOT_GRID_MINUTES } from '../model/bookingPolicy'

export const MIN_ADVANCE_MINUTES_FLOOR = 0
export const MIN_ADVANCE_MINUTES_CEIL = 1440
export const MAX_ADVANCE_DAYS_FLOOR = 1
export const MAX_ADVANCE_DAYS_CEIL = 90
export const CANCELLATION_DEADLINE_MINUTES_FLOOR = 0
export const CANCELLATION_DEADLINE_MINUTES_CEIL = 10080

export function validateMinAdvanceMinutes(value: number): string | undefined {
  if (
    !Number.isInteger(value) ||
    value < MIN_ADVANCE_MINUTES_FLOOR ||
    value > MIN_ADVANCE_MINUTES_CEIL
  ) {
    return `Escribe un número entero entre ${MIN_ADVANCE_MINUTES_FLOOR} y ${MIN_ADVANCE_MINUTES_CEIL} minutos.`
  }
  return undefined
}

export function validateMaxAdvanceDays(value: number): string | undefined {
  if (!Number.isInteger(value) || value < MAX_ADVANCE_DAYS_FLOOR || value > MAX_ADVANCE_DAYS_CEIL) {
    return `Escribe un número entero entre ${MAX_ADVANCE_DAYS_FLOOR} y ${MAX_ADVANCE_DAYS_CEIL} días.`
  }
  return undefined
}

export function validateSlotGridMinutes(value: number): string | undefined {
  if (!ALLOWED_SLOT_GRID_MINUTES.includes(value as (typeof ALLOWED_SLOT_GRID_MINUTES)[number])) {
    return `Elige uno de los pasos permitidos: ${ALLOWED_SLOT_GRID_MINUTES.join(', ')} minutos.`
  }
  return undefined
}

export function validateCancellationDeadlineMinutes(value: number): string | undefined {
  if (
    !Number.isInteger(value) ||
    value < CANCELLATION_DEADLINE_MINUTES_FLOOR ||
    value > CANCELLATION_DEADLINE_MINUTES_CEIL
  ) {
    return `Escribe un número entero entre ${CANCELLATION_DEADLINE_MINUTES_FLOOR} y ${CANCELLATION_DEADLINE_MINUTES_CEIL} minutos.`
  }
  return undefined
}

/**
 * CA-093-02: exigir motivo para una cancelación tardía que el cliente ni
 * siquiera puede hacer es una combinación incoherente (DEC-083).
 */
export function validateLateCancellationCoherence(
  clientAllowed: boolean,
  reasonRequired: boolean,
): string | undefined {
  if (reasonRequired && !clientAllowed) {
    return 'No puedes exigir motivo si el cliente no puede cancelar tarde.'
  }
  return undefined
}

/**
 * Mismo criterio que barbershop_min_advance_vs_window_ck: la anticipación
 * mínima no puede alcanzar ni superar toda la ventana pública de reserva.
 */
export function validateAdvanceWithinWindow(
  minAdvanceMinutes: number,
  maxAdvanceDays: number,
): string | undefined {
  if (minAdvanceMinutes >= maxAdvanceDays * 1440) {
    return 'La anticipación mínima no puede alcanzar ni superar la ventana máxima.'
  }
  return undefined
}
