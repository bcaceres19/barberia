/**
 * Pruebas de ToastRegion (DEC-095): la región refleja la cola, abre un solo
 * aviso a la vez, pausa el tiempo con el cursor, ejecuta la acción y
 * descarta con Escape o con "Descartar todos".
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { TOAST_DURATION_MS, pushToast, resetToasts, toastState } from '@/shared/model/toastStore'
import ToastRegion from '../ToastRegion.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

beforeEach(() => {
  vi.useFakeTimers()
  resetToasts()
})

afterEach(() => {
  resetToasts()
  vi.useRealTimers()
})

describe('ToastRegion', () => {
  it('is a persistent labelled region, even without toasts', () => {
    const wrapper = mount(ToastRegion)

    expect(wrapper.attributes('role')).toBe('region')
    expect(wrapper.attributes('aria-label')).toBe('Avisos')
    expect(wrapper.findAll('.base-toast')).toHaveLength(0)
  })

  it('anchors to the window or to its host container', () => {
    expect(mount(ToastRegion).classes()).toContain('toast-region--viewport')
    expect(mount(ToastRegion, { props: { placement: 'host' } }).classes()).toContain(
      'toast-region--host',
    )
  })

  it('renders the queue newest first and removes a toast when its time ends', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({ variant: 'success', title: 'Primero' })
    pushToast({ variant: 'danger', title: 'Segundo' })
    await wrapper.vm.$nextTick()

    expect(wrapper.findAll('.base-toast__title').map((node) => node.text())).toEqual([
      'Segundo',
      'Primero',
    ])

    vi.advanceTimersByTime(TOAST_DURATION_MS.success)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.base-toast__title').map((node) => node.text())).toEqual(['Segundo'])
  })

  it('opens one toast at a time and holds the countdown while open', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({ variant: 'success', title: 'Uno', detail: 'Detalle uno' })
    pushToast({ variant: 'warning', title: 'Dos', detail: 'Detalle dos' })
    await wrapper.vm.$nextTick()

    const heads = wrapper.findAll('.base-toast__head')
    await heads[0].trigger('click')
    await heads[1].trigger('click')

    expect(wrapper.findAll('.base-toast__panel--open')).toHaveLength(1)
    expect(heads[0].attributes('aria-expanded')).toBe('false')
    expect(heads[1].attributes('aria-expanded')).toBe('true')

    vi.advanceTimersByTime(TOAST_DURATION_MS.danger * 2)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.base-toast')).toHaveLength(1)
  })

  it('pauses the countdown while the cursor is over a toast', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({ variant: 'info', title: 'Aviso' })
    await wrapper.vm.$nextTick()

    await wrapper.find('.base-toast').trigger('mouseenter')
    vi.advanceTimersByTime(60_000)
    expect(toastState.items).toHaveLength(1)

    await wrapper.find('.base-toast').trigger('mouseleave')
    vi.advanceTimersByTime(TOAST_DURATION_MS.info)
    expect(toastState.items).toHaveLength(0)
  })

  it('runs the action once and dismisses the toast', async () => {
    const run = vi.fn()
    const wrapper = mount(ToastRegion)
    pushToast({
      variant: 'danger',
      title: 'No pudimos guardar',
      detail: 'Inténtalo de nuevo.',
      action: { label: 'Reintentar', icon: 'retry', run },
    })
    await wrapper.vm.$nextTick()

    await wrapper.find('.base-toast__head').trigger('click')
    await wrapper.find('.base-toast__action').trigger('click')

    expect(run).toHaveBeenCalledOnce()
    expect(toastState.items).toHaveLength(0)
  })

  it('lets an action publish a follow-up toast without losing it', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({
      variant: 'danger',
      title: 'No pudimos guardar',
      action: {
        label: 'Reintentar',
        run: () => pushToast({ variant: 'success', title: 'Servicio guardado' }),
      },
    })
    await wrapper.vm.$nextTick()

    await wrapper.find('.base-toast__head').trigger('click')
    await wrapper.find('.base-toast__action').trigger('click')

    expect(toastState.items.map((item) => item.title)).toEqual(['Servicio guardado'])
  })

  it('dismisses the focused toast with Escape', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({ variant: 'success', title: 'Uno' })
    pushToast({ variant: 'info', title: 'Dos' })
    await wrapper.vm.$nextTick()

    await wrapper.findAll('.toast-region__item')[0].trigger('keydown', { key: 'Escape' })

    expect(toastState.items.map((item) => item.title)).toEqual(['Uno'])
  })

  it('offers "Descartar todos" with a count only when there is more than one toast', async () => {
    const wrapper = mount(ToastRegion)
    pushToast({ variant: 'success', title: 'Uno' })
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.toast-region__clear').exists()).toBe(false)

    pushToast({ variant: 'info', title: 'Dos' })
    await wrapper.vm.$nextTick()
    const clear = wrapper.find('.toast-region__clear')
    expect(clear.text()).toContain('Descartar todos')
    expect(clear.find('.toast-region__count').text()).toBe('2')

    await clear.trigger('click')
    expect(toastState.items).toHaveLength(0)
  })

  it('has no axe violations with a mixed queue', async () => {
    // axe usa temporizadores propios: con el reloj simulado nunca resuelve.
    vi.useRealTimers()
    const wrapper = mount(ToastRegion, { attachTo: document.body })
    pushToast({ variant: 'success', title: 'Servicio guardado', detail: 'Ya está en el catálogo.' })
    pushToast({
      variant: 'danger',
      title: 'No pudimos guardar el servicio',
      detail: 'Revisa tu conexión.',
      reference: 'req_8d41c2',
      action: { label: 'Reintentar', icon: 'retry', run: () => undefined },
    })
    await wrapper.vm.$nextTick()

    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    wrapper.unmount()
  })
})
