import { describe, expect, it } from 'vitest'
import { buildVocabulary, DEFAULT_BRAND } from '@/shared/model'
import { FULL_NAME_MAX_LENGTH, validateFullName } from '../staffValidation'

const v = buildVocabulary(DEFAULT_BRAND)

describe('staffValidation.validateFullName', () => {
  it('names the professional with the configured word (DEC-119)', () => {
    const nails = buildVocabulary({
      ...DEFAULT_BRAND,
      professionalTerm: 'manicurista',
      professionalTermPlural: 'manicuristas',
      professionalTermGender: 'feminine',
    })
    expect(validateFullName('', nails)).toBe('Escribe el nombre de la manicurista.')
  })

  it('rejects an empty name', () => {
    expect(validateFullName('', v)).toBe('Escribe el nombre del barbero.')
  })

  it('rejects a whitespace-only name', () => {
    expect(validateFullName('   \t\n  ', v)).toBe('Escribe el nombre del barbero.')
  })

  it('rejects a name longer than 120 characters', () => {
    expect(validateFullName('a'.repeat(FULL_NAME_MAX_LENGTH + 1), v)).toBe(
      `El nombre no puede superar ${FULL_NAME_MAX_LENGTH} caracteres.`,
    )
  })

  it('accepts a name exactly 120 characters long', () => {
    expect(validateFullName('a'.repeat(FULL_NAME_MAX_LENGTH), v)).toBeUndefined()
  })

  it('accepts a normal unicode name', () => {
    expect(validateFullName('José Núñez', v)).toBeUndefined()
  })

  it('does not require two words', () => {
    expect(validateFullName('Cher', v)).toBeUndefined()
  })
})
