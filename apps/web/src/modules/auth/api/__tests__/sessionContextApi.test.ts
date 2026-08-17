/**
 * Pruebas del cliente tipado (`sessionContextApi` sobre
 * `shared/api/httpClient`, HU-012/DEC-060): mapeo por `status`, mismo
 * criterio que `loginApi.test.ts`. El recorrido de red real se prueba en
 * `e2e/panel.spec.ts` y en `apps/api` (`cmd/api/session_context_integration_test.go`).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock },
}))

const { fetchSessionContext } = await import('../sessionContextApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function unauthorized() {
  return {
    data: undefined,
    error: {
      type: '/api/v1/problems/unauthorized',
      title: 'No autorizado',
      status: 401,
      code: 'unauthorized',
      requestId: 'req-1',
    },
    response: new Response(null, { status: 401 }),
  }
}

describe('sessionContextApi.fetchSessionContext (cliente tipado)', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('maps a 200 SessionContextResponse to an authenticated outcome', async () => {
    getMock.mockResolvedValueOnce(
      ok({
        barbershop: { id: 'shop-1', name: 'Barbería de prueba' },
        expiresAt: '2026-09-12T12:00:00Z',
      }),
    )

    const outcome = await fetchSessionContext()

    expect(outcome).toEqual({
      kind: 'authenticated',
      barbershopId: 'shop-1',
      barbershopName: 'Barbería de prueba',
      expiresAt: '2026-09-12T12:00:00Z',
    })
  })

  it('maps 401 to unauthenticated', async () => {
    getMock.mockResolvedValueOnce(unauthorized())

    const outcome = await fetchSessionContext()

    expect(outcome).toEqual({ kind: 'unauthenticated' })
  })

  it('maps an undocumented status to unexpected-error', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal-error', title: 'Error interno' },
      response: new Response(null, { status: 500 }),
    })

    const outcome = await fetchSessionContext()

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response at all) to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await fetchSessionContext()

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('calls exactly the session-context operation', async () => {
    getMock.mockResolvedValueOnce(
      ok({ barbershop: { id: 'shop-1', name: 'Barbería' }, expiresAt: '2026-09-12T12:00:00Z' }),
    )

    await fetchSessionContext()

    expect(getMock).toHaveBeenCalledTimes(1)
    expect(getMock.mock.calls[0]?.[0]).toBe('/private/auth/session')
  })
})
