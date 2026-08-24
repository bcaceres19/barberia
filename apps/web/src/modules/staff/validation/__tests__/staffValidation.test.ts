import { describe, expect, it } from 'vitest'
import { FULL_NAME_MAX_LENGTH, validateFullName } from '../staffValidation'

describe('staffValidation.validateFullName', () => {
  it('rejects an empty name', () => {
    expect(validateFullName('')).toBe('Escribe el nombre del barbero.')
  })

  it('rejects a whitespace-only name', () => {
    expect(validateFullName('   \t\n  ')).toBe('Escribe el nombre del barbero.')
  })

  it('rejects a name longer than 120 characters', () => {
    expect(validateFullName('a'.repeat(FULL_NAME_MAX_LENGTH + 1))).toBe(
      `El nombre no puede superar ${FULL_NAME_MAX_LENGTH} caracteres.`,
    )
  })

  it('accepts a name exactly 120 characters long', () => {
    expect(validateFullName('a'.repeat(FULL_NAME_MAX_LENGTH))).toBeUndefined()
  })

  it('accepts a normal unicode name', () => {
    expect(validateFullName('José Núñez')).toBeUndefined()
  })

  it('does not require two words', () => {
    expect(validateFullName('Cher')).toBeUndefined()
  })
})
