/**
 * Tests for BaseInput component
 * Verifies: types, states, label/hint/error, events, accessibility
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BaseInput from '../BaseInput.vue'

// Ver BaseButton.test.ts: color-contrast se desactiva por la ausencia de
// Canvas2D en jsdom; el contraste ya está verificado en la tabla aprobada
// de estandar-diseno-visual.md §4.3.
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('BaseInput', () => {
  describe('Rendering', () => {
    it('renders input element', () => {
      const wrapper = mount(BaseInput)
      expect(wrapper.find('input').exists()).toBe(true)
    })

    it('renders label when provided', () => {
      const wrapper = mount(BaseInput, { props: { label: 'Email' } })
      expect(wrapper.find('label').text()).toBe('Email')
    })

    it('renders hint when provided', () => {
      const wrapper = mount(BaseInput, { props: { hint: 'Enter your email' } })
      expect(wrapper.find('.base-input__hint').text()).toBe('Enter your email')
    })

    it('renders error when provided', () => {
      const wrapper = mount(BaseInput, { props: { error: 'Invalid email' } })
      expect(wrapper.find('.base-input__error').text()).toBe('Invalid email')
    })

    it('applies type attribute', () => {
      const types = ['text', 'email', 'password', 'number', 'tel', 'url', 'search'] as const
      types.forEach((type) => {
        const wrapper = mount(BaseInput, { props: { type } })
        expect(wrapper.find('input').attributes('type')).toBe(type)
      })
    })

    it('applies placeholder', () => {
      const wrapper = mount(BaseInput, { props: { placeholder: 'Enter text' } })
      expect(wrapper.find('input').attributes('placeholder')).toBe('Enter text')
    })

    it('renders leading slot', () => {
      const wrapper = mount(BaseInput, { slots: { leading: '@', default: 'Text' } })
      expect(wrapper.find('.base-input__icon--leading').exists()).toBe(true)
      expect(wrapper.find('.base-input__icon--leading').text()).toBe('@')
    })

    it('renders trailing slot', () => {
      const wrapper = mount(BaseInput, { slots: { trailing: '🔍', default: 'Text' } })
      expect(wrapper.find('.base-input__icon--trailing').exists()).toBe(true)
      expect(wrapper.find('.base-input__icon--trailing').text()).toBe('🔍')
    })
  })

  describe('v-model', () => {
    it('binds modelValue to input value', () => {
      const wrapper = mount(BaseInput, { props: { modelValue: 'test value' } })
      expect((wrapper.find('input').element as HTMLInputElement).value).toBe('test value')
    })

    it('emits update:modelValue on input', async () => {
      const wrapper = mount(BaseInput)
      const input = wrapper.find('input')
      await input.setValue('new value')
      expect(wrapper.emitted('update:modelValue')).toBeTruthy()
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('new value')
    })

    it('emits input event on input', async () => {
      const wrapper = mount(BaseInput)
      const input = wrapper.find('input')
      await input.setValue('new value')
      expect(wrapper.emitted('input')).toBeTruthy()
    })

    it('emits change event on change', async () => {
      const wrapper = mount(BaseInput)
      const input = wrapper.find('input')
      await input.setValue('new value')
      await input.trigger('change')
      expect(wrapper.emitted('change')).toBeTruthy()
    })

    it('emits blur event on blur', async () => {
      const wrapper = mount(BaseInput)
      const input = wrapper.find('input')
      await input.trigger('blur')
      expect(wrapper.emitted('blur')).toBeTruthy()
    })

    it('emits focus event on focus', async () => {
      const wrapper = mount(BaseInput)
      const input = wrapper.find('input')
      await input.trigger('focus')
      expect(wrapper.emitted('focus')).toBeTruthy()
    })
  })

  describe('States', () => {
    it('applies disabled class and attributes when disabled', () => {
      const wrapper = mount(BaseInput, { props: { disabled: true } })
      expect(wrapper.find('.base-input__wrapper').classes()).toContain(
        'base-input__wrapper--disabled',
      )
      expect(wrapper.find('input').attributes('disabled')).toBeDefined()
      expect(wrapper.find('input').attributes('aria-disabled')).toBe('true')
    })

    it('applies readonly class and attributes when readonly', () => {
      const wrapper = mount(BaseInput, { props: { readonly: true } })
      expect(wrapper.find('.base-input__wrapper').classes()).toContain(
        'base-input__wrapper--readonly',
      )
      expect(wrapper.find('input').attributes('readonly')).toBeDefined()
      expect(wrapper.find('input').attributes('aria-readonly')).toBe('true')
    })

    it('applies invalid class when error provided', () => {
      const wrapper = mount(BaseInput, { props: { error: 'Error' } })
      expect(wrapper.find('input').classes()).toContain('base-input--invalid')
      expect(wrapper.find('input').attributes('aria-invalid')).toBe('true')
    })

    it('shows hint when no error', () => {
      const wrapper = mount(BaseInput, { props: { hint: 'Hint', error: '' } })
      expect(wrapper.find('.base-input__hint').exists()).toBe(true)
    })

    it('shows error instead of hint when both provided', () => {
      const wrapper = mount(BaseInput, { props: { hint: 'Hint', error: 'Error' } })
      expect(wrapper.find('.base-input__hint').exists()).toBe(false)
      expect(wrapper.find('.base-input__error').exists()).toBe(true)
    })

    it('sets required attribute and aria-required', () => {
      const wrapper = mount(BaseInput, { props: { required: true } })
      expect(wrapper.find('input').attributes('required')).toBeDefined()
      expect(wrapper.find('input').attributes('aria-required')).toBe('true')
    })
  })

  describe('Accessibility', () => {
    it('links input to label via for/id', () => {
      const wrapper = mount(BaseInput, { props: { label: 'Email', id: 'email-input' } })
      expect(wrapper.find('label').attributes('for')).toBe('email-input')
      expect(wrapper.find('input').attributes('id')).toBe('email-input')
    })

    it('generates id if not provided', () => {
      const wrapper = mount(BaseInput, { props: { label: 'Email' } })
      const inputId = wrapper.find('input').attributes('id')
      const labelFor = wrapper.find('label').attributes('for')
      expect(inputId).toBeDefined()
      expect(inputId).toBe(labelFor)
    })

    it('sets aria-describedby for hint', () => {
      const wrapper = mount(BaseInput, { props: { hint: 'Hint', id: 'test-input' } })
      expect(wrapper.find('input').attributes('aria-describedby')).toContain('test-input-hint')
    })

    it('sets aria-describedby for error', () => {
      const wrapper = mount(BaseInput, { props: { error: 'Error', id: 'test-input' } })
      expect(wrapper.find('input').attributes('aria-describedby')).toContain('test-input-error')
    })

    it('sets aria-describedby for both hint and error', () => {
      const wrapper = mount(BaseInput, {
        props: { hint: 'Hint', error: 'Error', id: 'test-input' },
      })
      const describedBy = wrapper.find('input').attributes('aria-describedby')
      expect(describedBy).toContain('test-input-hint')
      expect(describedBy).toContain('test-input-error')
    })

    it('required label shows exactly one asterisk (issue #51: not two)', () => {
      const wrapper = mount(BaseInput, { props: { label: 'Email', required: true } })
      const label = wrapper.find('.base-input__label')
      expect(label.classes()).toContain('base-input__label--required')
      expect(label.text().replace(/\s+/g, ' ').trim()).toBe('Email *')
    })
  })

  describe('Validation attributes', () => {
    it('applies pattern attribute', () => {
      const wrapper = mount(BaseInput, { props: { pattern: '[a-z]+' } })
      expect(wrapper.find('input').attributes('pattern')).toBe('[a-z]+')
    })

    it('applies minlength attribute', () => {
      const wrapper = mount(BaseInput, { props: { minlength: 3 } })
      expect(wrapper.find('input').attributes('minlength')).toBe('3')
    })

    it('applies maxlength attribute', () => {
      const wrapper = mount(BaseInput, { props: { maxlength: 10 } })
      expect(wrapper.find('input').attributes('maxlength')).toBe('10')
    })

    it('applies min/max/step for number type', () => {
      const wrapper = mount(BaseInput, { props: { type: 'number', min: 0, max: 100, step: 5 } })
      expect(wrapper.find('input').attributes('min')).toBe('0')
      expect(wrapper.find('input').attributes('max')).toBe('100')
      expect(wrapper.find('input').attributes('step')).toBe('5')
    })
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones con label y ayuda', async () => {
      const wrapper = mount(BaseInput, {
        props: { label: 'Correo', hint: 'Usaremos este correo para confirmar el turno' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })

    it('sin violaciones en estado de error', async () => {
      const wrapper = mount(BaseInput, {
        props: { label: 'Correo', error: 'Ingresa un correo válido' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })

    it('sin violaciones deshabilitado', async () => {
      const wrapper = mount(BaseInput, {
        props: { label: 'Correo', disabled: true, modelValue: 'a@b.com' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })

  describe('Mostrar/ocultar contraseña', () => {
    it('no renderiza el botón en campos que no son password', () => {
      const wrapper = mount(BaseInput, { props: { type: 'text' } })
      expect(wrapper.find('.base-input__toggle').exists()).toBe(false)
    })

    it('renderiza el botón y alterna el tipo del input al hacer click', async () => {
      const wrapper = mount(BaseInput, { props: { type: 'password', modelValue: 'secreta' } })
      const toggle = wrapper.find('.base-input__toggle')
      expect(toggle.exists()).toBe(true)
      expect(wrapper.find('input').attributes('type')).toBe('password')
      expect(toggle.attributes('aria-label')).toBe('Mostrar contraseña')
      expect(toggle.attributes('aria-pressed')).toBe('false')

      await toggle.trigger('click')

      expect(wrapper.find('input').attributes('type')).toBe('text')
      expect(toggle.attributes('aria-label')).toBe('Ocultar contraseña')
      expect(toggle.attributes('aria-pressed')).toBe('true')

      await toggle.trigger('click')
      expect(wrapper.find('input').attributes('type')).toBe('password')
    })

    it('es type="button" para no enviar el formulario al hacer click', () => {
      const wrapper = mount(BaseInput, { props: { type: 'password' } })
      expect(wrapper.find('.base-input__toggle').attributes('type')).toBe('button')
    })

    it('se deshabilita junto con el campo', () => {
      const wrapper = mount(BaseInput, { props: { type: 'password', disabled: true } })
      expect(wrapper.find('.base-input__toggle').attributes('disabled')).toBeDefined()
    })

    it('el slot trailing no se renderiza cuando el campo es password (el botón tiene prioridad)', () => {
      const wrapper = mount(BaseInput, {
        props: { type: 'password' },
        slots: { trailing: '🔍' },
      })
      expect(wrapper.find('.base-input__toggle').exists()).toBe(true)
      expect(wrapper.find('.base-input__icon--trailing').exists()).toBe(false)
    })

    it('sin violaciones de accesibilidad con el botón visible', async () => {
      const wrapper = mount(BaseInput, {
        props: { type: 'password', label: 'Contraseña', modelValue: 'secreta' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })

  describe('Fallthrough de clase', () => {
    // class ya no es un prop propio (auditoría HU-009); el fallthrough
    // automático de Vue lo aplica al elemento raíz (base-input__wrapper).
    it('hereda una clase pasada por el consumidor', () => {
      const wrapper = mount(BaseInput, { attrs: { class: 'my-custom-class' } })
      expect(wrapper.classes()).toContain('my-custom-class')
    })
  })
})
