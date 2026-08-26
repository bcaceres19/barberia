import { describe, it, expect } from 'vitest'
import {
  MAX_REASON_LENGTH,
  validateEffectiveDate,
  validateExceptionShape,
  validateReason,
} from '../exceptionValidation'

describe('validateEffectiveDate', () => {
  it('accepts real calendar dates', () => {
    for (const value of ['2026-01-01', '2026-12-31', '2028-02-29']) {
      expect(validateEffectiveDate(value)).toBeUndefined()
    }
  })

  it('rejects malformed or nonexistent dates', () => {
    for (const value of ['2026-13-01', '2026-02-30', 'not-a-date', '', '2026/01/01', '26-01-01']) {
      expect(validateEffectiveDate(value)).toBeDefined()
    }
  })
})

describe('validateReason', () => {
  it('accepts an empty reason and the maximum length', () => {
    expect(validateReason('')).toBeUndefined()
    expect(validateReason('a'.repeat(MAX_REASON_LENGTH))).toBeUndefined()
  })

  it('rejects a reason over the maximum length', () => {
    expect(validateReason('a'.repeat(MAX_REASON_LENGTH + 1))).toBeDefined()
  })
})

describe('validateExceptionShape', () => {
  it('accepts a closed day without segments', () => {
    expect(validateExceptionShape(true, [])).toBeUndefined()
  })

  it('rejects an open day without segments', () => {
    expect(validateExceptionShape(false, [])).toBeDefined()
  })

  it('accepts an open day with non-overlapping segments, including a night shift', () => {
    expect(
      validateExceptionShape(false, [
        { startsTime: '08:00', durationMinutes: 120 },
        { startsTime: '14:00', durationMinutes: 120 },
      ]),
    ).toBeUndefined()
    expect(
      validateExceptionShape(false, [{ startsTime: '22:00', durationMinutes: 300 }]),
    ).toBeUndefined()
  })

  it('accepts contiguous segments (semiopen [start, end))', () => {
    expect(
      validateExceptionShape(false, [
        { startsTime: '08:00', durationMinutes: 120 },
        { startsTime: '10:00', durationMinutes: 30 },
      ]),
    ).toBeUndefined()
  })

  it('rejects overlapping segments', () => {
    expect(
      validateExceptionShape(false, [
        { startsTime: '08:00', durationMinutes: 120 },
        { startsTime: '09:00', durationMinutes: 60 },
      ]),
    ).toBeDefined()
  })

  it('rejects a segment with an invalid starts time or duration', () => {
    expect(
      validateExceptionShape(false, [{ startsTime: '8am', durationMinutes: 60 }]),
    ).toBeDefined()
    expect(
      validateExceptionShape(false, [{ startsTime: '08:00', durationMinutes: 0 }]),
    ).toBeDefined()
  })
})
