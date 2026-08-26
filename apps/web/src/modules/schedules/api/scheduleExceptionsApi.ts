// Único punto del módulo `schedules` que llama al cliente HTTP tipado para
// las operaciones de HU-041 (docs/03-desarrollo/estandar-frontend-vue.md
// §6): calendario de festivos, excepciones de jornada y festivos
// colombianos de referencia. Traduce las respuestas reales a los outcomes
// discriminados que la página consume; ningún componente ve `Problem`,
// `status` HTTP crudo ni cabeceras. Mismo criterio que schedulesApi.ts
// (HU-040).
import { httpClient } from '@/shared/api/httpClient'
import type {
  ColombianHoliday,
  ScheduleException,
  ScheduleExceptionPage,
  ScheduleExceptionSegment,
} from '../model/scheduleException'
import type {
  CreateScheduleExceptionOutcome,
  DeleteScheduleExceptionOutcome,
  FetchColombianHolidaysOutcome,
  FetchHolidayCalendarOutcome,
  FetchScheduleExceptionsOutcome,
  UpdateHolidayCalendarOutcome,
  UpdateScheduleExceptionOutcome,
} from '../model/exceptionOutcome'

// EXCEPTIONS_LIMIT cubre en una sola página cualquier configuración real de
// un barbero (unas pocas excepciones por año), evitando paginar
// internamente (mismo criterio que WORKING_HOURS_LIMIT).
const EXCEPTIONS_LIMIT = 50

export interface ScheduleExceptionSegmentInput {
  startsTime: string
  durationMinutes: number
}

export async function fetchHolidayCalendar(barberId: string): Promise<FetchHolidayCalendarOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/barbers/{barberId}/holiday-calendar',
      {
        params: { path: { barberId } },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', enabled: data.enabled }
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

export async function updateHolidayCalendar(
  barberId: string,
  enabled: boolean,
): Promise<UpdateHolidayCalendarOutcome> {
  try {
    const { data, response } = await httpClient.PATCH(
      '/private/barbers/{barberId}/holiday-calendar',
      {
        params: { path: { barberId } },
        body: { enabled },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', enabled: data.enabled }
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

export async function fetchScheduleExceptions(
  barberId: string,
): Promise<FetchScheduleExceptionsOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/barbers/{barberId}/schedule-exceptions',
      { params: { path: { barberId }, query: { limit: EXCEPTIONS_LIMIT } } },
    )

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

export async function createScheduleException(
  barberId: string,
  effectiveDate: string,
  isClosed: boolean,
  reason: string | null,
  segments: ScheduleExceptionSegmentInput[],
  idempotencyKey: string,
): Promise<CreateScheduleExceptionOutcome> {
  try {
    const { data, response, error } = await httpClient.POST(
      '/private/barbers/{barberId}/schedule-exceptions',
      {
        params: { path: { barberId }, header: { 'Idempotency-Key': idempotencyKey } },
        body: { effectiveDate, isClosed, reason, segments },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', exception: toException(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        // RN-IDE-01 (idempotency-conflict/idempotency-locked) frente a
        // CA-041-05 (conflict, fecha duplicada/solape): el cliente decide
        // por `code`, nunca por `detail` (docs/06-api/estandar-openapi.md
        // §5).
        return isProblemCode(error, 'conflict')
          ? { kind: 'date-conflict' }
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

export async function updateScheduleException(
  barberId: string,
  exceptionId: string,
  effectiveDate: string,
  isClosed: boolean,
  reason: string | null,
  segments: ScheduleExceptionSegmentInput[],
): Promise<UpdateScheduleExceptionOutcome> {
  try {
    const { data, response } = await httpClient.PATCH(
      '/private/barbers/{barberId}/schedule-exceptions/{exceptionId}',
      {
        params: { path: { barberId, exceptionId } },
        body: { effectiveDate, isClosed, reason, segments },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', exception: toException(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'date-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function deleteScheduleException(
  barberId: string,
  exceptionId: string,
): Promise<DeleteScheduleExceptionOutcome> {
  try {
    const { response } = await httpClient.DELETE(
      '/private/barbers/{barberId}/schedule-exceptions/{exceptionId}',
      { params: { path: { barberId, exceptionId } } },
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
    return { kind: 'unexpected-error' }
  }
}

// fetchColombianHolidays no depende de ningún barbero ni excepción: es un
// dato de referencia igual para toda la barbería (RN-BLQ-02).
export async function fetchColombianHolidays(year: number): Promise<FetchColombianHolidaysOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/schedule/colombian-holidays', {
      params: { query: { year } },
    })
    if (response.ok && data) {
      return { kind: 'success', items: data.items }
    }
    return { kind: 'unavailable' }
  } catch {
    return { kind: 'unavailable' }
  }
}

function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function toException(data: {
  id: string
  effectiveDate: string
  isClosed: boolean
  reason: string | null
  segments: ScheduleExceptionSegment[]
  createdAt: string
  updatedAt: string
}): ScheduleException {
  return {
    id: data.id,
    effectiveDate: data.effectiveDate,
    isClosed: data.isClosed,
    reason: data.reason,
    segments: data.segments.map((seg) => ({
      id: seg.id,
      startsTime: seg.startsTime,
      durationMinutes: seg.durationMinutes,
    })),
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
  }
}

function toPage(data: {
  items: Parameters<typeof toException>[0][]
  nextCursor: string | null
}): ScheduleExceptionPage {
  return { items: data.items.map(toException), nextCursor: data.nextCursor }
}

export type { ColombianHoliday }
