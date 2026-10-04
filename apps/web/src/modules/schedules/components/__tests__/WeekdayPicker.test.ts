import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import WeekdayPicker from '../WeekdayPicker.vue'

function mountPicker(
  props: Partial<{ modelValue: number; disabled: boolean; error: string }> = {},
) {
  return mount(WeekdayPicker, {
    attachTo: document.body,
    props: { modelValue: 1, label: 'Día', ...props },
  })
}

describe('WeekdayPicker', () => {
  it('exposes seven radios named by the full weekday, only the chosen one in the tab order', () => {
    const wrapper = mountPicker({ modelValue: 3 })
    const radios = wrapper.findAll('[role="radio"]')

    expect(radios.map((r) => r.attributes('aria-label'))).toEqual([
      'Lunes',
      'Martes',
      'Miércoles',
      'Jueves',
      'Viernes',
      'Sábado',
      'Domingo',
    ])
    expect(radios.filter((r) => r.attributes('tabindex') === '0')).toHaveLength(1)
    expect(radios[2]!.attributes('aria-checked')).toBe('true')
    wrapper.unmount()
  })

  it('emits the ISO weekday when a chip is pressed', async () => {
    const wrapper = mountPicker()
    await wrapper.findAll('[role="radio"]')[4]!.trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[5]])
    wrapper.unmount()
  })

  it('moves with the arrow keys, wrapping at both ends, and jumps with Home/End', async () => {
    const wrapper = mountPicker({ modelValue: 7 })
    const radios = wrapper.findAll('[role="radio"]')

    await radios[6]!.trigger('keydown', { key: 'ArrowRight' })
    await radios[0]!.trigger('keydown', { key: 'ArrowLeft' })
    await radios[2]!.trigger('keydown', { key: 'Home' })
    await radios[2]!.trigger('keydown', { key: 'End' })

    expect(wrapper.emitted('update:modelValue')).toEqual([[1], [7], [1], [7]])
    wrapper.unmount()
  })

  it('does nothing while disabled', async () => {
    const wrapper = mountPicker({ disabled: true })
    await wrapper.findAll('[role="radio"]')[1]!.trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('announces its error and has no axe violations', async () => {
    const wrapper = mountPicker({ error: 'Elige un día.' })

    expect(wrapper.get('[role="alert"]').text()).toBe('Elige un día.')
    expect(
      await axe(wrapper.element, { rules: { 'color-contrast': { enabled: false } } }),
    ).toHaveNoViolations()
    wrapper.unmount()
  })
})
