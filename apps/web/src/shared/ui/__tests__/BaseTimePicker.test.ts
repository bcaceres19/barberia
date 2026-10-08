import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BaseTimePicker from '../BaseTimePicker.vue'

function render(modelValue = '', disabled = false) {
  return mount(BaseTimePicker, {
    attachTo: document.body,
    props: { modelValue, label: 'Hora de inicio', disabled, required: true },
    global: { stubs: { teleport: true } },
  })
}
describe('BaseTimePicker', () => {
  it('does not change a blank time until applying a selection', async () => {
    const wrapper = render()
    expect(wrapper.get('.base-time-picker__trigger').text()).toContain('Elegir hora')
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    expect(wrapper.findAll('[role="spinbutton"]')).toHaveLength(2)
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    await wrapper.get('[role="spinbutton"][aria-valuemax="23"]').setValue('23')
    await wrapper.get('[role="spinbutton"][aria-valuemax="59"]').setValue('55')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.get('.base-time-picker__apply').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['23:55']])
    expect(document.activeElement).toBe(wrapper.get('.base-time-picker__trigger').element)
    wrapper.unmount()
  })
  it('moves directly from a completed hour to minutes and keeps partial typing intact', async () => {
    const wrapper = render()
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    const hour = wrapper.get<HTMLInputElement>('[aria-valuemax="23"]')
    const minute = wrapper.get<HTMLInputElement>('[aria-valuemax="59"]')
    await hour.setValue('2')
    expect(hour.element.value).toBe('2')
    await hour.setValue('23')
    expect(document.activeElement).toBe(minute.element)
    wrapper.unmount()
  })
  it('preserves exact minutes, including midnight, and validates manual precision', async () => {
    const wrapper = render('00:07')
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="23"]').element.value).toBe('00')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="59"]').element.value).toBe('07')
    await wrapper.get('[aria-valuemax="59"]').setValue('60')
    expect(wrapper.get('.base-time-picker__apply').attributes('disabled')).toBeDefined()
    await wrapper.get('[aria-valuemax="59"]').setValue('xy')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="59"]').element.value).toBe('')
    expect(wrapper.get('[aria-valuemax="59"]').attributes('aria-invalid')).toBe('true')
    await wrapper.get('[aria-valuemax="59"]').setValue('-3')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="59"]').element.value).toBe('3')
    await wrapper.get('[aria-valuemax="59"]').setValue('7')
    await wrapper.get('.base-time-picker__apply').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['00:07']])
    wrapper.unmount()
  })
  it('cancels the draft, isolates Escape and restores the original value on reopening', async () => {
    const wrapper = render('12:37')
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    await wrapper.get('[aria-valuemax="23"]').setValue('19')
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    let propagated = false
    const listener = () => {
      propagated = true
    }
    document.addEventListener('keydown', listener)
    wrapper.get('[role="dialog"]').element.dispatchEvent(event)
    document.removeEventListener('keydown', listener)
    await wrapper.vm.$nextTick()
    expect(event.defaultPrevented).toBe(true)
    expect(propagated).toBe(false)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="23"]').element.value).toBe('12')
    expect(wrapper.get<HTMLInputElement>('[aria-valuemax="59"]').element.value).toBe('37')
    wrapper.unmount()
  })
  it('supports arrows, Home and End, traps Tab and closes on outside click', async () => {
    const wrapper = render()
    await wrapper.get('.base-time-picker__trigger').trigger('keydown', { key: 'ArrowDown' })
    const input = wrapper.get<HTMLInputElement>('[aria-valuemax="23"]')
    await input.trigger('keydown', { key: 'ArrowUp' })
    expect(input.element.value).toBe('10')
    await input.trigger('keydown', { key: 'Home' })
    expect(input.element.value).toBe('00')
    await input.trigger('keydown', { key: 'End' })
    expect(input.element.value).toBe('23')
    await input.trigger('keydown', { key: 'ArrowUp' })
    expect(input.element.value).toBe('00')
    const last = wrapper.get('.base-time-picker__apply')
    ;(last.element as HTMLElement).focus()
    await last.trigger('keydown', { key: 'Tab' })
    expect(document.activeElement).toBe(wrapper.get('.base-time-picker__close').element)
    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })
  it('adjusts exact minutes with the keyboard and applies with Enter', async () => {
    const wrapper = render('09:37')
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    const minute = wrapper.get<HTMLInputElement>('[aria-valuemax="59"]')
    await minute.trigger('keydown', { key: 'ArrowUp' })
    expect(minute.element.value).toBe('38')
    await minute.trigger('keydown', { key: 'ArrowDown' })
    expect(minute.element.value).toBe('37')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await minute.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')).toEqual([['09:37']])
    wrapper.unmount()
  })
  it('cannot open while disabled and closes when saving disables the field', async () => {
    const wrapper = render('', true)
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    await wrapper.setProps({ disabled: false })
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    await wrapper.setProps({ disabled: true })
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })
  it('has accessible names and no detectable axe violations', async () => {
    const wrapper = render('22:07')
    await wrapper.get('.base-time-picker__trigger').trigger('click')
    const options = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }
    expect(await axe(wrapper.element as HTMLElement, options)).toHaveNoViolations()
    expect(wrapper.get('[aria-valuemax="59"]').attributes('aria-valuenow')).toBe('7')
    wrapper.unmount()
  })
})
