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
