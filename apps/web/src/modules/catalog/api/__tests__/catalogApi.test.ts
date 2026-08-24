/**
 * Pruebas del cliente tipado (`catalogApi` sobre `shared/api/httpClient`):
 * mapeo por `status`/`code` (nunca por `detail`), mismo criterio que
 * `staff/api/__tests__/staffApi.test.ts`. El recorrido de red real vive en
 * un E2E propio de HU-022.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const postMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: postMock, PATCH: patchMock },
}))

const { fetchServices, createService, updateService } = await import('../catalogApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function created(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 201 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return { data: undefined, error: body, response: new Response(null, { status }) }
}

const serviceBody = {
  id: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  name: 'Corte clásico',
  description: 'Corte con máquina y tijera, incluye lavado.',
  durationMinutes: 30,
  price: '45000.00',
  currency: 'COP',
  createdAt: '2026-08-24T15:04:05Z',
  updatedAt: '2026-08-24T15:04:05Z',
}

describe('catalogApi.fetchServices', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a success outcome with items and nextCursor', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [serviceBody], nextCursor: null }))

    const outcome = await fetchServices()

    expect(outcome).toEqual({ kind: 'success', page: { items: [serviceBody], nextCursor: null } })
  })

  it('forwards the cursor as a query parameter when provided', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [], nextCursor: null }))

    await fetchServices('opaque-cursor')

    expect(getMock).toHaveBeenCalledWith('/private/services', {
      params: { query: { cursor: 'opaque-cursor' } },
    })
  })

  it('sends no cursor query parameter for the first page', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [], nextCursor: null }))

    await fetchServices()

    expect(getMock).toHaveBeenCalledWith('/private/services', { params: { query: {} } })
  })

  it('maps any non-2xx status to unexpected-error (401 handled globally)', async () => {
    getMock.mockResolvedValueOnce(problem(401))

    const outcome = await fetchServices()

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await fetchServices()

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('catalogApi.createService', () => {
  beforeEach(() => postMock.mockReset())

  it('sends name/durationMinutes/price and the Idempotency-Key header, omitting an empty description', async () => {
    postMock.mockResolvedValueOnce(created(serviceBody))

    await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(postMock).toHaveBeenCalledWith('/private/services', {
      params: { header: { 'Idempotency-Key': 'idem-key-1' } },
      body: { name: 'Corte clásico', durationMinutes: 30, price: '45000.00' },
    })
  })

  it('includes a trimmed description when present', async () => {
    postMock.mockResolvedValueOnce(created(serviceBody))

    await createService(
      { name: 'Corte clásico', description: '  Con lavado  ', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(postMock).toHaveBeenCalledWith('/private/services', {
      params: { header: { 'Idempotency-Key': 'idem-key-1' } },
      body: { name: 'Corte clásico', description: 'Con lavado', durationMinutes: 30, price: '45000.00' },
    })
  })

  it('maps a 201 success body to a success outcome', async () => {
    postMock.mockResolvedValueOnce(created(serviceBody))

    const outcome = await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(outcome).toEqual({ kind: 'success', service: serviceBody })
  })

  it('maps 422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422, { code: 'validation-error' }))

    const outcome = await createService(
      { name: '', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps 409 with code=conflict to name-conflict (DEC-067)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'conflict' }))

    const outcome = await createService(
      { name: 'Repetido', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(outcome).toEqual({ kind: 'name-conflict' })
  })

  it('maps 409 with code=idempotency-conflict to idempotency-conflict', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'idempotency-conflict' }))

    const outcome = await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'reused-key',
    )

    expect(outcome).toEqual({ kind: 'idempotency-conflict' })
  })

  it('maps 409 with code=idempotency-locked to idempotency-conflict (not a name conflict)', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'idempotency-locked' }))

    const outcome = await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'reused-key',
    )

    expect(outcome).toEqual({ kind: 'idempotency-conflict' })
  })

  it('maps a 500 to unexpected-error', async () => {
    postMock.mockResolvedValueOnce(problem(500, { code: 'internal-error' }))

    const outcome = await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await createService(
      { name: 'Corte clásico', description: '', durationMinutes: 30, price: '45000.00' },
      'idem-key-1',
    )

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('catalogApi.updateService', () => {
  beforeEach(() => patchMock.mockReset())

  it('sends the serviceId as a path parameter and the four catalog fields as the body', async () => {
    patchMock.mockResolvedValueOnce(ok(serviceBody))

    await updateService(serviceBody.id, {
      name: 'Nuevo nombre',
      description: 'Nueva descripción',
      durationMinutes: 45,
      price: '50000.00',
    })

    expect(patchMock).toHaveBeenCalledWith('/private/services/{serviceId}', {
      params: { path: { serviceId: serviceBody.id } },
      body: {
        name: 'Nuevo nombre',
        description: 'Nueva descripción',
        durationMinutes: 45,
        price: '50000.00',
      },
    })
  })

  it('maps a 200 success body to a success outcome', async () => {
    patchMock.mockResolvedValueOnce(ok(serviceBody))

    const outcome = await updateService(serviceBody.id, {
      name: 'Corte clásico',
      description: '',
      durationMinutes: 30,
      price: '45000.00',
    })

    expect(outcome).toEqual({ kind: 'success', service: serviceBody })
  })

  it('maps 404 to not-found', async () => {
    patchMock.mockResolvedValueOnce(problem(404, { code: 'not-found' }))

    const outcome = await updateService('unknown-id', {
      name: 'X',
      description: '',
      durationMinutes: 30,
      price: '1.00',
    })

    expect(outcome).toEqual({ kind: 'not-found' })
  })

  it('maps 409 to name-conflict', async () => {
    patchMock.mockResolvedValueOnce(problem(409, { code: 'conflict' }))

    const outcome = await updateService(serviceBody.id, {
      name: 'Repetido',
      description: '',
      durationMinutes: 30,
      price: '1.00',
    })

    expect(outcome).toEqual({ kind: 'name-conflict' })
  })

  it('maps 422 to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(422, { code: 'validation-error' }))

    const outcome = await updateService(serviceBody.id, {
      name: '',
      description: '',
      durationMinutes: 30,
      price: '1.00',
    })

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    patchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await updateService(serviceBody.id, {
      name: 'X',
      description: '',
      durationMinutes: 30,
      price: '1.00',
    })

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})
