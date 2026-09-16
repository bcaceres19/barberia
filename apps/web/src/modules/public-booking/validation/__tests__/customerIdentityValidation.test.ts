import { describe, it, expect } from 'vitest'
import {
  validateAttendeeName,
  validateCustomerEmail,
  validateCustomerFullName,
  validateCustomerNote,
  validateCustomerPhone,
} from '../customerIdentityValidation'

describe('customerIdentityValidation', () => {
  it('requires a non-empty full name, unlike the optional manual-booking rule', () => {
    expect(validateCustomerFullName('')).toBeDefined()
    expect(validateCustomerFullName('Ana Ríos')).toBeUndefined()
  })

  it('requires a phone in E.164 format', () => {
    expect(validateCustomerPhone('')).toBeDefined()
    expect(validateCustomerPhone('3001234567')).toBeDefined()
    expect(validateCustomerPhone('+573001234567')).toBeUndefined()
  })

  it('requires a well-formed email', () => {
    expect(validateCustomerEmail('')).toBeDefined()
    expect(validateCustomerEmail('no-arroba')).toBeDefined()
    expect(validateCustomerEmail('ana@example.com')).toBeUndefined()
  })

  it('allows an empty note but rejects one over the limit', () => {
    expect(validateCustomerNote('')).toBeUndefined()
    expect(validateCustomerNote('a'.repeat(501))).toBeDefined()
    expect(validateCustomerNote('a'.repeat(500))).toBeUndefined()
  })

  it('requires a non-empty attendee name when evaluated', () => {
    expect(validateAttendeeName('')).toBeDefined()
    expect(validateAttendeeName('   ')).toBeDefined()
    expect(validateAttendeeName('Mateo Ruiz')).toBeUndefined()
  })
})
