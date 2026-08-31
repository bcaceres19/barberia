// Pruebas de las utilidades de fecha civil de HU-063: aritmética de
// calendario (no de reloj), fin de mes/año, día civil "hoy" en la zona
// explícita de la barbería (nunca la del dispositivo) y formato completo en
// español a partir de una fecha ya resuelta.
import { describe, expect, it } from 'vitest'
import {
  formatCivilDateFull,
  getCivilDateInTimezone,
  isCivilDateString,
  shiftCivilDate,
} from '../civilDate'

describe('isCivilDateString', () => {
  it('accepts a real calendar date in AAAA-MM-DD', () => {
    expect(isCivilDateString('2026-08-31')).toBe(true)
  })

  it('rejects a day that does not exist in that month', () => {
    expect(isCivilDateString('2026-02-30')).toBe(false)
  })

  it('rejects other formats, including a full instant', () => {
    expect(isCivilDateString('2026-8-31')).toBe(false)
    expect(isCivilDateString('2026-08-31T00:00:00Z')).toBe(false)
    expect(isCivilDateString('not-a-date')).toBe(false)
  })
})

describe('shiftCivilDate', () => {
  it('advances and rewinds a single civil day', () => {
    expect(shiftCivilDate('2026-08-31', 1)).toBe('2026-09-01')
    expect(shiftCivilDate('2026-08-31', -1)).toBe('2026-08-30')
  })

  it('crosses a month and a year boundary', () => {
    expect(shiftCivilDate('2026-01-01', -1)).toBe('2025-12-31')
    expect(shiftCivilDate('2026-12-31', 1)).toBe('2027-01-01')
  })

  it('is exact across a leap-year February', () => {
    expect(shiftCivilDate('2028-02-28', 1)).toBe('2028-02-29')
    expect(shiftCivilDate('2028-02-29', 1)).toBe('2028-03-01')
  })

  it('advances exactly one civil day across a DST transition (23h/25h local day)', () => {
    // America/Bogota no observa DST, pero shiftCivilDate nunca consulta una
    // zona: la aritmética es sobre el triplete año/mes/día, así que un día
    // local de 23 o 25 horas en cualquier otra zona no la afecta.
    expect(shiftCivilDate('2026-03-08', 1)).toBe('2026-03-09')
    expect(shiftCivilDate('2026-11-01', 1)).toBe('2026-11-02')
  })
})

describe('getCivilDateInTimezone', () => {
  it('reads the civil date from the explicit timezone received, never the device one', () => {
    // Mismo instante, dos zonas con desfase suficiente para caer en días
    // civiles distintos.
    const at = new Date('2026-08-31T02:00:00Z')
    expect(getCivilDateInTimezone('America/Bogota', at)).toBe('2026-08-30')
    expect(getCivilDateInTimezone('Asia/Tokyo', at)).toBe('2026-08-31')
  })

  it('is correct right around a DST transition', () => {
    // 2026-03-08T06:30:00Z es 2026-03-08T01:30 en America/New_York, antes
    // de que el reloj salte a las 03:00 ese día (DST); sigue siendo el
    // mismo día civil.
    const beforeSpringForward = new Date('2026-03-08T06:30:00Z')
    expect(getCivilDateInTimezone('America/New_York', beforeSpringForward)).toBe('2026-03-08')
  })
})

describe('formatCivilDateFull', () => {
  it('formats a resolved civil date fully in Spanish, independent of any timezone', () => {
    expect(formatCivilDateFull('2026-08-31')).toBe('lunes, 31 de agosto de 2026')
  })

  it('never shifts the day, even for a date that would cross UTC+14 near midnight', () => {
    expect(formatCivilDateFull('2026-01-01')).toBe('jueves, 1 de enero de 2026')
    expect(formatCivilDateFull('2026-12-31')).toBe('jueves, 31 de diciembre de 2026')
  })
})
