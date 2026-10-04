/**
 * Catálogo de zonas del selector (DEC-110): lista IANA del navegador, búsqueda
 * por ciudad o desfase, reloj de la zona y desfase con horario de verano.
 */
import { describe, expect, it } from 'vitest'
import {
  cityOf,
  clockParts,
  detectDeviceTimezone,
  filterTimezones,
  isKnownTimezone,
  listTimezones,
  offsetLabel,
} from '../timezones'

// 2026-10-03 12:00 UTC: Bogotá no cambia de hora; Nueva York y Madrid están en verano.
const AT = new Date('2026-10-03T12:00:00Z')

describe('listTimezones', () => {
  it('lists IANA zones sorted, with Bogotá among them', () => {
    const zones = listTimezones()
    expect(zones).toContain('America/Bogota')
    expect(zones.length).toBeGreaterThan(50)
    expect(zones).toEqual([...zones].sort((a, b) => a.localeCompare(b, 'en')))
  })

  it('always includes the current zone of the barbershop, even if the browser does not list it', () => {
    expect(listTimezones(['Legacy/Zona_Vieja'])).toContain('Legacy/Zona_Vieja')
  })

  it('does not repeat a zone', () => {
    const zones = listTimezones(['America/Bogota'])
    expect(zones.filter((z) => z === 'America/Bogota')).toHaveLength(1)
  })
})

describe('offsetLabel', () => {
  it('formats whole hours, half hours and UTC', () => {
    expect(offsetLabel('America/Bogota', AT)).toBe('UTC−5')
    expect(offsetLabel('Asia/Kolkata', AT)).toBe('UTC+5:30')
    expect(offsetLabel('UTC', AT)).toBe('UTC')
  })

  it('follows daylight saving time at the given instant', () => {
    expect(offsetLabel('America/New_York', AT)).toBe('UTC−4')
    expect(offsetLabel('America/New_York', new Date('2026-12-03T12:00:00Z'))).toBe('UTC−5')
    expect(offsetLabel('Europe/Madrid', AT)).toBe('UTC+2')
  })

  it('answers an empty label for a zone the browser does not know', () => {
    expect(offsetLabel('No/Existe', AT)).toBe('')
  })
})

describe('clockParts', () => {
  it('reads the time and the day in the zone, never in the device zone', () => {
    const bogota = clockParts('America/Bogota', AT)!
    expect(bogota.hours).toBe('07')
    expect(bogota.minutes).toBe('00')
    expect(bogota.day).toBe('Sábado, 3 de octubre')

    const madrid = clockParts('Europe/Madrid', AT)!
    expect(madrid.hours).toBe('14')
  })

  it('crosses midnight correctly', () => {
    const tokyo = clockParts('Asia/Tokyo', new Date('2026-10-03T20:30:00Z'))!
    expect(tokyo.hours).toBe('05')
    expect(tokyo.day).toBe('Domingo, 4 de octubre')
  })

  it('answers null for a zone the browser does not know', () => {
    expect(clockParts('No/Existe', AT)).toBeNull()
  })
})

describe('filterTimezones', () => {
  const zones = [
    'America/Bogota',
    'America/New_York',
    'America/Argentina/Buenos_Aires',
    'Europe/Madrid',
    'Asia/Kolkata',
  ]

  it('returns everything for an empty search', () => {
    expect(filterTimezones(zones, '  ', AT)).toEqual(zones)
  })

  it('matches a city, ignoring case, accents and underscores', () => {
    expect(filterTimezones(zones, 'BOGOTÁ', AT)).toEqual(['America/Bogota'])
    expect(filterTimezones(zones, 'new york', AT)).toEqual(['America/New_York'])
    expect(filterTimezones(zones, 'buenos aires', AT)).toEqual(['America/Argentina/Buenos_Aires'])
  })

  it('matches the region and requires every word', () => {
    expect(filterTimezones(zones, 'america', AT)).toHaveLength(3)
    expect(filterTimezones(zones, 'america york', AT)).toEqual(['America/New_York'])
    expect(filterTimezones(zones, 'america madrid', AT)).toEqual([])
  })

  it('matches the offset, with a hyphen or a real minus sign', () => {
    expect(filterTimezones(zones, 'utc-5', AT)).toEqual(['America/Bogota'])
    expect(filterTimezones(zones, 'utc−5', AT)).toEqual(['America/Bogota'])
    expect(filterTimezones(zones, 'utc+5:30', AT)).toEqual(['Asia/Kolkata'])
  })
})

describe('helpers', () => {
  it('names a zone by its city', () => {
    expect(cityOf('America/Bogota')).toBe('Bogota')
    expect(cityOf('America/Argentina/Buenos_Aires')).toBe('Buenos Aires')
    expect(cityOf('UTC')).toBe('UTC')
  })

  it('recognizes IANA zones and rejects abbreviations and offsets', () => {
    expect(isKnownTimezone('America/Bogota')).toBe(true)
    expect(isKnownTimezone('COT')).toBe(false)
    expect(isKnownTimezone('UTC-5')).toBe(false)
    expect(isKnownTimezone('')).toBe(false)
  })

  it('detects the device zone only as a suggestion', () => {
    const zone = detectDeviceTimezone()
    expect(zone === null || typeof zone === 'string').toBe(true)
  })
})
