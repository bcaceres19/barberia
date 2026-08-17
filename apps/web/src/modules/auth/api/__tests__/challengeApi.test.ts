/**
 * Pruebas del cliente tipado del reto telefónico (HU-007, DEC-062): mapeo
 * de request/response por `status`, mismo patrón que `loginApi.test.ts`.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const postMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { POST: postMock },
}))

const { requestChallenge, verifyChallenge } = await import('../challengeApi')

function accepted(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 202 }) }
}

function noContent() {
  return { data: undefined, error: undefined, response: new Response(null, { status: 204 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return {
    data: undefined,
    error: {
      type: '/api/v1/problems/unauthorized',
      title: 'No autorizado',
      status,
      detail: 'código de verificación inválido o expirado',
      instance: 'req-123',
      code: 'unauthorized',
      requestId: 'req-123',
      ...body,
    },
    response: new Response(null, { status }),
  }
}

describe('challengeApi.requestChallenge', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 202 to accepted, regardless of body content', async () => {
    postMock.mockResolvedValueOnce(accepted({ message: 'genérico' }))

    const outcome = await requestChallenge('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'accepted' })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await requestChallenge('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('maps an unexpected status to unexpected-error and keeps requestId', async () => {
    postMock.mockResolvedValueOnce(problem(500, { requestId: 'req-500' }))

    const outcome = await requestChallenge('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'unexpected-error', requestId: 'req-500' })
  })

  it('sends exactly the email as the body', async () => {
    postMock.mockResolvedValueOnce(accepted({ message: 'x' }))

    await requestChallenge('barbero@ejemplo.test')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/challenge')
    expect(options.body).toEqual({ email: 'barbero@ejemplo.test' })
  })
})

describe('challengeApi.verifyChallenge', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 204 to verified', async () => {
    postMock.mockResolvedValueOnce(noContent())

    const outcome = await verifyChallenge('barbero@ejemplo.test', '482913')

    expect(outcome).toEqual({ kind: 'verified' })
  })

  it('maps a 401 to invalid-code regardless of Problem.detail wording', async () => {
    postMock.mockResolvedValueOnce(problem(401))

    const outcome = await verifyChallenge('barbero@ejemplo.test', '000000')

    expect(outcome).toEqual({ kind: 'invalid-code' })
  })

  it('maps 400/422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422))

    const outcome = await verifyChallenge('barbero@ejemplo.test', '12')

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await verifyChallenge('barbero@ejemplo.test', '482913')

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('sends exactly email and code as the body', async () => {
    postMock.mockResolvedValueOnce(noContent())

    await verifyChallenge('barbero@ejemplo.test', '482913')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/challenge/verify')
    expect(options.body).toEqual({ email: 'barbero@ejemplo.test', code: '482913' })
  })
})
