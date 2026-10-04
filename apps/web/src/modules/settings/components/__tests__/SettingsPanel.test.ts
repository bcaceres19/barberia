/**
 * SettingsPanel (DEC-110): sección con encabezado y alcance.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import SettingsPanel from '../SettingsPanel.vue'

describe('SettingsPanel', () => {
  it('is a region named by its own heading, with the content in the slot', () => {
    const wrapper = mount(SettingsPanel, {
      props: { id: 'marca', number: '02', title: 'Marca', description: 'Nombre y color.' },
      slots: { default: '<p>contenido</p>' },
    })

    const section = wrapper.get('section')
    expect(section.attributes('id')).toBe('marca')
    expect(section.attributes('aria-labelledby')).toBe('marca-title')
    expect(wrapper.get('#marca-title').text()).toBe('Marca')
    expect(wrapper.text()).toContain('Nombre y color.')
    expect(wrapper.text()).toContain('contenido')
  })

  it('hides the decorative number from assistive technology', () => {
    const wrapper = mount(SettingsPanel, { props: { id: 'x', number: '03', title: 'Hora' } })

    expect(wrapper.get('.settings-panel__number').attributes('aria-hidden')).toBe('true')
  })

  it('says where a change applies, in words and not only in color', () => {
    const wrapper = mount(SettingsPanel, {
      props: {
        id: 'x',
        number: '01',
        title: 'Pantalla',
        scope: 'device',
        scopeLabel: 'Solo este dispositivo',
      },
    })

    expect(wrapper.get('.settings-panel__scope--device').text()).toBe('Solo este dispositivo')
  })

  it('shows no scope when none is given', () => {
    const wrapper = mount(SettingsPanel, { props: { id: 'x', number: '06', title: 'Reservas' } })

    expect(wrapper.find('.settings-panel__scope').exists()).toBe(false)
  })

  it('has no axe violations', async () => {
    const wrapper = mount(SettingsPanel, {
      attachTo: document.body,
      props: {
        id: 'x',
        number: '01',
        title: 'Pantalla',
        scope: 'shop',
        scopeLabel: 'Para la barbería',
      },
      slots: { default: '<p>contenido</p>' },
    })
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
