/**
 * Tests for AuthSplitLayout (Fase 3 del rediseño Tailored Grid, issue #188;
 * marca reglada del issue #213). Verifies: brand panel is decorative
 * (aria-hidden), the content-column mark is decorative too, slot content
 * renders, default/custom tagline, the required caption, and axe.
 */
import { afterEach, describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import AuthSplitLayout from '../AuthSplitLayout.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountLayout(props: Record<string, unknown> = {}, slotHtml = '<h1>Accede a NAVA</h1>') {
  return mount(AuthSplitLayout, {
    props: { caption: 'ACCESO SEGURO', ...props },
    slots: { default: slotHtml },
  })
}

describe('AuthSplitLayout', () => {
  it('renders the brand mark as decorative (aria-hidden)', () => {
    const wrapper = mountLayout()
    expect(wrapper.find('.auth-split__brand').attributes('aria-hidden')).toBe('true')
  })

  it('renders the content-column mark as decorative too (issue #213)', () => {
    const wrapper = mountLayout()
    const mark = wrapper.get('.auth-split__mark')
    expect(mark.attributes('aria-hidden')).toBe('true')
    expect(mark.text()).toContain('NAVA')
  })

  it('adds an ornamental backdrop inside the decorative brand panel, without text', () => {
    const wrapper = mountLayout()
    const backdrop = wrapper.get('.auth-split__brand .brand-backdrop')
    expect(backdrop.attributes('aria-hidden')).toBe('true')
    expect(backdrop.text()).toBe('')
  })

  describe('entrada escalonada de la tarjeta', () => {
    afterEach(() => vi.useRealTimers())

    it('se retira tras la entrada para que alertas tardías no hereden sus retrasos', async () => {
      vi.useFakeTimers()
      const wrapper = mountLayout()
      expect(wrapper.get('.auth-split__card').classes()).toContain('auth-split__card--intro')
      await vi.advanceTimersByTimeAsync(1700)
      expect(wrapper.get('.auth-split__card').classes()).not.toContain('auth-split__card--intro')
    })
  })

  it('renders the default tagline', () => {
    const wrapper = mountLayout()
    expect(wrapper.text()).toContain('Gestión precisa para tu negocio.')
  })

  it('renders a custom tagline when provided', () => {
    const wrapper = mountLayout(
      { tagline: 'Recupera el acceso a tu cuenta.' },
      '<h1>Solicita tu código</h1>',
    )
    expect(wrapper.text()).toContain('Recupera el acceso a tu cuenta.')
  })

  it('renders the caption next to NAVA in the brand panel and again at the foot of the card for mobile', () => {
    const wrapper = mountLayout({ caption: 'RECUPERACIÓN SEGURA' })
    expect(wrapper.get('.auth-split__caption').text()).toBe('RECUPERACIÓN SEGURA · NAVA')
    expect(wrapper.get('.auth-split__mobile-caption').text()).toBe('RECUPERACIÓN SEGURA · NAVA')
  })

  it('renders slot content as the real page heading', () => {
    const wrapper = mountLayout()
    expect(wrapper.get('h1').text()).toBe('Accede a NAVA')
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones con el contenido real de una pantalla', async () => {
      const wrapper = mountLayout(
        {},
        '<h1>Accede a NAVA</h1><label for="c">Correo</label><input id="c" />',
      )
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })
})
