// Pista local de la política de cancelación (HU-099, RN-CAN-01, RN-CAN-02) para
// decidir qué ofrecer ANTES de llamar al API. Es solo una pista con el reloj del
// dispositivo: la decisión real es del servidor con su propio reloj, y su rechazo
// (`window-closed`, `reason-required`) siempre prevalece sobre esta pista.
import type { CustomerAppointment } from './customerAppointmentOutcome'

/** Tope del motivo (appointment.cancellation_reason, 500 caracteres). */
export const CANCELLATION_REASON_MAX_LENGTH = 500

export type CancellationAvailability =
  /** Dentro del plazo: se puede cancelar sin motivo obligatorio. */
  | { mode: 'within-deadline'; reasonRequired: false }
  /** Pasado el plazo pero la barbería permite cancelar; puede exigir motivo. */
  | { mode: 'late-allowed'; reasonRequired: boolean }
  /** El plazo venció y la barbería no permite cancelar tarde: contactar. */
  | { mode: 'unavailable'; reason: 'late-not-allowed' | 'started' }

/**
 * Misma regla que el servidor: hasta `startsAt - plazo` INCLUSIVE es dentro de
 * plazo; con plazo 0 toda cancelación es tardía (DEC-083); desde `startsAt` ya no
 * procede. `nowMs` se inyecta para poder probar antes, exacto y después.
 */
export function cancellationAvailability(
  appointment: Pick<
    CustomerAppointment,
    | 'startsAt'
    | 'cancellationDeadlineMinutes'
    | 'lateCancellationClientAllowed'
    | 'lateCancellationReasonRequired'
  >,
  nowMs: number,
): CancellationAvailability {
  const startsAtMs = Date.parse(appointment.startsAt)
  if (nowMs >= startsAtMs) return { mode: 'unavailable', reason: 'started' }

  const deadlineMs = startsAtMs - appointment.cancellationDeadlineMinutes * 60_000
  if (appointment.cancellationDeadlineMinutes > 0 && nowMs <= deadlineMs) {
    return { mode: 'within-deadline', reasonRequired: false }
  }
  if (!appointment.lateCancellationClientAllowed) {
    return { mode: 'unavailable', reason: 'late-not-allowed' }
  }
  return { mode: 'late-allowed', reasonRequired: appointment.lateCancellationReasonRequired }
}
