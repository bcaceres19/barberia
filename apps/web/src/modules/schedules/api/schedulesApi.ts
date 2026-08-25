// Único punto del módulo `schedules` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/POST /private/barbers/{barberId}/working-hours y
// GET/PATCH/DELETE .../working-hours/{workingHourId} a los outcomes
// discriminados que la página consume; ningún componente ve `Problem`,
// `status` HTTP crudo ni cabeceras.
//
// fetchBarberSummaries llama DIRECTAMENTE al cliente HTTP compartido (nunca
// a staffApi.ts, privado de su propio módulo): mismo criterio que
// barberServicesApi.fetchBarberSummaries.
import { httpClient } from '@/shared/api/httpClient'
import type { BarberSummary, WorkingHour, WorkingHourPage } from '../model/workingHour'
import type {
  CreateWorkingHourOutcome,
  DeleteWorkingHourOutcome,
  FetchBarberSummariesOutcome,
  FetchBarbershopTimezoneOutcome,
  FetchWorkingHoursOutcome,
  UpdateWorkingHourOutcome,
} from '../model/scheduleOutcome'

// PICKER_LIMIT es el máximo que el contrato admite por página
// (docs/06-api/estandar-openapi.md §6.10), mismo criterio que
// barberServicesApi.PICKER_LIMIT.
const PICKER_LIMIT = 50

// WORKING_HOURS_LIMIT cubre en una sola página cualquier configuración real
// de un barbero (como mucho unos pocos tramos por día, 7 días): evita
// paginar internamente una pantalla que muestra la semana completa a la vez
// (CA-040-01).
const WORKING_HOURS_LIMIT = 50

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

// fetchBarbershopTimezone llama DIRECTAMENTE al cliente HTTP compartido
// (nunca a settingsApi.ts, privado de su propio módulo, mismo criterio que
// fetchBarberSummaries): CA-040-06 exige mostrar la zona IANA de la
// barbería junto a la hora, nunca la del dispositivo.
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

export async function fetchWorkingHours(barberId: string): Promise<FetchWorkingHoursOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers/{barberId}/working-hours', {
      params: { path: { barberId }, query: { limit: WORKING_HOURS_LIMIT } },
    })

    if (response.ok && data) {
      return { kind: 'success', page: toPage(data) }
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

export async function createWorkingHour(
  barberId: string,
  isoWeekday: number,
  startsTime: string,
  durationMinutes: number,
  idempotencyKey: string,
): Promise<CreateWorkingHourOutcome> {
  try {
    const { data, response, error } = await httpClient.POST(
      '/private/barbers/{barberId}/working-hours',
      {
        params: { path: { barberId }, header: { 'Idempotency-Key': idempotencyKey } },
        body: { isoWeekday, startsTime, durationMinutes },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', workingHour: toWorkingHour(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        // RN-IDE-01 (idempotency-conflict/idempotency-locked) frente a
        // CA-040-04 (conflict, solape): el cliente decide por `code`, nunca
        // por `detail` (docs/06-api/estandar-openapi.md §5).
        return isProblemCode(error, 'conflict')
          ? { kind: 'overlap-conflict' }
          : { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function updateWorkingHour(
  barberId: string,
  workingHourId: string,
  isoWeekday: number,
  startsTime: string,
  durationMinutes: number,
): Promise<UpdateWorkingHourOutcome> {
  try {
    const { data, response } = await httpClient.PATCH(
      '/private/barbers/{barberId}/working-hours/{workingHourId}',
      {
        params: { path: { barberId, workingHourId } },
        body: { isoWeekday, startsTime, durationMinutes },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', workingHour: toWorkingHour(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'overlap-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function deleteWorkingHour(
  barberId: string,
  workingHourId: string,
): Promise<DeleteWorkingHourOutcome> {
  try {
    const { response } = await httpClient.DELETE(
      '/private/barbers/{barberId}/working-hours/{workingHourId}',
      { params: { path: { barberId, workingHourId } } },
    )

    if (response.ok) {
      return { kind: 'success' }
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

function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function toWorkingHour(data: {
  id: string
  isoWeekday: number
  startsTime: string
  durationMinutes: number
  createdAt: string
  updatedAt: string
}): WorkingHour {
  return {
    id: data.id,
    isoWeekday: data.isoWeekday,
    startsTime: data.startsTime,
    durationMinutes: data.durationMinutes,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
  }
}

function toPage(data: {
  items: {
    id: string
    isoWeekday: number
    startsTime: string
    durationMinutes: number
    createdAt: string
    updatedAt: string
  }[]
  nextCursor: string | null
}): WorkingHourPage {
  return { items: data.items.map(toWorkingHour), nextCursor: data.nextCursor }
}

export type { BarberSummary }
