import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import { useToast } from '../useToast'

beforeEach(() => {
  vi.useFakeTimers()
  resetToasts()
})

afterEach(() => {
  resetToasts()
  vi.useRealTimers()
})

describe('useToast', () => {
  it('maps each helper to its variant', () => {
    const toast = useToast()

    toast.success('Guardado')
    toast.info('Nota')
    toast.warning('Atención')
    toast.error('Falló')

    expect(toastState.items.map((item) => item.variant)).toEqual([
      'danger',
      'warning',
      'info',
      'success',
    ])
  })

  it('forwards detail, reference and action', () => {
    const run = vi.fn()

    useToast().error('No pudimos guardar', {
      detail: 'Inténtalo de nuevo.',
      reference: 'req_1',
      action: { label: 'Reintentar', icon: 'retry', run },
    })

    const [item] = toastState.items
    expect(item.detail).toBe('Inténtalo de nuevo.')
    expect(item.reference).toBe('req_1')
    expect(item.action?.label).toBe('Reintentar')
    item.action?.run()
    expect(run).toHaveBeenCalledOnce()
  })

  it('dismisses by id and dismisses all', () => {
    const toast = useToast()
    const id = toast.success('Uno')
    toast.info('Dos')

    toast.dismiss(id)
    expect(toastState.items).toHaveLength(1)

    toast.dismissAll()
    expect(toastState.items).toHaveLength(0)
  })
})
