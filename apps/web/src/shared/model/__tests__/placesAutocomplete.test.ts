import { describe, expect, it, vi } from 'vitest'
import {
  createPlacesClient,
  isPlacesAvailable,
  MIN_QUERY_LENGTH,
  toSelection,
} from '../placesAutocomplete'

function jsonResponse(body: unknown, ok = true): Response {
  return { ok, json: async () => body } as Response
}

describe('isPlacesAvailable', () => {
  it('needs a browser key or the dev server', () => {
    expect(isPlacesAvailable({ key: 'k', dev: false })).toBe(true)
    expect(isPlacesAvailable({ dev: true })).toBe(true)
    expect(isPlacesAvailable({ dev: false })).toBe(false)
  })
})

describe('toSelection', () => {
  it('rejects anything without a real address and point', () => {
    expect(toSelection('  ', 4.6, -74)).toBeNull()
    expect(toSelection('Calle 1', undefined, -74)).toBeNull()
    expect(toSelection('Calle 1', 91, -74)).toBeNull()
    expect(toSelection('Calle 1', 4.6, Number.NaN)).toBeNull()
  })

  it('normalizes the address text', () => {
    expect(toSelection('  Calle 10   # 5-20 ', 4.6, -74)).toEqual({
      address: 'Calle 10 # 5-20',
      latitude: 4.6,
      longitude: -74,
    })
  })
})

describe('Google Places client', () => {
  const env = { key: 'browser-key', dev: false }

  it('asks nothing until the query is long enough', async () => {
    const fetchFn = vi.fn()
    const client = createPlacesClient(env, fetchFn as unknown as typeof fetch)!

    expect(await client.suggest('ab')).toEqual([])
    expect(fetchFn).not.toHaveBeenCalled()
    expect(MIN_QUERY_LENGTH).toBe(3)
  })

  it('maps predictions, caps them at five and sends the key as a header, never in the URL', async () => {
    const predictions = Array.from({ length: 7 }, (_, i) => ({
      placePrediction: { placeId: `p${i}`, text: { text: `Calle ${i}` } },
    }))
    const fetchFn = vi.fn().mockResolvedValue(jsonResponse({ suggestions: predictions }))
    const client = createPlacesClient(env, fetchFn as unknown as typeof fetch)!

    const found = await client.suggest('calle')

    expect(found).toHaveLength(5)
    expect(found[0]).toEqual({ placeId: 'p0', text: 'Calle 0' })
    const [url, init] = fetchFn.mock.calls[0]!
    expect(url).toBe('https://places.googleapis.com/v1/places:autocomplete')
    expect(String(url)).not.toContain('browser-key')
    expect(init.headers['X-Goog-Api-Key']).toBe('browser-key')
  })

  it('resolves a place to address and coordinates, asking only for those fields', async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      jsonResponse({
        formattedAddress: 'Cra. 7 #71-21, Bogotá',
        location: { latitude: 4.6486, longitude: -74.0628 },
      }),
    )
    const client = createPlacesClient(env, fetchFn as unknown as typeof fetch)!

    expect(await client.resolve('abc/123')).toEqual({
      address: 'Cra. 7 #71-21, Bogotá',
      latitude: 4.6486,
      longitude: -74.0628,
    })
    const [url, init] = fetchFn.mock.calls[0]!
    expect(url).toBe('https://places.googleapis.com/v1/places/abc%2F123')
    expect(init.headers['X-Goog-FieldMask']).toBe('formattedAddress,location')
  })

  it('a place without coordinates is not a selection', async () => {
    const fetchFn = vi.fn().mockResolvedValue(jsonResponse({ formattedAddress: 'Algún lugar' }))
    const client = createPlacesClient(env, fetchFn as unknown as typeof fetch)!

    expect(await client.resolve('x')).toBeNull()
  })

  it('throws on a failed request so the field can say so', async () => {
    const fetchFn = vi.fn().mockResolvedValue(jsonResponse({}, false))
    const client = createPlacesClient(env, fetchFn as unknown as typeof fetch)!

    await expect(client.suggest('calle')).rejects.toThrow()
    await expect(client.resolve('x')).rejects.toThrow()
  })
})

describe('simulated client (development without a key)', () => {
  it('is only built in development, and never talks to the network', async () => {
    expect(createPlacesClient({ dev: false })).toBeNull()
    const fetchFn = vi.fn()
    const client = createPlacesClient({ dev: true }, fetchFn as unknown as typeof fetch)!

    const [first] = await client.suggest('Calle 10 # 5-20')
    const selection = await client.resolve(first!.placeId)

    expect(selection).toMatchObject({ address: first!.text })
    expect(selection!.latitude).toBeGreaterThan(4)
    expect(fetchFn).not.toHaveBeenCalled()
  })
})
