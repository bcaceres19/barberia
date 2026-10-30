import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  createMapPickerEngine,
  isValidPoint,
  loadGoogleMaps,
  roundPoint,
  samePoint,
} from '../mapPicker'

afterEach(() => {
  Reflect.deleteProperty(window, 'google')
  document.head.querySelectorAll('script').forEach((s) => s.remove())
  vi.restoreAllMocks()
})

describe('point helpers', () => {
  it('rounds to 6 decimals', () => {
    expect(roundPoint({ latitude: 4.12345678, longitude: -74.98765432 })).toEqual({
      latitude: 4.123457,
      longitude: -74.987654,
    })
  })

  it('treats points within ~1 m as the same', () => {
    const a = { latitude: 4.6, longitude: -74 }
    expect(samePoint(a, { latitude: 4.600001, longitude: -74.000001 })).toBe(true)
    expect(samePoint(a, { latitude: 4.601, longitude: -74 })).toBe(false)
  })

  it('validates ranges and finiteness', () => {
    expect(isValidPoint({ latitude: 4.6, longitude: -74 })).toBe(true)
    expect(isValidPoint({ latitude: 91, longitude: 0 })).toBe(false)
    expect(isValidPoint({ latitude: 0, longitude: 181 })).toBe(false)
    expect(isValidPoint({ latitude: Number.NaN, longitude: 0 })).toBe(false)
  })
})

describe('createMapPickerEngine', () => {
  it('uses Google with a key, the simulator in dev, and nothing otherwise', () => {
    expect(createMapPickerEngine({ key: 'k', dev: false })).not.toBeNull()
    expect(createMapPickerEngine({ dev: true })).not.toBeNull()
    expect(createMapPickerEngine({ dev: false })).toBeNull()
  })
})

describe('simulated engine', () => {
  it('moves the point only when the person presses a direction', async () => {
    const engine = createMapPickerEngine({ dev: true })!
    const el = document.createElement('div')
    const onMove = vi.fn()
    const handle = await engine.mount(el, { latitude: 4.7, longitude: -74.07 }, 11, onMove)

    expect(onMove).not.toHaveBeenCalled()
    el.querySelector<HTMLButtonElement>('[aria-label="Mover el mapa al norte"]')!.click()
    el.querySelector<HTMLButtonElement>('[aria-label="Mover el mapa al este"]')!.click()

    expect(onMove).toHaveBeenCalledTimes(2)
    expect(handle.getCenter().latitude).toBeCloseTo(4.701, 6)
    expect(handle.getCenter().longitude).toBeCloseTo(-74.069, 6)

    handle.setCenter({ latitude: 1, longitude: 2 })
    expect(onMove).toHaveBeenCalledTimes(2)
    expect(handle.getCenter()).toEqual({ latitude: 1, longitude: 2 })

    handle.destroy()
    expect(el.querySelector('button')).toBeNull()
  })

  it('describes a point without network', async () => {
    const engine = createMapPickerEngine({ dev: true })!
    expect(await engine.describe({ latitude: 4.65, longitude: -74.06 })).toBe(
      'Punto marcado cerca de 4.6500, -74.0600',
    )
  })
})

describe('Google engine', () => {
  function installFakeGoogle() {
    const listeners: Record<string, (e?: unknown) => void> = {}
    let center = { lat: 0, lng: 0 }
    const map = {
      getCenter: () => ({ lat: () => center.lat, lng: () => center.lng }),
      setCenter: vi.fn((c: { lat: number; lng: number }) => (center = c)),
      setZoom: vi.fn(),
      panTo: vi.fn(),
      addListener: vi.fn((event: string, handler: (e?: unknown) => void) => {
        listeners[event] = handler
        return { remove: vi.fn() }
      }),
    }
    const Map = vi.fn(function (_el: unknown, options: { center: { lat: number; lng: number } }) {
      center = options.center
      return map
    })
    const geocode = vi.fn()
    const Geocoder = vi.fn(function () {
      return { geocode }
    })
    ;(window as unknown as { google: unknown }).google = {
      maps: {
        importLibrary: vi.fn(async (name: string) => (name === 'maps' ? { Map } : { Geocoder })),
      },
    }
    return { map, Map, geocode, listeners, move: (c: { lat: number; lng: number }) => (center = c) }
  }

  const engine = () => createMapPickerEngine({ key: 'browser-key', dev: false })!

  it('creates a map without default controls and reports only user moves', async () => {
    const fake = installFakeGoogle()
    const onMove = vi.fn()
    const handle = await engine().mount(
      document.createElement('div'),
      { latitude: 4.7, longitude: -74.07 },
      15,
      onMove,
    )

    expect(fake.Map.mock.calls[0]![1]).toMatchObject({
      center: { lat: 4.7, lng: -74.07 },
      zoom: 15,
      disableDefaultUI: true,
      zoomControl: true,
    })

    // idle sin cambio de centro: no es una acción de la persona.
    fake.listeners.idle!()
    expect(onMove).not.toHaveBeenCalled()

    // La persona arrastra el mapa.
    fake.move({ lat: 4.71, lng: -74.08 })
    fake.listeners.idle!()
    expect(onMove).toHaveBeenCalledWith({ latitude: 4.71, longitude: -74.08 })

    // Un movimiento por código no se confunde con uno de la persona.
    handle.setCenter({ latitude: 4.8, longitude: -74.1 }, 17)
    fake.listeners.idle!()
    expect(onMove).toHaveBeenCalledTimes(1)
    expect(fake.map.setZoom).toHaveBeenCalledWith(17)
  })

  it('pans to a tapped spot', async () => {
    const fake = installFakeGoogle()
    await engine().mount(document.createElement('div'), { latitude: 1, longitude: 2 }, 11, vi.fn())

    fake.listeners.click!({ latLng: { lat: () => 3, lng: () => 4 } })
    expect(fake.map.panTo).toHaveBeenCalledWith({ lat: 3, lng: 4 })
  })

  it('describes a point with reverse geocoding and tidies the text', async () => {
    const fake = installFakeGoogle()
    fake.geocode.mockResolvedValue({
      results: [{ formatted_address: '  Cra 7  #26-20,  Bogotá ' }],
    })

    expect(await engine().describe({ latitude: 4.6, longitude: -74.07 })).toBe(
      'Cra 7 #26-20, Bogotá',
    )
    expect(fake.geocode).toHaveBeenCalledWith({
      location: { lat: 4.6, lng: -74.07 },
      language: 'es',
    })
  })

  it('skips Plus Code results and falls back to null when only those come back', async () => {
    const fake = installFakeGoogle()
    fake.geocode.mockResolvedValue({
      results: [
        { formatted_address: '67P7JW7R+G8', types: ['plus_code'] },
        { formatted_address: 'Cra. 7 # 26-10, Bogotá', types: ['street_address'] },
      ],
    })
    expect(await engine().describe({ latitude: 4.6, longitude: -74.07 })).toBe(
      'Cra. 7 # 26-10, Bogotá',
    )

    fake.geocode.mockResolvedValue({
      results: [{ formatted_address: '67P7JW7R+G8', types: ['plus_code'] }],
    })
    expect(await engine().describe({ latitude: 4.6, longitude: -74.07 })).toBeNull()
  })

  it('returns null when there is no address or the request was aborted', async () => {
    const fake = installFakeGoogle()
    fake.geocode.mockResolvedValue({ results: [] })
    expect(await engine().describe({ latitude: 0, longitude: 0 })).toBeNull()

    fake.geocode.mockResolvedValue({ results: [{ formatted_address: 'Calle 1' }] })
    const controller = new AbortController()
    controller.abort()
    expect(await engine().describe({ latitude: 0, longitude: 0 }, controller.signal)).toBeNull()
  })
})

describe('loadGoogleMaps', () => {
  it('injects one script with the key, loads once and retries after a failure', async () => {
    const first = loadGoogleMaps('browser-key')
    const second = loadGoogleMaps('browser-key')
    const scripts = document.head.querySelectorAll('script')
    expect(scripts).toHaveLength(1)
    expect(scripts[0]!.src).toContain('https://maps.googleapis.com/maps/api/js?')
    expect(scripts[0]!.src).toContain('key=browser-key')
    expect(scripts[0]!.src).toContain('loading=async')

    scripts[0]!.onerror?.(new Event('error'))
    await expect(first).rejects.toThrow()
    await expect(second).rejects.toThrow()

    const retry = loadGoogleMaps('browser-key')
    const retried = document.head.querySelectorAll('script')
    expect(retried).toHaveLength(2)
    ;(window as unknown as { google: unknown }).google = { maps: { importLibrary: vi.fn() } }
    callReady(retried[1]!)
    await expect(retry).resolves.toBeDefined()
  })

  // Regresión: con `loading=async` el evento `load` llega antes de que exista `importLibrary`
  // (en Google real el mapa no cargaba en frío). Solo `callback` indica que la API está lista.
  it('waits for the API callback, not the script load event', async () => {
    const { loadGoogleMaps } = await freshModule()
    const loaded = loadGoogleMaps('browser-key')
    const script = document.head.querySelector('script')!
    script.onload?.(new Event('load'))
    let settled = false
    void loaded.then(
      () => (settled = true),
      () => (settled = true),
    )
    await Promise.resolve()
    expect(settled).toBe(false)

    ;(window as unknown as { google: unknown }).google = { maps: { importLibrary: vi.fn() } }
    callReady(script)
    await expect(loaded).resolves.toBeDefined()
  })

  it('rejects if the API reports ready without importLibrary', async () => {
    const { loadGoogleMaps } = await freshModule()
    const loaded = loadGoogleMaps('browser-key')
    callReady(document.head.querySelector('script')!)
    await expect(loaded).rejects.toThrow('maps-api-unavailable')
  })
})

/** El cargador recuerda la promesa de la primera carga: cada prueba pide un módulo limpio. */
async function freshModule() {
  vi.resetModules()
  return import('../mapPicker')
}

/** Invoca el `callback` que la URL del script le pide a Google, como haría la API al quedar lista. */
function callReady(script: HTMLScriptElement) {
  const name = new URL(script.src).searchParams.get('callback')!
  ;(window as unknown as Record<string, () => void>)[name]!()
}
