/**
 * Prueba de integración de enrutamiento (HU-012, DEC-060): sin sesión real,
 * cualquier ruta privada redirige a `/acceso` conservando el destino
 * pretendido (CA-012-02); con sesión vigente, entra directamente
 * (CA-012-01). `fetchSessionContext` se sustituye por un doble de prueba:
 * el recorrido real contra el API real vive en `e2e/panel.spec.ts`.
 * `/recuperar-acceso` es el flujo real de HU-011 (CA-010-08, antes
 * provisional bajo `DP-UX-06`).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { authRoutes } from '../index'
import { resetForFreshLogin } from '../model/sessionStore'
import type { SessionContextOutcome } from '../model/sessionContextOutcome'

const fetchSessionContextMock = vi.hoisted(() => vi.fn())
vi.mock('../api/sessionContextApi', () => ({ fetchSessionContext: fetchSessionContextMock }))

function buildRouter() {
  return createRouter({ history: createMemoryHistory(), routes: authRoutes })
}

const authenticated: SessionContextOutcome = {
  kind: 'authenticated',
  barbershopId: 'shop-1',
  barbershopName: 'Barbería de prueba',
  expiresAt: new Date(Date.now() + 60_000).toISOString(),
}

describe('authRoutes', () => {
  beforeEach(() => {
    fetchSessionContextMock.mockReset()
    // El store es un singleton compartido entre pruebas del mismo archivo:
    // lo devuelve a `checking` para que cada guard vuelva a consultar el
    // doble de prueba en vez de reutilizar el resultado de la prueba
    // anterior.
    resetForFreshLogin()
  })

  it('registers /acceso, /panel and /recuperar-acceso', () => {
    const paths = authRoutes.map((route) => route.path)
    expect(paths).toEqual(expect.arrayContaining(['/acceso', '/panel', '/recuperar-acceso']))
  })

  it('redirects /panel to /acceso when there is no real session (CA-012-02)', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'unauthenticated' })
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('preserves the intended destination as a query param (CA-012-02)', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'unauthenticated' })
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.query.redirect).toBe('/panel')
  })

  it('allows /panel when the real session is valid (CA-012-01)', async () => {
    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('reaches the private shell (connection-lost state) instead of bouncing when the server is unreachable (CA-012-05)', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'network-error' })
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    // No redirige a /acceso: el cascarón renderiza el estado recuperable
    // en la propia ruta privada, sin bucle de redirección.
    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('exposes a real, matched destination for /recuperar-acceso (no dead link)', async () => {
    const router = buildRouter()
    await router.push('/recuperar-acceso')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('recuperar-acceso')
    expect(router.currentRoute.value.matched.length).toBeGreaterThan(0)
  })
})
