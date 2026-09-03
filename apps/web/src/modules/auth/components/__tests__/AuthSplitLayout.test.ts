/**
 * Tests for AuthSplitLayout (Fase 3 del rediseño Tailored Grid, issue #188)
 * Verifies: brand panel is decorative (aria-hidden), slot content renders,
 * default and custom tagline, axe.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import AuthSplitLayout from '../AuthSplitLayout.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('AuthSplitLayout', () => {
  it('renders the brand mark as decorative (aria-hidden)', () => {
    const wrapper = mount(AuthSplitLayout, { slots: { default: '<h1>Accede a NAVA</h1>' } })
    expect(wrapper.find('.auth-split__brand').attributes('aria-hidden')).toBe('true')
  })

  it('renders the default tagline', () => {
    const wrapper = mount(AuthSplitLayout, { slots: { default: '<h1>Accede a NAVA</h1>' } })
    expect(wrapper.text()).toContain('Gestión precisa para tu barbería.')
  })

  it('renders a custom tagline when provided', () => {
    const wrapper = mount(AuthSplitLayout, {
      props: { tagline: 'Recupera el acceso a tu cuenta.' },
      slots: { default: '<h1>Solicita tu código</h1>' },
    })
    expect(wrapper.text()).toContain('Recupera el acceso a tu cuenta.')
  })

  it('renders slot content as the real page heading', () => {
    const wrapper = mount(AuthSplitLayout, { slots: { default: '<h1>Accede a NAVA</h1>' } })
    expect(wrapper.get('h1').text()).toBe('Accede a NAVA')
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones con el contenido real de una pantalla', async () => {
      const wrapper = mount(AuthSplitLayout, {
        slots: {
          default: '<h1>Accede a NAVA</h1><label for="c">Correo</label><input id="c" />',
        },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })
})
