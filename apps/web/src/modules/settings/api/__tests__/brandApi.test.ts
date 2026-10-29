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

const location = {
  hasPhysicalLocation: true,
  address: 'Calle 10 # 5-20, Bogotá',
  latitude: 4.598056,
  longitude: -74.076111,
}
// El servidor siempre devuelve la marca junto con la ubicación (DEC-130).
const body = { ...brand, ...location }

describe('brandApi.fetchBrand', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 body to a success outcome and asks the right path', async () => {
    getMock.mockResolvedValueOnce(ok(body))

    expect(await fetchBrand()).toEqual({ kind: 'success', brand, location })
    expect(getMock).toHaveBeenCalledWith('/private/settings/brand')
  })

  it('discards an accent this client does not know instead of painting it', async () => {
    getMock.mockResolvedValueOnce(ok({ ...body, accent: 'neon' }))

    expect(await fetchBrand()).toEqual({ kind: 'unexpected-error' })
  })

  it('reads the panel profile the server confirmed', async () => {
    getMock.mockResolvedValueOnce(ok({ ...body, panelProfile: 'solo' }))

    expect(await fetchBrand()).toEqual({
      kind: 'success',
      brand: { ...brand, panelProfile: 'solo' },
      location,
    })
  })

  it('falls back to the full panel when the profile is one this client does not know', async () => {
    getMock.mockResolvedValueOnce(ok({ ...body, panelProfile: 'team' }))

    expect(await fetchBrand()).toEqual({ kind: 'success', brand, location })
  })

  it('reads a barbershop without address as null coordinates, never as empty text', async () => {
    const empty = { hasPhysicalLocation: true, address: null, latitude: null, longitude: null }
    getMock.mockResolvedValueOnce(ok({ ...brand, ...empty }))

    expect(await fetchBrand()).toEqual({ kind: 'success', brand, location: empty })
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

  it('sends exactly the six brand fields, never the panel, and returns what the server confirmed', async () => {
    patchMock.mockResolvedValueOnce(ok({ ...body, businessTerm: 'salón de belleza' }))

    const outcome = await saveBrand(brand)

    expect(patchMock.mock.calls[0]![1].body).not.toHaveProperty('panelProfile')
    expect(patchMock).toHaveBeenCalledWith('/private/settings/brand', {
      body: expect.objectContaining({ accent: brand.accent }),
    })
    expect(Object.keys(patchMock.mock.calls[0]![1].body).sort()).toEqual(
      [
        'accent',
        'businessTerm',
        'businessTermGender',
        'professionalTerm',
        'professionalTermGender',
        'professionalTermPlural',
      ].sort(),
    )
    expect(outcome).toEqual({
      kind: 'success',
      brand: { ...brand, businessTerm: 'salón de belleza' },
      location,
    })
  })

  it('sends only the location fields that changed, and null removes the address (DEC-130)', async () => {
    patchMock.mockResolvedValue(ok(body))

    await saveBrand(brand, { hasPhysicalLocation: false })
    expect(patchMock.mock.calls[0]![1].body).toMatchObject({ hasPhysicalLocation: false })
    expect(patchMock.mock.calls[0]![1].body).not.toHaveProperty('address')

    await saveBrand(brand, { address: 'Calle 1 # 2-3' })
    expect(patchMock.mock.calls[1]![1].body).toMatchObject({ address: 'Calle 1 # 2-3' })
    expect(patchMock.mock.calls[1]![1].body).not.toHaveProperty('hasPhysicalLocation')

    await saveBrand(brand, { address: null })
    expect(patchMock.mock.calls[2]![1].body).toHaveProperty('address', null)
  })

  it('never sends coordinates: the server computes them', async () => {
    patchMock.mockResolvedValueOnce(ok(body))

    await saveBrand(brand, { address: 'Calle 1 # 2-3' })

    const sent = patchMock.mock.calls[0]![1].body
    expect(sent).not.toHaveProperty('latitude')
    expect(sent).not.toHaveProperty('longitude')
  })

  it('maps a 422 to location-error when an address was sent, else to validation-error', async () => {
    patchMock.mockResolvedValue(problem(422))

    expect(await saveBrand(brand, { address: 'Ninguna parte 999' })).toEqual({
      kind: 'location-error',
    })
    expect(await saveBrand(brand, { hasPhysicalLocation: false })).toEqual({
      kind: 'validation-error',
    })
    expect(await saveBrand(brand)).toEqual({ kind: 'validation-error' })
  })

  it.each([400, 422])('maps a %s to validation-error without a location change', async (status) => {
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
