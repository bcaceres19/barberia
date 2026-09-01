// Detalle e historial de un turno (HU-064). Vocabulario cerrado, mismo
// texto que `booking.EventType`/`booking.ActorType` en el backend (nunca
// números, DEC-016).
import type { AppointmentOrigin, AppointmentStatus } from './dailyAgenda'

// AppointmentDetail (CA-064-01 a CA-064-04) es la lectura completa de un
// turno: nunca trae `customerId` ni ningún identificador de actor
// (fuera de alcance del contrato). versionToken es opaco: la interfaz nunca
// lo decodifica ni lo muestra, solo lo conserva para una precondición
// futura.
export type AppointmentDetail = {
  id: string
  barberId: string
  barberFullName: string
  attendeeName: string
  customerFullName: string
  customerPhone: string | null
  customerEmail: string | null
  customerNote: string | null
  startsAt: string
  endsAt: string
  status: AppointmentStatus
  origin: AppointmentOrigin
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
  versionToken: string
  createdAt: string
}

export type HistoryEventType =
  | 'appointment_created'
  | 'appointment_rescheduled'
  | 'appointment_service_changed'
  | 'appointment_completed'
  | 'appointment_cancelled_by_customer'
  | 'appointment_cancelled_by_barber'
  | 'appointment_no_show'
  | 'appointment_status_corrected'

export type HistoryActorType = 'staff' | 'customer' | 'system'

export type HistoryChange = {
  fieldName: string
  previousValue: string | null
  newValue: string | null
}

// HistoryEntry (CA-064-05): actorLabel ya llega resuelto y seguro desde el
// servidor (nunca correo ni ID interno, RN-HIS-01); la interfaz solo lo
// muestra tal cual.
export type HistoryEntry = {
  id: string
  eventType: HistoryEventType
  actorType: HistoryActorType
  actorLabel: string
  reason: string | null
  occurredAt: string
  changes: HistoryChange[]
}

// Etiquetas en español de los ocho eventos cerrados (DEC-041): la interfaz
// nunca muestra el texto técnico en inglés de la API.
export const HISTORY_EVENT_LABELS: Record<HistoryEventType, string> = {
  appointment_created: 'Turno creado',
  appointment_rescheduled: 'Turno reprogramado',
  appointment_service_changed: 'Servicio cambiado',
  appointment_completed: 'Turno completado',
  appointment_cancelled_by_customer: 'Cancelado por el cliente',
  appointment_cancelled_by_barber: 'Cancelado por el barbero',
  appointment_no_show: 'Cliente no se presentó',
  appointment_status_corrected: 'Estado corregido',
}

// Etiquetas de los campos que appointment_history_change puede traer
// (RN-HIS-01): un mapeo exhaustivo y tipado, nunca el nombre técnico crudo
// en pantalla (trabajo requerido §3.4).
export const HISTORY_FIELD_LABELS: Record<string, string> = {
  status: 'Estado',
  starts_at: 'Hora de inicio',
  ends_at: 'Hora de fin',
  barber_id: 'Barbero',
  service_id: 'Servicio',
  service_name_snapshot: 'Servicio',
  duration_minutes_snapshot: 'Duración',
  price_amount_snapshot: 'Precio',
  cancellation_reason: 'Motivo de cancelación',
}

export function historyFieldLabel(fieldName: string): string {
  return HISTORY_FIELD_LABELS[fieldName] ?? fieldName
}
