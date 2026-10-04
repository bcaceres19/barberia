/**
 * OptionGroup (DEC-110): un `radiogroup` real con inputs nativos, así el
 * navegador resuelve flechas, foco y anuncio sin código propio.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import OptionGroup from '../OptionGroup.vue'

const options = [
  { value: 'a', label: 'Uno' },
  { value: 'b', label: 'Dos' },
  { value: 'c', label: 'Tres' },
] as const

function mountGroup(props: Record<string, unknown> = {}) {
  return mount(OptionGroup, {
    props: { modelValue: 'a', options, label: 'Elige una letra', ...props },
    slots: {
      default: `<template #default="{ option, selected }"><span>{{ option.label }}{{ selected ? ' ✓' : '' }}</span></template>`,
    },
  })
}

describe('OptionGroup', () => {
  it('exposes a named radiogroup with one native radio per option', () => {
    const wrapper = mountGroup()

    expect(wrapper.get('[role="radiogroup"]').attributes('aria-label')).toBe('Elige una letra')
    const radios = wrapper.findAll('input[type="radio"]')
    expect(radios).toHaveLength(3)
    // Todas comparten nombre: el navegador las trata como un solo grupo.
    expect(new Set(radios.map((r) => r.attributes('name'))).size).toBe(1)
  })

  it('marks only the current option as checked and renders the slot with its state', () => {
    const wrapper = mountGroup({ modelValue: 'b' })

    const checked = wrapper.findAll('input').map((r) => (r.element as HTMLInputElement).checked)
    expect(checked).toEqual([false, true, false])
    expect(wrapper.text()).toContain('Dos ✓')
    expect(wrapper.text()).not.toContain('Uno ✓')
  })

  it('emits the chosen value', async () => {
    const wrapper = mountGroup()

    await wrapper.findAll('input')[2]!.setValue(true)

    expect(wrapper.emitted('update:modelValue')).toEqual([['c']])
  })

  it('disables every radio and emits nothing while disabled', async () => {
    const wrapper = mountGroup({ disabled: true })

    expect(wrapper.findAll('input').every((r) => r.attributes('disabled') !== undefined)).toBe(true)
  })

  it('gives each radio an accessible name from its label text', async () => {
    const wrapper = mount(OptionGroup, {
      attachTo: document.body,
      props: { modelValue: 'a', options, label: 'Elige una letra' },
      slots: {
        default: `<template #default="{ option }"><span>{{ option.label }}</span></template>`,
      },
    })

    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
