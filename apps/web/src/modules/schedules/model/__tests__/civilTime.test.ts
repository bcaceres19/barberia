import { describe, expect, it } from 'vitest'
import { civilDateTimeToInstant, formatInstantInTimezone } from '../civilTime'

describe('civilDateTimeToInstant', () => {
  it('convierte hora civil de Bogotá (UTC-5, sin horario de verano) a instante UTC', () => {
    // 15:00 en America/Bogota (siempre UTC-5, sin DST) es 20:00 UTC.
    expect(civilDateTimeToInstant('2026-07-20', '15:00', 'America/Bogota')).toBe(
      '2026-07-20T20:00:00.000Z',
    )
  })

  it('convierte hora civil que cruza medianoche hacia el día UTC siguiente', () => {
    // 22:00 en Bogotá (UTC-5) cae a las 03:00 UTC del día siguiente.
    expect(civilDateTimeToInstant('2026-07-20', '22:00', 'America/Bogota')).toBe(
      '2026-07-21T03:00:00.000Z',
    )
  })

  it('respeta el horario de verano de una zona que sí lo observa (Madrid, verano UTC+2)', () => {
    expect(civilDateTimeToInstant('2026-07-20', '15:00', 'Europe/Madrid')).toBe(
      '2026-07-20T13:00:00.000Z',
    )
  })

  it('respeta el horario estándar de una zona que observa DST (Madrid, invierno UTC+1)', () => {
    expect(civilDateTimeToInstant('2026-01-15', '15:00', 'Europe/Madrid')).toBe(
      '2026-01-15T14:00:00.000Z',
    )
  })

  it('lanza sobre una fecha u hora con forma inválida', () => {
    expect(() => civilDateTimeToInstant('no-es-fecha', '15:00', 'America/Bogota')).toThrow()
  })
})

describe('formatInstantInTimezone', () => {
  it('es la inversa de civilDateTimeToInstant para America/Bogota', () => {
    const instant = civilDateTimeToInstant('2026-07-20', '15:00', 'America/Bogota')
    expect(formatInstantInTimezone(instant, 'America/Bogota')).toEqual({
      date: '2026-07-20',
      time: '15:00',
    })
  })

  it('es la inversa de civilDateTimeToInstant a través de un cruce de medianoche', () => {
    const instant = civilDateTimeToInstant('2026-07-20', '22:30', 'America/Bogota')
    expect(formatInstantInTimezone(instant, 'America/Bogota')).toEqual({
      date: '2026-07-20',
      time: '22:30',
    })
  })
})
