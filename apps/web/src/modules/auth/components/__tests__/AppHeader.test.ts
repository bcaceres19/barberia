/**
 * Pruebas de AppHeader (HU-012): la barbería activa siempre visible
 * (CA-012-04) y el cierre de sesión (CA-012-07): deshabilitado durante el
 * envío, evita doble envío, limpia el estado local y navega a acceso
 * incluso si el servidor responde con un fallo ambiguo.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'

const postMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: vi.fn(), POST: postMock },
}))

const { default: AppHeader } = await import('../AppHeader.vue')
const { sessionState, resetForFreshLogin } = await import('../../model/sessionStore')

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/acceso', name: 'acceso', component: { template: '<div>acceso</div>' } },
      { path: '/panel', name: 'panel', component: { template: '<div>panel</div>' } },
      {
        path: '/panel/nuevo-turno',
        name: 'agenda-nuevo-turno',
        component: { template: '<div>nuevo turno</div>' },
      },
    ],
  })
}

async function mountHeader(barbershopName = 'Barbería de prueba') {
  const router = buildRouter()
  await router.push({ name: 'panel' })
  await router.isReady()
  const wrapper = mount(AppHeader, {
    props: { barbershopName },
    global: { plugins: [router] },
  })
  return { wrapper, router }
}

describe('AppHeader', () => {
  beforeEach(() => {
    postMock.mockReset()
    resetForFreshLogin()
  })

  it('always shows the active barbershop name (CA-012-04)', async () => {
    const { wrapper } = await mountHeader('Barbería El Corte')
    expect(wrapper.get('[data-testid="barbershop-name"]').text()).toBe('Barbería El Corte')
  })

  it('shows the NAVA wordmark and a global "Nuevo turno" action linking to the manual creation route', async () => {
    const { wrapper } = await mountHeader()
    expect(wrapper.text()).toContain('NAVA')
    const cta = wrapper.get('a[href="/panel/nuevo-turno"]')
    expect(cta.text()).toBe('Nuevo turno')
  })

  it('logs out, revokes the server session and navigates to acceso (CA-012-07)', async () => {
    postMock.mockResolvedValueOnce({ response: new Response(null, { status: 204 }) })
    const { wrapper, router } = await mountHeader()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(postMock).toHaveBeenCalledWith('/private/auth/logout')
    expect(router.currentRoute.value.name).toBe('acceso')
    expect(sessionState.bootstrap.status).toBe('checking')
  })

  it('disables the control during the request and prevents a second submission', async () => {
    let resolveLogout: (v: unknown) => void = () => {}
    postMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveLogout = resolve
        }),
    )
    const { wrapper } = await mountHeader()

    const button = wrapper.get('button')
    await button.trigger('click')
    expect(button.attributes('disabled')).toBeDefined()

    await button.trigger('click')
    expect(postMock).toHaveBeenCalledTimes(1)

    resolveLogout({ response: new Response(null, { status: 204 }) })
    await flushPromises()
  })

  it('still clears local state and navigates away on an ambiguous network failure', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const { wrapper, router } = await mountHeader()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('has no axe violations', async () => {
    const { wrapper } = await mountHeader()
    const results = await axe(wrapper.element, { rules: { region: { enabled: false } } })
    expect(results).toHaveNoViolations()
  })
})
