/**
 * Pruebas del cliente tipado (`loginApi` sobre `shared/api/httpClient`):
 * mapeo de request/response/problem derivados del bundle OpenAPI real por
 * `status` (nunca por `detail`), y prueba de forma de la solicitud (sin
 * credenciales en la URL, cuerpo exacto del contrato). El recorrido de
 * red real (fetch de verdad contra el API real en local) se prueba en
 * `e2e/acceso.spec.ts`; esta suite fija la frontera con `httpClient` para
 * poder simular cada `status` documentado y uno no documentado (429) sin
 * un servidor real, exactamente igual a como `openapi-fetch` los entrega
 * (`{ data, error, response }`).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const postMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { POST: postMock },
}))

const { login } = await import('../loginApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function problem(status: number, body: Record<string, unknown>, headers: HeadersInit = {}) {
  return {
    data: undefined,
    error: body,
    response: new Response(null, { status, headers }),
  }
}

function problemBody(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    type: '/api/v1/problems/unauthorized',
    title: 'No autorizado',
    status: 401,
    detail: 'correo o contraseña incorrectos',
    instance: 'req-123',
    code: 'unauthorized',
    requestId: 'req-123',
    ...overrides,
  }
}

describe('loginApi.login (cliente tipado)', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('maps a 200 LoginSuccess body to a success outcome with expiresAt', async () => {
    postMock.mockResolvedValueOnce(ok({ expiresAt: '2026-09-12T12:00:00Z' }))

    const outcome = await login({ email: 'barbero@ejemplo.test', password: 'contraseña-real' })

    expect(outcome).toEqual({ kind: 'success', expiresAt: '2026-09-12T12:00:00Z' })
  })

  it('maps 401 to invalid-credentials regardless of Problem.detail wording', async () => {
    postMock.mockResolvedValueOnce(problem(401, problemBody({ detail: 'cualquier motivo' })))

    const outcome = await login({ email: 'inexistente@ejemplo.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'invalid-credentials' })
  })

  it('maps 422 (ValidationProblem) to validation-error', async () => {
    postMock.mockResolvedValueOnce(
      problem(
        422,
        problemBody({
          type: '/api/v1/problems/validation-error',
          title: 'Error de validación',
          status: 422,
          code: 'validation-error',
        }),
      ),
    )

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps 400 (InvalidRequestProblem) to validation-error', async () => {
    postMock.mockResolvedValueOnce(
      problem(
        400,
        problemBody({
          type: '/api/v1/problems/invalid-request',
          title: 'Solicitud inválida',
          status: 400,
          code: 'invalid-request',
        }),
      ),
    )

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps an undocumented 429 to rate-limited and reads Retry-After', async () => {
    postMock.mockResolvedValueOnce(
      problem(
        429,
        problemBody({
          type: '/api/v1/problems/rate-limited',
          title: 'Demasiadas solicitudes',
          status: 429,
          code: 'rate-limited',
        }),
        { 'Retry-After': '120' },
      ),
    )

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'rate-limited', retryAfterSeconds: 120 })
  })

  it('maps a 429 without Retry-After to rate-limited with no retry hint', async () => {
    postMock.mockResolvedValueOnce(
      problem(
        429,
        problemBody({ status: 429, code: 'rate-limited', title: 'Demasiadas solicitudes' }),
      ),
    )

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'rate-limited', retryAfterSeconds: undefined })
  })

  it('maps a 500 (InternalErrorProblem) to unexpected-error and keeps requestId', async () => {
    postMock.mockResolvedValueOnce(
      problem(
        500,
        problemBody({
          type: '/api/v1/problems/internal-error',
          title: 'Error interno',
          status: 500,
          code: 'internal-error',
          detail: undefined,
          requestId: 'req-abc-500',
        }),
      ),
    )

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'unexpected-error', requestId: 'req-abc-500' })
  })

  it('maps a rejected request (no HTTP response at all) to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await login({ email: 'x@y.test', password: 'x' })

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('sends exactly the LoginRequest shape as the body, never in the URL (CA-010-07)', async () => {
    postMock.mockResolvedValueOnce(ok({ expiresAt: '2026-09-12T12:00:00Z' }))

    await login({ email: 'barbero@ejemplo.test', password: 'clave-super-secreta' })

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/login')
    expect(options.body).toEqual({
      email: 'barbero@ejemplo.test',
      password: 'clave-super-secreta',
    })
  })
})
