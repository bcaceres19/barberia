/**
 * Pruebas de AppNav (HU-012; dock NAVA desde la Fase 2, issue #135):
 * navegación semántica, marca la entrada activa con clase y con el
 * aria-current nativo de RouterLink, icono decorativo por entrada y
 * composición de extraItems tras la base "Agenda".
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'
import AppNav from '../AppNav.vue'
import type { NavItem } from '@/shared/navigation/navItem'

async function mountNav(initialRoute = '/panel', extraItems?: NavItem[]) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/panel', name: 'panel', component: { template: '<div />' } }],
  })
  await router.push(initialRoute)
  await router.isReady()
  return mount(AppNav, { props: { extraItems }, global: { plugins: [router] } })
}

describe('AppNav', () => {
  it('renders a semantic nav landmark with a labelled link (Fase 2: "Agenda", not "Panel")', async () => {
    const wrapper = await mountNav()
    expect(wrapper.get('nav').attributes('aria-label')).toBeTruthy()
    expect(wrapper.get('a').text()).toBe('Agenda')
  })

  it('marks the current route as active, including native aria-current from RouterLink', async () => {
    const wrapper = await mountNav('/panel')
    const link = wrapper.get('a')
    expect(link.classes()).toContain('app-nav__link--active')
    expect(link.attributes('aria-current')).toBe('page')
  })

  it('renders a decorative icon (aria-hidden) alongside each label', async () => {
    const wrapper = await mountNav()
    const icon = wrapper.get('.app-nav__icon')
    expect(icon.attributes('aria-hidden')).toBe('true')
    expect(icon.find('svg').exists()).toBe(true)
  })

  it('composes extra items after the base "Agenda" entry', async () => {
    const wrapper = await mountNav('/panel', [{ to: { name: 'panel' }, label: 'Servicios' }])
    const links = wrapper.findAll('a')
    expect(links.map((l) => l.text())).toEqual(['Agenda', 'Servicios'])
  })

  it('has no axe violations', async () => {
    const wrapper = await mountNav()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
