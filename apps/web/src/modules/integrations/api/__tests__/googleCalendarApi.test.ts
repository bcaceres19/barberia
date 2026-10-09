/**
 * Pruebas del cliente tipado de la integración con Google Calendar: mapeo por
 * `status` (nunca por `detail`) a outcomes discriminados, igual que
 * `staff/api/__tests__/staffApi.test.ts`. El cliente nunca envía tokens ni
 * identificadores de usuario o barbero.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const postMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
const deleteMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: postMock, PATCH: patchMock, DELETE: deleteMock },
}))

const { fetchConnection, startConnect, completeCallback, syncNow, disconnect, updateReminder } =
  await import('../googleCalendarApi')

const wire = {
  enabled: true,
  barberLinked: true,
  status: 'connected',
  accountEmail: 'barbero@ejemplo.test',
  reminderMinutes: 30,
  connectedAt: '2026-10-10T15:00:00Z',
  lastSyncedAt: null,
  pendingSyncJobs: 2,
  failedSyncJobs: 0,
}

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}
function status(code: number) {
  return { data: undefined, error: {}, response: new Response(null, { status: code }) }
}

describe('googleCalendarApi.fetchConnection', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 to the connection', async () => {
    getMock.mockResolvedValueOnce(ok(wire))
    expect(await fetchConnection()).toEqual({ kind: 'success', connection: wire })
    expect(getMock).toHaveBeenCalledWith('/private/integrations/google-calendar')
  })

  it('maps failures', async () => {
    getMock.mockResolvedValueOnce(status(500))
    expect(await fetchConnection()).toEqual({ kind: 'unexpected-error' })
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    expect(await fetchConnection()).toEqual({ kind: 'network-error' })
  })
})

describe('googleCalendarApi.startConnect', () => {
  beforeEach(() => postMock.mockReset())

  it('returns the authorization URL', async () => {
    postMock.mockResolvedValueOnce(ok({ authorizationUrl: 'https://accounts.example.test/auth' }))
    expect(await startConnect()).toEqual({
      kind: 'success',
      authorizationUrl: 'https://accounts.example.test/auth',
    })
    expect(postMock).toHaveBeenCalledWith('/private/integrations/google-calendar/connect')
  })

  it.each([
    [409, 'unavailable'],
    [500, 'unexpected-error'],
  ])('maps status %i to %s', async (code, kind) => {
    postMock.mockResolvedValueOnce(status(code))
    expect(await startConnect()).toEqual({ kind })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    expect(await startConnect()).toEqual({ kind: 'network-error' })
  })
})

describe('googleCalendarApi.completeCallback', () => {
  beforeEach(() => postMock.mockReset())

  it('forwards only what Google returned and maps the result', async () => {
    postMock.mockResolvedValueOnce(ok({ result: 'connected' }))
    expect(await completeCallback({ state: 's1', code: 'c1' })).toEqual({
      kind: 'success',
      result: 'connected',
    })
    expect(postMock).toHaveBeenCalledWith('/private/integrations/google-calendar/callback', {
      body: { state: 's1', code: 'c1' },
    })
  })

  it('omits empty code and error, and forwards a denial', async () => {
    postMock.mockResolvedValueOnce(ok({ result: 'denied' }))
    await completeCallback({ state: 's2', code: '', error: 'access_denied' })
    expect(postMock).toHaveBeenCalledWith('/private/integrations/google-calendar/callback', {
      body: { state: 's2', error: 'access_denied' },
    })
  })

  it.each([
    [400, 'invalid'],
    [409, 'unavailable'],
    [500, 'unexpected-error'],
  ])('maps status %i to %s', async (code, kind) => {
    postMock.mockResolvedValueOnce(status(code))
    expect(await completeCallback({ state: 's' })).toEqual({ kind })
  })
})

describe('googleCalendarApi.syncNow / disconnect / updateReminder', () => {
  beforeEach(() => {
    postMock.mockReset()
    deleteMock.mockReset()
    patchMock.mockReset()
  })

  it('syncNow maps the queue counters and failures', async () => {
    postMock.mockResolvedValueOnce(ok({ pendingSyncJobs: 3, failedSyncJobs: 0 }))
    expect(await syncNow()).toEqual({ kind: 'success', pendingSyncJobs: 3, failedSyncJobs: 0 })
    expect(postMock).toHaveBeenCalledWith('/private/integrations/google-calendar/sync')

    postMock.mockResolvedValueOnce(status(404))
    expect(await syncNow()).toEqual({ kind: 'not-connected' })
    postMock.mockResolvedValueOnce(status(409))
    expect(await syncNow()).toEqual({ kind: 'unavailable' })
  })

  it('disconnect maps a 204 to success', async () => {
    deleteMock.mockResolvedValueOnce({
      data: undefined,
      response: new Response(null, { status: 204 }),
    })
    expect(await disconnect()).toEqual({ kind: 'success' })
    deleteMock.mockResolvedValueOnce(status(500))
    expect(await disconnect()).toEqual({ kind: 'unexpected-error' })
    deleteMock.mockRejectedValueOnce(new TypeError('x'))
    expect(await disconnect()).toEqual({ kind: 'network-error' })
  })

  it('updateReminder sends the minutes (or null) and maps outcomes', async () => {
    patchMock.mockResolvedValueOnce(ok({ ...wire, reminderMinutes: 15 }))
    expect(await updateReminder(15)).toEqual({
      kind: 'success',
      connection: { ...wire, reminderMinutes: 15 },
    })
    expect(patchMock).toHaveBeenCalledWith('/private/integrations/google-calendar', {
      body: { reminderMinutes: 15 },
    })

    patchMock.mockResolvedValueOnce(ok({ ...wire, reminderMinutes: null }))
    await updateReminder(null)
    expect(patchMock).toHaveBeenLastCalledWith('/private/integrations/google-calendar', {
      body: { reminderMinutes: null },
    })

    patchMock.mockResolvedValueOnce(status(422))
    expect(await updateReminder(99999)).toEqual({ kind: 'validation-error' })
    patchMock.mockResolvedValueOnce(status(404))
    expect(await updateReminder(5)).toEqual({ kind: 'not-connected' })
  })
})
