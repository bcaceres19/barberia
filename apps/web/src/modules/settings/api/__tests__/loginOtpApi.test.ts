/**
 * Cliente tipado del canal del código de acceso (DEC-139): mapeo por `status`,
 * un canal fuera de la lista se descarta y el cuerpo es solo el canal.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, PATCH: patchMock },
}))

const { fetchLoginOtpChannel, saveLoginOtpChannel } = await import('../loginOtpApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}
function problem(status: number) {
  return { data: undefined, error: {}, response: new Response(null, { status }) }
}

describe('loginOtpApi.fetchLoginOtpChannel', () => {
  beforeEach(() => getMock.mockReset())

  it.each(['email', 'whatsapp', 'both'])(
    'maps a 200 with %s to a success outcome',
    async (channel) => {
      getMock.mockResolvedValueOnce(ok({ channel }))

      expect(await fetchLoginOtpChannel()).toEqual({ kind: 'success', channel })
      expect(getMock).toHaveBeenCalledWith('/private/settings/login-otp')
    },
  )

  it.each(['sms', '', 'Email'])(
    'discards a channel the server would never emit (%j)',
    async (channel) => {
      getMock.mockResolvedValueOnce(ok({ channel }))

      expect(await fetchLoginOtpChannel()).toEqual({ kind: 'unexpected-error' })
    },
  )

  it.each([401, 404, 500])('maps a %i response to an unexpected error', async (status) => {
    getMock.mockResolvedValueOnce(problem(status))

    expect(await fetchLoginOtpChannel()).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a transport failure to a network error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await fetchLoginOtpChannel()).toEqual({ kind: 'network-error' })
  })
})

describe('loginOtpApi.saveLoginOtpChannel', () => {
  beforeEach(() => patchMock.mockReset())

  it('sends only the channel — never a destination — and returns what the server stored', async () => {
    patchMock.mockResolvedValueOnce(ok({ channel: 'both' }))

    expect(await saveLoginOtpChannel('both')).toEqual({ kind: 'success', channel: 'both' })
    expect(patchMock).toHaveBeenCalledWith('/private/settings/login-otp', {
      body: { channel: 'both' },
    })
  })

  it.each([400, 401, 404, 422, 500])(
    'maps a %i response to an unexpected error',
    async (status) => {
      patchMock.mockResolvedValueOnce(problem(status))

      expect(await saveLoginOtpChannel('email')).toEqual({ kind: 'unexpected-error' })
    },
  )

  it('maps a transport failure to a network error', async () => {
    patchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await saveLoginOtpChannel('email')).toEqual({ kind: 'network-error' })
  })
})
