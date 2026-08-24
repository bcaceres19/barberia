import { describe, it, expect } from 'vitest'
import {
  validateName,
  validateDescription,
  validateDurationMinutes,
  validatePrice,
  NAME_MAX_LENGTH,
  DESCRIPTION_MAX_LENGTH,
} from '../catalogValidation'

describe('validateName', () => {
  it('rejects an empty or whitespace-only name', () => {
    expect(validateName('')).toBeDefined()
    expect(validateName('   ')).toBeDefined()
  })

  it('accepts a name up to NAME_MAX_LENGTH characters', () => {
    expect(validateName('a'.repeat(NAME_MAX_LENGTH))).toBeUndefined()
  })

  it('rejects a name longer than NAME_MAX_LENGTH characters', () => {
    expect(validateName('a'.repeat(NAME_MAX_LENGTH + 1))).toBeDefined()
  })
})

describe('validateDescription', () => {
  it('accepts an empty description (optional field)', () => {
    expect(validateDescription('')).toBeUndefined()
  })

  it('accepts a description up to DESCRIPTION_MAX_LENGTH characters', () => {
    expect(validateDescription('a'.repeat(DESCRIPTION_MAX_LENGTH))).toBeUndefined()
  })

  it('rejects a description longer than DESCRIPTION_MAX_LENGTH characters', () => {
    expect(validateDescription('a'.repeat(DESCRIPTION_MAX_LENGTH + 1))).toBeDefined()
  })
})

describe('validateDurationMinutes', () => {
  it('accepts 25, 30, 45 and 90 minutes without a closed list (RN-SER-02)', () => {
    for (const minutes of ['25', '30', '45', '90']) {
      expect(validateDurationMinutes(minutes)).toBeUndefined()
    }
  })

  it('accepts the boundary values 1 and 1440', () => {
    expect(validateDurationMinutes('1')).toBeUndefined()
    expect(validateDurationMinutes('1440')).toBeUndefined()
  })

  it('rejects zero, negative, fractional, and out-of-range values', () => {
    for (const raw of ['0', '-5', '30.5', '1441', '', 'abc']) {
      expect(validateDurationMinutes(raw)).toBeDefined()
    }
  })
})

describe('validatePrice', () => {
  it('accepts a valid decimal with up to two fraction digits', () => {
    for (const price of ['45000', '45000.5', '45000.00', '0.01']) {
      expect(validatePrice(price)).toBeUndefined()
    }
  })

  it('rejects zero and negative amounts (DEC-067: no free services)', () => {
    expect(validatePrice('0')).toBeDefined()
    expect(validatePrice('0.00')).toBeDefined()
    expect(validatePrice('-1')).toBeDefined()
  })

  it('rejects an invalid format', () => {
    for (const raw of ['', 'abc', '45,000', '45000.123', '1e10', '+45000']) {
      expect(validatePrice(raw)).toBeDefined()
    }
  })
})
