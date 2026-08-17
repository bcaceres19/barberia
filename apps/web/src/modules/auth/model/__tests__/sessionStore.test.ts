/**
 * Pruebas del coordinador de estado de sesión (HU-012). `fetchSessionContext`
 * se sustituye por un doble de prueba; CA-012-03 (única limpieza/redirección
 * ante varios 401) se prueba aquí a nivel del coordinador, y a nivel HTTP
 * completo en apps/api (cmd/api) y en e2e/panel.spec.ts.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { SessionContextOutcome } from '../sessionContextOutcome'

const fetchSessionContextMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/sessionContextApi', () => ({ fetchSessionContext: fetchSessionContextMock }))

const {
  sessionState,
  ensureBootstrapped,
  resetForFreshLogin,
  retryBootstrap,
  onUnauthorized,
  reportUnauthorized,
} = await import('../sessionStore')

const authenticated: SessionContextOutcome = {
  kind: 'authenticated',
  barbershopId: 'shop-1',
  barbershopName: 'Barbería de prueba',
  expiresAt: new Date(Date.now() + 60_000).toISOString(),
}

describe('sessionStore', () => {
  beforeEach(() => {
    fetchSessionContextMock.mockReset()
    resetForFreshLogin()
  })

  it('starts in checking and moves to authenticated on success', async () => {
    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    await ensureBootstrapped()
    expect(sessionState.bootstrap).toEqual({ status: 'authenticated', ...authenticated })
  })

  it('moves to unauthenticated when the server says so', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'unauthenticated' })
    await ensureBootstrapped()
    expect(sessionState.bootstrap).toEqual({ status: 'unauthenticated' })
  })

  it('moves to connection-lost on a network or unexpected error', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'network-error' })
    await ensureBootstrapped()
    expect(sessionState.bootstrap).toEqual({ status: 'connection-lost' })
  })

  it('does not re-query the server once already resolved (not checking)', async () => {
    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    await ensureBootstrapped()
    await ensureBootstrapped()
    expect(fetchSessionContextMock).toHaveBeenCalledTimes(1)
  })

  it('shares one in-flight request across concurrent callers', async () => {
    let resolveOutcome: (value: SessionContextOutcome) => void = () => {}
    fetchSessionContextMock.mockImplementationOnce(
      () =>
        new Promise<SessionContextOutcome>((resolve) => {
          resolveOutcome = resolve
        }),
    )
    const first = ensureBootstrapped()
    const second = ensureBootstrapped()
    resolveOutcome(authenticated)
    await Promise.all([first, second])
    expect(fetchSessionContextMock).toHaveBeenCalledTimes(1)
  })

  it('resetForFreshLogin forces the next bootstrap to query again', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'unauthenticated' })
    await ensureBootstrapped()
    expect(sessionState.bootstrap.status).toBe('unauthenticated')

    resetForFreshLogin()
    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    await ensureBootstrapped()
    expect(sessionState.bootstrap.status).toBe('authenticated')
    expect(fetchSessionContextMock).toHaveBeenCalledTimes(2)
  })

  it('retryBootstrap always queries again, never reuses a cached connection-lost result', async () => {
    fetchSessionContextMock.mockResolvedValueOnce({ kind: 'network-error' })
    await ensureBootstrapped()
    expect(sessionState.bootstrap.status).toBe('connection-lost')

    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    await retryBootstrap()
    expect(sessionState.bootstrap.status).toBe('authenticated')
  })

  // --- CA-012-03: coordinación única de 401 --------------------------------

  it('reportUnauthorized notifies listeners exactly once for several near-simultaneous 401s', async () => {
    fetchSessionContextMock.mockResolvedValueOnce(authenticated)
    await ensureBootstrapped() // arranca autenticado, como si ya hubiera sesión

    let notifications = 0
    const unsubscribe = onUnauthorized(() => {
      notifications += 1
    })

    reportUnauthorized()
    reportUnauthorized()
    reportUnauthorized()

    expect(notifications).toBe(1)
    expect(sessionState.bootstrap).toEqual({ status: 'unauthenticated' })
    unsubscribe()
  })

  it('reportUnauthorized is a no-op once already unauthenticated', () => {
    let notifications = 0
    const unsubscribe = onUnauthorized(() => {
      notifications += 1
    })

    reportUnauthorized() // checking -> unauthenticated: primera notificación
    reportUnauthorized() // ya unauthenticated: no debe notificar de nuevo

    expect(notifications).toBe(1)
    unsubscribe()
  })

  it('onUnauthorized unsubscribe stops future notifications', () => {
    let notifications = 0
    const unsubscribe = onUnauthorized(() => {
      notifications += 1
    })
    unsubscribe()

    reportUnauthorized()

    expect(notifications).toBe(0)
  })
})
