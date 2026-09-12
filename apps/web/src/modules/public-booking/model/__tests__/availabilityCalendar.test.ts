/**
 * Pruebas de groupSlotsByCivilDate (HU-095): agrupación por día civil en la
 * zona de la barbería, orden cronológico de días conservado sin un `sort`
 * adicional, y un instante que cruza medianoche en UTC pero no en la zona
 * de la barbería queda en el día civil correcto (RN-DIS-07).
 */
import { describe, it, expect } from 'vitest'
import { groupSlotsByCivilDate } from '../availabilityCalendar'

describe('groupSlotsByCivilDate', () => {
  it('returns an empty array for an empty slot list', () => {
    expect(groupSlotsByCivilDate([], 'America/Bogota')).toEqual([])
  })

  it('groups slots of the same civil date into a single day', () => {
    const days = groupSlotsByCivilDate(
      [{ startsAt: '2026-09-15T19:00:00Z' }, { startsAt: '2026-09-15T19:15:00Z' }],
      'America/Bogota',
    )
    expect(days).toHaveLength(1)
    expect(days[0]!.civilDate).toBe('2026-09-15')
    expect(days[0]!.slots).toHaveLength(2)
  })

  it('preserves chronological day order without an extra sort', () => {
    const days = groupSlotsByCivilDate(
      [
        { startsAt: '2026-09-15T19:00:00Z' },
        { startsAt: '2026-09-16T14:00:00Z' },
        { startsAt: '2026-09-17T14:00:00Z' },
      ],
      'America/Bogota',
    )
    expect(days.map((d) => d.civilDate)).toEqual(['2026-09-15', '2026-09-16', '2026-09-17'])
  })

  it('resolves the civil date using the barbershop timezone, not UTC (RN-DIS-07)', () => {
    // 21:30 UTC del 15 de septiembre es 09:30 del 16 en Asia/Tokyo (UTC+9):
    // en esa zona el turno cae en el día siguiente al de UTC.
    const days = groupSlotsByCivilDate([{ startsAt: '2026-09-15T21:30:00Z' }], 'Asia/Tokyo')
    expect(days[0]!.civilDate).toBe('2026-09-16')
  })

  it('keeps two slots in the same civil day distinct even if their UTC dates differ', () => {
    // 23:30 UTC del 15 y 00:15 UTC del 16 son ambos 2026-09-15 en
    // America/Bogota (UTC-5): un solo día, dos franjas.
    const days = groupSlotsByCivilDate(
      [{ startsAt: '2026-09-15T23:30:00Z' }, { startsAt: '2026-09-16T00:15:00Z' }],
      'America/Bogota',
    )
    expect(days).toHaveLength(1)
    expect(days[0]!.civilDate).toBe('2026-09-15')
    expect(days[0]!.slots).toHaveLength(2)
  })
})
