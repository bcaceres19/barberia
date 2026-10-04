import { describe, expect, it } from 'vitest'
import {
  civilDateTile,
  computeScale,
  endLabel,
  formatHours,
  minutesToTime,
  nowInTimezone,
  scaleTicks,
  segmentGeometry,
  timeToMinutes,
} from '../weekBoard'

describe('weekBoard', () => {
  it('convierte entre HH:MM y minutos del día', () => {
    expect(timeToMinutes('09:30')).toBe(570)
    expect(minutesToTime(570)).toBe('09:30')
    expect(minutesToTime(1500)).toBe('01:00')
  })

  it('mantiene la regla de 07 a 21 h salvo que un tramo se salga', () => {
    expect(computeScale([])).toEqual({ startMinute: 420, endMinute: 1260 })
    expect(computeScale([{ startsTime: '09:00', durationMinutes: 480 }])).toEqual({
      startMinute: 420,
      endMinute: 1260,
    })
    expect(computeScale([{ startsTime: '05:30', durationMinutes: 60 }]).startMinute).toBe(300)
    expect(computeScale([{ startsTime: '18:00', durationMinutes: 300 }]).endMinute).toBe(1380)
  })

  it('recorta un tramo nocturno al fin del día y lo marca', () => {
    const scale = computeScale([{ startsTime: '20:00', durationMinutes: 480 }])
    expect(scale.endMinute).toBe(1440)
    const geometry = segmentGeometry('20:00', 480, scale)
    expect(geometry.crossesMidnight).toBe(true)
    expect(geometry.from + geometry.span).toBeCloseTo(100)
    expect(endLabel('20:00', 480)).toBe('04:00 (+1)')
  })

  it('ignora una hora a medio escribir al calcular la regla', () => {
    expect(computeScale([{ startsTime: '', durationMinutes: 60 }])).toEqual({
      startMinute: 420,
      endMinute: 1260,
    })
  })

  it('posiciona un tramo como porcentaje de la regla', () => {
    const scale = { startMinute: 420, endMinute: 1260 }
    const geometry = segmentGeometry('09:00', 240, scale)
    expect(geometry.from).toBeCloseTo((120 / 840) * 100)
    expect(geometry.span).toBeCloseTo((240 / 840) * 100)
    expect(geometry.crossesMidnight).toBe(false)
  })

  it('rotula la regla cada dos horas en una jornada normal', () => {
    const ticks = scaleTicks({ startMinute: 420, endMinute: 1260 })
    expect(ticks.map((t) => t.label)).toEqual(['07', '09', '11', '13', '15', '17', '19', '21'])
    expect(ticks[0]!.percent).toBe(0)
    expect(ticks.at(-1)!.percent).toBe(100)
  })

  it('formatea horas legibles', () => {
    expect(formatHours(480)).toBe('8 h')
    expect(formatHours(450)).toBe('7 h 30 min')
    expect(formatHours(45)).toBe('45 min')
  })

  it('lee el día y la hora en la zona de la barbería, no en la del dispositivo', () => {
    // 2026-10-05 03:30 UTC es domingo 4 a las 22:30 en Bogotá (UTC-5).
    const now = nowInTimezone('America/Bogota', new Date('2026-10-05T03:30:00Z'))
    expect(now).toEqual({ isoWeekday: 7, minute: 22 * 60 + 30 })
  })

  it('parte una fecha civil en día y mes abreviado sin tocar la zona del dispositivo', () => {
    expect(civilDateTile('2026-12-08')).toEqual({ day: '08', month: 'DIC' })
    expect(civilDateTile('2027-01-01')).toEqual({ day: '01', month: 'ENE' })
    expect(civilDateTile('')).toEqual({ day: '', month: '' })
  })

  it('ignora una duración inválida al calcular la regla', () => {
    expect(computeScale([{ startsTime: '09:00', durationMinutes: Number.NaN }])).toEqual({
      startMinute: 420,
      endMinute: 1260,
    })
  })
})
