// Único punto del módulo `agenda` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de POST /private/appointments y de los selectores de barbero/
// servicio/zona a los outcomes discriminados que la página consume; ningún
// componente ve `Problem`, `status` HTTP crudo ni cabeceras.
//
// fetchBarberSummaries/fetchBarbershopTimezone llaman DIRECTAMENTE al
// cliente HTTP compartido (nunca a schedulesApi.ts/settingsApi.ts,
// privados de sus propios módulos): mismo criterio que
// schedulesApi.fetchBarberSummaries frente a staffApi.
import { httpClient } from '@/shared/api/httpClient'
import type { AppointmentDetail, HistoryEntry } from '../model/appointmentDetail'
import type {
  CancelAppointmentOutcome,
  CloseAppointmentOutcome,
  CreateManualAppointmentOutcome,
  CreatedManualAppointment,
  FetchAppointmentDetailOutcome,
  FetchAppointmentHistoryOutcome,
  FetchAssignedServicesOutcome,
  FetchBarberSummariesOutcome,
  FetchBarbershopTimezoneOutcome,
  FetchDailyAgendaOutcome,
  RescheduleAppointmentOutcome,
} from '../model/appointmentOutcome'
import type { DailyAgendaEntry } from '../model/dailyAgenda'

const PICKER_LIMIT = 50

export async function fetchBarberSummaries(): Promise<FetchBarberSummariesOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers', {
      params: { query: { limit: PICKER_LIMIT } },
    })
    if (response.ok && data) {
      return { kind: 'success', items: data.items.map((b) => ({ id: b.id, fullName: b.fullName })) }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

// fetchAssignedServices lista solo los servicios activos ya asignados a
// barberId (DEC-072): el formulario nunca ofrece un servicio que el
// servidor rechazaría.
export async function fetchAssignedServices(
  barberId: string,
): Promise<FetchAssignedServicesOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers/{barberId}/services', {
      params: { path: { barberId }, query: { limit: PICKER_LIMIT } },
    })
    if (response.ok && data) {
      const servicesById = await fetchServiceNames(data.items.map((a) => a.serviceId))
      return {
        kind: 'success',
        items: data.items
          .map((a) => ({ id: a.serviceId, name: servicesById.get(a.serviceId) ?? a.serviceId }))
          .filter((s) => servicesById.has(s.id)),
      }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// fetchServiceNames resuelve el catálogo completo una sola vez y filtra en
// memoria a los activos: GET /private/barbers/{barberId}/services (HU-023)
// solo expone barberId/serviceId/createdAt (CA-023-07), nunca el nombre, a
// propósito, para que un consumidor cruce por identificador contra el
// catálogo real.
async function fetchServiceNames(serviceIds: string[]): Promise<Map<string, string>> {
  const byId = new Map<string, string>()
  if (serviceIds.length === 0) return byId
  try {
    const { data, response } = await httpClient.GET('/private/services', {
      params: { query: { limit: PICKER_LIMIT } },
    })
    if (response.ok && data) {
      for (const s of data.items) {
        if (s.isActive) byId.set(s.id, s.name)
      }
    }
  } catch {
    // Un catálogo ilegible deja el selector vacío para ese barbero
    // (fetchAssignedServices filtra por servicesById.has); la página ya
    // trata una lista vacía como estado explícito, no como error oculto.
  }
  return byId
}

export async function fetchBarbershopTimezone(): Promise<FetchBarbershopTimezoneOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/barbershop')
    if (response.ok && data) {
      return { kind: 'success', timezone: data.timezone }
    }
    return { kind: 'unavailable' }
  } catch {
    return { kind: 'unavailable' }
  }
}

// fetchDailyAgenda (HU-062): sin `date`, el servidor decide "hoy" en la
// zona de la barbería (CA-062-01); esta función nunca calcula ni envía una
// fecha por defecto propia.
export async function fetchDailyAgenda(
  barberId: string,
  date?: string,
): Promise<FetchDailyAgendaOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/barbers/{barberId}/appointments/daily-agenda',
      {
        params: { path: { barberId }, query: date ? { date } : {} },
      },
    )
    if (response.ok && data) {
      return { kind: 'success', items: data.items.map(toDailyAgendaEntry) }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toDailyAgendaEntry(data: {
  id: string
  attendeeName: string
  startsAt: string
  endsAt: string
  status: string
  origin: string
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
}): DailyAgendaEntry {
  return {
    id: data.id,
    attendeeName: data.attendeeName,
    startsAt: data.startsAt,
    endsAt: data.endsAt,
    status: data.status as DailyAgendaEntry['status'],
    origin: data.origin as DailyAgendaEntry['origin'],
    serviceName: data.serviceName,
    durationMinutes: data.durationMinutes,
    priceAmount: data.priceAmount,
    currency: data.currency,
  }
}

// fetchAppointmentDetail (HU-064, CA-064-01 a CA-064-04).
export async function fetchAppointmentDetail(
  appointmentId: string,
): Promise<FetchAppointmentDetailOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/appointments/{appointmentId}', {
      params: { path: { appointmentId } },
    })
    if (response.ok && data) {
      return { kind: 'success', detail: toAppointmentDetail(data) }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toAppointmentDetail(data: {
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
  status: string
  origin: string
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
  versionToken: string
  createdAt: string
}): AppointmentDetail {
  return {
    id: data.id,
    barberId: data.barberId,
    barberFullName: data.barberFullName,
    attendeeName: data.attendeeName,
    customerFullName: data.customerFullName,
    customerPhone: data.customerPhone,
    customerEmail: data.customerEmail,
    customerNote: data.customerNote,
    startsAt: data.startsAt,
    endsAt: data.endsAt,
    status: data.status as AppointmentDetail['status'],
    origin: data.origin as AppointmentDetail['origin'],
    serviceName: data.serviceName,
    durationMinutes: data.durationMinutes,
    priceAmount: data.priceAmount,
    currency: data.currency,
    versionToken: data.versionToken,
    createdAt: data.createdAt,
  }
}

// fetchAppointmentHistory (HU-064, CA-064-05): cursor vacío pide la primera
// página, mismo criterio que fetchDailyAgenda frente a `date`.
export async function fetchAppointmentHistory(
  appointmentId: string,
  cursor?: string,
): Promise<FetchAppointmentHistoryOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/appointments/{appointmentId}/history',
      {
        params: { path: { appointmentId }, query: cursor ? { cursor } : {} },
      },
    )
    if (response.ok && data) {
      return {
        kind: 'success',
        items: data.items.map(toHistoryEntry),
        nextCursor: data.nextCursor,
      }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toHistoryEntry(data: {
  id: string
  eventType: string
  actorType: string
  actorLabel: string
  reason: string | null
  occurredAt: string
  changes: { fieldName: string; previousValue: string | null; newValue: string | null }[]
}): HistoryEntry {
  return {
    id: data.id,
    eventType: data.eventType as HistoryEntry['eventType'],
    actorType: data.actorType as HistoryEntry['actorType'],
    actorLabel: data.actorLabel,
    reason: data.reason,
    occurredAt: data.occurredAt,
    changes: data.changes.map((c) => ({
      fieldName: c.fieldName,
      previousValue: c.previousValue,
      newValue: c.newValue,
    })),
  }
}

// rescheduleAppointment (HU-065, T2): versionToken es el token opaco leído
// del detalle (HU-064), enviado como precondición `If-Match`; el servidor
// responde `version-conflict` si la representación ya cambió.
export async function rescheduleAppointment(
  appointmentId: string,
  input: { startsAt: string },
  versionToken: string,
  idempotencyKey: string,
): Promise<RescheduleAppointmentOutcome> {
  try {
    const { data, response, error } = await httpClient.POST(
      '/private/appointments/{appointmentId}/reschedule',
      {
        params: {
          path: { appointmentId },
          header: { 'Idempotency-Key': idempotencyKey, 'If-Match': versionToken },
        },
        body: { startsAt: input.startsAt },
      },
    )
    if (response.ok && data) {
      return { kind: 'success', startsAt: data.startsAt }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        if (isProblemCode(error, 'version-conflict')) return { kind: 'version-conflict' }
        if (isProblemCode(error, 'invalid-state')) return { kind: 'invalid-state' }
        if (
          isProblemCode(error, 'idempotency-conflict') ||
          isProblemCode(error, 'idempotency-locked')
        ) {
          return { kind: 'idempotency-conflict' }
        }
        return { kind: 'conflict', detail: problemDetail(error) }
      case 422:
        return { kind: 'validation-error', detail: problemDetail(error) }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// cancelAppointmentByBarber (HU-066, T6): versionToken es el token opaco
// leído del detalle (HU-064), enviado como precondición `If-Match`; el
// servidor responde `version-conflict` si la representación ya cambió.
// Sin cuerpo de solicitud: T6 no acepta ningún campo (el servidor deriva
// tenant/actor/estado destino).
export async function cancelAppointmentByBarber(
  appointmentId: string,
  versionToken: string,
  idempotencyKey: string,
): Promise<CancelAppointmentOutcome> {
  try {
    const { response, error } = await httpClient.POST(
      '/private/appointments/{appointmentId}/cancel',
      {
        params: {
          path: { appointmentId },
          header: { 'Idempotency-Key': idempotencyKey, 'If-Match': versionToken },
        },
      },
    )
    if (response.ok) {
      return { kind: 'success' }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        if (isProblemCode(error, 'version-conflict')) return { kind: 'version-conflict' }
        if (isProblemCode(error, 'invalid-state')) return { kind: 'invalid-state' }
        if (
          isProblemCode(error, 'idempotency-conflict') ||
          isProblemCode(error, 'idempotency-locked')
        ) {
          return { kind: 'idempotency-conflict' }
        }
        return { kind: 'unexpected-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// closeAppointment es el helper privado compartido por completeAppointment
// y markAppointmentNoShow (HU-067, T4 manual y T7): ambas rutas comparten
// forma de cabeceras, cuerpo vacío y traducción de errores, solo difieren
// en el path. versionToken es el token opaco leído del detalle (HU-064),
// enviado como precondición `If-Match`.
async function closeAppointment(
  path:
    | '/private/appointments/{appointmentId}/complete'
    | '/private/appointments/{appointmentId}/no-show',
  appointmentId: string,
  versionToken: string,
  idempotencyKey: string,
): Promise<CloseAppointmentOutcome> {
  try {
    const { response, error } = await httpClient.POST(path, {
      params: {
        path: { appointmentId },
        header: { 'Idempotency-Key': idempotencyKey, 'If-Match': versionToken },
      },
    })
    if (response.ok) {
      return { kind: 'success' }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        if (isProblemCode(error, 'version-conflict')) return { kind: 'version-conflict' }
        if (isProblemCode(error, 'invalid-state')) return { kind: 'invalid-state' }
        if (
          isProblemCode(error, 'idempotency-conflict') ||
          isProblemCode(error, 'idempotency-locked')
        ) {
          return { kind: 'idempotency-conflict' }
        }
        return { kind: 'unexpected-error' }
      case 422:
        return { kind: 'validation-error', detail: problemDetail(error) }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// completeAppointment (HU-067, T4 manual): cierra un turno `confirmed` cuyo
// `startsAt` ya pasó como `completed`. Sin cuerpo de solicitud: el servidor
// deriva tenant/actor/estado destino.
export async function completeAppointment(
  appointmentId: string,
  versionToken: string,
  idempotencyKey: string,
): Promise<CloseAppointmentOutcome> {
  return closeAppointment(
    '/private/appointments/{appointmentId}/complete',
    appointmentId,
    versionToken,
    idempotencyKey,
  )
}

// markAppointmentNoShow (HU-067, T7): cierra un turno `confirmed` cuyo
// `startsAt` ya pasó como `no_show`. Mismo criterio que completeAppointment.
export async function markAppointmentNoShow(
  appointmentId: string,
  versionToken: string,
  idempotencyKey: string,
): Promise<CloseAppointmentOutcome> {
  return closeAppointment(
    '/private/appointments/{appointmentId}/no-show',
    appointmentId,
    versionToken,
    idempotencyKey,
  )
}

export async function createManualAppointment(
  input: {
    barberId: string
    serviceId: string
    attendeeName: string
    customerFullName: string
    customerPhone: string | null
    customerEmail: string | null
    customerNote: string | null
    startsAt: string
  },
  idempotencyKey: string,
): Promise<CreateManualAppointmentOutcome> {
  try {
    const { data, response, error } = await httpClient.POST('/private/appointments', {
      params: { header: { 'Idempotency-Key': idempotencyKey } },
      body: {
        barberId: input.barberId,
        serviceId: input.serviceId,
        attendeeName: input.attendeeName,
        customerFullName: input.customerFullName,
        customerPhone: input.customerPhone,
        customerEmail: input.customerEmail,
        customerNote: input.customerNote,
        startsAt: input.startsAt,
      },
    })
    if (response.ok && data) {
      return { kind: 'success', appointment: toAppointment(data) }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        if (
          isProblemCode(error, 'idempotency-conflict') ||
          isProblemCode(error, 'idempotency-locked')
        ) {
          return { kind: 'idempotency-conflict' }
        }
        return { kind: 'conflict', detail: problemDetail(error) }
      case 422:
        return { kind: 'validation-error', detail: problemDetail(error) }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function problemDetail(error: unknown): string {
  if (
    error &&
    typeof error === 'object' &&
    typeof (error as { detail?: string }).detail === 'string'
  ) {
    return (error as { detail: string }).detail
  }
  return 'No pudimos completar la operación.'
}

function toAppointment(data: {
  id: string
  barberId: string
  serviceId: string
  attendeeName: string
  startsAt: string
  endsAt: string
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
}): CreatedManualAppointment {
  return {
    id: data.id,
    barberId: data.barberId,
    serviceId: data.serviceId,
    attendeeName: data.attendeeName,
    startsAt: data.startsAt,
    endsAt: data.endsAt,
    serviceName: data.serviceName,
    durationMinutes: data.durationMinutes,
    priceAmount: data.priceAmount,
    currency: data.currency,
  }
}
