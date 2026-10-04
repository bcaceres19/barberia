/**
 * Tests for EmptyScene (DEC-113): una escena por tipo de vacío, slots de
 * titular/ayuda/acción, ilustración decorativa y sin violaciones axe.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import EmptyScene from '../EmptyScene.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('EmptyScene', () => {
  it.each(['agenda', 'services', 'team', 'assignments', 'search'] as const)(
    'draws the %s scene and marks it as a modifier class',
    (scene) => {
      const wrapper = mount(EmptyScene, { props: { scene } })
      expect(wrapper.classes()).toContain(`empty-scene--${scene}`)
      expect(wrapper.find('svg').exists()).toBe(true)
    },
  )

  it('keeps the illustration decorative for assistive technology', () => {
    const wrapper = mount(EmptyScene, { props: { scene: 'team' } })
    expect(wrapper.find('.empty-scene__art').attributes('aria-hidden')).toBe('true')
  })

  it('renders the title, the hint and the action from its slots', () => {
    const wrapper = mount(EmptyScene, {
      props: { scene: 'services' },
      slots: {
        title: 'Aún no tienes servicios registrados.',
        hint: 'Crea el primero.',
        action: '<button type="button">Agregar servicio</button>',
      },
    })
    expect(wrapper.text()).toContain('Aún no tienes servicios registrados.')
    expect(wrapper.text()).toContain('Crea el primero.')
    expect(wrapper.find('button').text()).toBe('Agregar servicio')
  })

  it('omits the hint and the action when their slots are empty', () => {
    const wrapper = mount(EmptyScene, { props: { scene: 'search' }, slots: { title: 'Nada.' } })
    expect(wrapper.find('.empty-scene__hint').exists()).toBe(false)
    expect(wrapper.find('.empty-scene__action').exists()).toBe(false)
  })

  it('has no axe violations with a title, a hint and an action', async () => {
    const wrapper = mount(EmptyScene, {
      props: { scene: 'agenda' },
      slots: {
        title: 'No hay turnos hoy.',
        hint: 'Aparecerán aquí.',
        action: '<button type="button">Nuevo turno</button>',
      },
      attachTo: document.body,
    })
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    wrapper.unmount()
  })
})
