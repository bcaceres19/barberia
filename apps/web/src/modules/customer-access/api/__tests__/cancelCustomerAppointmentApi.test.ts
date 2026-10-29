import { beforeEach, describe, expect, it, vi } from 'vitest'

const postMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: vi.fn(), POST: postMock } }))

const { cancelCustomerAppointment } = await import('../cancelCustomerAppointmentApi')

function respond(status: number, error?: unknown) {
  postMock.mockResolvedValueOnce({
    data: status === 200 ? { status: 'cancelled_by_customer' } : undefined,
    error,
    response: { ok: status >= 200 && status < 300, status } as Response,
  })
}

const problem = (status: number, code: string) => ({
  type: `/api/v1/problems/${code}`,
  title: 'x',
  status,
  code,
  requestId: 'req-1',
})

beforeEach(() => postMock.mockReset())

describe('cancelCustomerAppointment', () => {
  it('envía token en la ruta, Idempotency-Key e If-Match en cabeceras y el motivo en el cuerpo', async () => {
    respond(200)
    const outcome = await cancelCustomerAppointment('tok', 'v-1', 'key-1', { reason: 'no puedo' })

    expect(outcome).toEqual({ kind: 'success' })
    expect(postMock).toHaveBeenCalledWith('/customer/appointments/{token}/cancel', {
      params: {
        path: { token: 'tok' },
        header: { 'Idempotency-Key': 'key-1', 'If-Match': 'v-1' },
      },
      body: { reason: 'no puedo' },
    })
  })

  it.each([
    [422, 'cancellation-window-closed', 'window-closed'],
    [422, 'cancellation-reason-required', 'reason-required'],
    [422, 'validation-error', 'validation-error'],
    [400, 'invalid-request', 'validation-error'],
    [409, 'version-conflict', 'changed'],
    [409, 'invalid-state', 'changed'],
    [409, 'idempotency-conflict', 'changed'],
    [409, 'idempotency-locked', 'idempotency-in-progress'],
    [404, 'not-found', 'not-found'],
  ])('mapea %i/%s a %s por status y code, nunca por detail', async (status, code, kind) => {
    respond(status, problem(status, code))
    expect((await cancelCustomerAppointment('tok', 'v', 'k', { reason: null })).kind).toBe(kind)
  })

  it('un 500 es inesperado y conserva el requestId para soporte', async () => {
    respond(500, problem(500, 'internal-error'))
    expect(await cancelCustomerAppointment('tok', 'v', 'k', { reason: null })).toEqual({
      kind: 'unexpected-error',
      requestId: 'req-1',
    })
  })

  it('un fallo de fetch es error de red', async () => {
    postMock.mockRejectedValueOnce(new TypeError('fetch failed'))
    expect(await cancelCustomerAppointment('tok', 'v', 'k', { reason: null })).toEqual({
      kind: 'network-error',
    })
  })
})
