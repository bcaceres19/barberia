import { describe, expect, it } from 'vitest'
import { affectedBetween, civilDaysTouched } from '../affectedAppointments'

const entry = (id: string, startsAt: string, endsAt: string, status = 'confirmed') => ({
  id,
  attendeeName: 'Ana',
  serviceName: 'Corte',
  startsAt,
  endsAt,
  status,
})

describe('affectedBetween (HU-069)', () => {
  const from = '2099-01-06T14:00:00Z'
  const to = '2099-01-06T16:00:00Z'

  it('keeps only confirmed appointments that overlap the half-open interval', () => {
    const result = affectedBetween(
      [
        entry('inside', '2099-01-06T14:30:00Z', '2099-01-06T15:00:00Z'),
        entry('touches-end', '2099-01-06T16:00:00Z', '2099-01-06T16:30:00Z'),
        entry('touches-start', '2099-01-06T13:30:00Z', '2099-01-06T14:00:00Z'),
        entry('straddles', '2099-01-06T15:30:00Z', '2099-01-06T16:30:00Z'),
        entry('cancelled', '2099-01-06T14:30:00Z', '2099-01-06T15:00:00Z', 'cancelled_by_customer'),
        entry('completed', '2099-01-06T14:30:00Z', '2099-01-06T15:00:00Z', 'completed'),
      ],
      from,
      to,
    )
    expect(result.map((item) => item.id)).toEqual(['inside', 'straddles'])
  })

  it('drops duplicates (a turn that crosses midnight appears in two daily agendas) and sorts', () => {
    const crossing = entry('x', '2099-01-06T22:00:00Z', '2099-01-07T01:00:00Z')
    const result = affectedBetween(
      [crossing, entry('early', '2099-01-06T20:00:00Z', '2099-01-06T21:00:00Z'), crossing],
      '2099-01-06T19:00:00Z',
      '2099-01-07T02:00:00Z',
    )
    expect(result.map((item) => item.id)).toEqual(['early', 'x'])
  })
})

describe('civilDaysTouched', () => {
  it('lists the civil days of the shop timezone the interval touches', () => {
    // 23:00 del día 6 a 01:00 del día 7 en Bogotá (UTC-5).
    expect(
      civilDaysTouched('2099-01-07T04:00:00Z', '2099-01-07T06:00:00Z', 'America/Bogota'),
    ).toEqual(['2099-01-06', '2099-01-07'])
  })

  it('does not touch the next day when the block ends exactly at midnight', () => {
    expect(
      civilDaysTouched('2099-01-06T15:00:00Z', '2099-01-07T05:00:00Z', 'America/Bogota'),
    ).toEqual(['2099-01-06'])
  })
})
