import { describe, it, expect } from 'vitest'
import {
  validateContactEmail,
  validateContactPhone,
  validateName,
  validateTimezone,
} from '../settingsValidation'

describe('validateName', () => {
  it('rejects empty and whitespace-only names', () => {
    expect(validateName('')).toBeTruthy()
    expect(validateName('   ')).toBeTruthy()
  })

  it('rejects a name longer than 120 characters', () => {
    expect(validateName('a'.repeat(121))).toBeTruthy()
  })

  it('accepts a valid name', () => {
    expect(validateName('Barbería Ejemplo')).toBeUndefined()
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
