// Prefijos telefónicos que ofrece el formulario de la reserva pública. El
// contrato del API sigue siendo un único teléfono E.164 (`+573001234567`); esta
// lista solo permite al cliente escribir el prefijo en su propia casilla y ver
// de qué país es. Un prefijo compartido por varios países (+1) se ofrece una
// sola vez, así prefijo → país es una función y la bandera nunca es ambigua.

export interface PhoneCountry {
  /** Código ISO 3166-1 alfa-2; de él se deriva la bandera. */
  iso: string
  name: string
  /** Prefijo con `+`, tal como se concatena al número nacional. */
  dial: string
}

export const DEFAULT_PHONE_COUNTRY = 'CO'

// Colombia abre la lista (es el mercado de la app); el resto va por nombre.
export const PHONE_COUNTRIES: readonly PhoneCountry[] = [
  { iso: 'CO', name: 'Colombia', dial: '+57' },
  { iso: 'AR', name: 'Argentina', dial: '+54' },
  { iso: 'BO', name: 'Bolivia', dial: '+591' },
  { iso: 'BR', name: 'Brasil', dial: '+55' },
  { iso: 'CL', name: 'Chile', dial: '+56' },
  { iso: 'CR', name: 'Costa Rica', dial: '+506' },
  { iso: 'CU', name: 'Cuba', dial: '+53' },
  { iso: 'EC', name: 'Ecuador', dial: '+593' },
  { iso: 'SV', name: 'El Salvador', dial: '+503' },
  { iso: 'ES', name: 'España', dial: '+34' },
  { iso: 'US', name: 'Estados Unidos y Canadá', dial: '+1' },
  { iso: 'FR', name: 'Francia', dial: '+33' },
  { iso: 'DE', name: 'Alemania', dial: '+49' },
  { iso: 'GT', name: 'Guatemala', dial: '+502' },
  { iso: 'HN', name: 'Honduras', dial: '+504' },
  { iso: 'IT', name: 'Italia', dial: '+39' },
  { iso: 'MX', name: 'México', dial: '+52' },
  { iso: 'NI', name: 'Nicaragua', dial: '+505' },
  { iso: 'PA', name: 'Panamá', dial: '+507' },
  { iso: 'PY', name: 'Paraguay', dial: '+595' },
  { iso: 'PE', name: 'Perú', dial: '+51' },
  { iso: 'PT', name: 'Portugal', dial: '+351' },
  { iso: 'GB', name: 'Reino Unido', dial: '+44' },
  { iso: 'UY', name: 'Uruguay', dial: '+598' },
  { iso: 'VE', name: 'Venezuela', dial: '+58' },
]

const BY_ISO = new Map(PHONE_COUNTRIES.map((country) => [country.iso, country]))

export function phoneCountry(iso: string): PhoneCountry {
  return BY_ISO.get(iso) ?? BY_ISO.get(DEFAULT_PHONE_COUNTRY)!
}

// La bandera es el par de indicadores regionales del código ISO. Donde el
// sistema no dibuja banderas (Windows) se ven las dos letras del país, que
// siguen identificándolo.
export function flagEmoji(iso: string): string {
  return String.fromCodePoint(...[...iso.toUpperCase()].map((char) => 0x1f1a5 + char.charCodeAt(0)))
}

export interface SplitPhone {
  country: string
  national: string
}

// splitInternationalPhone separa un número pegado con prefijo (`+57 300…` o
// `0057 300…`). Prueba primero el prefijo más largo: +593 gana a +59. Si no
// reconoce el prefijo devuelve null y el llamador conserva lo escrito.
export function splitInternationalPhone(raw: string): SplitPhone | null {
  const trimmed = raw.trim()
  let digits: string
  if (trimmed.startsWith('+')) {
    digits = trimmed.slice(1).replace(/\D/g, '')
  } else if (trimmed.startsWith('00')) {
    digits = trimmed.slice(2).replace(/\D/g, '')
  } else {
    return null
  }
  const match = [...PHONE_COUNTRIES]
    .sort((a, b) => b.dial.length - a.dial.length)
    .find((country) => digits.startsWith(country.dial.slice(1)))
  if (!match) return null
  return { country: match.iso, national: digits.slice(match.dial.length - 1) }
}

// nationalDigits deja solo los dígitos: espacios, guiones y paréntesis que el
// cliente copie de su agenda no viajan al API.
export function nationalDigits(value: string): string {
  return value.replace(/\D/g, '')
}

export function composePhone(iso: string, national: string): string {
  return `${phoneCountry(iso).dial}${nationalDigits(national)}`
}
