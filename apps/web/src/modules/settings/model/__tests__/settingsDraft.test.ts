/**
 * Ayudas puras del borrador de Configuración (DEC-110): qué cambió, qué se
 * envía y el plural sugerido.
 */
import { describe, expect, it } from 'vitest'
import { DEFAULT_BRAND } from '@/shared/model'
import {
  brandForSave,
  BUSINESS_PRESETS,
  PROFESSIONAL_PRESETS,
  sameBarbershop,
  sameBrand,
  suggestPlural,
} from '../settingsDraft'

const shop = {
  name: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@ejemplo.test',
  contactPhone: '+573001234567',
}

describe('sameBarbershop', () => {
  it('ignores surrounding spaces, so typing and deleting a space is not a change', () => {
    expect(sameBarbershop({ ...shop, name: '  Barbería Ejemplo ' }, shop)).toBe(true)
  })

  it.each(['name', 'timezone', 'contactEmail', 'contactPhone'] as const)(
    'detects a change in %s',
    (field) => {
      expect(sameBarbershop({ ...shop, [field]: 'otro' }, shop)).toBe(false)
    },
  )
})

describe('sameBrand', () => {
  it('compares the words as the server will store them', () => {
    expect(sameBrand({ ...DEFAULT_BRAND, businessTerm: '  Barbería ' }, DEFAULT_BRAND)).toBe(true)
    expect(sameBrand({ ...DEFAULT_BRAND, professionalTermPlural: 'BARBEROS' }, DEFAULT_BRAND)).toBe(
      true,
    )
  })

  it.each([
    ['accent', { accent: 'ruby' as const }],
    ['businessTerm', { businessTerm: 'estudio' }],
    ['businessTermGender', { businessTermGender: 'masculine' as const }],
    ['professionalTerm', { professionalTerm: 'estilista' }],
    ['professionalTermPlural', { professionalTermPlural: 'estilistas' }],
    ['professionalTermGender', { professionalTermGender: 'feminine' as const }],
  ])('detects a change in %s', (_field, change) => {
    expect(sameBrand({ ...DEFAULT_BRAND, ...change }, DEFAULT_BRAND)).toBe(false)
  })
})

describe('brandForSave', () => {
  it('normalizes the three words and leaves the rest untouched', () => {
    expect(
      brandForSave({
        ...DEFAULT_BRAND,
        accent: 'copper',
        businessTerm: '  Salón   de Belleza ',
        professionalTerm: 'Estilista',
        professionalTermPlural: ' ESTILISTAS ',
      }),
    ).toEqual({
      ...DEFAULT_BRAND,
      accent: 'copper',
      businessTerm: 'salón de belleza',
      professionalTerm: 'estilista',
      professionalTermPlural: 'estilistas',
    })
  })
})

describe('suggestPlural', () => {
  it.each([
    ['barbero', 'barberos'],
    ['estilista', 'estilistas'],
    ['peluquera', 'peluqueras'],
    ['pintor', 'pintores'],
    ['masajista', 'masajistas'],
    ['lápiz', 'lápices'],
    ['lunes', 'lunes'],
    ['colibrí', 'colibríes'],
    ['  Barbero ', 'barberos'],
    ['', ''],
  ])('%j → %j', (singular, plural) => {
    expect(suggestPlural(singular)).toBe(plural)
  })
})

describe('presets', () => {
  it('start with the words the interface always used', () => {
    expect(BUSINESS_PRESETS[0]).toEqual({ term: 'barbería', gender: 'feminine' })
    expect(PROFESSIONAL_PRESETS[0]).toEqual({
      singular: 'barbero',
      plural: 'barberos',
      gender: 'masculine',
    })
  })

  it('suggest a plural that matches the one they carry', () => {
    for (const preset of PROFESSIONAL_PRESETS) {
      expect(suggestPlural(preset.singular)).toBe(preset.plural)
    }
  })
})
