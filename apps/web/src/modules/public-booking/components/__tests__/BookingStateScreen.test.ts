/**
 * Pruebas de BookingStateScreen (DEC-111): las cuatro pantallas de estado de
 * página de los pasos públicos conservan el copy, los roles y el reintento
 * que cada página repetía por su cuenta.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BookingStateScreen from '../BookingStateScreen.vue'

function mountScreen(
  status: 'loading' | 'not-found' | 'network-error' | 'unexpected-error',
  requestId?: string,
) {
  return mount(BookingStateScreen, {
    props: { status, requestId, title: 'Elegir servicio', loadingHeadline: 'Cargando servicios…' },
  })
}

describe('BookingStateScreen', () => {
  it('always provides the hidden page heading and one <main>', () => {
    for (const status of ['loading', 'not-found', 'network-error', 'unexpected-error'] as const) {
      const wrapper = mountScreen(status)
      expect(wrapper.find('main').exists()).toBe(true)
      expect(wrapper.find('h1').text()).toBe('Elegir servicio')
    }
  })

  it('announces loading politely, without a retry action', () => {
    const wrapper = mountScreen('loading')
    expect(wrapper.text()).toContain('Cargando servicios…')
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('shows the neutral not-found copy as an alert with a retry', () => {
    const wrapper = mountScreen('not-found')
    expect(wrapper.find('[role="alert"]').text()).toContain('No encontramos ese enlace')
    expect(wrapper.text()).toContain('Revisa que copiaste la dirección completa')
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('shows the connection copy as a warning alert', () => {
    const wrapper = mountScreen('network-error')
    expect(wrapper.find('[role="alert"]').text()).toContain('No pudimos conectar')
    expect(wrapper.text()).toContain('Atención')
  })

  it('shares the support request id only when the server provided one', () => {
    expect(mountScreen('unexpected-error', 'req-abc-123').text()).toContain('req-abc-123')
    const bare = mountScreen('unexpected-error')
    expect(bare.text()).toContain('Inténtalo de nuevo en unos segundos.')
    expect(bare.text()).not.toContain('código')
  })

  it('emits retry from the action in every failure state', async () => {
    for (const status of ['not-found', 'network-error', 'unexpected-error'] as const) {
      const wrapper = mountScreen(status)
      await wrapper.find('button').trigger('click')
      expect(wrapper.emitted('retry')).toHaveLength(1)
    }
  })

  it('has no accessibility violations in any state', async () => {
    for (const status of ['loading', 'not-found', 'network-error', 'unexpected-error'] as const) {
      const wrapper = mount(BookingStateScreen, {
        props: { status, title: 'Elegir servicio', loadingHeadline: 'Cargando…' },
        attachTo: document.body,
      })
      expect(await axe(wrapper.element)).toHaveNoViolations()
      wrapper.unmount()
    }
  })
})
