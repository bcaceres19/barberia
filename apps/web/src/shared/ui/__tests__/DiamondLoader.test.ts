/**
 * Tests for DiamondLoader: texto accesible único, decoración oculta a la
 * tecnología de asistencia, rotación de frases y respeto de
 * prefers-reduced-motion.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import DiamondLoader from '../DiamondLoader.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function stubMatchMedia(reduced: boolean) {
  vi.stubGlobal(
    'matchMedia',
    vi.fn().mockReturnValue({ matches: reduced, addEventListener: vi.fn() }),
  )
  window.matchMedia = globalThis.matchMedia
}

describe('DiamondLoader', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    stubMatchMedia(false)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('exposes only the label to assistive tech and hides graphic and phrases', () => {
    const wrapper = mount(DiamondLoader, { props: { label: 'Cargando barberos…' } })
    expect(wrapper.find('.diamond-loader__label').text()).toBe('Cargando barberos…')
    expect(wrapper.find('.diamond-loader__stage').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('.diamond-loader__text').attributes('aria-hidden')).toBe('true')
    wrapper.unmount()
  })

  it('applies the layout modifier', () => {
    const wrapper = mount(DiamondLoader, { props: { label: 'Cargando', layout: 'inline' } })
    expect(wrapper.classes()).toContain('diamond-loader--inline')
    wrapper.unmount()
  })

  it('rotates through the phrases and wraps around', async () => {
    const wrapper = mount(DiamondLoader, {
      props: { label: 'Cargando', phrases: ['Uno', 'Dos'] },
    })
    expect(wrapper.find('.diamond-loader__phrase').text()).toBe('Uno')
    await vi.advanceTimersByTimeAsync(2600)
    expect(wrapper.find('.diamond-loader__phrase').text()).toBe('Dos')
    await vi.advanceTimersByTimeAsync(2600)
    expect(wrapper.find('.diamond-loader__phrase').text()).toBe('Uno')
    wrapper.unmount()
  })

  it('keeps the first phrase when the user prefers reduced motion', async () => {
    stubMatchMedia(true)
    const wrapper = mount(DiamondLoader, {
      props: { label: 'Cargando', phrases: ['Uno', 'Dos'] },
    })
    await vi.advanceTimersByTimeAsync(10000)
    expect(wrapper.find('.diamond-loader__phrase').text()).toBe('Uno')
    wrapper.unmount()
  })

  it('stops its timer on unmount', () => {
    const wrapper = mount(DiamondLoader, { props: { label: 'Cargando' } })
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })

  it('has no obvious accessibility violations', async () => {
    vi.useRealTimers()
    const wrapper = mount(DiamondLoader, {
      props: { label: 'Cargando' },
      attachTo: document.body,
    })
    expect(await axe(wrapper.element as HTMLElement, axeOptions)).toHaveNoViolations()
    wrapper.unmount()
  })
})
