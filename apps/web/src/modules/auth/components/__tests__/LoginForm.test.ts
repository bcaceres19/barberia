/**
 * Pruebas de LoginForm: render inicial, etiquetas/ayuda, validación de
 * forma, resumen de errores con foco, estados servidor (props), enlace de
 * recuperación y accesibilidad (vitest-axe). LoginForm no llama al API ni
 * conoce RFC 9457: esas pruebas viven en LoginPage.test.ts.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import LoginForm from '../LoginForm.vue'
import type { LoginServerErrorSummary } from '../LoginForm.vue'

// Igual que el resto de la suite de HU-009: jsdom no implementa Canvas2D,
// así que color-contrast no puede medir contraste real ahí (ya verificado
// en la tabla aprobada de estandar-diseno-visual.md §4.3).
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

const RouterLinkStub = {
  props: ['to'],
  template: "<a :href=\"typeof to === 'string' ? to : '#'\"><slot /></a>",
}

interface MountFormOptions {
  email?: string
  password?: string
  submitting?: boolean
  serverError?: LoginServerErrorSummary | null
  recoveryHref?: string
  attachToBody?: boolean
}

function mountForm(options: MountFormOptions = {}) {
  const { attachToBody, ...props } = options
  return mount(LoginForm, {
    props: {
      email: '',
      password: '',
      submitting: false,
      recoveryHref: '/recuperar-acceso',
      ...props,
    },
    global: {
      stubs: { RouterLink: RouterLinkStub },
    },
    attachTo: attachToBody ? document.body : undefined,
  })
}

describe('LoginForm', () => {
  describe('Render inicial', () => {
    it('renders email and password fields with accessible labels', () => {
      const wrapper = mountForm()
      expect(wrapper.find('label').exists()).toBe(true)
      const emailInput = wrapper.get('input[name="email"]')
      const passwordInput = wrapper.get('input[name="password"]')
      expect(emailInput.attributes('type')).toBe('email')
      expect(passwordInput.attributes('type')).toBe('password')
    })

    it('sets autocomplete attributes required by CA-010-07 hygiene', () => {
      const wrapper = mountForm()
      expect(wrapper.get('input[name="email"]').attributes('autocomplete')).toBe('username')
      expect(wrapper.get('input[name="password"]').attributes('autocomplete')).toBe(
        'current-password',
      )
    })

    it('renders a single primary submit action', () => {
      const wrapper = mountForm()
      const button = wrapper.get('button[type="submit"]')
      expect(button.text()).toBe('Iniciar sesión')
    })

    it('renders a visible, operable recovery link (CA-010-08)', () => {
      const wrapper = mountForm()
      const link = wrapper.get('a')
      expect(link.attributes('href')).toBe('/recuperar-acceso')
      expect(link.attributes('href')).not.toBe('#')
      expect(link.text().length).toBeGreaterThan(0)
    })

    it('does not show field errors before any submit attempt', () => {
      const wrapper = mountForm()
      expect(wrapper.find('.base-input__error').exists()).toBe(false)
    })
  })

  describe('Validación de forma en cliente', () => {
    it('shows an error-associated message per empty field on submit', async () => {
      const wrapper = mountForm()
      await wrapper.get('form').trigger('submit')

      const emailInput = wrapper.get('input[name="email"]')
      const passwordInput = wrapper.get('input[name="password"]')
      expect(emailInput.attributes('aria-invalid')).toBe('true')
      expect(passwordInput.attributes('aria-invalid')).toBe('true')
      expect(emailInput.attributes('aria-describedby')).toBeTruthy()
      expect(passwordInput.attributes('aria-describedby')).toBeTruthy()
    })

    it('does not emit submit while the form is locally invalid', async () => {
      const wrapper = mountForm()
      await wrapper.get('form').trigger('submit')
      expect(wrapper.emitted('submit')).toBeFalsy()
    })

    it('rejects a malformed email shape without contacting the server', async () => {
      const wrapper = mountForm({ email: 'no-es-un-correo', password: 'algo' })
      await wrapper.get('form').trigger('submit')
      expect(wrapper.emitted('submit')).toBeFalsy()
      expect(wrapper.get('input[name="email"]').attributes('aria-invalid')).toBe('true')
    })

    it('moves focus to the error summary when there is more than one error', async () => {
      const wrapper = mountForm({ attachToBody: true })
      await wrapper.get('form').trigger('submit')
      await wrapper.vm.$nextTick()
      const summary = wrapper.get('[role="alert"].login-form__summary')
      expect(document.activeElement).toBe(summary.element)
      wrapper.unmount()
    })

    it('emits a typed submit intent when the form is locally valid', async () => {
      const wrapper = mountForm({ email: 'barbero@ejemplo.test', password: 'contraseña-valida' })
      await wrapper.get('form').trigger('submit')
      expect(wrapper.emitted('submit')).toBeTruthy()
      expect(wrapper.emitted('submit')?.length).toBe(1)
    })
  })

  describe('Doble envío', () => {
    it('a second Enter/submit before re-render still emits at most one submit per call', async () => {
      const wrapper = mountForm({ email: 'barbero@ejemplo.test', password: 'contraseña-valida' })
      const form = wrapper.get('form')
      await form.trigger('submit')
      await form.trigger('submit')
      // Cada envío válido emite exactamente un evento; la deduplicación de
      // solicitudes reales ocurre en LoginPage (guardia sobre el estado de
      // envío), verificada en LoginPage.test.ts.
      expect(wrapper.emitted('submit')?.length).toBe(2)
    })

    it('disables inputs and shows loading affordance while submitting', () => {
      const wrapper = mountForm({ submitting: true })
      expect(wrapper.get('input[name="email"]').attributes('disabled')).toBeDefined()
      expect(wrapper.get('input[name="password"]').attributes('disabled')).toBeDefined()
      const button = wrapper.get('button[type="submit"]')
      expect(button.attributes('aria-busy')).toBe('true')
      expect(button.text()).toBe('Iniciando sesión…')
    })
  })

  describe('Estados de servidor (props)', () => {
    it('renders invalid-credentials summary without revealing which field was wrong', () => {
      const wrapper = mountForm({
        serverError: {
          tone: 'danger',
          title: 'No pudimos iniciar tu sesión',
          message: 'Revisa tu correo y contraseña e inténtalo de nuevo.',
        },
      })
      const alert = wrapper.get('[role="alert"].base-alert')
      expect(alert.text()).toContain('Revisa tu correo y contraseña')
    })

    it('renders a retry action for a recoverable network error (CA-010-03)', async () => {
      const wrapper = mountForm({
        serverError: {
          tone: 'warning',
          title: 'No pudimos conectar',
          message: 'Revisa tu conexión e inténtalo de nuevo.',
          actionLabel: 'Reintentar',
        },
      })
      const retryButton = wrapper.get('button:not([type="submit"])')
      expect(retryButton.text()).toBe('Reintentar')
      await retryButton.trigger('click')
      expect(wrapper.emitted('retry')).toBeTruthy()
    })

    it('renders a 429 rate-limit explanation without claiming HU-007 exists', () => {
      const wrapper = mountForm({
        serverError: {
          tone: 'warning',
          title: 'Demasiados intentos',
          message: 'Espera un momento antes de volver a intentarlo.',
        },
      })
      expect(wrapper.text()).toContain('Demasiados intentos')
    })

    it('renders an unexpected-error state with a safe, generic message', () => {
      const wrapper = mountForm({
        serverError: {
          tone: 'danger',
          title: 'Ocurrió un error inesperado',
          message: 'Inténtalo de nuevo en unos segundos.',
        },
      })
      expect(wrapper.text()).toContain('Ocurrió un error inesperado')
    })
  })

  describe('Accesibilidad', () => {
    it('has no axe violations in the default state', async () => {
      const wrapper = mountForm()
      const results = await axe(wrapper.element, axeOptions)
      expect(results).toHaveNoViolations()
    })

    it('has no axe violations with field errors and a server error visible', async () => {
      const wrapper = mountForm({
        serverError: {
          tone: 'danger',
          title: 'No pudimos iniciar tu sesión',
          message: 'Revisa tu correo y contraseña e inténtalo de nuevo.',
        },
      })
      await wrapper.get('form').trigger('submit')
      const results = await axe(wrapper.element, axeOptions)
      expect(results).toHaveNoViolations()
    })

    it('keeps a coherent tab order: email, password, submit, recovery link', () => {
      const wrapper = mountForm()
      const focusable = wrapper.findAll('input, button, a')
      const order = focusable.map((el) => `${el.element.tagName}:${el.attributes('name') ?? ''}`)
      expect(order).toEqual(['INPUT:email', 'INPUT:password', 'BUTTON:', 'A:'])
    })
  })
})
