/**
 * Pruebas del cliente tipado de series de bloqueo (HU-042, #100): las
 * operaciones de fechas explícitas, excepciones, edición por alcance y alta
 * por fechas se mapean por `status`, nunca por `detail`.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const postMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
const deleteMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: vi.fn(), POST: postMock, PATCH: patchMock, DELETE: deleteMock },
}))

const {
  createDateListSeries,
  updateTimeBlockSeries,
  addSeriesDate,
  removeSeriesDate,
  addSeriesException,
  removeSeriesException,
} = await import('../timeBlocksApi')

function result(status: number, body?: unknown) {
  return {
    data: status < 300 ? body : undefined,
    error: status >= 300 ? {} : undefined,
    response: new Response(null, { status }),
  }
}

const seriesBody = {
  id: 'ser-1',
  blockType: 'vacation',
  recurrenceKind: 'date_list',
  isoWeekday: null,
  startsTime: '00:00',
  durationMinutes: 1440,
  effectiveFrom: '2027-12-01',
  effectiveUntil: '2027-12-31',
  reason: null,
  deletedAt: null,
  createdAt: '2026-10-01T12:00:00Z',
  updatedAt: '2026-10-01T12:00:00Z',
  dates: [{ blockDate: '2027-12-15' }],
  exceptions: [],
}

const dateListInput = {
  blockType: 'vacation' as const,
  startsTime: '00:00',
  durationMinutes: 1440,
  effectiveFrom: '2027-12-01',
  effectiveUntil: '2027-12-31',
  reason: null,
  explicitDates: ['2027-12-15'],
}

describe('timeBlocksApi.createDateListSeries', () => {
  beforeEach(() => postMock.mockReset())

  it('sends recurrenceKind date_list with the explicit dates and the idempotency key', async () => {
    postMock.mockResolvedValueOnce(result(201, seriesBody))

    const outcome = await createDateListSeries('b-1', dateListInput, 'key-1')

    expect(outcome).toMatchObject({ kind: 'success', series: { id: 'ser-1' } })
    expect(postMock).toHaveBeenCalledWith('/private/barbers/{barberId}/time-block-series', {
      params: { path: { barberId: 'b-1' }, header: { 'Idempotency-Key': 'key-1' } },
      body: { recurrenceKind: 'date_list', ...dateListInput },
    })
  })

  it.each([
    [404, 'not-found'],
    [409, 'idempotency-conflict'],
    [422, 'validation-error'],
    [500, 'unexpected-error'],
  ])('maps %i to %s', async (status, kind) => {
    postMock.mockResolvedValueOnce(result(status))
    expect(await createDateListSeries('b-1', dateListInput, 'key-1')).toEqual({ kind })
  })

  it('maps a network failure to network-error', async () => {
    postMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await createDateListSeries('b-1', dateListInput, 'key-1')).toEqual({
      kind: 'network-error',
    })
  })
})

describe('timeBlocksApi.updateTimeBlockSeries', () => {
  beforeEach(() => patchMock.mockReset())
  const input = {
    blockType: 'lunch' as const,
    startsTime: '13:30',
    durationMinutes: 45,
    effectiveFrom: '2026-01-04',
    effectiveUntil: null,
    reason: null,
  }

  it('sends scope whole without effectiveDate', async () => {
    patchMock.mockResolvedValueOnce(result(200, seriesBody))

    await updateTimeBlockSeries('b-1', 'ser-1', input, null)

    expect(patchMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/time-block-series/{seriesId}',
      {
        params: { path: { barberId: 'b-1', seriesId: 'ser-1' } },
        body: { scope: 'whole', ...input },
      },
    )
  })

  it('sends scope this_and_following with the cut-off date', async () => {
    patchMock.mockResolvedValueOnce(result(200, seriesBody))

    const outcome = await updateTimeBlockSeries('b-1', 'ser-1', input, {
      effectiveDate: '2026-06-01',
    })

    expect(outcome).toMatchObject({ kind: 'success' })
    expect(patchMock.mock.calls[0]![1].body).toEqual({
      scope: 'this_and_following',
      effectiveDate: '2026-06-01',
      ...input,
    })
  })

  it.each([
    [404, 'not-found'],
    [422, 'validation-error'],
    [500, 'unexpected-error'],
  ])('maps %i to %s', async (status, kind) => {
    patchMock.mockResolvedValueOnce(result(status))
    expect(await updateTimeBlockSeries('b-1', 'ser-1', input, null)).toEqual({ kind })
  })

  it('maps a network failure to network-error', async () => {
    patchMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await updateTimeBlockSeries('b-1', 'ser-1', input, null)).toEqual({
      kind: 'network-error',
    })
  })
})

describe('timeBlocksApi fechas explícitas y excepciones', () => {
  beforeEach(() => {
    postMock.mockReset()
    deleteMock.mockReset()
  })

  it('adds an explicit date', async () => {
    postMock.mockResolvedValueOnce(result(204))

    expect(await addSeriesDate('b-1', 'ser-1', '2027-12-20')).toEqual({ kind: 'success' })
    expect(postMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/time-block-series/{seriesId}/dates',
      {
        params: { path: { barberId: 'b-1', seriesId: 'ser-1' } },
        body: { blockDate: '2027-12-20' },
      },
    )
  })

  it.each([
    [404, 'not-found'],
    [409, 'duplicate'],
    [422, 'validation-error'],
    [500, 'unexpected-error'],
  ])('maps %i when adding a date to %s', async (status, kind) => {
    postMock.mockResolvedValueOnce(result(status))
    expect(await addSeriesDate('b-1', 'ser-1', '2027-12-20')).toEqual({ kind })
  })

  it('removes an explicit date', async () => {
    deleteMock.mockResolvedValueOnce(result(204))

    expect(await removeSeriesDate('b-1', 'ser-1', '2027-12-15')).toEqual({ kind: 'success' })
    expect(deleteMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/time-block-series/{seriesId}/dates/{blockDate}',
      { params: { path: { barberId: 'b-1', seriesId: 'ser-1', blockDate: '2027-12-15' } } },
    )
  })

  it('adds an exception with its reason', async () => {
    postMock.mockResolvedValueOnce(result(204))

    expect(await addSeriesException('b-1', 'ser-1', '2027-01-08', 'Libró')).toEqual({
      kind: 'success',
    })
    expect(postMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/time-block-series/{seriesId}/exceptions',
      {
        params: { path: { barberId: 'b-1', seriesId: 'ser-1' } },
        body: { excludedDate: '2027-01-08', reason: 'Libró' },
      },
    )
  })

  it('maps a duplicate exception to duplicate', async () => {
    postMock.mockResolvedValueOnce(result(409))
    expect(await addSeriesException('b-1', 'ser-1', '2027-01-08', null)).toEqual({
      kind: 'duplicate',
    })
  })

  it('removes an exception and maps a missing one to not-found', async () => {
    deleteMock.mockResolvedValueOnce(result(204)).mockResolvedValueOnce(result(404))

    expect(await removeSeriesException('b-1', 'ser-1', '2027-01-08')).toEqual({ kind: 'success' })
    expect(await removeSeriesException('b-1', 'ser-1', '2027-01-08')).toEqual({
      kind: 'not-found',
    })
    expect(deleteMock.mock.calls[0]![0]).toBe(
      '/private/barbers/{barberId}/time-block-series/{seriesId}/exceptions/{excludedDate}',
    )
  })

  it('maps network failures to network-error', async () => {
    postMock.mockRejectedValueOnce(new Error('fetch failed'))
    deleteMock.mockRejectedValueOnce(new Error('fetch failed'))

    expect(await addSeriesDate('b-1', 'ser-1', '2027-12-20')).toEqual({ kind: 'network-error' })
    expect(await removeSeriesDate('b-1', 'ser-1', '2027-12-15')).toEqual({
      kind: 'network-error',
    })
  })
})
