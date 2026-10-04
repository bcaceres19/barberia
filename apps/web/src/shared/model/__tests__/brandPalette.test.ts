/**
 * Contraste de la paleta de acentos y del tema Marfil (DEC-110, WCAG 2.2 AA).
 * Esta prueba es la garantía de que ningún acento ni par de texto llegue a
 * pantalla sin haber sido medido: si alguien cambia un color o añade un
 * acento sin contraste suficiente, falla aquí.
 */
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import {
  accentByKey,
  BRAND_ACCENTS,
  contrastRatio,
  DEFAULT_BRAND_ACCENT,
  isBrandAccentKey,
  SURFACE_INK,
  SURFACE_IVORY,
} from '../brandPalette'

const AA_TEXT = 4.5

describe('contrastRatio', () => {
  it('matches the WCAG reference values', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 1)
    expect(contrastRatio('#ffffff', '#ffffff')).toBeCloseTo(1, 5)
    // Simétrico: no importa cuál sea el color claro.
    expect(contrastRatio('#101b2b', '#f4f0e7')).toBeCloseTo(contrastRatio('#f4f0e7', '#101b2b'), 8)
  })
})

describe('BRAND_ACCENTS', () => {
  it('has the six closed keys the contract and the database allow', () => {
    expect(BRAND_ACCENTS.map((a) => a.key)).toEqual([
      'brass',
      'emerald',
      'sapphire',
      'ruby',
      'amethyst',
      'copper',
    ])
  })

  it('starts on brass, the NAVA brass, and recognizes only its own keys', () => {
    expect(DEFAULT_BRAND_ACCENT).toBe('brass')
    expect(isBrandAccentKey('emerald')).toBe(true)
    expect(isBrandAccentKey('#b8955a')).toBe(false)
    expect(isBrandAccentKey('Brass')).toBe(false)
    expect(isBrandAccentKey(null)).toBe(false)
  })

  it.each(BRAND_ACCENTS.map((a) => [a.key, a] as const))(
    '%s reads as text and as a button fill on the Tinta surface (AA)',
    (_key, accent) => {
      // Como texto o filete sobre tinta...
      expect(contrastRatio(accent.ink, SURFACE_INK)).toBeGreaterThanOrEqual(AA_TEXT)
      // ...y como relleno de un botón con texto del color de la superficie.
      expect(contrastRatio(SURFACE_INK, accent.ink)).toBeGreaterThanOrEqual(AA_TEXT)
    },
  )

  it.each(BRAND_ACCENTS.map((a) => [a.key, a] as const))(
    '%s reads as text and as a button fill on the Marfil surface (AA)',
    (_key, accent) => {
      expect(contrastRatio(accent.ivory, SURFACE_IVORY)).toBeGreaterThanOrEqual(AA_TEXT)
      expect(contrastRatio(SURFACE_IVORY, accent.ivory)).toBeGreaterThanOrEqual(AA_TEXT)
    },
  )

  it('keeps every accent readable over a raised field and white in Marfil too', () => {
    for (const accent of BRAND_ACCENTS) {
      expect(contrastRatio(accent.ink, '#16243a')).toBeGreaterThanOrEqual(AA_TEXT)
      expect(contrastRatio(accent.ink, '#0b1420')).toBeGreaterThanOrEqual(AA_TEXT)
      expect(contrastRatio(accent.ivory, '#ffffff')).toBeGreaterThanOrEqual(AA_TEXT)
    }
  })

  it('gives each accent distinct colors, so a choice is always visible', () => {
    expect(new Set(BRAND_ACCENTS.map((a) => a.ink)).size).toBe(BRAND_ACCENTS.length)
    expect(new Set(BRAND_ACCENTS.map((a) => a.ivory)).size).toBe(BRAND_ACCENTS.length)
  })

  it('keeps brass on Tinta identical to the token that styled the app before DEC-110', () => {
    const tokens = readFileSync(resolve(__dirname, '../../../styles/tokens.css'), 'utf-8')
    expect(tokens).toContain(`--color-brand-accent-surface: ${accentByKey('brass').ink};`)
    expect(tokens).toContain(`--color-accent-brass: ${accentByKey('brass').ivory};`)
  })
})

describe('tema Marfil (tokens.css)', () => {
  const tokens = readFileSync(resolve(__dirname, '../../../styles/tokens.css'), 'utf-8')
  const block = /:root\[data-app-theme='ivory'\] \{([^}]*)\}/.exec(tokens)?.[1] ?? ''
  const value = (name: string) => new RegExp(`${name}: (#[0-9a-f]{6});`).exec(block)?.[1] as string

  it('defines the ivory block', () => {
    expect(block).not.toBe('')
  })

  it.each([
    '--color-on-strong',
    '--color-on-strong-muted',
    '--color-on-strong-soft',
    '--color-danger-on-strong',
    '--color-warning-on-strong',
    '--color-success-on-strong',
    '--color-pending-on-strong',
    '--color-no-show-on-strong',
    '--color-info-on-strong',
  ])('%s reads on the ivory surface and on the darker chrome (AA)', (name) => {
    const color = value(name)
    expect(color).toBeDefined()
    expect(contrastRatio(color, value('--color-surface-strong'))).toBeGreaterThanOrEqual(AA_TEXT)
    expect(contrastRatio(color, value('--color-chrome-surface'))).toBeGreaterThanOrEqual(AA_TEXT)
  })

  it('uses the same ivory surface the palette is measured against', () => {
    expect(value('--color-surface-strong')).toBe(SURFACE_IVORY)
  })
})
