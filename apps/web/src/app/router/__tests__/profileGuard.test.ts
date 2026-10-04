import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import { DEFAULT_BRAND, resetBrand, setBrand } from '@/shared/model'
import { installProfileRedirect, profileRedirect } from '../profileGuard'

describe('profileRedirect (DEC-115)', () => {
  it('sends «Servicios por barbero» to «Servicios» for the solo profile', () => {
    expect(profileRedirect({ name: 'barber-services' }, 'solo')).toEqual({
      name: 'catalog-servicios',
    })
  })

  it('leaves every route alone in the full panel', () => {
    expect(profileRedirect({ name: 'barber-services' }, 'shop')).toBeNull()
    expect(profileRedirect({ name: 'staff-barberos' }, 'shop')).toBeNull()
  })

  it.each([
    'panel',
    'catalog-servicios',
    'staff-barberos',
    'schedules-horarios',
    'configuracion-barberia',
  ])('keeps «%s» available in the solo profile', (name) => {
    expect(profileRedirect({ name }, 'solo')).toBeNull()
  })

  it('ignores unnamed routes', () => {
    expect(profileRedirect({ name: undefined }, 'solo')).toBeNull()
    expect(profileRedirect({ name: Symbol('x') }, 'solo')).toBeNull()
  })
})

describe('installProfileRedirect (DEC-115)', () => {
  async function routerAt(path: string) {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/panel/servicios', name: 'catalog-servicios', component: { template: '<div />' } },
        {
          path: '/panel/servicios-por-barbero',
          name: 'barber-services',
          component: { template: '<div />' },
        },
      ],
    })
    installProfileRedirect(router)
    await router.push(path)
    await router.isReady()
    return router
  }

  beforeEach(() => resetBrand())

  it('redirects on navigation when the profile is already known', async () => {
    setBrand({ ...DEFAULT_BRAND, panelProfile: 'solo' })

    const router = await routerAt('/panel/servicios-por-barbero')

    expect(router.currentRoute.value.name).toBe('catalog-servicios')
  })

  it('does not touch the full panel', async () => {
    const router = await routerAt('/panel/servicios-por-barbero')

    expect(router.currentRoute.value.name).toBe('barber-services')
  })

  it('moves a visitor who is already on a retired route when the profile arrives later', async () => {
    // Primera visita sin copia local: la marca llega con la sesión, después de la ruta.
    const router = await routerAt('/panel/servicios-por-barbero')
    expect(router.currentRoute.value.name).toBe('barber-services')

    setBrand({ ...DEFAULT_BRAND, panelProfile: 'solo' })
    await vi.waitFor(() => expect(router.currentRoute.value.name).toBe('catalog-servicios'))
  })

  it('also corrects the route when the profile arrives while the first navigation is still pending', async () => {
    // La sesión se comprueba antes de mostrar la ruta; la marca llega en ese lapso.
    let finishSession: () => void = () => {}
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/panel/servicios', name: 'catalog-servicios', component: { template: '<div />' } },
        {
          path: '/panel/servicios-por-barbero',
          name: 'barber-services',
          component: { template: '<div />' },
          beforeEnter: () => new Promise<void>((resolve) => (finishSession = resolve)),
        },
      ],
    })
    installProfileRedirect(router)
    const first = router.push('/panel/servicios-por-barbero')

    setBrand({ ...DEFAULT_BRAND, panelProfile: 'solo' })
    finishSession()
    await first

    await vi.waitFor(() => expect(router.currentRoute.value.name).toBe('catalog-servicios'))
  })

  it('leaves a visitor on an available route alone when the profile arrives', async () => {
    const router = await routerAt('/panel/servicios')

    setBrand({ ...DEFAULT_BRAND, panelProfile: 'solo' })
    await Promise.resolve()

    expect(router.currentRoute.value.name).toBe('catalog-servicios')
  })
})
