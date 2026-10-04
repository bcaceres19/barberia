// Marca y vocabulario de la barbería activa (DEC-110), compartidos por todas
// las pantallas privadas: el color de acento, y las palabras con las que se
// nombra al negocio y a su profesional. La fuente de verdad es el servidor
// (`GET /private/settings/brand`); aquí vive el último valor confirmado, más una
// copia en `localStorage` para pintar la marca en el primer fotograma de la
// siguiente visita sin esperar la red. Si la copia pertenece a otra barbería, el
// servidor la corrige en cuanto responde.
//
// Estado global real, mismo criterio que `toastStore`/`sessionStore`: un
// singleton reactivo de solo lectura hacia fuera.
import { computed, reactive, readonly, type DeepReadonly } from 'vue'
import { isBrandAccentKey, type BrandAccentKey } from './brandPalette'
import { buildVocabulary, DEFAULT_BRAND, type BrandSettings, type Gender } from './vocabulary'

const STORAGE_KEY = 'nava.brand.v1'

function isGender(value: unknown): value is Gender {
  return value === 'masculine' || value === 'feminine'
}
function isTerm(value: unknown): value is string {
  return typeof value === 'string' && value.length >= 2 && value.length <= 30
}

function readStored(): BrandSettings {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULT_BRAND }
    const c = JSON.parse(raw) as Record<string, unknown> | null
    if (
      c &&
      isBrandAccentKey(c.accent) &&
      isTerm(c.businessTerm) &&
      isGender(c.businessTermGender) &&
      isTerm(c.professionalTerm) &&
      isTerm(c.professionalTermPlural) &&
      isGender(c.professionalTermGender)
    ) {
      return {
        accent: c.accent,
        businessTerm: c.businessTerm,
        businessTermGender: c.businessTermGender,
        professionalTerm: c.professionalTerm,
        professionalTermPlural: c.professionalTermPlural,
        professionalTermGender: c.professionalTermGender,
      }
    }
  } catch {
    // Copia ilegible: se ignora y se espera al servidor.
  }
  return { ...DEFAULT_BRAND }
}

function persist(): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state.brand))
  } catch {
    // Sin almacenamiento solo se pierde el primer fotograma rápido.
  }
}

const state = reactive<{ brand: BrandSettings; accentPreview: BrandAccentKey | null }>({
  brand: readStored(),
  accentPreview: null,
})

export const brandState: DeepReadonly<{
  brand: BrandSettings
  accentPreview: BrandAccentKey | null
}> = readonly(state)

/** Acento que se pinta: la vista previa de Configuración o, si no hay, el guardado. */
export const effectiveAccent = computed<BrandAccentKey>(
  () => state.accentPreview ?? state.brand.accent,
)

/** Vocabulario ya concordado de la marca confirmada por el servidor. */
export const vocabulary = computed(() => buildVocabulary(state.brand))

/** Guarda la marca confirmada por el servidor (nunca un valor optimista). */
export function setBrand(brand: BrandSettings): void {
  state.brand = { ...brand }
  persist()
}

/** Muestra un acento sin guardarlo; `null` vuelve al confirmado. */
export function setAccentPreview(accent: BrandAccentKey | null): void {
  state.accentPreview = accent
}

/** Al cerrar sesión: la siguiente barbería no hereda la marca de la anterior. */
export function resetBrand(): void {
  state.brand = { ...DEFAULT_BRAND }
  state.accentPreview = null
  try {
    window.localStorage.removeItem(STORAGE_KEY)
  } catch {
    // Nada que limpiar.
  }
}
