/**
 * Pruebas de installSessionHandling (HU-012, trabajo requerido §5):
 * conecta el middleware de respuesta del cliente HTTP con la coordinación
 * única de 401 del store real, y con un router real (memoria) para
 * confirmar la redirección y la ausencia de bucle.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { flushPromises } from '@vue/test-utils'

const useMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { use: useMock } }))

const { installSessionHandling } = await import('../installSessionHandling')
const { resetForFreshLogin, sessionState } = await import('../../model/sessionStore')

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/acceso', name: 'acceso', component: { template: '<div />' } },
      { path: '/panel', name: 'panel', component: { template: '<div />' } },
    ],
  })
}

function capturedOnResponse() {
  const middleware = useMock.mock.calls[0]?.[0] as {
    onResponse: (args: { request: Request; response: Response }) => unknown
  }
  return middleware.onResponse
}

describe('installSessionHandling', () => {
  beforeEach(async () => {
    useMock.mockReset()
    resetForFreshLogin()
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    installSessionHandling(router)
  })

  it('registers exactly one response middleware on the shared http client', () => {
    expect(useMock).toHaveBeenCalledTimes(1)
    expect(typeof capturedOnResponse()).toBe('function')
  })

  it('redirects to acceso on a 401 from a private route (not the bootstrap call)', async () => {
    const router = buildRouter()
    await router.push('/panel')
    await router.isReady()
    useMock.mockClear()
    installSessionHandling(router)
    const onResponse = capturedOnResponse()

    onResponse({
      request: new Request('http://localhost/api/v1/private/auth/logout'),
      response: new Response(null, { status: 401 }),
    })
    await flushPromises()

    expect(sessionState.bootstrap).toEqual({ status: 'unauthenticated' })
    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('never redirects again once already on acceso (no cycle)', async () => {
    const router = buildRouter()
    await router.push({ name: 'acceso' })
    await router.isReady()
    useMock.mockClear()
    installSessionHandling(router)
    const onResponse = capturedOnResponse()
    const pushSpy = vi.spyOn(router, 'push')

    onResponse({
      request: new Request('http://localhost/api/v1/private/auth/logout'),
      response: new Response(null, { status: 401 }),
    })

    expect(pushSpy).not.toHaveBeenCalled()
  })

  it('ignores a 401 from the session-context bootstrap call (the guard already handles it)', () => {
    resetForFreshLogin()
    const onResponse = capturedOnResponse()

    onResponse({
      request: new Request('http://localhost/api/v1/private/auth/session'),
      response: new Response(null, { status: 401 }),
    })

    // El estado se queda en `checking`: este flujo no interviene, la propia
    // llamada de ensureBootstrapped es quien decide.
    expect(sessionState.bootstrap).toEqual({ status: 'checking' })
  })

  it('ignores a 401 from a public route (e.g. wrong login password)', () => {
    resetForFreshLogin()
    const onResponse = capturedOnResponse()

    onResponse({
      request: new Request('http://localhost/api/v1/public/auth/login'),
      response: new Response(null, { status: 401 }),
    })

    expect(sessionState.bootstrap).toEqual({ status: 'checking' })
  })

  it('ignores a non-401 response from a private route', () => {
    resetForFreshLogin()
    const onResponse = capturedOnResponse()

    onResponse({
      request: new Request('http://localhost/api/v1/private/auth/logout'),
      response: new Response(null, { status: 204 }),
    })

    expect(sessionState.bootstrap).toEqual({ status: 'checking' })
  })
})
