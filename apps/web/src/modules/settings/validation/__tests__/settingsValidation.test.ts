import { describe, it, expect } from 'vitest'
import { buildVocabulary, DEFAULT_BRAND } from '@/shared/model'
import {
  validateContactEmail,
  validateContactPhone,
  validateName,
  validateTimezone,
} from '../settingsValidation'

const v = buildVocabulary(DEFAULT_BRAND)

describe('validateName', () => {
  it('rejects empty and whitespace-only names', () => {
    expect(validateName('', v)).toBeTruthy()
    expect(validateName('   ', v)).toBeTruthy()
  })

  it('rejects a name longer than 120 characters', () => {
    expect(validateName('a'.repeat(121), v)).toBeTruthy()
  })

  it('names the business with the configured word (DEC-119)', () => {
    const studio = buildVocabulary({
      ...DEFAULT_BRAND,
      businessTerm: 'estudio',
      businessTermGender: 'masculine',
    })
    expect(validateName('', v)).toBe('Escribe el nombre de la barbería.')
    expect(validateName('', studio)).toBe('Escribe el nombre del estudio.')
  })

  it('accepts a valid name', () => {
    expect(validateName('Barbería Ejemplo', v)).toBeUndefined()
  })
})

describe('validateTimezone', () => {
  it('rejects empty timezone', () => {
    expect(validateTimezone('')).toBeTruthy()
  })

  it('does not reject an unrecognized-but-well-formed value (IANA membership is server-side, CA-020-03)', () => {
    // Deliberado: el cliente no valida contra un catálogo IANA (fuera de
    // alcance). "COT" pasa la validación de cliente y solo el servidor la
    // rechaza.
    expect(validateTimezone('COT')).toBeUndefined()
  })

  it('accepts a valid-looking timezone', () => {
    expect(validateTimezone('America/Bogota')).toBeUndefined()
  })
})

describe('validateContactEmail', () => {
  it('accepts empty (no contact)', () => {
    expect(validateContactEmail('')).toBeUndefined()
    expect(validateContactEmail('   ')).toBeUndefined()
  })

  it('rejects a malformed email', () => {
    expect(validateContactEmail('no-es-un-correo')).toBeTruthy()
  })

  it('accepts a well-formed email', () => {
    expect(validateContactEmail('contacto@ejemplo.test')).toBeUndefined()
  })
})

describe('validateContactPhone', () => {
  it('accepts empty (no contact)', () => {
    expect(validateContactPhone('')).toBeUndefined()
  })

  it('rejects a phone without the E.164 leading +', () => {
    expect(validateContactPhone('3001234567')).toBeTruthy()
  })

  it('rejects a phone that is too short', () => {
    expect(validateContactPhone('+57')).toBeTruthy()
  })

  it('accepts a well-formed E.164 phone', () => {
    expect(validateContactPhone('+573001234567')).toBeUndefined()
  })
})
