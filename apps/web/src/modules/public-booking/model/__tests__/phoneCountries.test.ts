import { describe, expect, it } from 'vitest'
import {
  PHONE_COUNTRIES,
  composePhone,
  flagEmoji,
  phoneCountry,
  splitInternationalPhone,
} from '../phoneCountries'

describe('phoneCountries', () => {
  it('dibuja la bandera a partir del código ISO', () => {
    expect(flagEmoji('CO')).toBe('🇨🇴')
    expect(flagEmoji('mx')).toBe('🇲🇽')
  })

  it('no repite prefijos: prefijo → país es una función', () => {
    const dials = PHONE_COUNTRIES.map((country) => country.dial)
    expect(new Set(dials).size).toBe(dials.length)
  })

  it('cae en Colombia ante un ISO desconocido', () => {
    expect(phoneCountry('ZZ').iso).toBe('CO')
  })

  it('compone el E.164 con prefijo y solo dígitos', () => {
    expect(composePhone('CO', '300 123-4567')).toBe('+573001234567')
    expect(composePhone('US', '(415) 555 0100')).toBe('+14155550100')
  })

  it('separa un número pegado con +, con 00 o con espacios', () => {
    expect(splitInternationalPhone('+57 300 123 4567')).toEqual({
      country: 'CO',
      national: '3001234567',
    })
    expect(splitInternationalPhone('0052 55 1234 5678')).toEqual({
      country: 'MX',
      national: '5512345678',
    })
  })

  it('elige el prefijo más largo: +593 es Ecuador, no un +59 inexistente', () => {
    expect(splitInternationalPhone('+593991234567')).toEqual({
      country: 'EC',
      national: '991234567',
    })
  })

  it('no toca lo que no trae prefijo ni lo que no reconoce', () => {
    expect(splitInternationalPhone('3001234567')).toBeNull()
    expect(splitInternationalPhone('+81 90 1234 5678')).toBeNull()
  })
})
