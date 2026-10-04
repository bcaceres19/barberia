import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import BaseSelect from '../BaseSelect.vue'

const options = [
  { value: 'break', label: 'Descanso' },
  { value: 'lunch', label: 'Almuerzo' },
]
function render(disabled = false) {
  return mount(BaseSelect, {
    props: { modelValue: 'break', options, label: 'Tipo de bloqueo', disabled },
    attachTo: document.body,
  })
}
describe('BaseSelect', () => {
  it('opens on the selected option, chooses with keyboard and restores focus', async () => {
    const wrapper = render()
    const trigger = wrapper.get('button')
    await trigger.trigger('keydown', { key: 'ArrowDown' })
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(document.activeElement?.textContent).toContain('Descanso')
    await wrapper.get('[role="listbox"]').trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement?.textContent).toContain('Almuerzo')
    await wrapper.get('[role="listbox"]').trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')).toEqual([['lunch']])
    expect(document.activeElement).toBe(trigger.element)
    expect(trigger.attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })
  it('handles Home, End and Escape without closing the containing dialog', async () => {
    const wrapper = render()
    await wrapper.get('button').trigger('click')
    const list = wrapper.get('[role="listbox"]')
    await list.trigger('keydown', { key: 'End' })
    expect(document.activeElement?.textContent).toContain('Almuerzo')
    await list.trigger('keydown', { key: 'Home' })
    expect(document.activeElement?.textContent).toContain('Descanso')
    const escape = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    let propagated = false
    const listener = () => {
      propagated = true
    }
    document.addEventListener('keydown', listener)
    list.element.dispatchEvent(escape)
    document.removeEventListener('keydown', listener)
    expect(propagated).toBe(false)
    expect(escape.defaultPrevented).toBe(true)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })
  it('closes on Tab and outside clicks and blocks changes while disabled', async () => {
    const wrapper = render()
    await wrapper.get('button').trigger('click')
    await wrapper.get('[role="listbox"]').trigger('keydown', { key: 'Tab' })
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    await wrapper.get('button').trigger('click')
    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    await wrapper.get('button').trigger('click')
    await wrapper.setProps({ disabled: true })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })
  it('preserves numeric option values', async () => {
    const wrapper = mount(BaseSelect, {
      props: {
        modelValue: 1,
        options: [
          { value: 1, label: 'Lunes' },
          { value: 3, label: 'Miércoles' },
        ],
        label: 'Día',
      },
    })
    await wrapper.get('button').trigger('click')
    await wrapper.findAll('[role="option"]')[1]!.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[3]])
    wrapper.unmount()
  })
})
