/**
 * Preferencias de pantalla de este dispositivo (DEC-110): valores iniciales,
 * persistencia en `localStorage`, validación de lo guardado y reinicio. El
 * almacén es un singleton, así que cada prueba lo importa de nuevo.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'

const KEY = 'nava.appearance.v1'

async function load() {
  vi.resetModules()
  return import('../appearanceStore')
}

beforeEach(() => {
  window.localStorage.clear()
})

describe('appearanceStore', () => {
  it('starts on the look the panel always had: Tinta, normal text, full motion', async () => {
    const store = await load()
    expect({ ...store.appearance }).toEqual({ theme: 'ink', textScale: 'normal', motion: 'full' })
    expect(store.isDefaultAppearance()).toBe(true)
  })

  it('persists every change and reads it back on the next visit', async () => {
    const first = await load()
    first.setTheme('ivory')
    first.setTextScale('large')
    first.setMotion('reduced')

    expect(JSON.parse(window.localStorage.getItem(KEY)!)).toEqual({
      theme: 'ivory',
      textScale: 'large',
      motion: 'reduced',
    })

    const second = await load()
    expect({ ...second.appearance }).toEqual({
      theme: 'ivory',
      textScale: 'large',
      motion: 'reduced',
    })
    expect(second.isDefaultAppearance()).toBe(false)
  })

  it('falls back to the initial values for corrupt or unknown stored data', async () => {
    window.localStorage.setItem(KEY, '{no-json')
    expect({ ...(await load()).appearance }).toMatchObject({ theme: 'ink' })

    window.localStorage.setItem(KEY, JSON.stringify(['ivory']))
    expect({ ...(await load()).appearance }).toMatchObject({ theme: 'ink' })

    // Un valor desconocido se descarta campo por campo; los válidos se conservan.
    window.localStorage.setItem(
      KEY,
      JSON.stringify({ theme: 'neon', textScale: 'xlarge', motion: 'wild' }),
    )
    expect({ ...(await load()).appearance }).toEqual({
      theme: 'ink',
      textScale: 'xlarge',
      motion: 'full',
    })
  })

  it('keeps working when storage is unavailable', async () => {
    const getItem = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked')
    })
    const setItem = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked')
    })

    const store = await load()
    expect(store.appearance.theme).toBe('ink')
    expect(() => store.setTheme('ivory')).not.toThrow()
    // La elección sigue valiendo durante la sesión aunque no se pueda guardar.
    expect(store.appearance.theme).toBe('ivory')

    getItem.mockRestore()
    setItem.mockRestore()
  })

  it('resets to the initial values and persists the reset', async () => {
    const store = await load()
    store.setTheme('auto')
    store.setTextScale('small')
    store.setMotion('reduced')

    store.resetAppearance()

    expect(store.isDefaultAppearance()).toBe(true)
    expect(JSON.parse(window.localStorage.getItem(KEY)!)).toEqual({
      theme: 'ink',
      textScale: 'normal',
      motion: 'full',
    })
  })

  it('offers the four text sizes, growing from small to extra large', async () => {
    const { TEXT_SCALES, TEXT_SCALE_KEYS } = await load()
    const factors = TEXT_SCALE_KEYS.map((key) => TEXT_SCALES[key].factor)
    expect(TEXT_SCALE_KEYS).toEqual(['small', 'normal', 'large', 'xlarge'])
    expect(factors).toEqual([...factors].sort((a, b) => a - b))
    expect(TEXT_SCALES.normal.factor).toBe(1)
  })
})
