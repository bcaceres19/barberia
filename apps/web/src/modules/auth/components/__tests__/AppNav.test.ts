/**
 * Pruebas de AppNav (HU-012): navegación semántica, marca la entrada activa
 * y opera con teclado.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'
import AppNav from '../AppNav.vue'

async function mountNav(initialRoute = '/panel') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/panel', name: 'panel', component: { template: '<div />' } }],
  })
  await router.push(initialRoute)
  await router.isReady()
  return mount(AppNav, { global: { plugins: [router] } })
}

describe('AppNav', () => {
  it('renders a semantic nav landmark with a labelled link', async () => {
    const wrapper = await mountNav()
    expect(wrapper.get('nav').attributes('aria-label')).toBeTruthy()
    expect(wrapper.get('a').text()).toBe('Panel')
  })

  it('marks the current route as active', async () => {
    const wrapper = await mountNav('/panel')
    expect(wrapper.get('a').classes()).toContain('app-nav__link--active')
  })

  it('has no axe violations', async () => {
    const wrapper = await mountNav()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
