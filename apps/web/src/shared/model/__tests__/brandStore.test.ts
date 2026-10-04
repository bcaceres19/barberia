/**
 * Marca confirmada de la barbería (DEC-110): copia local para el primer
 * fotograma, vista previa del acento sin guardar y limpieza al cerrar sesión.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_BRAND, type BrandSettings } from '../vocabulary'

const KEY = 'nava.brand.v1'

const salon: BrandSettings = {
  accent: 'emerald',
  businessTerm: 'salón',
  businessTermGender: 'masculine',
  professionalTerm: 'estilista',
  professionalTermPlural: 'estilistas',
  professionalTermGender: 'feminine',
  panelProfile: 'solo',
}

async function load() {
  vi.resetModules()
  return import('../brandStore')
}

beforeEach(() => {
  window.localStorage.clear()
})

describe('brandStore', () => {
  it('starts on the interface the barbershop always had', async () => {
    const store = await load()
    expect({ ...store.brandState.brand }).toEqual(DEFAULT_BRAND)
    expect(store.effectiveAccent.value).toBe('brass')
    expect(store.vocabulary.value.Professionals).toBe('Barberos')
  })

  it('publishes the confirmed brand to the vocabulary and keeps a local copy', async () => {
    const store = await load()
    store.setBrand(salon)

    expect(store.vocabulary.value.theBusiness).toBe('el salón')
    expect(store.vocabulary.value.Professionals).toBe('Estilistas')
    expect(store.effectiveAccent.value).toBe('emerald')
    expect(JSON.parse(window.localStorage.getItem(KEY)!)).toEqual(salon)
  })

  it('exposes the panel profile and reads a copy saved before profiles existed as the full panel', async () => {
    const store = await load()
    expect(store.panelProfile.value).toBe('shop')
    expect(store.isSoloProfile.value).toBe(false)

    store.setBrand(salon)
    expect(store.panelProfile.value).toBe('solo')
    expect(store.isSoloProfile.value).toBe(true)

    const legacy: Partial<BrandSettings> = { ...salon }
    delete legacy.panelProfile
    window.localStorage.setItem(KEY, JSON.stringify(legacy))
    const next = await load()
    expect(next.brandState.brand.accent).toBe('emerald')
    expect(next.panelProfile.value).toBe('shop')

    window.localStorage.setItem(KEY, JSON.stringify({ ...salon, panelProfile: 'team' }))
    expect((await load()).panelProfile.value).toBe('shop')
  })

  it('paints the stored copy on the next visit before the server answers', async () => {
    ;(await load()).setBrand(salon)

    const next = await load()
    expect({ ...next.brandState.brand }).toEqual(salon)
  })

  it('ignores a stored copy that is corrupt, partial or from a newer version', async () => {
    window.localStorage.setItem(KEY, '{no-json')
    expect({ ...(await load()).brandState.brand }).toEqual(DEFAULT_BRAND)

    window.localStorage.setItem(KEY, JSON.stringify({ ...salon, accent: 'neon' }))
    expect({ ...(await load()).brandState.brand }).toEqual(DEFAULT_BRAND)

    window.localStorage.setItem(KEY, JSON.stringify({ ...salon, businessTerm: 'x' }))
    expect({ ...(await load()).brandState.brand }).toEqual(DEFAULT_BRAND)

    window.localStorage.setItem(KEY, JSON.stringify({ accent: 'ruby' }))
    expect({ ...(await load()).brandState.brand }).toEqual(DEFAULT_BRAND)
  })

  it('previews an accent without confirming it, and goes back on null', async () => {
    const store = await load()
    store.setBrand(salon)

    store.setAccentPreview('ruby')
    expect(store.effectiveAccent.value).toBe('ruby')
    expect(store.brandState.brand.accent).toBe('emerald')
    expect(JSON.parse(window.localStorage.getItem(KEY)!).accent).toBe('emerald')

    store.setAccentPreview(null)
    expect(store.effectiveAccent.value).toBe('emerald')
  })

  it('forgets everything on logout so the next barbershop never inherits it', async () => {
    const store = await load()
    store.setBrand(salon)
    store.setAccentPreview('ruby')

    store.resetBrand()

    expect({ ...store.brandState.brand }).toEqual(DEFAULT_BRAND)
    expect(store.effectiveAccent.value).toBe('brass')
    expect(window.localStorage.getItem(KEY)).toBeNull()
  })

  it('keeps working when storage is unavailable', async () => {
    const setItem = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked')
    })
    const store = await load()

    expect(() => store.setBrand(salon)).not.toThrow()
    expect(store.brandState.brand.accent).toBe('emerald')
    setItem.mockRestore()
  })
})
