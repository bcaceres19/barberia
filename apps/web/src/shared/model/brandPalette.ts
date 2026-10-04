// Paleta cerrada de acentos de marca (DEC-110). La barbería elige una CLAVE; el
// color real vive solo aquí, con un valor por modo, para que ninguna
// combinación llegue a pantalla sin contraste comprobado
// (`__tests__/brandPalette.test.ts` mide cada par con la fórmula de WCAG 2.2).
//
// Cada acento cumple dos funciones a la vez: tinta de texto/filete sobre la
// superficie del modo y relleno de un botón con texto del color de la
// superficie. Por eso un solo valor por modo basta: si pasa AA como texto,
// pasa AA como relleno. `ink` (modo Tinta) está aclarado para leerse sobre
// #101b2b; `ivory` (modo Marfil) está oscurecido para leerse sobre #f4f0e7.
//
// Las claves son las de `barbershop_brand_accent_ck` y del enum `BrandAccent`
// del contrato OpenAPI; añadir una exige migración y contrato en el mismo
// cambio.

export type BrandAccentKey = 'brass' | 'emerald' | 'sapphire' | 'ruby' | 'amethyst' | 'copper'

export interface BrandAccent {
  key: BrandAccentKey
  label: string
  /** Valor sobre la superficie del modo Tinta (#101b2b). */
  ink: string
  /** Valor sobre la superficie del modo Marfil (#f4f0e7). */
  ivory: string
}

export const BRAND_ACCENTS: readonly BrandAccent[] = [
  { key: 'brass', label: 'Latón', ink: '#b8955a', ivory: '#765c2f' },
  { key: 'emerald', label: 'Esmeralda', ink: '#4cb58a', ivory: '#1c6a4d' },
  { key: 'sapphire', label: 'Zafiro', ink: '#6da8e6', ivory: '#27519a' },
  { key: 'ruby', label: 'Rubí', ink: '#e8808c', ivory: '#9a2b3c' },
  { key: 'amethyst', label: 'Amatista', ink: '#b094ea', ivory: '#5d3d9c' },
  { key: 'copper', label: 'Cobre', ink: '#df8a57', ivory: '#8b4a1e' },
]

export const DEFAULT_BRAND_ACCENT: BrandAccentKey = 'brass'

/** Superficies del panel contra las que se mide el contraste de cada acento. */
export const SURFACE_INK = '#101b2b'
export const SURFACE_IVORY = '#f4f0e7'

export function isBrandAccentKey(value: unknown): value is BrandAccentKey {
  return BRAND_ACCENTS.some((accent) => accent.key === value)
}

export function accentByKey(key: BrandAccentKey): BrandAccent {
  return BRAND_ACCENTS.find((accent) => accent.key === key) ?? BRAND_ACCENTS[0]!
}

function channel(value: number): number {
  const c = value / 255
  return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
}

function luminance(hex: string): number {
  const n = Number.parseInt(hex.slice(1), 16)
  return (
    0.2126 * channel((n >> 16) & 255) + 0.7152 * channel((n >> 8) & 255) + 0.0722 * channel(n & 255)
  )
}

/** Relación de contraste WCAG 2.x entre dos colores `#rrggbb`. */
export function contrastRatio(a: string, b: string): number {
  const [light, dark] = [luminance(a), luminance(b)].sort((x, y) => y - x) as [number, number]
  return (light + 0.05) / (dark + 0.05)
}
