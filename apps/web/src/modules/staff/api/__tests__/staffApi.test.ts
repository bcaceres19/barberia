/**
 * Pruebas del cliente tipado (`staffApi` sobre `shared/api/httpClient`):
 * mapeo por `status` (nunca por `detail`), mismo criterio que
 * `settings/api/__tests__/settingsApi.test.ts`. El recorrido de red real
 * vive en un E2E propio de HU-021.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const postMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
const putMock = vi.hoisted(() => vi.fn())
const deleteMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: {
    GET: getMock,
    POST: postMock,
    PATCH: patchMock,
    PUT: putMock,
    DELETE: deleteMock,
  },
}))

const {
  fetchBarbers,
  fetchBarberPage,
  createBarber,
  renameBarber,
  uploadBarberPhoto,
  removeBarberPhoto,
  barberPhotoUrl,
} = await import('../staffApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function created(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 201 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return { data: undefined, error: body, response: new Response(null, { status }) }
}

const barberBody = {
  id: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  fullName: 'Carlos Ramírez',
  createdAt: '2026-08-23T15:04:05Z',
  updatedAt: '2026-08-23T15:04:05Z',
  photoUpdatedAt: null,
}

describe('staffApi.fetchBarbers', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a success outcome with items and nextCursor', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [barberBody], nextCursor: null }))

    const outcome = await fetchBarbers()

    expect(outcome).toEqual({ kind: 'success', page: { items: [barberBody], nextCursor: null } })
  })

  it('forwards the cursor as a query parameter when provided', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [], nextCursor: null }))

    await fetchBarbers('opaque-cursor')

    expect(getMock).toHaveBeenCalledWith('/private/barbers', {
      params: { query: { cursor: 'opaque-cursor' } },
    })
  })

  it('sends no cursor query parameter for the first page', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [], nextCursor: null }))

    await fetchBarbers()

    expect(getMock).toHaveBeenCalledWith('/private/barbers', { params: { query: {} } })
  })

  it('maps any non-2xx status to unexpected-error (401 handled globally)', async () => {
    getMock.mockResolvedValueOnce(problem(401))

    const outcome = await fetchBarbers()

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await fetchBarbers()

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('staffApi.createBarber', () => {
  beforeEach(() => postMock.mockReset())

  it('sends fullName and the Idempotency-Key header', async () => {
    postMock.mockResolvedValueOnce(created(barberBody))

    await createBarber('Carlos Ramírez', 'idem-key-1')

    expect(postMock).toHaveBeenCalledWith('/private/barbers', {
      params: { header: { 'Idempotency-Key': 'idem-key-1' } },
      body: { fullName: 'Carlos Ramírez' },
    })
  })

  it('maps a 201 success body to a success outcome', async () => {
    postMock.mockResolvedValueOnce(created(barberBody))

    const outcome = await createBarber('Carlos Ramírez', 'idem-key-1')

    expect(outcome).toEqual({ kind: 'success', barber: barberBody })
  })

  it('maps 422 to validation-error', async () => {
    postMock.mockResolvedValueOnce(problem(422, { code: 'validation-error' }))

    const outcome = await createBarber('', 'idem-key-1')

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps 409 to idempotency-conflict', async () => {
    postMock.mockResolvedValueOnce(problem(409, { code: 'idempotency-conflict' }))

    const outcome = await createBarber('Carlos Ramírez', 'reused-key')

    expect(outcome).toEqual({ kind: 'idempotency-conflict' })
  })

  it('maps a 500 to unexpected-error', async () => {
    postMock.mockResolvedValueOnce(problem(500, { code: 'internal-error' }))

    const outcome = await createBarber('Carlos Ramírez', 'idem-key-1')

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await createBarber('Carlos Ramírez', 'idem-key-1')

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('staffApi.renameBarber', () => {
  beforeEach(() => patchMock.mockReset())

  it('sends the barberId as a path parameter and fullName as the body', async () => {
    patchMock.mockResolvedValueOnce(ok(barberBody))

    await renameBarber(barberBody.id, 'Nuevo Nombre')

    expect(patchMock).toHaveBeenCalledWith('/private/barbers/{barberId}', {
      params: { path: { barberId: barberBody.id } },
      body: { fullName: 'Nuevo Nombre' },
    })
  })

  it('maps a 200 success body to a success outcome', async () => {
    patchMock.mockResolvedValueOnce(ok(barberBody))

    const outcome = await renameBarber(barberBody.id, 'Carlos Ramírez')

    expect(outcome).toEqual({ kind: 'success', barber: barberBody })
  })

  it('maps 404 to not-found', async () => {
    patchMock.mockResolvedValueOnce(problem(404, { code: 'not-found' }))

    const outcome = await renameBarber('unknown-id', 'X')

    expect(outcome).toEqual({ kind: 'not-found' })
  })

  it('maps 422 to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(422, { code: 'validation-error' }))

    const outcome = await renameBarber(barberBody.id, '')

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    patchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await renameBarber(barberBody.id, 'X')

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('staffApi barber photo mapping (DEC-104)', () => {
  beforeEach(() => getMock.mockReset())

  it('keeps photoUpdatedAt when the barber has a photo', async () => {
    const withPhoto = { ...barberBody, photoUpdatedAt: '2026-09-30T12:00:00Z' }
    getMock.mockResolvedValueOnce(ok({ items: [withPhoto], nextCursor: null }))

    const outcome = await fetchBarbers()

    expect(outcome).toEqual({ kind: 'success', page: { items: [withPhoto], nextCursor: null } })
  })

  it('treats a missing photoUpdatedAt (a replayed pre-DEC-104 response) as no photo', async () => {
    const { photoUpdatedAt: _omitted, ...legacy } = barberBody
    void _omitted
    getMock.mockResolvedValueOnce(ok({ items: [legacy], nextCursor: null }))

    const outcome = await fetchBarbers()

    expect(outcome).toEqual({ kind: 'success', page: { items: [barberBody], nextCursor: null } })
  })
})

describe('staffApi.barberPhotoUrl', () => {
  it('returns null while the barber has no photo', () => {
    expect(barberPhotoUrl({ id: 'b-1', photoUpdatedAt: null })).toBeNull()
  })

  it('versions the URL with photoUpdatedAt so a new photo never reuses a stale cache entry', () => {
    const url = barberPhotoUrl({ id: 'b-1', photoUpdatedAt: '2026-09-30T12:00:00Z' })
    expect(url).toBe('/api/v1/private/barbers/b-1/photo?v=2026-09-30T12%3A00%3A00Z')
    expect(barberPhotoUrl({ id: 'b-1', photoUpdatedAt: '2026-10-01T09:00:00Z' })).not.toBe(url)
  })
})

describe('staffApi.uploadBarberPhoto', () => {
  beforeEach(() => putMock.mockReset())

  const blob = new Blob(['jpeg-bytes'], { type: 'image/jpeg' })

  it('sends the blob untouched with its own Content-Type and the barberId as a path parameter', async () => {
    putMock.mockResolvedValueOnce(ok({ ...barberBody, photoUpdatedAt: '2026-09-30T12:00:00Z' }))

    await uploadBarberPhoto(barberBody.id, blob)

    const [path, init] = putMock.mock.calls[0]!
    expect(path).toBe('/private/barbers/{barberId}/photo')
    expect(init.params).toEqual({ path: { barberId: barberBody.id } })
    expect(init.headers).toEqual({ 'Content-Type': 'image/jpeg' })
    expect(init.body).toBe(blob)
    // El serializador deja pasar el Blob: nunca lo convierte a JSON.
    expect(init.bodySerializer(blob)).toBe(blob)
  })

  it('maps a 200 body to a success outcome with the new photoUpdatedAt', async () => {
    const updated = { ...barberBody, photoUpdatedAt: '2026-09-30T12:00:00Z' }
    putMock.mockResolvedValueOnce(ok(updated))

    expect(await uploadBarberPhoto(barberBody.id, blob)).toEqual({
      kind: 'success',
      barber: updated,
    })
  })

  it.each([
    [404, 'not-found'],
    [422, 'validation-error'],
    [500, 'unexpected-error'],
  ])('maps status %i to %s', async (status, kind) => {
    putMock.mockResolvedValueOnce(problem(status))
    expect(await uploadBarberPhoto(barberBody.id, blob)).toEqual({ kind })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    putMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    expect(await uploadBarberPhoto(barberBody.id, blob)).toEqual({ kind: 'network-error' })
  })
})

describe('staffApi.removeBarberPhoto', () => {
  beforeEach(() => deleteMock.mockReset())

  it('maps a 204 to success and sends the barberId as a path parameter', async () => {
    deleteMock.mockResolvedValueOnce({
      data: undefined,
      response: new Response(null, { status: 204 }),
    })

    expect(await removeBarberPhoto(barberBody.id)).toEqual({ kind: 'success' })
    expect(deleteMock).toHaveBeenCalledWith('/private/barbers/{barberId}/photo', {
      params: { path: { barberId: barberBody.id } },
    })
  })

  it.each([
    [404, 'not-found'],
    [500, 'unexpected-error'],
  ])('maps status %i to %s', async (status, kind) => {
    deleteMock.mockResolvedValueOnce(problem(status))
    expect(await removeBarberPhoto(barberBody.id)).toEqual({ kind })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    deleteMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    expect(await removeBarberPhoto(barberBody.id)).toEqual({ kind: 'network-error' })
  })
})

describe('staffApi.fetchBarberPage (#288)', () => {
  beforeEach(() => getMock.mockReset())
  it('requests the numbered mode and maps exact page metadata', async () => {
    getMock.mockResolvedValueOnce(
      ok({ items: [barberBody], nextCursor: null, page: 2, pageSize: 3, total: 4, totalPages: 2 }),
    )
    expect(await fetchBarberPage(2, 3)).toEqual({
      kind: 'success',
      page: { items: [barberBody], page: 2, pageSize: 3, total: 4, totalPages: 2 },
    })
    expect(getMock).toHaveBeenCalledWith('/private/barbers', {
      params: { query: { page: 2, pageSize: 3 } },
    })
  })
  it('rejects a cursor response instead of inventing numbered metadata', async () => {
    getMock.mockResolvedValueOnce(ok({ items: [barberBody], nextCursor: null }))
    expect(await fetchBarberPage(1, 3)).toEqual({ kind: 'unexpected-error' })
  })
})
