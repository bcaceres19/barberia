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

// Variante de BaseBadge para cada uno de los cinco estados (CA-062-03):
// cada etiqueta lleva color Y texto, nunca solo color. Debe ir por la prop
// `variant`, no por una clase CSS externa: BaseBadge calcula
// --badge-surface/--badge-text/--badge-border a partir de `variant` y los
// aplica por `:style` en línea (shared/ui/BaseBadge.vue) — un estilo en
// línea siempre gana sobre cualquier clase CSS externa que intente
// redefinir esa misma custom property, así que una clase
// "base-badge--status-confirmed" añadida por fuera nunca tenía efecto
// (bug preexistente: cada insignia salía con la paleta neutra por
// defecto sin importar el estado, issue #189).
export const APPOINTMENT_STATUS_BADGE_VARIANT: Record<
  AppointmentStatus,
  'neutral' | 'success' | 'warning' | 'danger' | 'info'
> = {
  confirmed: 'info',
  completed: 'success',
  cancelled_by_customer: 'neutral',
  cancelled_by_barber: 'danger',
  no_show: 'warning',
}

// isTerminalStatus separa las citas activas de las terminales
// (RN-CIT-04/estados-citas.md §2.2): las terminales conservan su fila y
// legibilidad (nunca desaparecen), pero HU-062 puede atenuarlas.
export function isTerminalStatus(status: AppointmentStatus): boolean {
  return status !== 'confirmed'
}
