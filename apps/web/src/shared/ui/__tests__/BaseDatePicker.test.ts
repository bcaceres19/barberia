import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseDatePicker from '../BaseDatePicker.vue'

describe('BaseDatePicker · formularios', () => {
  it('opens a blank civil field on the provided tenant today without inventing a selection', async () => {
    const wrapper = mount(BaseDatePicker, {
      props: { modelValue: '', today: '2026-10-02' },
      global: { stubs: { teleport: true } },
    })
    expect(wrapper.get('button').text()).toContain('Elegir fecha')
    await wrapper.get('button').trigger('click')
    expect(wrapper.get('[data-date="2026-10-02"]').attributes('tabindex')).toBe('0')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.setProps({ disabled: true })
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })
  it('generates distinct trigger ids for multiple fields', () => {
    const wrapper = mount({
      components: { BaseDatePicker },
      template: '<div><BaseDatePicker model-value=""/><BaseDatePicker model-value=""/></div>',
    })
    const buttons = wrapper.findAll('button')
    expect(buttons[0]!.attributes('id')).not.toBe(buttons[1]!.attributes('id'))
    wrapper.unmount()
  })
  it('does not bubble Escape to the containing form dialog', async () => {
    const wrapper = mount(BaseDatePicker, {
      props: { modelValue: '2026-10-02' },
      attachTo: document.body,
      global: { stubs: { teleport: true } },
    })
    await wrapper.get('button').trigger('click')
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    let propagated = false
    const listener = () => {
      propagated = true
    }
    document.addEventListener('keydown', listener)
    wrapper.get('[role="dialog"]').element.dispatchEvent(event)
    document.removeEventListener('keydown', listener)
    expect(propagated).toBe(false)
    expect(event.defaultPrevented).toBe(true)
    wrapper.unmount()
  })
})
