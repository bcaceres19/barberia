/**
 * Prueba de integración de enrutamiento (HU-090): `/reservar/:slug` es una
 * ruta pública real, matched, sin ningún guard de sesión (CA-090-03) y sin
 * depender del cascarón privado.
 */
import { describe, it, expect } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { publicBookingRoutes } from '../index'

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [...publicBookingRoutes],
  })
}

describe('publicBookingRoutes', () => {
  it('registers /reservar/:slug', () => {
    const paths = publicBookingRoutes.map((route) => route.path)
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
    const paths = publicBookingRoutes.map((route) => route.path)
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
    const paths = publicBookingRoutes.map((route) => route.path)
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
})
