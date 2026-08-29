// Vocabulario cerrado de estados_citas.md §2, mismo texto que
// `booking.Status` en el backend (nunca números, DEC-016/estados-citas.md
// §11).
export type AppointmentStatus =
  'confirmed' | 'completed' | 'cancelled_by_customer' | 'cancelled_by_barber' | 'no_show'

export type AppointmentOrigin = 'public' | 'manual'

// DailyAgendaEntry (HU-062, CA-062-05) es la proyección mínima que la
// agenda diaria muestra: nunca teléfono, correo, nota ni customerId.
export type DailyAgendaEntry = {
  id: string
  attendeeName: string
  startsAt: string
  endsAt: string
  status: AppointmentStatus
  origin: AppointmentOrigin
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
}

// Etiquetas en español de los cinco estados (CA-062-03): la interfaz nunca
// muestra el texto técnico en inglés de la API.
export const APPOINTMENT_STATUS_LABELS: Record<AppointmentStatus, string> = {
  confirmed: 'Confirmado',
  completed: 'Completado',
  cancelled_by_customer: 'Cancelado por el cliente',
  cancelled_by_barber: 'Cancelado por el barbero',
  no_show: 'No se presentó',
}

// Clases de BaseBadge ya definidas en el sistema visual
// (shared/ui/BaseBadge.vue, "Estados de turno") para los cinco estados:
// cada etiqueta lleva color Y texto, nunca solo color (CA-062-03,
// estandar-diseno-visual.md).
export const APPOINTMENT_STATUS_BADGE_CLASS: Record<AppointmentStatus, string> = {
  confirmed: 'base-badge--status-confirmed',
  completed: 'base-badge--status-completed',
  cancelled_by_customer: 'base-badge--status-cancelled-customer',
  cancelled_by_barber: 'base-badge--status-cancelled-barber',
  no_show: 'base-badge--status-no-show',
}

// isTerminalStatus separa las citas activas de las terminales
// (RN-CIT-04/estados-citas.md §2.2): las terminales conservan su fila y
// legibilidad (nunca desaparecen), pero HU-062 puede atenuarlas.
export function isTerminalStatus(status: AppointmentStatus): boolean {
  return status !== 'confirmed'
}
