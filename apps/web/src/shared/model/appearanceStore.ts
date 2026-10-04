// Preferencias de pantalla de ESTE dispositivo (DEC-110): modo, tamaño de texto
// y animaciones. Son personales y del aparato -un mostrador con tableta pide
// letra grande aunque el portátil no- por eso viven en `localStorage` y no en
// el servidor. Estado global real, igual criterio que `toastStore`: un
// singleton reactivo, de solo lectura hacia fuera, sin dependencia nueva.
//
// Este módulo solo guarda y valida la elección; quien la pinta en <html> es
// `workspaceTheme.ts`, y solo mientras el cascarón privado está montado.
import { reactive, readonly, type DeepReadonly } from 'vue'

/** Tinta es el modo actual del panel (oscuro); Marfil, el claro; Automático sigue al sistema. */
export type ThemeMode = 'ink' | 'ivory' | 'auto'
export type TextScaleKey = 'small' | 'normal' | 'large' | 'xlarge'
export type MotionMode = 'full' | 'reduced'

export interface AppearancePrefs {
  theme: ThemeMode
  textScale: TextScaleKey
  motion: MotionMode
}

export const DEFAULT_APPEARANCE: AppearancePrefs = {
  theme: 'ink',
  textScale: 'normal',
  motion: 'full',
}

export const THEME_MODES: readonly ThemeMode[] = ['ink', 'ivory', 'auto']
export const MOTION_MODES: readonly MotionMode[] = ['full', 'reduced']

/** Factor de escala de cada paso; la interfaz se escala con `zoom` sobre <html>. */
export const TEXT_SCALES: Record<TextScaleKey, { label: string; factor: number }> = {
  small: { label: 'Pequeño', factor: 0.9 },
  normal: { label: 'Normal', factor: 1 },
  large: { label: 'Grande', factor: 1.125 },
  xlarge: { label: 'Muy grande', factor: 1.25 },
}
export const TEXT_SCALE_KEYS = Object.keys(TEXT_SCALES) as TextScaleKey[]

const STORAGE_KEY = 'nava.appearance.v1'

function isTheme(value: unknown): value is ThemeMode {
  return THEME_MODES.includes(value as ThemeMode)
}
function isTextScale(value: unknown): value is TextScaleKey {
  return TEXT_SCALE_KEYS.includes(value as TextScaleKey)
}
function isMotion(value: unknown): value is MotionMode {
  return MOTION_MODES.includes(value as MotionMode)
}

/** Lee y valida lo guardado; cualquier fallo (privado, bloqueado, corrupto) cae a los valores iniciales. */
function readStored(): AppearancePrefs {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULT_APPEARANCE }
    const parsed: unknown = JSON.parse(raw)
    if (typeof parsed !== 'object' || parsed === null) return { ...DEFAULT_APPEARANCE }
    const candidate = parsed as Record<string, unknown>
    return {
      theme: isTheme(candidate.theme) ? candidate.theme : DEFAULT_APPEARANCE.theme,
      textScale: isTextScale(candidate.textScale)
        ? candidate.textScale
        : DEFAULT_APPEARANCE.textScale,
      motion: isMotion(candidate.motion) ? candidate.motion : DEFAULT_APPEARANCE.motion,
    }
  } catch {
    return { ...DEFAULT_APPEARANCE }
  }
}

function persist(): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    // Sin almacenamiento la preferencia sigue valiendo hasta recargar.
  }
}

const state = reactive<AppearancePrefs>(readStored())

export const appearance: DeepReadonly<AppearancePrefs> = readonly(state)

export function setTheme(theme: ThemeMode): void {
  state.theme = theme
  persist()
}

export function setTextScale(textScale: TextScaleKey): void {
  state.textScale = textScale
  persist()
}

export function setMotion(motion: MotionMode): void {
  state.motion = motion
  persist()
}

export function resetAppearance(): void {
  Object.assign(state, DEFAULT_APPEARANCE)
  persist()
}

export function isDefaultAppearance(prefs: DeepReadonly<AppearancePrefs> = state): boolean {
  return (
    prefs.theme === DEFAULT_APPEARANCE.theme &&
    prefs.textScale === DEFAULT_APPEARANCE.textScale &&
    prefs.motion === DEFAULT_APPEARANCE.motion
  )
}
