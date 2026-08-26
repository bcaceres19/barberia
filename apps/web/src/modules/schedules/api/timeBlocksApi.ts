// Único punto del módulo `schedules` que llama al cliente HTTP tipado para
// HU-042 (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las
// respuestas reales de time-blocks/time-block-series a los outcomes
// discriminados que la página consume; ningún componente ve `Problem`,
// `status` HTTP crudo ni cabeceras. Cubre el alta/listado/retiro de
// bloqueos puntuales y series recurrentes; los sub-recursos de fechas
// explícitas/excepciones y la edición de series (scope whole/
// this_and_following) ya existen en el backend pero todavía no tienen
// pantalla propia (ver prompt HU-042, seguimiento pendiente).
import { httpClient } from '@/shared/api/httpClient'
import type {
  BlockType,
  TimeBlock,
  TimeBlockPage,
  TimeBlockSeries,
  TimeBlockSeriesPage,
} from '../model/timeBlock'
import type {
  CreateTimeBlockOutcome,
  CreateTimeBlockSeriesOutcome,
  DeleteTimeBlockOutcome,
  DeleteTimeBlockSeriesOutcome,
  FetchTimeBlockSeriesOutcome,
  FetchTimeBlocksOutcome,
} from '../model/blockOutcome'

// LIST_LIMIT cubre en una sola página cualquier configuración real de un
// barbero (mismo criterio que WORKING_HOURS_LIMIT).
const LIST_LIMIT = 50

export async function fetchTimeBlocks(barberId: string): Promise<FetchTimeBlocksOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers/{barberId}/time-blocks', {
      params: { path: { barberId }, query: { limit: LIST_LIMIT } },
    })
    if (response.ok && data) {
      return { kind: 'success', page: toTimeBlockPage(data) }
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

export async function createTimeBlock(
  barberId: string,
  blockType: BlockType,
  startsAt: string,
  endsAt: string,
  reason: string | null,
  idempotencyKey: string,
): Promise<CreateTimeBlockOutcome> {
  try {
    const { data, response } = await httpClient.POST('/private/barbers/{barberId}/time-blocks', {
      params: { path: { barberId }, header: { 'Idempotency-Key': idempotencyKey } },
      body: { blockType, startsAt, endsAt, reason },
    })
    if (response.ok && data) {
      return { kind: 'success', block: toTimeBlock(data) }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function deleteTimeBlock(
  barberId: string,
  blockId: string,
): Promise<DeleteTimeBlockOutcome> {
  try {
    const { response } = await httpClient.DELETE(
      '/private/barbers/{barberId}/time-blocks/{blockId}',
      {
        params: { path: { barberId, blockId } },
      },
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

export async function fetchTimeBlockSeries(barberId: string): Promise<FetchTimeBlockSeriesOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/barbers/{barberId}/time-block-series',
      {
        params: { path: { barberId }, query: { limit: LIST_LIMIT } },
      },
    )
    if (response.ok && data) {
      return { kind: 'success', page: toSeriesPage(data) }
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

export async function createTimeBlockSeries(
  barberId: string,
  blockType: BlockType,
  isoWeekday: number,
  startsTime: string,
  durationMinutes: number,
  effectiveFrom: string,
  reason: string | null,
  idempotencyKey: string,
): Promise<CreateTimeBlockSeriesOutcome> {
  try {
    const { data, response } = await httpClient.POST(
      '/private/barbers/{barberId}/time-block-series',
      {
        params: { path: { barberId }, header: { 'Idempotency-Key': idempotencyKey } },
        body: {
          blockType,
          recurrenceKind: 'weekly',
          isoWeekday,
          startsTime,
          durationMinutes,
          effectiveFrom,
          reason,
        },
      },
    )
    if (response.ok && data) {
      return { kind: 'success', series: toSeries(data) }
    }
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function deleteTimeBlockSeries(
  barberId: string,
  seriesId: string,
): Promise<DeleteTimeBlockSeriesOutcome> {
  try {
    const { response } = await httpClient.DELETE(
      '/private/barbers/{barberId}/time-block-series/{seriesId}',
      { params: { path: { barberId, seriesId } } },
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

function toTimeBlock(data: {
  id: string
  blockType: string
  source: string
  startsAt: string
  endsAt: string
  reason: string | null
  deletedAt: string | null
  deletedBy: string | null
  createdAt: string
  updatedAt: string
}): TimeBlock {
  return {
    id: data.id,
    blockType: data.blockType as BlockType,
    source: data.source as TimeBlock['source'],
    startsAt: data.startsAt,
    endsAt: data.endsAt,
    reason: data.reason,
    deletedAt: data.deletedAt,
    deletedBy: data.deletedBy,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
  }
}

function toTimeBlockPage(data: {
  items: Parameters<typeof toTimeBlock>[0][]
  nextCursor: string | null
}): TimeBlockPage {
  return { items: data.items.map(toTimeBlock), nextCursor: data.nextCursor }
}

function toSeries(data: {
  id: string
  blockType: string
  recurrenceKind: string
  isoWeekday: number | null
  startsTime: string
  durationMinutes: number
  effectiveFrom: string
  effectiveUntil: string | null
  reason: string | null
  deletedAt: string | null
  createdAt: string
  updatedAt: string
  dates: { blockDate: string }[]
  exceptions: { excludedDate: string; reason: string | null; createdAt: string }[]
}): TimeBlockSeries {
  return {
    id: data.id,
    blockType: data.blockType as BlockType,
    recurrenceKind: data.recurrenceKind as TimeBlockSeries['recurrenceKind'],
    isoWeekday: data.isoWeekday,
    startsTime: data.startsTime,
    durationMinutes: data.durationMinutes,
    effectiveFrom: data.effectiveFrom,
    effectiveUntil: data.effectiveUntil,
    reason: data.reason,
    deletedAt: data.deletedAt,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
    dates: data.dates,
    exceptions: data.exceptions,
  }
}

function toSeriesPage(data: {
  items: Parameters<typeof toSeries>[0][]
  nextCursor: string | null
}): TimeBlockSeriesPage {
  return { items: data.items.map(toSeries), nextCursor: data.nextCursor }
}
