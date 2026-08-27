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
import type {
  CreateManualAppointmentOutcome,
  CreatedManualAppointment,
  FetchAssignedServicesOutcome,
  FetchBarberSummariesOutcome,
  FetchBarbershopTimezoneOutcome,
} from '../model/appointmentOutcome'

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
