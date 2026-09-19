/**
 * Pruebas de la cola de avisos emergentes (DEC-095): tiempo por variante,
 * pausa con avisos abiertos/cursor/foco, acordeón, cupo máximo y descarte.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  TOAST_DURATION_MS,
  TOAST_MAX_VISIBLE,
  dismissAllToasts,
  dismissToast,
  pushToast,
  resetToasts,
  setToastHeld,
  toastState,
  toggleToast,
} from '../toastStore'

beforeEach(() => {
  vi.useFakeTimers()
  resetToasts()
})

afterEach(() => {
  resetToasts()
  vi.useRealTimers()
})

describe('toastStore', () => {
  it('publishes the newest toast first with its variant duration', () => {
    pushToast({ variant: 'success', title: 'Primero' })
    pushToast({ variant: 'danger', title: 'Segundo' })

    expect(toastState.items.map((item) => item.title)).toEqual(['Segundo', 'Primero'])
    expect(toastState.items[0].durationMs).toBe(TOAST_DURATION_MS.danger)
    expect(toastState.items[1].durationMs).toBe(TOAST_DURATION_MS.success)
  })

  it('gives an error more time on screen than a confirmation', () => {
    expect(TOAST_DURATION_MS.danger).toBeGreaterThan(TOAST_DURATION_MS.warning)
    expect(TOAST_DURATION_MS.warning).toBeGreaterThan(TOAST_DURATION_MS.info)
    expect(TOAST_DURATION_MS.info).toBeGreaterThan(TOAST_DURATION_MS.success)
  })

  it.each(['success', 'info', 'warning', 'danger'] as const)(
    'removes a %s toast after its duration',
    (variant) => {
      pushToast({ variant, title: 'Aviso' })

      vi.advanceTimersByTime(TOAST_DURATION_MS[variant] - 1)
      expect(toastState.items).toHaveLength(1)

      vi.advanceTimersByTime(1)
      expect(toastState.items).toHaveLength(0)
    },
  )

  it('keeps optional fields empty by default', () => {
    pushToast({ variant: 'info', title: 'Solo título' })

    const [item] = toastState.items
    expect(item.detail).toBe('')
    expect(item.reference).toBe('')
    expect(item.action).toBeNull()
    expect(item.expanded).toBe(false)
  })

  it('pauses the countdown while the toast is open and resumes what was left', () => {
    const id = pushToast({ variant: 'success', title: 'Aviso', detail: 'Detalle' })

    vi.advanceTimersByTime(3000)
    toggleToast(id)
    vi.advanceTimersByTime(60_000)
    expect(toastState.items).toHaveLength(1)

    toggleToast(id)
    vi.advanceTimersByTime(1999)
    expect(toastState.items).toHaveLength(1)
    vi.advanceTimersByTime(1)
    expect(toastState.items).toHaveLength(0)
  })

  it.each(['hovered', 'focused'] as const)('pauses the countdown while %s', (source) => {
    const id = pushToast({ variant: 'info', title: 'Aviso' })

    setToastHeld(id, source, true)
    vi.advanceTimersByTime(60_000)
    expect(toastState.items).toHaveLength(1)

    setToastHeld(id, source, false)
    vi.advanceTimersByTime(TOAST_DURATION_MS.info)
    expect(toastState.items).toHaveLength(0)
  })

  it('stays paused until every hold is released', () => {
    const id = pushToast({ variant: 'info', title: 'Aviso' })

    setToastHeld(id, 'hovered', true)
    setToastHeld(id, 'focused', true)
    setToastHeld(id, 'hovered', false)
    vi.advanceTimersByTime(60_000)
    expect(toastState.items).toHaveLength(1)

    setToastHeld(id, 'focused', false)
    vi.advanceTimersByTime(TOAST_DURATION_MS.info)
    expect(toastState.items).toHaveLength(0)
  })

  it('opens only one toast at a time (accordion)', () => {
    const first = pushToast({ variant: 'success', title: 'Uno' })
    const second = pushToast({ variant: 'warning', title: 'Dos' })

    toggleToast(first)
    toggleToast(second)

    const byId = new Map(toastState.items.map((item) => [item.id, item]))
    expect(byId.get(first)?.expanded).toBe(false)
    expect(byId.get(second)?.expanded).toBe(true)
  })

  it('lets the countdown of a closed toast resume when another one opens', () => {
    const first = pushToast({ variant: 'success', title: 'Uno' })
    const second = pushToast({ variant: 'success', title: 'Dos' })

    toggleToast(first)
    toggleToast(second)
    vi.advanceTimersByTime(TOAST_DURATION_MS.success)

    expect(toastState.items.map((item) => item.id)).toEqual([second])
  })

  it('evicts the oldest toast beyond the maximum', () => {
    for (let index = 1; index <= TOAST_MAX_VISIBLE + 1; index += 1) {
      pushToast({ variant: 'success', title: `Aviso ${index}` })
    }

    expect(toastState.items).toHaveLength(TOAST_MAX_VISIBLE)
    expect(toastState.items.map((item) => item.title)).not.toContain('Aviso 1')
    expect(toastState.items[0].title).toBe(`Aviso ${TOAST_MAX_VISIBLE + 1}`)
  })

  it('does not fire a stale timer for an evicted toast', () => {
    const evicted = pushToast({ variant: 'success', title: 'Antiguo' })
    for (let index = 0; index < TOAST_MAX_VISIBLE; index += 1) {
      pushToast({ variant: 'danger', title: `Reciente ${index}` })
    }

    vi.advanceTimersByTime(TOAST_DURATION_MS.success)

    expect(toastState.items.some((item) => item.id === evicted)).toBe(false)
    expect(toastState.items).toHaveLength(TOAST_MAX_VISIBLE)
  })

  it('dismisses one toast or all of them and ignores unknown ids', () => {
    const first = pushToast({ variant: 'success', title: 'Uno' })
    pushToast({ variant: 'info', title: 'Dos' })

    dismissToast(9999)
    expect(toastState.items).toHaveLength(2)

    dismissToast(first)
    expect(toastState.items.map((item) => item.title)).toEqual(['Dos'])

    dismissAllToasts()
    expect(toastState.items).toHaveLength(0)
    expect(vi.getTimerCount()).toBe(0)
  })
})
