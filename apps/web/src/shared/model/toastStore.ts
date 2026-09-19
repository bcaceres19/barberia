// Cola de avisos emergentes (toasts) compartida entre TODAS las pantallas.
// Es estado global real -cualquier módulo notifica y una sola región los
// muestra-, pero una sola cola no justifica una dependencia nueva: se sigue
// el mismo criterio que `sessionStore` (DEC-033, DEC-035): un singleton
// reactivo exportado desde un módulo, de solo lectura hacia fuera.
//
// El tiempo en pantalla vive aquí y no en el componente para que una
// pantalla que se desmonta (por ejemplo, un diálogo que se cierra tras
// guardar) no se lleve consigo el aviso que acaba de emitir.
import { reactive, readonly, type DeepReadonly } from 'vue'

export type ToastVariant = 'success' | 'info' | 'warning' | 'danger'

/** Icono de la acción del aviso; solo comunica qué hace el botón. */
export type ToastActionIcon = 'retry' | 'arrow'

export interface ToastAction {
  label: string
  icon?: ToastActionIcon
  /** Se ejecuta al activar la acción; el aviso se descarta después. */
  run: () => void
}

export interface ToastInput {
  variant: ToastVariant
  title: string
  /** Explicación accionable; se muestra al abrir el aviso. */
  detail?: string
  /** Referencia segura para soporte (por ejemplo `request_id`), nunca un dato personal. */
  reference?: string
  action?: ToastAction
}

export interface ToastItem {
  id: number
  variant: ToastVariant
  title: string
  detail: string
  reference: string
  action: ToastAction | null
  durationMs: number
  expanded: boolean
  hovered: boolean
  focused: boolean
}

/**
 * Tiempo en pantalla por variante (DEC-095): un error necesita más lectura
 * que una confirmación. La cuenta atrás se detiene mientras el aviso está
 * abierto, con el cursor encima o con foco de teclado (WCAG 2.2 AA 2.2.1).
 */
export const TOAST_DURATION_MS: Readonly<Record<ToastVariant, number>> = {
  success: 5000,
  info: 6000,
  warning: 8000,
  danger: 12000,
}

/** Máximo simultáneo: al superarlo se descarta el aviso más antiguo. */
export const TOAST_MAX_VISIBLE = 4

const state = reactive<{ items: ToastItem[] }>({ items: [] })

/** Vista de solo lectura, el aviso más reciente primero. */
export const toastState: DeepReadonly<{ items: ToastItem[] }> = readonly(state)

interface Countdown {
  remainingMs: number
  startedAt: number
  handle: ReturnType<typeof setTimeout> | null
}

const countdowns = new Map<number, Countdown>()
let nextId = 1

function isHeld(item: ToastItem): boolean {
  return item.expanded || item.hovered || item.focused
}

function stopCountdown(countdown: Countdown): void {
  if (countdown.handle === null) return
  clearTimeout(countdown.handle)
  countdown.handle = null
  countdown.remainingMs -= Date.now() - countdown.startedAt
}

function startCountdown(id: number, countdown: Countdown): void {
  if (countdown.handle !== null) return
  countdown.startedAt = Date.now()
  countdown.handle = setTimeout(() => dismissToast(id), Math.max(countdown.remainingMs, 0))
}

function syncCountdown(item: ToastItem): void {
  const countdown = countdowns.get(item.id)
  if (!countdown) return
  if (isHeld(item)) stopCountdown(countdown)
  else startCountdown(item.id, countdown)
}

function forget(id: number): void {
  const countdown = countdowns.get(id)
  if (countdown?.handle) clearTimeout(countdown.handle)
  countdowns.delete(id)
}

/** Publica un aviso y devuelve su identificador. */
export function pushToast(input: ToastInput): number {
  const item: ToastItem = {
    id: nextId++,
    variant: input.variant,
    title: input.title,
    detail: input.detail ?? '',
    reference: input.reference ?? '',
    action: input.action ?? null,
    durationMs: TOAST_DURATION_MS[input.variant],
    expanded: false,
    hovered: false,
    focused: false,
  }
  countdowns.set(item.id, { remainingMs: item.durationMs, startedAt: 0, handle: null })
  state.items.unshift(item)
  for (const evicted of state.items.splice(TOAST_MAX_VISIBLE)) forget(evicted.id)
  syncCountdown(item)
  return item.id
}

export function dismissToast(id: number): void {
  const index = state.items.findIndex((item) => item.id === id)
  if (index === -1) return
  state.items.splice(index, 1)
  forget(id)
}

export function dismissAllToasts(): void {
  for (const item of state.items) forget(item.id)
  state.items.splice(0)
}

/** Acordeón: abrir un aviso cierra el que estuviera abierto. */
export function toggleToast(id: number): void {
  for (const item of state.items) {
    item.expanded = item.id === id ? !item.expanded : false
    syncCountdown(item)
  }
}

export function setToastHeld(id: number, source: 'hovered' | 'focused', value: boolean): void {
  const item = state.items.find((candidate) => candidate.id === id)
  if (!item || item[source] === value) return
  item[source] = value
  syncCountdown(item)
}

/** Vacía la cola y sus temporizadores; solo para pruebas. */
export function resetToasts(): void {
  dismissAllToasts()
  nextId = 1
}
