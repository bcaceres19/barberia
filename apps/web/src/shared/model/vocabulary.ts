// Vocabulario de la barbería (DEC-110): la palabra con la que llama a su
// negocio y a su profesional ("barbería" / "salón", "barbero" / "estilista"),
// con su plural y su género gramatical para que los artículos concuerden
// ("este estilista", "esta estilista"). Es presentación pura: no renombra la
// tabla `barber`, los contratos de la API ni los términos del dominio
// (`turno`, estados), que siguen siendo los autorizados.
//
// Con los valores iniciales cada frase sale idéntica a la interfaz anterior a
// DEC-110, de modo que una barbería que no toca esta configuración no ve
// ningún cambio de texto.

import type { BrandAccentKey } from './brandPalette'

export type Gender = 'masculine' | 'feminine'

export interface BrandSettings {
  accent: BrandAccentKey
  businessTerm: string
  businessTermGender: Gender
  professionalTerm: string
  professionalTermPlural: string
  professionalTermGender: Gender
}

export const DEFAULT_BRAND: BrandSettings = {
  accent: 'brass',
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
}

/** Largo permitido de un término; igual que el contrato y la base. */
export const TERM_MIN_LENGTH = 2
export const TERM_MAX_LENGTH = 30

// Misma forma que el servicio Go: empieza con letra y sigue con letras,
// espacios, guion o apóstrofo.
const TERM_SHAPE = /^\p{L}[\p{L} '-]*$/u

/** Recorta, colapsa espacios y pasa a minúsculas, igual que el servidor. */
export function normalizeTerm(raw: string): string {
  return raw.trim().split(/\s+/).filter(Boolean).join(' ').toLowerCase()
}

export function validateTerm(raw: string): string | undefined {
  const term = normalizeTerm(raw)
  if (!term) return 'Escribe una palabra.'
  const length = [...term].length
  if (length < TERM_MIN_LENGTH) return `Usa al menos ${TERM_MIN_LENGTH} letras.`
  if (length > TERM_MAX_LENGTH) return `Usa como máximo ${TERM_MAX_LENGTH} caracteres.`
  if (!TERM_SHAPE.test(term)) return 'Usa solo letras, espacios, guion o apóstrofo.'
  return undefined
}

export function capitalize(text: string): string {
  return text.charAt(0).toLocaleUpperCase('es') + text.slice(1)
}

interface Articles {
  the: string
  a: string
  this: string
}

const ARTICLES: Record<Gender, Articles> = {
  masculine: { the: 'el', a: 'un', this: 'este' },
  feminine: { the: 'la', a: 'una', this: 'esta' },
}

/** Frases ya concordadas; las que empiezan en mayúscula llevan el prefijo `cap`. */
export interface Vocabulary {
  business: string
  Business: string
  /** "la barbería" / "el salón". */
  theBusiness: string
  /** "de la barbería" / "del salón". */
  ofTheBusiness: string
  /** "esta barbería" / "este salón". */
  thisBusiness: string

  professional: string
  Professional: string
  professionals: string
  Professionals: string
  /** "un barbero" / "una estilista". */
  aProfessional: string
  /** "el barbero" / "la estilista". */
  theProfessional: string
  /** "los barberos" / "las estilistas". */
  theProfessionals: string
  /** "este barbero" / "esta estilista". */
  thisProfessional: string
  /** "de este barbero" / "de esta estilista". */
  ofThisProfessional: string
  /** "al barbero" / "a la estilista". */
  toTheProfessional: string
  /** "del barbero" / "de la estilista". */
  ofTheProfessional: string
  /** "barberos registrados" / "estilistas registradas": el participio concuerda con el plural. */
  professionalsRegistered: string
  /** Terminación para concordar participios y adjetivos con el profesional: "o" o "a"
   * ("agregad" + "o" / "a"). */
  professionalEnding: 'o' | 'a'
  /** Igual que `professionalEnding`, para la palabra del negocio ("guardad" + "a"). */
  businessEnding: 'o' | 'a'
  /** "uno" / "una": el pronombre que sustituye a un profesional ("Agrega uno"). */
  oneProfessional: 'uno' | 'una'
  /** "otro" / "otra": "Elige otro barbero". */
  anotherProfessional: 'otro' | 'otra'
  /** "el único barbero" / "la única estilista". */
  theOnlyProfessional: string
}

function plural(article: string): string {
  return article === 'el' ? 'los' : 'las'
}

export function buildVocabulary(brand: BrandSettings): Vocabulary {
  const b = ARTICLES[brand.businessTermGender]
  const p = ARTICLES[brand.professionalTermGender]
  const business = brand.businessTerm
  const professional = brand.professionalTerm
  const professionals = brand.professionalTermPlural
  return {
    business,
    Business: capitalize(business),
    theBusiness: `${b.the} ${business}`,
    ofTheBusiness: b.the === 'el' ? `del ${business}` : `de ${b.the} ${business}`,
    thisBusiness: `${b.this} ${business}`,

    professional,
    Professional: capitalize(professional),
    professionals,
    Professionals: capitalize(professionals),
    aProfessional: `${p.a} ${professional}`,
    theProfessional: `${p.the} ${professional}`,
    theProfessionals: `${plural(p.the)} ${professionals}`,
    thisProfessional: `${p.this} ${professional}`,
    ofThisProfessional: `de ${p.this} ${professional}`,
    toTheProfessional: p.the === 'el' ? `al ${professional}` : `a ${p.the} ${professional}`,
    ofTheProfessional: p.the === 'el' ? `del ${professional}` : `de ${p.the} ${professional}`,
    professionalsRegistered: `${professionals} ${p.the === 'el' ? 'registrados' : 'registradas'}`,
    professionalEnding: p.the === 'el' ? 'o' : 'a',
    businessEnding: b.the === 'el' ? 'o' : 'a',
    oneProfessional: p.the === 'el' ? 'uno' : 'una',
    anotherProfessional: p.the === 'el' ? 'otro' : 'otra',
    theOnlyProfessional: `${p.the === 'el' ? 'el único' : 'la única'} ${professional}`,
  }
}
