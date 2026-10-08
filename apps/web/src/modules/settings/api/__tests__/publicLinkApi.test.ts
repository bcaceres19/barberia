/**
 * Cliente tipado del enlace público (DEC-117): mapeo por `status`, nunca por
 * `detail`; un slug que no cumple el formato de la base se descarta.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock },
}))

const { fetchPublicLink } = await import('../publicLinkApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}
function problem(status: number) {
  return { data: undefined, error: {}, response: new Response(null, { status }) }
}

describe('publicLinkApi.fetchPublicLink', () => {
  beforeEach(() => getMock.mockReset())

  it.each(['cortefinok7x2m9q4', 'corte-fino-k7x2m9'])(
    'maps a 200 body to a success outcome (new slug, or a legacy one with hyphens): %s',
    async (slug) => {
      getMock.mockResolvedValueOnce(ok({ slug }))

      expect(await fetchPublicLink()).toEqual({ kind: 'success', slug })
      expect(getMock).toHaveBeenCalledWith('/private/settings/public-link')
    },
  )

  it.each(['', 'ab', '-corte-fino', 'Corte-Fino-k7x2m9', 'corte fino', 'a'.repeat(41)])(
    'discards a slug the database would never emit (%j)',
    async (slug) => {
      getMock.mockResolvedValueOnce(ok({ slug }))

      expect(await fetchPublicLink()).toEqual({ kind: 'unexpected-error' })
    },
  )

  it.each([401, 404, 500])('maps a %i response to an unexpected error', async (status) => {
    getMock.mockResolvedValueOnce(problem(status))

    expect(await fetchPublicLink()).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a transport failure to a network error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await fetchPublicLink()).toEqual({ kind: 'network-error' })
  })
})
