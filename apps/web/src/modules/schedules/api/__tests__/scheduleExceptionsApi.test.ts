/**
 * Pruebas del cliente tipado (`scheduleExceptionsApi` sobre
 * `shared/api/httpClient`): mapeo por `status`/`code` (nunca por `detail`),
 * mismo criterio que `schedulesApi.test.ts` (HU-040). El recorrido de red
 * real vive en un E2E propio de HU-041.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const postMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
const deleteMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: postMock, PATCH: patchMock, DELETE: deleteMock },
}))

const {
  fetchHolidayCalendar,
  updateHolidayCalendar,
  fetchScheduleExceptions,
  createScheduleException,
  updateScheduleException,
  deleteScheduleException,
  fetchColombianHolidays,
} = await import('../scheduleExceptionsApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function created(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 201 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return { data: undefined, error: body, response: new Response(null, { status }) }
}

const exceptionBody = {
  id: 'exc-1',
  effectiveDate: '2026-12-08',
  isClosed: true,
  reason: null,
  segments: [],
  createdAt: '2026-08-25T15:04:05Z',
  updatedAt: '2026-08-25T15:04:05Z',
}

describe('scheduleExceptionsApi.fetchHolidayCalendar', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to its enabled flag', async () => {
    getMock.mockResolvedValueOnce(ok({ enabled: true }))

    const outcome = await fetchHolidayCalendar('b-1')

    expect(outcome).toEqual({ kind: 'success', enabled: true })
    expect(getMock).toHaveBeenCalledWith('/private/barbers/{barberId}/holiday-calendar', {
      params: { path: { barberId: 'b-1' } },
    })
  })

  it('maps a 404 to not-found (RN-TEN-01)', async () => {
    getMock.mockResolvedValueOnce(problem(404))
    expect(await fetchHolidayCalendar('unknown')).toEqual({ kind: 'not-found' })
  })

  it('maps a network failure to network-error', async () => {
    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchHolidayCalendar('b-1')).toEqual({ kind: 'network-error' })
  })
})

describe('scheduleExceptionsApi.updateHolidayCalendar', () => {
  beforeEach(() => patchMock.mockReset())

  it('maps a 200 success body to its new enabled flag', async () => {
    patchMock.mockResolvedValueOnce(ok({ enabled: true }))

    const outcome = await updateHolidayCalendar('b-1', true)

    expect(outcome).toEqual({ kind: 'success', enabled: true })
    expect(patchMock).toHaveBeenCalledWith('/private/barbers/{barberId}/holiday-calendar', {
      params: { path: { barberId: 'b-1' } },
      body: { enabled: true },
    })
  })

  it('maps a 404 to not-found', async () => {
    patchMock.mockResolvedValueOnce(problem(404))
    expect(await updateHolidayCalendar('unknown', true)).toEqual({ kind: 'not-found' })
  })
})

describe('scheduleExceptionsApi.fetchScheduleExceptions', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a page of exceptions', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [exceptionBody], nextCursor: null }))

    const outcome = await fetchScheduleExceptions('b-1')

    expect(outcome).toEqual({
      kind: 'success',
      page: { items: [exceptionBody], nextCursor: null },
    })
    expect(getMock).toHaveBeenCalledWith('/private/barbers/{barberId}/schedule-exceptions', {
      params: { path: { barberId: 'b-1' }, query: { limit: 50 } },
    })
  })

  it('maps a 404 to not-found', async () => {
    getMock.mockResolvedValueOnce(problem(404))
    expect(await fetchScheduleExceptions('unknown')).toEqual({ kind: 'not-found' })
  })

  it('maps a network failure to network-error', async () => {
    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchScheduleExceptions('b-1')).toEqual({ kind: 'network-error' })
  })
})

describe('scheduleExceptionsApi.createScheduleException', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 201 success body to a success outcome, sending Idempotency-Key', async () => {
    postMock.mockResolvedValueOnce(created(exceptionBody))

    const outcome = await createScheduleException('b-1', '2026-12-08', true, null, [], 'key-1')

    expect(outcome).toEqual({ kind: 'success', exception: exceptionBody })
    expect(postMock).toHaveBeenCalledWith('/private/barbers/{barberId}/schedule-exceptions', {
      params: { path: { barberId: 'b-1' }, header: { 'Idempotency-Key': 'key-1' } },
      body: { effectiveDate: '2026-12-08', isClosed: true, reason: null, segments: [] },
    })
  })

  it('maps a 404 to not-found', async () => {
    postMock.mockResolvedValueOnce(problem(404))
    expect(await createScheduleException('unknown', '2026-12-08', true, null, [], 'k')).toEqual({
      kind: 'not-found',
    })
  })

  it('maps a 409 with code=conflict to date-conflict (CA-041-05)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'conflict' }))
    expect(await createScheduleException('b-1', '2026-12-08', true, null, [], 'k')).toEqual({
      kind: 'date-conflict',
    })
  })

  it('maps a 409 with code=idempotency-conflict to idempotency-conflict (RN-IDE-01)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'idempotency-conflict' }))
    expect(await createScheduleException('b-1', '2026-12-08', true, null, [], 'k')).toEqual({
      kind: 'idempotency-conflict',
    })
  })

  it('maps a 422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422))
    expect(await createScheduleException('b-1', '2026-12-08', true, null, [], 'k')).toEqual({
      kind: 'validation-error',
    })
  })

  it('maps a network failure to network-error', async () => {
    postMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await createScheduleException('b-1', '2026-12-08', true, null, [], 'k')).toEqual({
      kind: 'network-error',
    })
  })
})

describe('scheduleExceptionsApi.updateScheduleException', () => {
  beforeEach(() => patchMock.mockReset())

  it('maps a 200 success body to a success outcome', async () => {
    patchMock.mockResolvedValueOnce(ok(exceptionBody))

    const outcome = await updateScheduleException('b-1', 'exc-1', '2026-12-08', true, null, [])

    expect(outcome).toEqual({ kind: 'success', exception: exceptionBody })
    expect(patchMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/schedule-exceptions/{exceptionId}',
      {
        params: { path: { barberId: 'b-1', exceptionId: 'exc-1' } },
        body: { effectiveDate: '2026-12-08', isClosed: true, reason: null, segments: [] },
      },
    )
  })

  it('maps a 404 to not-found', async () => {
    patchMock.mockResolvedValueOnce(problem(404))
    expect(await updateScheduleException('b-1', 'unknown', '2026-12-08', true, null, [])).toEqual({
      kind: 'not-found',
    })
  })

  it('maps a 409 to date-conflict', async () => {
    patchMock.mockResolvedValueOnce(problem(409))
    expect(await updateScheduleException('b-1', 'exc-1', '2026-12-08', true, null, [])).toEqual({
      kind: 'date-conflict',
    })
  })

  it('maps a 422 to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(422))
    expect(await updateScheduleException('b-1', 'exc-1', '2026-12-08', true, null, [])).toEqual({
      kind: 'validation-error',
    })
  })
})

describe('scheduleExceptionsApi.deleteScheduleException', () => {
  beforeEach(() => deleteMock.mockReset())

  it('maps a 204 success to a success outcome', async () => {
    deleteMock.mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    })
    expect(await deleteScheduleException('b-1', 'exc-1')).toEqual({ kind: 'success' })
  })

  it('maps a 404 to not-found (safe on retry, CA-041-06)', async () => {
    deleteMock.mockResolvedValueOnce(problem(404))
    expect(await deleteScheduleException('b-1', 'unknown')).toEqual({ kind: 'not-found' })
  })
})

describe('scheduleExceptionsApi.fetchColombianHolidays', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to its list of holidays', async () => {
    const items = [{ date: '2026-07-20', name: 'Día de la Independencia' }]
    getMock.mockResolvedValueOnce(ok({ items }))

    const outcome = await fetchColombianHolidays(2026)

    expect(outcome).toEqual({ kind: 'success', items })
    expect(getMock).toHaveBeenCalledWith('/private/schedule/colombian-holidays', {
      params: { query: { year: 2026 } },
    })
  })

  it('maps any failure to unavailable, never blocking the schedule from loading', async () => {
    getMock.mockResolvedValueOnce(problem(400))
    expect(await fetchColombianHolidays(2026)).toEqual({ kind: 'unavailable' })

    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchColombianHolidays(2026)).toEqual({ kind: 'unavailable' })
  })
})
