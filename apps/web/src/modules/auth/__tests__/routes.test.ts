/**
 * Prueba de integración de enrutamiento (DEC-056): sin sesión recordada,
 * `/panel` redirige a `/acceso`; con sesión recordada (vigente), navega y
 * renderiza el marcador de posición autenticado mínimo, sin cabecera ni
 * navegación general. `/recuperar-acceso` es un destino real, no roto
 * (CA-010-08, `DP-UX-06`).
 */
import { describe, it, expect, beforeEach } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { authRoutes } from '../index'
import { rememberSessionUntil } from '../model/sessionMarker'

function buildRouter() {
  return createRouter({ history: createMemoryHistory(), routes: authRoutes })
}

describe('authRoutes', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
  })

  it('registers /acceso, /panel and /recuperar-acceso', () => {
    const paths = authRoutes.map((route) => route.path)
    expect(paths).toEqual(expect.arrayContaining(['/acceso', '/panel', '/recuperar-acceso']))
  })

  it('redirects /panel to /acceso when there is no remembered session', async () => {
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('allows /panel when a valid session is remembered', async () => {
    rememberSessionUntil(new Date(Date.now() + 60_000).toISOString())
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('redirects /panel to /acceso when the remembered session already expired', async () => {
    rememberSessionUntil(new Date(Date.now() - 60_000).toISOString())
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('exposes a real, matched destination for /recuperar-acceso (no dead link)', async () => {
    const router = buildRouter()
    await router.push('/recuperar-acceso')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('recuperar-acceso')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
  })
})
