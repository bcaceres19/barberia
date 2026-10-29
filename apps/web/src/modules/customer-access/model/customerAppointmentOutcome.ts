// Modelo de dominio de la lectura del turno del cliente por token (HU-098).
// `CustomerAppointment` es la forma que la página consume;
// `GetCustomerAppointmentOutcome` es lo único que
// `api/getCustomerAppointmentApi.ts` entrega. Ninguno de los dos expone la
// forma RFC 9457 (`Problem`) del contrato (docs/03-desarrollo/
// estandar-frontend-vue.md §6): un token con forma inválida, inexistente,
// vencido o revocado llegan TODOS como `not-found` (CA-098-02), nunca
// distinguidos.

import type { Vocabulary, VocabularyTerms } from '@/shared/model'

/** Estado vigente del turno (docs/02-requisitos/estados-citas.md). Se
 * duplica aquí en vez de importar `modules/agenda` (CA-002-06, mismo
 * criterio que publicbooking frente a booking en el backend): los módulos
 * de frontend no se importan entre sí, solo desde `shared`. */
export type AppointmentStatus =
  'confirmed' | 'completed' | 'cancelled_by_customer' | 'cancelled_by_barber' | 'no_show'

const FIXED_STATUS_LABELS: Record<Exclude<AppointmentStatus, 'cancelled_by_barber'>, string> = {
  confirmed: 'Confirmado',
  completed: 'Completado',
  cancelled_by_customer: 'Cancelado por el cliente',
  no_show: 'No se presentó',
}

// El estado técnico `cancelled_by_barber` no se renombra (DEC-016); su etiqueta
// nombra a quien atendía con la palabra de la barbería (DEC-119).
export function appointmentStatusLabel(status: AppointmentStatus, v: Vocabulary): string {
  return status === 'cancelled_by_barber'
    ? `Cancelado ${v.byTheProfessional}`
    : FIXED_STATUS_LABELS[status]
}

// Variante de BaseBadge para cada estado (RN-CNF-02, CA-098-04): el color
// nunca es la única señal, el texto de appointmentStatusLabel siempre lo
// acompaña. Ninguna variante insinúa que la falta de respuesta cancele el
// turno: `confirmed` es 'info', nunca 'warning'/'danger'.
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

/** La reseña que el cliente dejó sobre su turno (HU-100): sin identificadores. */
export interface CustomerReview {
  rating: number
  comment: string | null
  createdAt: string
}

/** Lectura mínima del turno del cliente (CA-098-01, CA-098-03): nunca un
 * identificador interno, contacto del cliente, nota ni historial. */
export interface CustomerAppointment {
  barbershopName: string
  timezone: string
  attendeeName: string
  serviceName: string
  durationMinutes: number
  barberName: string
  startsAt: string
  endsAt: string
  status: AppointmentStatus
  cancellationDeadlineMinutes: number
  lateCancellationClientAllowed: boolean
  lateCancellationReasonRequired: boolean
  /** Token opaco de versión: se reenvía tal cual como `If-Match` al cancelar (HU-099). */
  versionToken: string
  /** Contacto público opcional de la barbería: adonde acudir si la política rechaza cancelar. */
  barbershopContactEmail: string | null
  barbershopContactPhone: string | null
  /** Palabras con las que la barbería nombra a su negocio y a su profesional (DEC-119). */
  vocabulary: VocabularyTerms
  /** «Calificar» solo se ofrece con el turno completado y sin reseña (HU-100, CA-100-01). */
  canReview: boolean
  /** La reseña que el cliente ya dejó sobre este turno, o `null`. */
  review: CustomerReview | null
}

/** Resultado discriminado de resolver un token de acceso contra el API
 * real. `not-found` cubre las cuatro causas indistinguibles de CA-098-02
 * (forma inválida, inexistente, vencido, revocado); `network-error` y
 * `unexpected-error` son las mismas fronteras defensivas que el resto de la
 * aplicación ya usa (ver public-booking/model/barbershopProfileOutcome.ts). */
export type GetCustomerAppointmentOutcome =
  | { kind: 'success'; appointment: CustomerAppointment }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }

/** Lo que el cliente envía al calificar (HU-100, CA-100-02). */
export interface ReviewInput {
  rating: number
  /** Texto libre ya recortado; `null` si no escribió nada. */
  comment: string | null
}

/**
 * Resultado de calificar el turno (HU-100). `not-reviewable` es el `409`
 * uniforme (turno no completado, cancelado o ya calificado); `not-found` el `404`
 * del enlace; `validation-error` el `422`. Ninguno expone la forma RFC 9457.
 */
export type CreateReviewOutcome =
  | { kind: 'success'; review: CustomerReview }
  | { kind: 'validation-error' }
  | { kind: 'not-reviewable' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }

/** Rechazo de política de la cancelación que el servidor devolvió y la interfaz explica. */
export type CancellationRejection = 'window-closed' | 'reason-required' | null

/** Lo que el cliente envía al cancelar (HU-099). */
export interface CancelAppointmentInput {
  /** Texto libre ya recortado; `null` si no escribió nada. */
  reason: string | null
}

/**
 * Resultado de cancelar el turno (HU-099), discriminado por `status`/`code` y
 * nunca por `detail`. `window-closed` y `reason-required` son los dos rechazos
 * de política (422); `changed` cubre los 409 de versión obsoleta o estado ya no
 * confirmado (la página recarga para mostrar lo vigente); `not-found` es el 404
 * uniforme del enlace (también el de un enlace ya revocado por una cancelación
 * previa). `validation-error` es el motivo demasiado largo o un 400.
 */
export type CancelAppointmentOutcome =
  | { kind: 'success' }
  | { kind: 'window-closed' }
  | { kind: 'reason-required' }
  | { kind: 'validation-error' }
  | { kind: 'changed' }
  | { kind: 'idempotency-in-progress' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
