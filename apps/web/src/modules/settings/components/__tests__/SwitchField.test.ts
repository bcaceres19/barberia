/**
 * SwitchField (DEC-110): interruptor con `role="switch"`, estado anunciado
 * con `aria-checked` y también con texto.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import SwitchField from '../SwitchField.vue'

function mountSwitch(props: Record<string, unknown> = {}) {
  return mount(SwitchField, {
    attachTo: document.body,
    props: {
      modelValue: false,
      label: 'Reducir animaciones',
      hint: 'Quita el movimiento.',
      ...props,
    },
  })
}

describe('SwitchField', () => {
  it('is a named switch described by its hint', () => {
    const wrapper = mountSwitch()
    const control = wrapper.get('[role="switch"]')

    expect(control.attributes('aria-checked')).toBe('false')
    const labelId = control.attributes('aria-labelledby')!
    expect(wrapper.get(`#${labelId}`).text()).toBe('Reducir animaciones')
    expect(wrapper.get(`#${control.attributes('aria-describedby')!}`).text()).toBe(
      'Quita el movimiento.',
    )
    wrapper.unmount()
  })

  it('states the value with text as well as color', async () => {
    const wrapper = mountSwitch()
    expect(wrapper.text()).toContain('Desactivado')

    await wrapper.setProps({ modelValue: true })

    expect(wrapper.get('[role="switch"]').attributes('aria-checked')).toBe('true')
    expect(wrapper.text()).toContain('Activado')
    wrapper.unmount()
  })

  it('emits the opposite value on click and on Enter/Space through the native button', async () => {
    const wrapper = mountSwitch({ modelValue: true })

    await wrapper.get('[role="switch"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[false]])
    wrapper.unmount()
  })

  it('does nothing when disabled', async () => {
    const wrapper = mountSwitch({ disabled: true })

    await wrapper.get('[role="switch"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('has no axe violations in either state', async () => {
    const wrapper = mountSwitch()
    expect(await axe(wrapper.element)).toHaveNoViolations()
    await wrapper.setProps({ modelValue: true })
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
