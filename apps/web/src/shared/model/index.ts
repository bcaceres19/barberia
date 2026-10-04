// Modelos y utilidades de dominio compartidas de verdad entre módulos, según
// docs/03-desarrollo/estandar-frontend-vue.md. Cada pieza se importa por su
// archivo; este índice reúne solo la marca y las preferencias de pantalla
// (DEC-110), que todo el panel privado consume.
export {
  appearance,
  DEFAULT_APPEARANCE,
  isDefaultAppearance,
  MOTION_MODES,
  resetAppearance,
  setMotion,
  setTextScale,
  setTheme,
  TEXT_SCALES,
  TEXT_SCALE_KEYS,
  THEME_MODES,
} from './appearanceStore'
export type { AppearancePrefs, MotionMode, TextScaleKey, ThemeMode } from './appearanceStore'
export {
  accentByKey,
  BRAND_ACCENTS,
  contrastRatio,
  DEFAULT_BRAND_ACCENT,
  isBrandAccentKey,
  SURFACE_INK,
  SURFACE_IVORY,
} from './brandPalette'
export type { BrandAccent, BrandAccentKey } from './brandPalette'
export {
  brandState,
  effectiveAccent,
  resetBrand,
  setAccentPreview,
  setBrand,
  vocabulary,
} from './brandStore'
export {
  buildVocabulary,
  capitalize,
  DEFAULT_BRAND,
  normalizeTerm,
  TERM_MAX_LENGTH,
  TERM_MIN_LENGTH,
  validateTerm,
} from './vocabulary'
export type { BrandSettings, Gender, Vocabulary } from './vocabulary'
export {
  attachWorkspaceTheme,
  effectiveZoom,
  MIN_EFFECTIVE_WIDTH,
  resolvedTheme,
} from './workspaceTheme'
