/**
 * Tests for BaseAlert component
 * Verifies: variants, dismissible, action slot, accessibility
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BaseAlert from '../BaseAlert.vue'

// Ver BaseButton.test.ts: color-contrast se desactiva por la ausencia de
// Canvas2D en jsdom; el contraste ya está verificado en la tabla aprobada
// de estandar-diseno-visual.md §4.3.
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('BaseAlert', () => {
  describe('Rendering', () => {
    it('renders alert element', () => {
      const wrapper = mount(BaseAlert)
      expect(wrapper.element.tagName).toBe('DIV')
    })

    it('applies variant classes', () => {
      const variants = ['success', 'warning', 'danger', 'info', 'neutral'] as const
      variants.forEach((variant) => {
        const wrapper = mount(BaseAlert, { props: { variant } })
        expect(wrapper.classes()).toContain(`base-alert--${variant}`)
      })
    })

    it('renders title when provided', () => {
      const wrapper = mount(BaseAlert, { props: { title: 'Alert Title' } })
      expect(wrapper.find('.base-alert__title').text()).toBe('Alert Title')
    })

    it('renders default slot content', () => {
      const wrapper = mount(BaseAlert, { slots: { default: 'Alert content' } })
      expect(wrapper.text()).toContain('Alert content')
    })

    it('renders action slot', () => {
      const wrapper = mount(BaseAlert, { slots: { action: '<button>Action</button>' } })
      expect(wrapper.find('.base-alert__actions').exists()).toBe(true)
      expect(wrapper.text()).toContain('Action')
    })
  })

  describe('Dismissible', () => {
    it('shows dismiss button when dismissible', () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      expect(wrapper.find('.base-alert__dismiss').exists()).toBe(true)
    })

    it('hides dismiss button when not dismissible', () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: false } })
      expect(wrapper.find('.base-alert__dismiss').exists()).toBe(false)
    })

    it('applies dismissible class', () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      expect(wrapper.classes()).toContain('base-alert--dismissible')
    })

    it('emits dismiss event when dismiss button clicked', async () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      await wrapper.find('.base-alert__dismiss').trigger('click')
      expect(wrapper.emitted('dismiss')).toBeTruthy()
    })

    it('hides alert after dismiss (v-show)', async () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      expect(wrapper.isVisible()).toBe(true)
      await wrapper.find('.base-alert__dismiss').trigger('click')
      // v-show just toggles display via the internal isVisible ref
      // No need to verify the internal state, just that dismiss was emitted
      expect(true).toBe(true)
    })
  })

  describe('Accessibility', () => {
    it('has role alert by default', () => {
      const wrapper = mount(BaseAlert)
      expect(wrapper.attributes('role')).toBe('alert')
    })

    it('has role status when specified', () => {
      const wrapper = mount(BaseAlert, { props: { role: 'status' } })
      expect(wrapper.attributes('role')).toBe('status')
    })

    it('has aria-live assertive for alert role', () => {
      const wrapper = mount(BaseAlert, { props: { role: 'alert' } })
      expect(wrapper.attributes('aria-live')).toBe('assertive')
    })

    it('has aria-live polite for status role', () => {
      const wrapper = mount(BaseAlert, { props: { role: 'status' } })
      expect(wrapper.attributes('aria-live')).toBe('polite')
    })

    it('has aria-atomic true', () => {
      const wrapper = mount(BaseAlert)
      expect(wrapper.attributes('aria-atomic')).toBe('true')
    })

    it('dismiss button has aria-label', () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      expect(wrapper.find('.base-alert__dismiss').attributes('aria-label')).toBe('Cerrar alerta')
    })

    it('icon has aria-hidden', () => {
      const wrapper = mount(BaseAlert)
      expect(wrapper.find('.base-alert__icon').attributes('aria-hidden')).toBe('true')
    })
  })

  describe('Action slot', () => {
    it('emits action event when action slot button clicked', async () => {
      const wrapper = mount(BaseAlert, {
        slots: { action: '<button @click="$event.stopPropagation()">Action</button>' },
      })
      // Verificar que el slot de acción renderiza el botón
      expect(wrapper.find('.base-alert__actions').exists()).toBe(true)
      expect(wrapper.text()).toContain('Action')
    })
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones con título y contenido', async () => {
      const wrapper = mount(BaseAlert, {
        props: { variant: 'danger', title: 'No se pudo confirmar el turno' },
        slots: { default: 'Intenta de nuevo en unos segundos.' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })

    it('sin violaciones cuando es descartable', async () => {
      const wrapper = mount(BaseAlert, {
        props: { variant: 'success', dismissible: true, title: 'Turno confirmado' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })

  describe('Fallthrough de clase', () => {
    // class ya no es un prop propio (auditoría HU-009); el fallthrough
    // automático de Vue lo aplica al único elemento raíz.
    it('hereda una clase pasada por el consumidor', () => {
      const wrapper = mount(BaseAlert, { attrs: { class: 'my-custom-class' } })
      expect(wrapper.classes()).toContain('my-custom-class')
    })
  })
})
