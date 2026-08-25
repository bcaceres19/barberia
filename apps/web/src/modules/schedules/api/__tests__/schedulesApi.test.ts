/**
 * Pruebas del cliente tipado (`schedulesApi` sobre `shared/api/httpClient`):
 * mapeo por `status`/`code` (nunca por `detail`), mismo criterio que
 * `barberServices/api/__tests__/barberServicesApi.test.ts`. El recorrido de
 * red real vive en un E2E propio de HU-040.
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
  fetchBarberSummaries,
  fetchBarbershopTimezone,
  fetchWorkingHours,
  createWorkingHour,
  updateWorkingHour,
  deleteWorkingHour,
} = await import('../schedulesApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function created(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 201 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return { data: undefined, error: body, response: new Response(null, { status }) }
}

const workingHourBody = {
  id: 'wh-1',
  isoWeekday: 1,
  startsTime: '08:00',
  durationMinutes: 480,
  createdAt: '2026-08-25T15:04:05Z',
  updatedAt: '2026-08-25T15:04:05Z',
}

describe('schedulesApi.fetchBarberSummaries', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a minimal summary, dropping timestamps', async () => {
    getMock.mockResolvedValueOnce(
      ok({
        items: [{ id: 'b-1', fullName: 'Carlos Ramírez', createdAt: '', updatedAt: '' }],
        nextCursor: null,
      }),
    )

    const outcome = await fetchBarberSummaries()

    expect(outcome).toEqual({ kind: 'success', items: [{ id: 'b-1', fullName: 'Carlos Ramírez' }] })
    expect(getMock).toHaveBeenCalledWith('/private/barbers', { params: { query: { limit: 50 } } })
  })

  it('maps a network failure to network-error', async () => {
    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchBarberSummaries()).toEqual({ kind: 'network-error' })
  })
})

describe('schedulesApi.fetchBarbershopTimezone', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to its IANA timezone (CA-040-06)', async () => {
    getMock.mockResolvedValueOnce(ok({ name: 'Barbería A', timezone: 'America/Bogota' }))

    const outcome = await fetchBarbershopTimezone()

    expect(outcome).toEqual({ kind: 'success', timezone: 'America/Bogota' })
    expect(getMock).toHaveBeenCalledWith('/private/settings/barbershop')
  })

  it('maps any failure to unavailable, never blocking the schedule from loading', async () => {
    getMock.mockResolvedValueOnce(problem(500))
    expect(await fetchBarbershopTimezone()).toEqual({ kind: 'unavailable' })

    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchBarbershopTimezone()).toEqual({ kind: 'unavailable' })
  })
})

describe('schedulesApi.fetchWorkingHours', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a page of working hours', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [workingHourBody], nextCursor: null }))

    const outcome = await fetchWorkingHours('b-1')

    expect(outcome).toEqual({
      kind: 'success',
      page: { items: [workingHourBody], nextCursor: null },
    })
    expect(getMock).toHaveBeenCalledWith('/private/barbers/{barberId}/working-hours', {
      params: { path: { barberId: 'b-1' }, query: { limit: 50 } },
    })
  })

  it('maps a 404 to not-found (RN-TEN-01: unknown or cross-tenant barber)', async () => {
    getMock.mockResolvedValueOnce(problem(404))
    expect(await fetchWorkingHours('unknown')).toEqual({ kind: 'not-found' })
  })

  it('maps a network failure to network-error', async () => {
    getMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await fetchWorkingHours('b-1')).toEqual({ kind: 'network-error' })
  })
})

describe('schedulesApi.createWorkingHour', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 201 success body to a success outcome, sending Idempotency-Key', async () => {
    postMock.mockResolvedValueOnce(created(workingHourBody))

    const outcome = await createWorkingHour('b-1', 1, '08:00', 480, 'key-1')

    expect(outcome).toEqual({ kind: 'success', workingHour: workingHourBody })
    expect(postMock).toHaveBeenCalledWith('/private/barbers/{barberId}/working-hours', {
      params: { path: { barberId: 'b-1' }, header: { 'Idempotency-Key': 'key-1' } },
      body: { isoWeekday: 1, startsTime: '08:00', durationMinutes: 480 },
    })
  })

  it('maps a 404 to not-found', async () => {
    postMock.mockResolvedValueOnce(problem(404))
    expect(await createWorkingHour('unknown', 1, '08:00', 60, 'k')).toEqual({ kind: 'not-found' })
  })

  it('maps a 409 with code=conflict to overlap-conflict (CA-040-04)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'conflict' }))
    expect(await createWorkingHour('b-1', 1, '08:00', 60, 'k')).toEqual({
      kind: 'overlap-conflict',
    })
  })

  it('maps a 409 with code=idempotency-conflict to idempotency-conflict (RN-IDE-01)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'idempotency-conflict' }))
    expect(await createWorkingHour('b-1', 1, '08:00', 60, 'k')).toEqual({
      kind: 'idempotency-conflict',
    })
  })

  it('maps a 422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422))
    expect(await createWorkingHour('b-1', 1, '08:00', 60, 'k')).toEqual({
      kind: 'validation-error',
    })
  })

  it('maps a network failure to network-error', async () => {
    postMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await createWorkingHour('b-1', 1, '08:00', 60, 'k')).toEqual({ kind: 'network-error' })
  })
})

describe('schedulesApi.updateWorkingHour', () => {
  beforeEach(() => patchMock.mockReset())

  it('maps a 200 success body to a success outcome', async () => {
    patchMock.mockResolvedValueOnce(ok(workingHourBody))

    const outcome = await updateWorkingHour('b-1', 'wh-1', 1, '09:00', 60)

    expect(outcome).toEqual({ kind: 'success', workingHour: workingHourBody })
    expect(patchMock).toHaveBeenCalledWith(
      '/private/barbers/{barberId}/working-hours/{workingHourId}',
      {
        params: { path: { barberId: 'b-1', workingHourId: 'wh-1' } },
        body: { isoWeekday: 1, startsTime: '09:00', durationMinutes: 60 },
      },
    )
  })

  it('maps a 404 to not-found', async () => {
    patchMock.mockResolvedValueOnce(problem(404))
    expect(await updateWorkingHour('b-1', 'unknown', 1, '08:00', 60)).toEqual({ kind: 'not-found' })
  })

  it('maps a 409 to overlap-conflict', async () => {
    patchMock.mockResolvedValueOnce(problem(409))
    expect(await updateWorkingHour('b-1', 'wh-1', 1, '08:00', 60)).toEqual({
      kind: 'overlap-conflict',
    })
  })

  it('maps a 422 to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(422))
    expect(await updateWorkingHour('b-1', 'wh-1', 1, '08:00', 60)).toEqual({
      kind: 'validation-error',
    })
  })
})

describe('schedulesApi.deleteWorkingHour', () => {
  beforeEach(() => deleteMock.mockReset())

  it('maps a 204 success to a success outcome', async () => {
    deleteMock.mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    })
    expect(await deleteWorkingHour('b-1', 'wh-1')).toEqual({ kind: 'success' })
  })

  it('maps a 404 to not-found (safe on retry, CA-040-05)', async () => {
    deleteMock.mockResolvedValueOnce(problem(404))
    expect(await deleteWorkingHour('b-1', 'unknown')).toEqual({ kind: 'not-found' })
  })

  it('maps a network failure to network-error', async () => {
    deleteMock.mockRejectedValueOnce(new Error('fetch failed'))
    expect(await deleteWorkingHour('b-1', 'wh-1')).toEqual({ kind: 'network-error' })
  })
})
