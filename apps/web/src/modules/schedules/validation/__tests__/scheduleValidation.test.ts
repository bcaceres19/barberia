import { describe, it, expect } from 'vitest'
import {
  validateDurationMinutes,
  validateISOWeekday,
  validateStartsTime,
} from '../scheduleValidation'

describe('validateISOWeekday', () => {
  it('accepts 1 through 7', () => {
    for (let day = 1; day <= 7; day++) {
      expect(validateISOWeekday(day)).toBeUndefined()
    }
  })

  it('rejects 0, 8 and non-integers', () => {
    expect(validateISOWeekday(0)).toBeDefined()
    expect(validateISOWeekday(8)).toBeDefined()
    expect(validateISOWeekday(1.5)).toBeDefined()
  })
})

describe('validateStartsTime', () => {
  it('accepts HH:MM 24-hour values', () => {
    for (const value of ['00:00', '08:00', '23:59']) {
      expect(validateStartsTime(value)).toBeUndefined()
    }
  })

  it('rejects malformed values', () => {
    for (const value of ['8:00', '24:00', '23:60', '8am', '']) {
      expect(validateStartsTime(value)).toBeDefined()
    }
  })
})

describe('validateDurationMinutes', () => {
  it('accepts the inclusive bounds 1 and 1440', () => {
    expect(validateDurationMinutes(1)).toBeUndefined()
    expect(validateDurationMinutes(1440)).toBeUndefined()
  })

  it('rejects 0, negative, over 1440, and non-integers', () => {
    expect(validateDurationMinutes(0)).toBeDefined()
    expect(validateDurationMinutes(-1)).toBeDefined()
    expect(validateDurationMinutes(1441)).toBeDefined()
    expect(validateDurationMinutes(60.5)).toBeDefined()
  })
})
