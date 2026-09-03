/**
 * Tests for PageHeader component (estandar-diseno-visual.md §6.6)
 * Verifies: title as unique h1, subtitle, back link, actions slot, axe.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory, type RouteLocationRaw } from 'vue-router'
import { axe } from 'vitest-axe'
import PageHeader from '../PageHeader.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

interface HeaderProps {
  title: string
  subtitle?: string
  backTo?: RouteLocationRaw
  backLabel?: string
}

async function mountHeader(props: HeaderProps, slots?: Record<string, string>) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'inicio', component: { template: '<div />' } },
      { path: '/panel', name: 'panel', component: { template: '<div />' } },
    ],
  })
  await router.push('/panel')
  await router.isReady()
  return mount(PageHeader, { props, slots, global: { plugins: [router] } })
}

describe('PageHeader', () => {
  it('renders the title as the only h1', async () => {
    const wrapper = await mountHeader({ title: 'Agenda' })
    const headings = wrapper.findAll('h1')
    expect(headings).toHaveLength(1)
    expect(headings[0].text()).toContain('Agenda')
  })

  it('renders subtitle when provided', async () => {
    const wrapper = await mountHeader({ title: 'Agenda', subtitle: 'miércoles, 2 de septiembre' })
    expect(wrapper.text()).toContain('miércoles, 2 de septiembre')
  })

  it('omits subtitle when not provided', async () => {
    const wrapper = await mountHeader({ title: 'Agenda' })
    expect(wrapper.find('.page-header__subtitle').exists()).toBe(false)
  })

  it('renders no back link without backTo', async () => {
    const wrapper = await mountHeader({ title: 'Turno' })
    expect(wrapper.find('.page-header__back').exists()).toBe(false)
  })

  it('renders a back link to backTo with backLabel', async () => {
    const wrapper = await mountHeader({
      title: 'Turno',
      backTo: { name: 'panel' },
      backLabel: 'Volver a la agenda',
    })
    const back = wrapper.get('.page-header__back')
    expect(back.text()).toContain('Volver a la agenda')
    expect(back.attributes('href')).toBe('/panel')
  })

  it('renders the actions slot only when provided', async () => {
    const withActions = await mountHeader(
      { title: 'Servicios' },
      { actions: '<button>Agregar servicio</button>' },
    )
    expect(withActions.find('.page-header__actions').exists()).toBe(true)
    expect(withActions.text()).toContain('Agregar servicio')

    const withoutActions = await mountHeader({ title: 'Servicios' })
    expect(withoutActions.find('.page-header__actions').exists()).toBe(false)
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones con título, subtítulo, retorno y acciones', async () => {
      const wrapper = await mountHeader(
        { title: 'Turno', subtitle: 'Contexto', backTo: { name: 'panel' }, backLabel: 'Volver' },
        { actions: '<button>Reprogramar turno</button>' },
      )
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })
})
