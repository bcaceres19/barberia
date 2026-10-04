/**
 * Pruebas de BookingProgress (DEC-111): un rombo por paso; el actual lleva
 * `aria-current="step"`, los anteriores se anuncian como completados y un
 * paso mayor que el último deja todos completados. El estado nunca depende
 * solo del color.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BookingProgress from '../BookingProgress.vue'

describe('BookingProgress', () => {
  it('lists the four steps in order inside a labelled navigation', () => {
    const wrapper = mount(BookingProgress, { props: { step: 1 } })
    expect(wrapper.find('nav').attributes('aria-label')).toBe('Progreso de la reserva')
    const labels = wrapper.findAll('li').map((item) => item.find('.pb-progress__label').text())
    expect(labels).toEqual(['Servicio', 'Barbero', 'Horario', 'Tus datos'])
  })

  it('marks only the current step with aria-current and the earlier ones as completed', () => {
    const wrapper = mount(BookingProgress, { props: { step: 3 } })
    const items = wrapper.findAll('li')
    expect(items.map((item) => item.attributes('aria-current'))).toEqual([
      undefined,
      undefined,
      'step',
      undefined,
    ])
    expect(items[0]!.text()).toContain('(completado)')
    expect(items[1]!.text()).toContain('(completado)')
    expect(items[2]!.text()).not.toContain('(completado)')
    expect(items[3]!.text()).not.toContain('(completado)')
  })

  it('completes every step when the step is past the last one (booking confirmed)', () => {
    const wrapper = mount(BookingProgress, { props: { step: 5 } })
    const items = wrapper.findAll('li')
    expect(items.every((item) => item.attributes('aria-current') === undefined)).toBe(true)
    expect(items.every((item) => item.text().includes('(completado)'))).toBe(true)
  })

  it('draws the traveled stretch from the shown step, not from the first paint', async () => {
    const wrapper = mount(BookingProgress, { props: { step: 4 } })
    // Antes del primer fotograma la regla parte de cero…
    expect(wrapper.find('nav').attributes('style')).toContain('--pb-fill: 0')
    await new Promise((resolve) => requestAnimationFrame(() => resolve(null)))
    await wrapper.vm.$nextTick()
    // …y después llega al último rombo.
    expect(wrapper.find('nav').attributes('style')).toContain('--pb-fill: 1')
  })

  it('has no accessibility violations', async () => {
    const wrapper = mount(BookingProgress, { props: { step: 2 }, attachTo: document.body })
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
