/**
 * Tests for OtpInput component (issue #212): ranuras de código reglado.
 * Verifica el contrato del trabajo requerido §3 — valor lógico único,
 * pegado con distribución automática, avance, retroceso, flechas,
 * filtrado a dígitos, nombre accesible del grupo y estados.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import OtpInput from '../OtpInput.vue'

// Ver BaseButton.test.ts: color-contrast se desactiva por la ausencia de
// Canvas2D en jsdom; el contraste ya está verificado en la tabla aprobada
// de estandar-diseno-visual.md §4.3.
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function slots(wrapper: ReturnType<typeof mount>) {
  return wrapper.findAll('.otp-input__slot')
}

describe('OtpInput', () => {
  describe('Rendering', () => {
    it('renders six slots', () => {
      const wrapper = mount(OtpInput)
      expect(slots(wrapper)).toHaveLength(6)
    })

    it('renders the label', () => {
      const wrapper = mount(OtpInput, { props: { label: 'Código de 6 dígitos' } })
      expect(wrapper.find('.otp-input__label').text()).toBe('Código de 6 dígitos')
    })

    it('distributes an existing modelValue across the slots', () => {
      const wrapper = mount(OtpInput, { props: { modelValue: '482913' } })
      const values = slots(wrapper).map((s) => (s.element as HTMLInputElement).value)
      expect(values).toEqual(['4', '8', '2', '9', '1', '3'])
    })

    it('each slot has inputmode numeric and autocomplete one-time-code', () => {
      const wrapper = mount(OtpInput)
      slots(wrapper).forEach((slot) => {
        expect(slot.attributes('inputmode')).toBe('numeric')
        expect(slot.attributes('autocomplete')).toBe('one-time-code')
        expect(slot.attributes('maxlength')).toBe('1')
      })
    })
  })

  describe('Valor lógico único', () => {
    it('emits the full six-digit string on update:modelValue, not per-slot values', async () => {
      const wrapper = mount(OtpInput)
      await slots(wrapper)[0].setValue('7')
      const emitted = wrapper.emitted('update:modelValue')
      expect(emitted).toBeTruthy()
      expect(emitted?.[0]?.[0]).toBe('7')
    })

    it('filters non-digit characters typed into a slot', async () => {
      const wrapper = mount(OtpInput)
      const slot = slots(wrapper)[0].element as HTMLInputElement
      slot.value = 'a'
      await slots(wrapper)[0].trigger('input')
      // Un carácter no numérico deja la ranura vacía y no emite un dígito.
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('')
    })

    it('emits complete when the sixth digit is filled', async () => {
      const wrapper = mount(OtpInput, { props: { modelValue: '48291' } })
      await slots(wrapper)[5].setValue('3')
      expect(wrapper.emitted('complete')?.[0]?.[0]).toBe('482913')
    })
  })

  describe('Pegado con distribución automática', () => {
    it('distributes a pasted six-digit code across all slots from the first one', async () => {
      const wrapper = mount(OtpInput)
      const clipboardData = { getData: () => '482913' }
      await slots(wrapper)[0].trigger('paste', { clipboardData })
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('482913')
    })

    it('strips non-digit characters from the pasted text', async () => {
      const wrapper = mount(OtpInput)
      const clipboardData = { getData: () => '48-29 13' }
      await slots(wrapper)[0].trigger('paste', { clipboardData })
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('482913')
    })

    it('pasting into a middle slot fills from that position onward', async () => {
      const wrapper = mount(OtpInput, { props: { modelValue: '48' } })
      const clipboardData = { getData: () => '2913' }
      await slots(wrapper)[2].trigger('paste', { clipboardData })
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('482913')
    })
  })

  describe('Avance, retroceso y flechas', () => {
    it('advances focus to the next slot after typing a digit', async () => {
      const wrapper = mount(OtpInput, { attachTo: document.body })
      await slots(wrapper)[0].setValue('4')
      await wrapper.vm.$nextTick()
      expect(document.activeElement).toBe(slots(wrapper)[1].element)
      wrapper.unmount()
    })

    it('Backspace on an empty slot clears and focuses the previous slot', async () => {
      const wrapper = mount(OtpInput, { props: { modelValue: '48' }, attachTo: document.body })
      const thirdSlot = slots(wrapper)[2]
      await thirdSlot.trigger('keydown', { key: 'Backspace' })
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('4')
      await wrapper.vm.$nextTick()
      expect(document.activeElement).toBe(slots(wrapper)[1].element)
      wrapper.unmount()
    })

    it('ArrowLeft/ArrowRight move focus between slots', async () => {
      const wrapper = mount(OtpInput, { attachTo: document.body })
      const secondSlot = slots(wrapper)[1]
      await secondSlot.trigger('keydown', { key: 'ArrowRight' })
      await wrapper.vm.$nextTick()
      expect(document.activeElement).toBe(slots(wrapper)[2].element)

      await slots(wrapper)[2].trigger('keydown', { key: 'ArrowLeft' })
      await wrapper.vm.$nextTick()
      expect(document.activeElement).toBe(slots(wrapper)[1].element)
      wrapper.unmount()
    })
  })

  describe('Estados', () => {
    it('applies the invalid state to every slot when error is set', () => {
      const wrapper = mount(OtpInput, { props: { error: 'El código no es válido.' } })
      slots(wrapper).forEach((slot) => {
        expect(slot.classes()).toContain('otp-input__slot--invalid')
        expect(slot.attributes('aria-invalid')).toBe('true')
      })
      expect(wrapper.find('.otp-input__error').text()).toBe('El código no es válido.')
    })

    it('disables every slot when disabled is true', () => {
      const wrapper = mount(OtpInput, { props: { disabled: true } })
      slots(wrapper).forEach((slot) => {
        expect(slot.attributes('disabled')).toBeDefined()
      })
    })

    it('marks a filled slot with the filled class', () => {
      const wrapper = mount(OtpInput, { props: { modelValue: '4' } })
      expect(slots(wrapper)[0].classes()).toContain('otp-input__slot--filled')
      expect(slots(wrapper)[1].classes()).not.toContain('otp-input__slot--filled')
    })
  })

  describe('Accesibilidad', () => {
    it('group is labelled by the visible label', () => {
      const wrapper = mount(OtpInput, { props: { label: 'Código de 6 dígitos' } })
      const group = wrapper.find('[role="group"]')
      const labelId = group.attributes('aria-labelledby')
      expect(labelId).toBeDefined()
      expect(wrapper.find(`#${labelId}`).text()).toBe('Código de 6 dígitos')
    })

    it('each slot has its own coherent accessible name', () => {
      const wrapper = mount(OtpInput)
      const values = slots(wrapper).map((s) => s.attributes('aria-label'))
      expect(values).toEqual([
        'Dígito 1 de 6',
        'Dígito 2 de 6',
        'Dígito 3 de 6',
        'Dígito 4 de 6',
        'Dígito 5 de 6',
        'Dígito 6 de 6',
      ])
    })

    it('sin violaciones en reposo con label', async () => {
      const wrapper = mount(OtpInput, { props: { label: 'Código de 6 dígitos' } })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })

    it('sin violaciones en estado de error', async () => {
      const wrapper = mount(OtpInput, {
        props: {
          label: 'Código de 6 dígitos',
          error: 'El código no es válido.',
          modelValue: '482913',
        },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })

    it('sin violaciones deshabilitado', async () => {
      const wrapper = mount(OtpInput, {
        props: { label: 'Código de 6 dígitos', disabled: true, modelValue: '482913' },
      })
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })
})
