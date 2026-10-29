import { describe, expect, it } from 'vitest'
import { isMapAvailable, mapEmbedUrl, mapsLinkUrl } from '../mapEmbed'

describe('mapEmbed', () => {
  it('con clave de navegador usa Maps Embed API centrado en el punto', () => {
    const url = mapEmbedUrl(4.598056, -74.076111, { key: 'browser-key', dev: false })
    expect(url).toBe(
      'https://www.google.com/maps/embed/v1/place?key=browser-key&q=4.598056%2C-74.076111&zoom=16',
    )
  })

  it('sin clave en producción no hay mapa (DEC-130: la sección queda oculta)', () => {
    expect(mapEmbedUrl(4.6, -74.1, { dev: false })).toBeNull()
    expect(isMapAvailable({ dev: false })).toBe(false)
  })

  it('sin clave en desarrollo usa el incrustado sin clave para poder verlo', () => {
    const url = mapEmbedUrl(4.6, -74.1, { dev: true })
    expect(url).toBe('https://maps.google.com/maps?q=4.6%2C-74.1&z=16&output=embed')
    expect(isMapAvailable({ dev: true })).toBe(true)
  })

  it('con clave siempre hay mapa, también en producción', () => {
    expect(isMapAvailable({ key: 'k', dev: false })).toBe(true)
  })

  it('rechaza puntos fuera de rango o no finitos', () => {
    const env = { key: 'k', dev: false }
    expect(mapEmbedUrl(91, 0, env)).toBeNull()
    expect(mapEmbedUrl(0, -181, env)).toBeNull()
    expect(mapEmbedUrl(Number.NaN, 0, env)).toBeNull()
    expect(mapsLinkUrl(0, Number.POSITIVE_INFINITY)).toBeNull()
  })

  it('escapa la clave: no puede inyectar parámetros en la URL', () => {
    const url = mapEmbedUrl(1, 2, { key: 'a&b=c', dev: false })
    expect(url).toContain('key=a%26b%3Dc')
  })

  it('el enlace de «abrir en Google Maps» no necesita clave', () => {
    expect(mapsLinkUrl(4.598056, -74.076111)).toBe(
      'https://www.google.com/maps/search/?api=1&query=4.598056%2C-74.076111',
    )
  })
})
