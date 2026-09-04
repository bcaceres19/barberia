/**
 * Tests for PageState component (issue #189, atlas panel-agenda-eventos).
 * Verifies: variant→role mapping, spinner only on loading, slots, headline
 * passthrough.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import PageState from '../PageState.vue'
import BaseSpinner from '../BaseSpinner.vue'
import BaseButton from '../BaseButton.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('PageState', () => {
  it('shows BaseSpinner and no divider when loading', () => {
    const wrapper = mount(PageState, {
      props: { variant: 'loading', headline: 'Cargando barberos…' },
    })
    expect(wrapper.findComponent(BaseSpinner).exists()).toBe(true)
    expect(wrapper.find('.page-state__divider').exists()).toBe(false)
  })

  it('shows the divider and no spinner for a non-loading variant', () => {
    const wrapper = mount(PageState, {
      props: { variant: 'warning', headline: 'No pudimos cargar esta sección' },
    })
    expect(wrapper.findComponent(BaseSpinner).exists()).toBe(false)
    expect(wrapper.find('.page-state__divider').exists()).toBe(true)
  })

  it('renders the headline unmodified, never rewording it', () => {
    const headline = 'No pudimos cargar la agenda de este barbero'
    const wrapper = mount(PageState, { props: { variant: 'warning', headline } })
    expect(wrapper.get('.page-state__headline').text()).toBe(headline)
  })

  it('omits the status label when not provided (informational states)', () => {
    const wrapper = mount(PageState, {
      props: { variant: 'info', headline: 'Aún no tienes barberos registrados.' },
    })
    expect(wrapper.find('.page-state__status').exists()).toBe(false)
  })

  it('renders the status label when provided', () => {
    const wrapper = mount(PageState, {
      props: { variant: 'warning', statusLabel: 'Atención', headline: 'Error' },
    })
    expect(wrapper.get('.page-state__status').text()).toBe('Atención')
  })

  it('renders default and action slots only when provided', () => {
    const wrapper = mount(PageState, {
      props: { variant: 'warning', headline: 'Error' },
      slots: { default: 'Revisa tu conexión.', action: '<button>Reintentar</button>' },
    })
    expect(wrapper.get('.page-state__body').text()).toBe('Revisa tu conexión.')
    expect(wrapper.get('.page-state__action').text()).toContain('Reintentar')
  })

  it.each([
    ['loading', undefined, 'status'],
    ['info', undefined, 'status'],
    ['warning', undefined, 'alert'],
    ['danger', undefined, 'alert'],
    ['warning', 'status', 'status'],
  ] as const)('resolves role for variant=%s (override=%s) to %s', (variant, role, expected) => {
    const wrapper = mount(PageState, { props: { variant, role, headline: 'Titular' } })
    expect(wrapper.attributes('role')).toBe(expected)
    expect(wrapper.attributes('aria-live')).toBe(expected === 'alert' ? 'assertive' : 'polite')
  })

  it('has no obvious accessibility violations with an action', async () => {
    const wrapper = mount(
      {
        components: { PageState, BaseButton },
        template: `
          <PageState variant="warning" status-label="Atención" headline="No pudimos cargar esta sección">
            Revisa tu conexión e inténtalo de nuevo.
            <template #action>
              <BaseButton type="button" variant="secondary">Reintentar</BaseButton>
            </template>
          </PageState>
        `,
      },
      {},
    )
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })
})
