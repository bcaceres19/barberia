/**
 * Ayudas puras de la vista previa de la política de reserva (HU-093).
 */
import { describe, expect, it } from 'vitest'
import { blockedShareOfWindow, formatDuration, sampleSlotLabels } from '../bookingPolicy'

describe('formatDuration', () => {
  it('writes minutes, hours and days compactly', () => {
    expect(formatDuration(0)).toBe('0 min')
    expect(formatDuration(45)).toBe('45 min')
    expect(formatDuration(60)).toBe('1 h')
    expect(formatDuration(90)).toBe('1 h 30 min')
    expect(formatDuration(1440)).toBe('1 día')
    expect(formatDuration(2 * 1440 + 3 * 60)).toBe('2 días 3 h')
  })

  it('treats a half-typed or invalid value as zero', () => {
    expect(formatDuration(Number.NaN)).toBe('0 min')
    expect(formatDuration(-5)).toBe('0 min')
  })
})

describe('blockedShareOfWindow', () => {
  it('is the share of the window closed by the minimum advance', () => {
    expect(blockedShareOfWindow(720, 1)).toBeCloseTo(0.5)
    expect(blockedShareOfWindow(0, 3)).toBe(0)
  })

  it('clamps out-of-range and non-numeric input', () => {
    expect(blockedShareOfWindow(9999, 1)).toBe(1)
    expect(blockedShareOfWindow(60, 0)).toBe(0)
    expect(blockedShareOfWindow(Number.NaN, 3)).toBe(0)
  })
})

describe('sampleSlotLabels', () => {
  it('lists the first slots from 10:00 on the chosen grid', () => {
    expect(sampleSlotLabels(30, 4)).toEqual(['10:00', '10:30', '11:00', '11:30'])
    expect(sampleSlotLabels(60)).toEqual(['10:00', '11:00', '12:00', '13:00'])
  })

  it('caps the number of slots and ignores an invalid grid', () => {
    expect(sampleSlotLabels(5)).toHaveLength(8)
    expect(sampleSlotLabels(0)).toEqual([])
  })
})
