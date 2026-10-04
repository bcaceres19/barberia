/**
 * Cliente tipado de marca y vocabulario (DEC-110): mapeo por `status`, nunca
 * por `detail`; una clave de acento desconocida se descarta en vez de pintarse.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, PATCH: patchMock },
}))

const { fetchBrand, saveBrand } = await import('../brandApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}
function problem(status: number) {
  return { data: undefined, error: {}, response: new Response(null, { status }) }
}

const brand = {
  accent: 'emerald' as const,
  businessTerm: 'salón',
  businessTermGender: 'masculine' as const,
  professionalTerm: 'estilista',
  professionalTermPlural: 'estilistas',
  professionalTermGender: 'feminine' as const,
  panelProfile: 'shop' as const,
}

describe('brandApi.fetchBrand', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 body to a success outcome and asks the right path', async () => {
    getMock.mockResolvedValueOnce(ok(brand))

    expect(await fetchBrand()).toEqual({ kind: 'success', brand })
    expect(getMock).toHaveBeenCalledWith('/private/settings/brand')
  })

  it('discards an accent this client does not know instead of painting it', async () => {
    getMock.mockResolvedValueOnce(ok({ ...brand, accent: 'neon' }))

    expect(await fetchBrand()).toEqual({ kind: 'unexpected-error' })
  })

  it('reads the panel profile the server confirmed', async () => {
    getMock.mockResolvedValueOnce(ok({ ...brand, panelProfile: 'solo' }))

    expect(await fetchBrand()).toEqual({
      kind: 'success',
      brand: { ...brand, panelProfile: 'solo' },
    })
  })

  it('falls back to the full panel when the profile is one this client does not know', async () => {
    getMock.mockResolvedValueOnce(ok({ ...brand, panelProfile: 'team' }))

    expect(await fetchBrand()).toEqual({ kind: 'success', brand })
  })

  it.each([401, 404, 500])('maps a %s to unexpected-error (401 is handled globally)', async (s) => {
    getMock.mockResolvedValueOnce(problem(s))

    expect(await fetchBrand()).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a thrown fetch to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await fetchBrand()).toEqual({ kind: 'network-error' })
  })
})

describe('brandApi.saveBrand', () => {
  beforeEach(() => patchMock.mockReset())

  it('sends exactly the seven fields and returns what the server confirmed', async () => {
    patchMock.mockResolvedValueOnce(ok({ ...brand, businessTerm: 'salón de belleza' }))

    const outcome = await saveBrand(brand)

    expect(patchMock).toHaveBeenCalledWith('/private/settings/brand', { body: brand })
    expect(Object.keys(patchMock.mock.calls[0]![1].body).sort()).toEqual(
      [
        'accent',
        'businessTerm',
        'businessTermGender',
        'professionalTerm',
        'professionalTermGender',
        'professionalTermPlural',
        'panelProfile',
      ].sort(),
    )
    expect(outcome).toEqual({
      kind: 'success',
      brand: { ...brand, businessTerm: 'salón de belleza' },
    })
  })

  it.each([400, 422])('maps a %s to validation-error', async (status) => {
    patchMock.mockResolvedValueOnce(problem(status))

    expect(await saveBrand(brand)).toEqual({ kind: 'validation-error' })
  })

  it.each([401, 404, 500])('maps a %s to unexpected-error', async (status) => {
    patchMock.mockResolvedValueOnce(problem(status))

    expect(await saveBrand(brand)).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a thrown fetch to network-error', async () => {
    patchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await saveBrand(brand)).toEqual({ kind: 'network-error' })
  })
})
