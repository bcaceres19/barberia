import { describe, expect, it } from 'vitest'
import { cancellationAvailability } from '../cancellationAvailability'

// Misma tabla que el servidor (booking.CustomerCancellationPolicy.Evaluate,
// HU-099): antes / exacto / después del límite × política tardía × motivo.
const STARTS_AT = '2027-03-15T15:00:00Z'
const startsMs = Date.parse(STARTS_AT)
const limitMs = startsMs - 20 * 60_000

function policy(over: Partial<Parameters<typeof cancellationAvailability>[0]> = {}) {
  return {
    startsAt: STARTS_AT,
    cancellationDeadlineMinutes: 20,
    lateCancellationClientAllowed: true,
    lateCancellationReasonRequired: true,
    ...over,
  }
}

describe('cancellationAvailability', () => {
  it('antes del límite es dentro de plazo y no exige motivo', () => {
    expect(cancellationAvailability(policy(), limitMs - 1)).toEqual({
      mode: 'within-deadline',
      reasonRequired: false,
    })
  })

  it('exactamente en el límite sigue siendo dentro de plazo', () => {
    expect(cancellationAvailability(policy(), limitMs).mode).toBe('within-deadline')
  })

  it('un instante después del límite es tardía y hereda el motivo obligatorio', () => {
    expect(cancellationAvailability(policy(), limitMs + 1)).toEqual({
      mode: 'late-allowed',
      reasonRequired: true,
    })
    expect(
      cancellationAvailability(policy({ lateCancellationReasonRequired: false }), limitMs + 1),
    ).toEqual({ mode: 'late-allowed', reasonRequired: false })
  })

  it('tardía con la política que no la permite no ofrece cancelar', () => {
    expect(
      cancellationAvailability(
        policy({ lateCancellationClientAllowed: false, lateCancellationReasonRequired: false }),
        limitMs + 1,
      ),
    ).toEqual({ mode: 'unavailable', reason: 'late-not-allowed' })
  })

  it('con el turno ya iniciado nunca procede, aun con la política tardía permitida', () => {
    expect(cancellationAvailability(policy(), startsMs)).toEqual({
      mode: 'unavailable',
      reason: 'started',
    })
    expect(cancellationAvailability(policy(), startsMs + 60_000).mode).toBe('unavailable')
  })

  it('plazo 0 trata todo como tardío (DEC-083)', () => {
    const farBefore = startsMs - 24 * 3_600_000
    expect(
      cancellationAvailability(policy({ cancellationDeadlineMinutes: 0 }), farBefore).mode,
    ).toBe('late-allowed')
    expect(
      cancellationAvailability(
        policy({
          cancellationDeadlineMinutes: 0,
          lateCancellationClientAllowed: false,
          lateCancellationReasonRequired: false,
        }),
        farBefore,
      ),
    ).toEqual({ mode: 'unavailable', reason: 'late-not-allowed' })
  })
})
