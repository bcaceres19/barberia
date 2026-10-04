// Borrador de la pantalla de Configuración: ayudas puras para saber qué cambió,
// qué se envía y qué sugerir (DEC-110). Sin Vue ni red, para probarlas sin montar nada.
import type { BrandSettings, Gender } from '@/shared/model'
import { normalizeTerm } from '@/shared/model'
import type { BarbershopSettingsFormValues } from './barbershopSettings'

export function sameBarbershop(
  a: BarbershopSettingsFormValues,
  b: BarbershopSettingsFormValues,
): boolean {
  return (
    a.name.trim() === b.name.trim() &&
    a.timezone.trim() === b.timezone.trim() &&
    a.contactEmail.trim() === b.contactEmail.trim() &&
    a.contactPhone.trim() === b.contactPhone.trim()
  )
}

/** Compara la marca como la guardaría el servidor: términos normalizados. */
export function sameBrand(a: BrandSettings, b: BrandSettings): boolean {
  return (
    a.accent === b.accent &&
    a.panelProfile === b.panelProfile &&
    a.businessTermGender === b.businessTermGender &&
    a.professionalTermGender === b.professionalTermGender &&
    normalizeTerm(a.businessTerm) === normalizeTerm(b.businessTerm) &&
    normalizeTerm(a.professionalTerm) === normalizeTerm(b.professionalTerm) &&
    normalizeTerm(a.professionalTermPlural) === normalizeTerm(b.professionalTermPlural)
  )
}

/** Marca lista para enviar: lo mismo que normalizará el servidor. */
export function brandForSave(brand: BrandSettings): BrandSettings {
  return {
    ...brand,
    businessTerm: normalizeTerm(brand.businessTerm),
    professionalTerm: normalizeTerm(brand.professionalTerm),
    professionalTermPlural: normalizeTerm(brand.professionalTermPlural),
  }
}

/**
 * Plural sugerido en español para el campo "plural" mientras se escribe el
 * singular. Solo una sugerencia editable: el plural español no es regular.
 */
export function suggestPlural(singular: string): string {
  const word = normalizeTerm(singular)
  if (!word) return ''
  const last = word.slice(-1)
  if ('aeiouáéó'.includes(last)) return `${word}s`
  if (last === 'z') return `${word.slice(0, -1)}ces`
  if (last === 's') return word // ya invariable: "lunes", "crisis"
  if (last === 'í' || last === 'ú') return `${word}es`
  return `${word}es`
}

export interface BusinessPreset {
  term: string
  gender: Gender
}
export interface ProfessionalPreset {
  singular: string
  plural: string
  gender: Gender
}

export const BUSINESS_PRESETS: readonly BusinessPreset[] = [
  { term: 'barbería', gender: 'feminine' },
  { term: 'salón de belleza', gender: 'masculine' },
  { term: 'peluquería', gender: 'feminine' },
  { term: 'estudio', gender: 'masculine' },
  { term: 'spa', gender: 'masculine' },
]

export const PROFESSIONAL_PRESETS: readonly ProfessionalPreset[] = [
  { singular: 'barbero', plural: 'barberos', gender: 'masculine' },
  { singular: 'barbera', plural: 'barberas', gender: 'feminine' },
  { singular: 'estilista', plural: 'estilistas', gender: 'feminine' },
  { singular: 'peluquero', plural: 'peluqueros', gender: 'masculine' },
  { singular: 'peluquera', plural: 'peluqueras', gender: 'feminine' },
]
