/**
 * Prueba de integración de enrutamiento (HU-090): `/reservar/:slug` es una
 * ruta pública real, matched, sin ningún guard de sesión (CA-090-03) y sin
 * depender del cascarón privado.
 */
import { describe, it, expect } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { publicBookingRoutes } from '../index'

// Las cinco pantallas son hijas de un único cascarón (`PublicBookingLayout`):
// la ruta raíz del registro es la del cascarón y cada pantalla conserva su
// ruta absoluta. Los `paths` registrados son los del cascarón más los de sus
// hijas.
const registeredPaths = [
  ...publicBookingRoutes.map((route) => route.path),
  ...publicBookingRoutes.flatMap((route) => (route.children ?? []).map((child) => child.path)),
]

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [...publicBookingRoutes],
  })
}

describe('publicBookingRoutes', () => {
  it('hosts every screen under one shared shell, with the step in the route meta', async () => {
    expect(publicBookingRoutes).toHaveLength(1)
    const router = buildRouter()
    const expected: Array<[string, number | undefined]> = [
      ['/reservar/a', undefined],
      ['/reservar/a/servicios', 1],
      ['/reservar/a/servicios/s/barbero', 2],
      ['/reservar/a/servicios/s/barbero/b/horario', 3],
      ['/reservar/a/servicios/s/barbero/b/horario/2026-09-15T19:00:00Z/cliente', 4],
    ]
    for (const [path, step] of expected) {
      await router.push(path)
      const current = router.currentRoute.value
      // Cascarón + pantalla: dos registros coincidentes en cada ruta.
      expect(current.matched).toHaveLength(2)
      expect(current.meta.bookingStep).toBe(step)
    }
  })

  it('registers /reservar/:slug', () => {
    const paths = registeredPaths
    expect(paths).toContain('/reservar/:slug')
  })

  it('resolves a real, matched destination without any session guard', async () => {
    const router = buildRouter()
    await router.push('/reservar/barberia-ejemplo')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('reserva-publica-entrada')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
  })

  it('passes the route param as a component prop (props: true)', async () => {
    const router = buildRouter()
    await router.push('/reservar/barberia-ejemplo')
    await router.isReady()
    expect(router.currentRoute.value.params.slug).toBe('barberia-ejemplo')
  })

  it('registers /reservar/:slug/servicios/:serviceId/barbero (HU-092)', () => {
    const paths = registeredPaths
    expect(paths).toContain('/reservar/:slug/servicios/:serviceId/barbero')
  })

  it('resolves the barber selection route with both params as props, without any session guard', async () => {
    const router = buildRouter()
    await router.push(
      '/reservar/barberia-ejemplo/servicios/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4/barbero',
    )
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('reserva-publica-barbero')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
    expect(router.currentRoute.value.params.slug).toBe('barberia-ejemplo')
    expect(router.currentRoute.value.params.serviceId).toBe('8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4')
  })

  it('registers /reservar/:slug/servicios/:serviceId/barbero/:barberId/horario (HU-095)', () => {
    const paths = registeredPaths
    expect(paths).toContain('/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario')
  })

  it('resolves the availability route with the three params as props, without any session guard', async () => {
    const router = buildRouter()
    await router.push(
      '/reservar/barberia-ejemplo/servicios/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4/barbero/a1111111-1111-1111-1111-111111111111/horario',
    )
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('reserva-publica-horario')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
    expect(router.currentRoute.value.params.slug).toBe('barberia-ejemplo')
    expect(router.currentRoute.value.params.serviceId).toBe('8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4')
    expect(router.currentRoute.value.params.barberId).toBe('a1111111-1111-1111-1111-111111111111')
  })

  it('registers /reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente (HU-096)', () => {
    const paths = registeredPaths
    expect(paths).toContain(
      '/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente',
    )
  })

  it('resolves the customer details route with the four params as props, without any session guard', async () => {
    const router = buildRouter()
    await router.push(
      '/reservar/barberia-ejemplo/servicios/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4/barbero/a1111111-1111-1111-1111-111111111111/horario/2026-09-20T14%3A30%3A00Z/cliente',
    )
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('reserva-publica-cliente')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
    expect(router.currentRoute.value.params.slug).toBe('barberia-ejemplo')
    expect(router.currentRoute.value.params.serviceId).toBe('8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4')
    expect(router.currentRoute.value.params.barberId).toBe('a1111111-1111-1111-1111-111111111111')
    expect(router.currentRoute.value.params.startsAt).toBe('2026-09-20T14:30:00Z')
  })
})
