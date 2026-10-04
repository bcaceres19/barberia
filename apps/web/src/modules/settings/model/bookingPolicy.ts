// Espejo tipado del cuerpo real del contrato (HU-093, CA-093-01):
// exactamente los seis campos de configuración más versionToken, el único
// dato de concurrencia expuesto.
export interface BookingPolicy {
  minAdvanceMinutes: number
  maxAdvanceDays: number
  slotGridMinutes: number
  cancellationDeadlineMinutes: number
  lateCancellationClientAllowed: boolean
  lateCancellationReasonRequired: boolean
  versionToken: string
}

/** Forma editable del formulario: los seis campos de configuración, sin
 * versionToken (viaja aparte, como precondición de la siguiente
 * escritura, nunca como un campo editable del formulario). */
export interface BookingPolicyFormValues {
  minAdvanceMinutes: number
  maxAdvanceDays: number
  slotGridMinutes: number
  cancellationDeadlineMinutes: number
  lateCancellationClientAllowed: boolean
  lateCancellationReasonRequired: boolean
}

/** Valores permitidos de la rejilla (DEC-083): un conjunto discreto, no un
 * rango continuo. El formulario ofrece un `<select>`, nunca un número
 * libre. */
export const ALLOWED_SLOT_GRID_MINUTES = [5, 10, 15, 20, 30, 60] as const

export function toFormValues(policy: BookingPolicy): BookingPolicyFormValues {
  return {
    minAdvanceMinutes: policy.minAdvanceMinutes,
    maxAdvanceDays: policy.maxAdvanceDays,
    slotGridMinutes: policy.slotGridMinutes,
    cancellationDeadlineMinutes: policy.cancellationDeadlineMinutes,
    lateCancellationClientAllowed: policy.lateCancellationClientAllowed,
    lateCancellationReasonRequired: policy.lateCancellationReasonRequired,
  }
}

/** Duración legible para la vista previa: «45 min», «1 h 30 min», «2 días 3 h». */
export function formatDuration(minutes: number): string {
  const total = Number.isFinite(minutes) ? Math.max(0, Math.round(minutes)) : 0
  const days = Math.floor(total / 1440)
  const hours = Math.floor((total % 1440) / 60)
  const mins = total % 60
  const parts: string[] = []
  if (days > 0) parts.push(`${days} ${days === 1 ? 'día' : 'días'}`)
  if (hours > 0) parts.push(`${hours} h`)
  if (mins > 0) parts.push(`${mins} min`)
  return parts.length > 0 ? parts.join(' ') : '0 min'
}

/** Qué fracción (0–1) de la ventana de reserva queda cerrada por la
 * anticipación mínima. Valores no numéricos o fuera de rango se acotan: el
 * formulario puede estar a medio escribir. */
export function blockedShareOfWindow(minAdvanceMinutes: number, maxAdvanceDays: number): number {
  const windowMinutes = Number.isFinite(maxAdvanceDays) ? maxAdvanceDays * 1440 : 0
  if (windowMinutes <= 0 || !Number.isFinite(minAdvanceMinutes)) return 0
  return Math.min(1, Math.max(0, minAdvanceMinutes / windowMinutes))
}

/** Primeras franjas de una mañana de ejemplo (desde las 10:00) con la rejilla
 * elegida: lo mismo que ve el cliente en la tira de horarios. */
export function sampleSlotLabels(slotGridMinutes: number, limit = 8): string[] {
  if (!Number.isFinite(slotGridMinutes) || slotGridMinutes <= 0) return []
  const startMinutes = 10 * 60
  const spanMinutes = 3 * 60
  const count = Math.min(limit, Math.floor(spanMinutes / slotGridMinutes) + 1)
  return Array.from({ length: count }, (_, index) => {
    const total = startMinutes + index * slotGridMinutes
    const hours = String(Math.floor(total / 60)).padStart(2, '0')
    const mins = String(total % 60).padStart(2, '0')
    return `${hours}:${mins}`
  })
}
