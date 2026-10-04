/**
 * SaveBar (DEC-110): resumen de lo pendiente y las dos acciones.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import SaveBar from '../SaveBar.vue'

describe('SaveBar', () => {
  it('announces what is pending politely', () => {
    const wrapper = mount(SaveBar, { props: { summary: 'Marca y vocabulario', saving: false } })

    expect(wrapper.attributes('role')).toBe('status')
    expect(wrapper.text()).toContain('Cambios sin guardar')
    expect(wrapper.text()).toContain('Marca y vocabulario')
  })

  it('emits discard and save', async () => {
    const wrapper = mount(SaveBar, { props: { summary: 'Datos básicos', saving: false } })
    const [discard, save] = wrapper.findAll('button')

    await discard!.trigger('click')
    await save!.trigger('click')

    expect(wrapper.emitted('discard')).toHaveLength(1)
    expect(wrapper.emitted('save')).toHaveLength(1)
  })

  it('blocks both actions while saving', () => {
    const wrapper = mount(SaveBar, { props: { summary: 'Datos básicos', saving: true } })

    expect(wrapper.findAll('button').every((b) => b.attributes('disabled') !== undefined)).toBe(
      true,
    )
  })

  it('has no axe violations', async () => {
    const wrapper = mount(SaveBar, {
      attachTo: document.body,
      props: { summary: 'Datos básicos', saving: false },
    })
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
