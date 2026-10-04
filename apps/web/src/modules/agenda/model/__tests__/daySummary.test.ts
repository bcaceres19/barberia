import { describe, expect, it } from 'vitest'
import type { DailyAgendaEntry } from '../dailyAgenda'
import { summarizeDay } from '../daySummary'

function entry(
  id: string,
  start: string,
  end: string,
  status: DailyAgendaEntry['status'] = 'confirmed',
) {
  return {
    id,
    attendeeName: `Cliente ${id}`,
    startsAt: `2026-10-04T${start}:00-05:00`,
    endsAt: `2026-10-04T${end}:00-05:00`,
    status,
    origin: 'manual',
    serviceName: `Servicio ${id}`,
    durationMinutes: 30,
    priceAmount: '20000.00',
    currency: 'COP',
  } satisfies DailyAgendaEntry
}

const at = (time: string) => new Date(`2026-10-04T${time}:00-05:00`).getTime()

describe('summarizeDay', () => {
  it('has nothing to say about an empty day', () => {
    expect(summarizeDay([], at('10:00'))).toEqual({
      total: 0,
      remaining: 0,
      inProgress: null,
      next: null,
    })
  })

  it('points at the next confirmed appointment before the day starts', () => {
    const day = [entry('b', '11:30', '12:30'), entry('a', '10:00', '10:40')]
    const summary = summarizeDay(day, at('08:00'))

    expect(summary.total).toBe(2)
    expect(summary.remaining).toBe(2)
    expect(summary.inProgress).toBeNull()
    // Aunque lleguen desordenados, «siguiente» es el primero en el tiempo.
    expect(summary.next?.id).toBe('a')
  })

  it('reports the appointment in progress and the one after it', () => {
    const day = [entry('a', '10:00', '10:40'), entry('b', '11:30', '12:30')]
    const summary = summarizeDay(day, at('10:15'))

    expect(summary.inProgress?.id).toBe('a')
    expect(summary.next?.id).toBe('b')
    expect(summary.remaining).toBe(2)
  })

  it('treats the end as exclusive: at 10:40 the appointment is over', () => {
    const day = [entry('a', '10:00', '10:40'), entry('b', '11:30', '12:30')]
    const summary = summarizeDay(day, at('10:40'))

    expect(summary.inProgress).toBeNull()
    expect(summary.remaining).toBe(1)
    expect(summary.next?.id).toBe('b')
  })

  it('ignores completed, cancelled and no-show appointments for what is left', () => {
    const day = [
      entry('a', '10:00', '10:40', 'completed'),
      entry('b', '11:30', '12:30', 'cancelled_by_customer'),
      entry('c', '13:00', '13:30', 'no_show'),
      entry('d', '15:00', '15:30'),
    ]
    const summary = summarizeDay(day, at('09:00'))

    expect(summary.total).toBe(4)
    expect(summary.remaining).toBe(1)
    expect(summary.next?.id).toBe('d')
  })

  it('does not count an unresolved confirmed appointment that already ended', () => {
    const summary = summarizeDay([entry('a', '10:00', '10:40')], at('16:00'))

    expect(summary.remaining).toBe(0)
    expect(summary.inProgress).toBeNull()
    expect(summary.next).toBeNull()
  })
})
