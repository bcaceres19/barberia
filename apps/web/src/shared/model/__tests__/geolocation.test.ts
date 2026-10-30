import { describe, expect, it, vi } from 'vitest'
import { isGeolocationSupported, locateUser } from '../geolocation'

type Geo = Pick<Geolocation, 'getCurrentPosition'>

function geoOk(latitude: number, longitude: number, accuracy: number): Geo {
  return {
    getCurrentPosition: vi.fn((ok: PositionCallback) =>
      ok({ coords: { latitude, longitude, accuracy } } as GeolocationPosition),
    ),
  }
}

function geoFails(code: number): Geo {
  return {
    getCurrentPosition: vi.fn((_ok: PositionCallback, fail?: PositionErrorCallback | null) =>
      fail?.({ code } as GeolocationPositionError),
    ),
  }
}

describe('isGeolocationSupported', () => {
  it('needs a getCurrentPosition function', () => {
    expect(isGeolocationSupported(geoOk(1, 1, 1))).toBe(true)
    expect(isGeolocationSupported({} as Geo)).toBe(false)
  })
})

describe('locateUser', () => {
  it('returns the rounded point and the accuracy', async () => {
    const geo = geoOk(4.6124421999, -74.0694181111, 18.4)
    const result = await locateUser(geo)

    expect(result).toEqual({
      kind: 'success',
      point: { latitude: 4.612442, longitude: -74.069418 },
      accuracyMeters: 18,
    })
  })

  it('asks for a fresh high-accuracy fix with a timeout (CA-138-05)', async () => {
    const geo = geoOk(4.6, -74, 5)
    await locateUser(geo, 7000)

    expect(geo.getCurrentPosition).toHaveBeenCalledWith(
      expect.any(Function),
      expect.any(Function),
      {
        enableHighAccuracy: true,
        timeout: 7000,
        maximumAge: 0,
      },
    )
  })

  it.each([
    [1, 'denied'],
    [2, 'unavailable'],
    [3, 'timeout'],
    [99, 'unavailable'],
  ])('maps error code %i to %s', async (code, kind) => {
    expect(await locateUser(geoFails(code))).toEqual({ kind })
  })

  it('is unsupported without the API, without throwing', async () => {
    expect(await locateUser({} as Geo)).toEqual({ kind: 'unsupported' })
  })

  it('treats an impossible position as unavailable', async () => {
    expect(await locateUser(geoOk(120, -74, 5))).toEqual({ kind: 'unavailable' })
  })

  it('reports an unknown accuracy as null', async () => {
    const result = await locateUser(geoOk(4.6, -74, Number.NaN))
    expect(result).toMatchObject({ kind: 'success', accuracyMeters: null })
  })
})
