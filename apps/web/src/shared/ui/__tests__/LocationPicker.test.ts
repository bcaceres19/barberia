import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import LocationPicker from '../LocationPicker.vue'
import type { LocateResult } from '../../model/geolocation'
import type { LatLng, MapPickerEngine, MapPickerHandle } from '../../model/mapPicker'

interface FakeMap {
  engine: MapPickerEngine
  handle: MapPickerHandle
  mount: ReturnType<typeof vi.fn>
  describe: ReturnType<typeof vi.fn>
  userMoves: (center: LatLng) => void
}

function fakeMap(describeText: string | null = 'Calle 1 # 2-3, Bogotá'): FakeMap {
  let onUserMove: (center: LatLng) => void = () => {}
  const handle: MapPickerHandle = {
    getCenter: vi.fn(() => ({ latitude: 0, longitude: 0 })),
    setCenter: vi.fn(),
    destroy: vi.fn(),
  }
  const mountFn = vi.fn(async (_el: HTMLElement, _s: LatLng, _z: number, cb: typeof onUserMove) => {
    onUserMove = cb
    return handle
  })
  const describeFn = vi.fn(async () => describeText)
  return {
    engine: { mount: mountFn, describe: describeFn },
    handle,
    mount: mountFn,
    describe: describeFn,
    userMoves: (c) => onUserMove(c),
  }
}

function mountIt(
  map: FakeMap | null,
  props: Record<string, unknown> = {},
  locate: () => Promise<LocateResult> = async () => ({ kind: 'unavailable' }),
) {
  return mount(LocationPicker, {
    attachTo: document.body,
    props: {
      modelValue: null,
      engine: map ? map.engine : null,
      locate,
      canLocate: true,
      ...props,
    },
  })
}

const lastEmit = (w: ReturnType<typeof mountIt>) => {
  const all = w.emitted('update:modelValue')!
  return all[all.length - 1]![0]
}

describe('LocationPicker', () => {
  it('opens the map on the given start with the city zoom when nothing is known', async () => {
    const map = fakeMap()
    mountIt(map, { start: { latitude: 6.2, longitude: -75.5 } })
    await flushPromises()

    expect(map.mount.mock.calls[0]!.slice(1, 3)).toEqual([{ latitude: 6.2, longitude: -75.5 }, 15])
  })

  it('opens close on an already chosen point', async () => {
    const map = fakeMap()
    mountIt(map, { modelValue: { address: 'Calle 9', latitude: 4.6, longitude: -74.1 } })
    await flushPromises()

    expect(map.mount.mock.calls[0]!.slice(1, 3)).toEqual([{ latitude: 4.6, longitude: -74.1 }, 17])
  })

  it('emits the point with a fallback text at once and then the approximate address (CA-138-01)', async () => {
    const map = fakeMap()
    const wrapper = mountIt(map)
    await flushPromises()

    map.userMoves({ latitude: 4.6123456789, longitude: -74.0698765432 })
    const emitted = wrapper.emitted('update:modelValue')!
    expect(emitted[0]![0]).toEqual({
      address: 'Ubicación marcada en el mapa',
      latitude: 4.612346,
      longitude: -74.069877,
    })
    await flushPromises()
    expect(lastEmit(wrapper)).toEqual({
      address: 'Calle 1 # 2-3, Bogotá',
      latitude: 4.612346,
      longitude: -74.069877,
    })
  })

  it('keeps the fallback text when there is no approximate address or geocoding fails', async () => {
    const none = fakeMap(null)
    const w1 = mountIt(none)
    await flushPromises()
    none.userMoves({ latitude: 4.6, longitude: -74 })
    await flushPromises()
    expect(w1.emitted('update:modelValue')).toHaveLength(1)

    const failing = fakeMap()
    failing.describe.mockRejectedValue(new Error('geocoder down'))
    const w2 = mountIt(failing)
    await flushPromises()
    failing.userMoves({ latitude: 4.6, longitude: -74 })
    await flushPromises()
    expect(w2.emitted('update:modelValue')).toHaveLength(1)
  })

  it('drops a late description of a point the person already moved away from', async () => {
    const map = fakeMap()
    let resolveFirst: (text: string) => void = () => {}
    map.describe
      .mockImplementationOnce(() => new Promise<string>((r) => (resolveFirst = r)))
      .mockResolvedValueOnce('Segunda calle')
    const wrapper = mountIt(map)
    await flushPromises()

    map.userMoves({ latitude: 1, longitude: 1 })
    map.userMoves({ latitude: 2, longitude: 2 })
    await flushPromises()
    resolveFirst('Primera calle')
    await flushPromises()

    expect(lastEmit(wrapper)).toMatchObject({ address: 'Segunda calle', latitude: 2 })
  })

  it('trims very long addresses to the 200 characters the server accepts', async () => {
    const map = fakeMap('x'.repeat(260))
    const wrapper = mountIt(map)
    await flushPromises()
    map.userMoves({ latitude: 4.6, longitude: -74 })
    await flushPromises()

    expect((lastEmit(wrapper) as { address: string }).address).toHaveLength(200)
  })

  it('recenters on a point chosen elsewhere without emitting it back (suggestion picked)', async () => {
    const map = fakeMap()
    const wrapper = mountIt(map)
    await flushPromises()

    await wrapper.setProps({ modelValue: { address: 'Cra 7', latitude: 4.65, longitude: -74.06 } })
    expect(map.handle.setCenter).toHaveBeenCalledWith({ latitude: 4.65, longitude: -74.06 }, 17)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('does not jump the map when the parent echoes the point it just emitted', async () => {
    const map = fakeMap()
    const wrapper = mountIt(map)
    await flushPromises()
    map.userMoves({ latitude: 4.6, longitude: -74 })
    await wrapper.setProps({
      modelValue: { address: 'Ubicación marcada en el mapa', latitude: 4.6, longitude: -74 },
    })

    expect(map.handle.setCenter).not.toHaveBeenCalled()
  })

  describe('Usar mi ubicación', () => {
    it('centers on the detected position, marks it and says how precise it is (CA-138-02)', async () => {
      const map = fakeMap()
      const locate = vi.fn(async (): Promise<LocateResult> => ({
        kind: 'success',
        point: { latitude: 4.65, longitude: -74.06 },
        accuracyMeters: 25,
      }))
      const wrapper = mountIt(map, {}, locate)
      await flushPromises()
      expect(locate).not.toHaveBeenCalled()

      await wrapper.get('button').trigger('click')
      await flushPromises()

      expect(map.handle.setCenter).toHaveBeenCalledWith({ latitude: 4.65, longitude: -74.06 }, 17)
      expect(lastEmit(wrapper)).toMatchObject({ latitude: 4.65, longitude: -74.06 })
      expect(wrapper.text()).toContain('Ubicamos tu posición. Ajusta el pin si hace falta.')
    })

    it('warns when the fix is coarse', async () => {
      const wrapper = mountIt(fakeMap(), {}, async () => ({
        kind: 'success',
        point: { latitude: 4.65, longitude: -74.06 },
        accuracyMeters: 850,
      }))
      await flushPromises()
      await wrapper.get('button').trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('precisión de unos 850 m')
    })

    it.each([
      ['denied', 'No tenemos permiso para ver tu ubicación'],
      ['unavailable', 'No pudimos detectar tu ubicación'],
      ['timeout', 'Tardó demasiado en detectar tu ubicación'],
      ['unsupported', 'Este navegador no puede detectar tu ubicación'],
    ] as const)('explains a %s outcome and marks nothing (CA-138-04)', async (kind, message) => {
      const wrapper = mountIt(fakeMap(), {}, async () => ({ kind }))
      await flushPromises()
      await wrapper.get('button').trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain(message)
      expect(wrapper.emitted('update:modelValue')).toBeUndefined()
      expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
    })

    it('is busy while detecting and ignores a second press', async () => {
      let finish: (r: LocateResult) => void = () => {}
      const locate = vi.fn(() => new Promise<LocateResult>((r) => (finish = r)))
      const wrapper = mountIt(fakeMap(), {}, locate)
      await flushPromises()

      await wrapper.get('button').trigger('click')
      expect(wrapper.get('button').text()).toContain('Buscando tu ubicación')
      await wrapper.get('button').trigger('click')
      expect(locate).toHaveBeenCalledTimes(1)

      finish({ kind: 'denied' })
      await flushPromises()
      expect(wrapper.get('button').text()).toContain('Usar mi ubicación')
    })

    it('hides the button when the browser cannot detect the position', async () => {
      const wrapper = mountIt(fakeMap(), { canLocate: false })
      await flushPromises()
      expect(wrapper.find('button').exists()).toBe(false)
    })

    it('still detects the position when the map failed to load (CA-138-06)', async () => {
      const wrapper = mountIt(null, {}, async () => ({
        kind: 'success',
        point: { latitude: 4.65, longitude: -74.06 },
        accuracyMeters: 10,
      }))
      await flushPromises()
      expect(wrapper.text()).toContain('No pudimos cargar el mapa')

      await wrapper.get('button').trigger('click')
      await flushPromises()
      expect(lastEmit(wrapper)).toMatchObject({
        address: 'Ubicación marcada en el mapa',
        latitude: 4.65,
      })
    })
  })

  it('shows the map load failure when the engine cannot mount', async () => {
    const map = fakeMap()
    map.mount.mockRejectedValue(new Error('script blocked'))
    const wrapper = mountIt(map)
    await flushPromises()

    expect(wrapper.text()).toContain(
      'No pudimos cargar el mapa. Busca tu dirección o usa tu ubicación.',
    )
    expect(wrapper.find('.location-picker__pin').exists()).toBe(false)
  })

  it('draws the center pin only once the map is ready', async () => {
    const wrapper = mountIt(fakeMap())
    expect(wrapper.find('.location-picker__pin').exists()).toBe(false)
    await flushPromises()
    expect(wrapper.find('.location-picker__pin').exists()).toBe(true)
  })

  it('announces the chosen location and shows an error linked to the map', async () => {
    const wrapper = mountIt(fakeMap(), {
      modelValue: { address: 'Calle 9', latitude: 4.6, longitude: -74.1 },
      error: 'Marca tu ubicación en el mapa.',
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Ubicación marcada: Calle 9')
    const alert = wrapper.get('[role="alert"]')
    expect(wrapper.get('[role="group"]').attributes('aria-describedby')).toContain(
      alert.attributes('id'),
    )
  })

  it('destroys the map on unmount and ignores a late mount', async () => {
    const map = fakeMap()
    const wrapper = mountIt(map)
    await flushPromises()
    wrapper.unmount()
    expect(map.handle.destroy).toHaveBeenCalledTimes(1)

    const late = fakeMap()
    let release: (h: MapPickerHandle) => void = () => {}
    late.mount.mockImplementation(() => new Promise<MapPickerHandle>((r) => (release = r)))
    const w2 = mountIt(late)
    w2.unmount()
    release(late.handle)
    await flushPromises()
    expect(late.handle.destroy).toHaveBeenCalledTimes(1)
  })

  it('has no accessibility violations', async () => {
    const wrapper = mountIt(fakeMap(), {
      modelValue: { address: 'Calle 9', latitude: 4.6, longitude: -74.1 },
    })
    await flushPromises()
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
