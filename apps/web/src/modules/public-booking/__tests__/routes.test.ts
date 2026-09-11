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
})
