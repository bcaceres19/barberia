/**
 * SettingsIndex (DEC-110): navegación por anclas con la sección vigente marcada.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import SettingsIndex from '../SettingsIndex.vue'

const items = [
  { id: 'uno', number: '01', label: 'Pantalla' },
  { id: 'dos', number: '02', label: 'Marca' },
] as const

describe('SettingsIndex', () => {
  it('is a labelled navigation of real links to each section', () => {
    const wrapper = mount(SettingsIndex, { props: { items, active: 'uno' } })

    expect(wrapper.get('nav').attributes('aria-label')).toBe('Secciones de configuración')
    expect(wrapper.findAll('a').map((a) => a.attributes('href'))).toEqual(['#uno', '#dos'])
  })

  it('marks only the active section with aria-current', () => {
    const wrapper = mount(SettingsIndex, { props: { items, active: 'dos' } })

    const current = wrapper.findAll('a').map((a) => a.attributes('aria-current'))
    expect(current).toEqual([undefined, 'location'])
    expect(wrapper.get('[aria-current]').text()).toContain('Marca')
  })

  it('emits the section instead of jumping, so the page can scroll smoothly', async () => {
    const wrapper = mount(SettingsIndex, { props: { items, active: 'uno' } })

    await wrapper.findAll('a')[1]!.trigger('click')

    expect(wrapper.emitted('select')).toEqual([['dos']])
  })

  it('has no axe violations', async () => {
    const wrapper = mount(SettingsIndex, {
      attachTo: document.body,
      props: { items, active: 'uno' },
    })
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
