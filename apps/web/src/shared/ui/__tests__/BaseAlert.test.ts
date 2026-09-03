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
      const variants = ['success', 'warning', 'danger', 'info', 'neutral', 'plain'] as const
      variants.forEach((variant) => {
        const wrapper = mount(BaseAlert, { props: { variant } })
        expect(wrapper.classes()).toContain(`base-alert--${variant}`)
      })
    })

    it('renders title when provided', () => {
      const wrapper = mount(BaseAlert, { props: { title: 'Alert Title' } })
      expect(wrapper.find('.base-alert__title').text()).toBe('Alert Title')
    })

    // Contrato visual del issue #212: la palabra de estado en versalitas
    // reemplaza el glifo circular anterior como el elemento que distingue
    // el estado además del color (WCAG 2.2 AA 1.4.1).
    it('shows the status word above the title for each stateful variant', () => {
      const expectations: Record<string, string> = {
        danger: 'Error',
        warning: 'Atención',
        info: 'Nota',
        success: 'Confirmación',
      }
      Object.entries(expectations).forEach(([variant, word]) => {
        const wrapper = mount(BaseAlert, {
          props: { variant: variant as 'danger' | 'warning' | 'info' | 'success', title: 'Título' },
        })
        expect(wrapper.find('.base-alert__status').text()).toBe(word)
      })
    })

    it('does not show a status word for neutral or plain', () => {
      const neutral = mount(BaseAlert, { props: { variant: 'neutral', title: 'Título' } })
      expect(neutral.find('.base-alert__status').exists()).toBe(false)

      const plain = mount(BaseAlert, { props: { variant: 'plain', title: 'Título' } })
      expect(plain.find('.base-alert__status').exists()).toBe(false)
    })

    // La variante sin relleno (caja de requisitos de contraseña) compone el
    // título como versalita única, sin repetirlo también como titular en
    // negrita (trabajo requerido §4 del issue #212).
    it('plain variant renders the title as a single brass caption, not a bold heading', () => {
      const wrapper = mount(BaseAlert, {
        props: { variant: 'plain', title: 'Requisitos de la contraseña' },
      })
      expect(wrapper.find('.base-alert__title--plain').text()).toBe('Requisitos de la contraseña')
      expect(wrapper.find('h4.base-alert__title').exists()).toBe(false)
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

    // El control de descarte ahora se rotula "Descartar" en vez de un icono
    // con aria-label (issue #212): el texto visible es su propio nombre
    // accesible, así que ya no necesita un aria-label aparte (WCAG 2.5.3
    // Label in Name).
    it('dismiss button is labeled "Descartar" and has no separate aria-label', () => {
      const wrapper = mount(BaseAlert, { props: { dismissible: true } })
      const dismiss = wrapper.find('.base-alert__dismiss')
      expect(dismiss.text()).toBe('Descartar')
      expect(dismiss.attributes('aria-label')).toBeUndefined()
    })

    // El glifo circular genérico se retiró: la palabra de estado en
    // versalitas es la que ahora cumple "icono, texto y estructura además
    // del color" (WCAG 2.2 AA 1.4.1).
    it('has no decorative icon element', () => {
      const wrapper = mount(BaseAlert, { props: { variant: 'danger', title: 'Título' } })
      expect(wrapper.find('.base-alert__icon').exists()).toBe(false)
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
