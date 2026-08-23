/**
 * Pruebas del cliente tipado de recuperación (HU-008, DEC-064/DEC-065):
 * mapeo de request/response por `status`, mismo patrón que
 * `challengeApi.test.ts`.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const postMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { POST: postMock },
}))

const { requestRecovery, verifyRecovery, resetRecoveryPassword } = await import('../recoveryApi')

function ok(status: number, body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status }) }
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
      detail: 'motivo interno, nunca debe determinar el mapeo',
      instance: 'req-123',
      code: 'unauthorized',
      requestId: 'req-123',
      ...body,
    },
    response: new Response(null, { status }),
  }
}

describe('recoveryApi.requestRecovery', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 202 to accepted, regardless of body content', async () => {
    postMock.mockResolvedValueOnce(ok(202, { message: 'genérico' }))

    const outcome = await requestRecovery('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'accepted' })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await requestRecovery('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('maps an unexpected status to unexpected-error and keeps requestId', async () => {
    postMock.mockResolvedValueOnce(problem(500, { requestId: 'req-500' }))

    const outcome = await requestRecovery('barbero@ejemplo.test')

    expect(outcome).toEqual({ kind: 'unexpected-error', requestId: 'req-500' })
  })

  it('sends exactly the email as the body', async () => {
    postMock.mockResolvedValueOnce(ok(202, { message: 'x' }))

    await requestRecovery('barbero@ejemplo.test')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/recovery/request')
    expect(options.body).toEqual({ email: 'barbero@ejemplo.test' })
  })
})

describe('recoveryApi.verifyRecovery', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 200 to verified and carries resetToken/maskedPhone/maskedEmail', async () => {
    postMock.mockResolvedValueOnce(
      ok(200, {
        resetToken: 'token-abc',
        maskedPhone: '+57 *** *** 12',
        maskedEmail: 'b***@c***.test',
      }),
    )

    const outcome = await verifyRecovery('barbero@ejemplo.test', '482913')

    expect(outcome).toEqual({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    })
  })

  it('maps a 401 to invalid-code regardless of Problem.detail wording (incorrect, expired or exhausted are indistinguishable, DEC-064/DEC-065)', async () => {
    postMock.mockResolvedValueOnce(problem(401))

    const outcome = await verifyRecovery('barbero@ejemplo.test', '000000')

    expect(outcome).toEqual({ kind: 'invalid-code' })
  })

  it('maps 400/422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422))

    const outcome = await verifyRecovery('barbero@ejemplo.test', '12')

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await verifyRecovery('barbero@ejemplo.test', '482913')

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('sends exactly email and code as the body', async () => {
    postMock.mockResolvedValueOnce(ok(200, { resetToken: 't', maskedPhone: 'p', maskedEmail: 'e' }))

    await verifyRecovery('barbero@ejemplo.test', '482913')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/recovery/verify')
    expect(options.body).toEqual({ email: 'barbero@ejemplo.test', code: '482913' })
  })
})

describe('recoveryApi.resetRecoveryPassword', () => {
  beforeEach(() => postMock.mockReset())

  it('maps a 204 to success', async () => {
    postMock.mockResolvedValueOnce(noContent())

    const outcome = await resetRecoveryPassword(
      'barbero@ejemplo.test',
      'token-abc',
      'contraseña-nueva-valida',
    )

    expect(outcome).toEqual({ kind: 'success' })
  })

  it('maps a 401 to invalid-token (unknown, expired, consumed or wrong-account token)', async () => {
    postMock.mockResolvedValueOnce(problem(401))

    const outcome = await resetRecoveryPassword('barbero@ejemplo.test', 'token-vencido', 'x')

    expect(outcome).toEqual({ kind: 'invalid-token' })
  })

  it('maps a 422 to policy-violation', async () => {
    postMock.mockResolvedValueOnce(problem(422))

    const outcome = await resetRecoveryPassword('barbero@ejemplo.test', 'token-abc', 'corto')

    expect(outcome).toEqual({ kind: 'policy-violation' })
  })

  it('maps a 400 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(400))

    const outcome = await resetRecoveryPassword('barbero@ejemplo.test', '', '')

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a rejected request to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await resetRecoveryPassword(
      'barbero@ejemplo.test',
      'token-abc',
      'contraseña-nueva-valida',
    )

    expect(outcome).toEqual({ kind: 'network-error' })
  })

  it('sends exactly email, resetToken and newPassword as the body', async () => {
    postMock.mockResolvedValueOnce(noContent())

    await resetRecoveryPassword('barbero@ejemplo.test', 'token-abc', 'contraseña-nueva-valida')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [path, options] = postMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/public/auth/recovery/reset-password')
    expect(options.body).toEqual({
      email: 'barbero@ejemplo.test',
      resetToken: 'token-abc',
      newPassword: 'contraseña-nueva-valida',
    })
  })
})
