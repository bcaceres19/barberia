/**
 * Pruebas de PrivateShell (HU-012): estados globales de arranque/carga,
 * conexión perdida con Reintentar y composición de cabecera+navegación+
 * contenido solo cuando hay sesión autenticada. `sessionStore` se
 * sustituye por un doble de prueba controlable.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'
import type { BootstrapState } from '../../model/sessionStore'

const mockState = vi.hoisted(() => ({ bootstrap: { status: 'checking' } as BootstrapState }))
const retryBootstrapMock = vi.hoisted(() => vi.fn())

vi.mock('../../model/sessionStore', () => ({
  sessionState: mockState,
  retryBootstrap: retryBootstrapMock,
}))
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: vi.fn(), POST: vi.fn() } }))

const { default: PrivateShell } = await import('../PrivateShell.vue')

async function mountShell() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/panel', name: 'panel', component: { template: '<p>contenido del panel</p>' } },
      { path: '/acceso', name: 'acceso', component: { template: '<div />' } },
      {
        path: '/panel/nuevo-turno',
        name: 'agenda-nuevo-turno',
        component: { template: '<div />' },
      },
    ],
  })
  await router.push('/panel')
  await router.isReady()
  return mount(PrivateShell, { global: { plugins: [router] } })
}

describe('PrivateShell', () => {
  beforeEach(() => {
    retryBootstrapMock.mockReset()
  })

  it('shows a non-blank checking state without header or nav', async () => {
    mockState.bootstrap = { status: 'checking' }
    const wrapper = await mountShell()

    expect(wrapper.text().length).toBeGreaterThan(0)
    expect(wrapper.find('header').exists()).toBe(false)
    expect(wrapper.find('nav').exists()).toBe(false)
  })

  it('renders header, nav and content when authenticated (CA-012-04)', async () => {
    mockState.bootstrap = {
      status: 'authenticated',
      barbershopId: 'shop-1',
      barbershopName: 'Barbería El Corte',
      expiresAt: new Date().toISOString(),
    }
    const wrapper = await mountShell()

    expect(wrapper.find('header').exists()).toBe(true)
    expect(wrapper.find('nav').exists()).toBe(true)
    expect(wrapper.text()).toContain('Barbería El Corte')
    expect(wrapper.text()).toContain('contenido del panel')
  })

  it('shows a recoverable connection-lost state with Reintentar (CA-012-05)', async () => {
    mockState.bootstrap = { status: 'connection-lost' }
    const wrapper = await mountShell()

    expect(wrapper.text()).toContain('No pudimos conectar')
    await wrapper.get('button').trigger('click')
    expect(retryBootstrapMock).toHaveBeenCalledTimes(1)
  })

  it('never shows a blank screen for unauthenticated (transient, guard already redirects)', async () => {
    mockState.bootstrap = { status: 'unauthenticated' }
    const wrapper = await mountShell()

    expect(wrapper.html().trim().length).toBeGreaterThan(0)
  })

  it('has no axe violations in the authenticated state', async () => {
    mockState.bootstrap = {
      status: 'authenticated',
      barbershopId: 'shop-1',
      barbershopName: 'Barbería El Corte',
      expiresAt: new Date().toISOString(),
    }
    const wrapper = await mountShell()
    const results = await axe(wrapper.element, { rules: { region: { enabled: false } } })
    expect(results).toHaveNoViolations()
  })
})
