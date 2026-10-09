import { describe, expect, it } from 'vitest'
import { MAX_REMINDER_MINUTES, validateReminderMinutes } from '../reminder'

describe('validateReminderMinutes', () => {
  it.each([
    ['0', 0],
    ['30', 30],
    [' 45 ', 45],
    [String(MAX_REMINDER_MINUTES), MAX_REMINDER_MINUTES],
  ])('accepts %j', (raw, minutes) => {
    expect(validateReminderMinutes(raw)).toEqual({ ok: true, minutes })
  })

  it.each([
    ['', 'Escribe cuántos minutos'],
    ['   ', 'Escribe cuántos minutos'],
    ['abc', 'solo números enteros'],
    ['-5', 'solo números enteros'],
    ['2.5', 'solo números enteros'],
    ['30 min', 'solo números enteros'],
    [String(MAX_REMINDER_MINUTES + 1), 'de 0 a 40320'],
  ])('rejects %j with a clear message', (raw, fragment) => {
    const result = validateReminderMinutes(raw)
    expect(result.ok).toBe(false)
    if (!result.ok) expect(result.message).toContain(fragment)
  })
})
