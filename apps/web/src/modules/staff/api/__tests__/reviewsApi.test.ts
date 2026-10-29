/**
 * Cliente de moderación de reseñas (HU-100, DEC-128): mapeo por `status`, nunca por `detail`.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: getMock, PATCH: patchMock } }))

const { listBarberReviews, setReviewHidden } = await import('../reviewsApi')

const ok = (body: unknown) => ({
  data: body,
  error: undefined,
  response: new Response(null, { status: 200 }),
})
const problem = (status: number) => ({
  data: undefined,
  error: {},
  response: new Response(null, { status }),
})

const review = {
  id: 'b0000000-0000-0000-0000-000000000001',
  rating: 4,
  comment: null,
  hiddenAt: null,
  createdAt: '2027-03-15T16:10:00Z',
}

describe('reviewsApi.listBarberReviews', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 body, keeping hidden reviews', async () => {
    const hidden = {
      ...review,
      id: 'b0000000-0000-0000-0000-000000000002',
      hiddenAt: '2027-03-16T10:00:00Z',
    }
    getMock.mockResolvedValueOnce(ok({ rating: 4, ratingCount: 1, items: [review, hidden] }))

    expect(await listBarberReviews('barber-1')).toEqual({
      kind: 'success',
      reviews: { rating: 4, ratingCount: 1, items: [review, hidden] },
    })
  })

  it.each([
    [404, 'not-found'],
    [500, 'unexpected-error'],
  ])('maps a %s to %s', async (status, kind) => {
    getMock.mockResolvedValueOnce(problem(status))

    expect(await listBarberReviews('barber-1')).toEqual({ kind })
  })

  it('maps a thrown fetch to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await listBarberReviews('barber-1')).toEqual({ kind: 'network-error' })
  })
})

describe('reviewsApi.setReviewHidden', () => {
  beforeEach(() => patchMock.mockReset())

  it('sends only { hidden } and returns the updated review', async () => {
    patchMock.mockResolvedValueOnce(ok({ ...review, hiddenAt: '2027-03-16T10:00:00Z' }))

    const outcome = await setReviewHidden('barber-1', review.id, true)

    expect(patchMock).toHaveBeenCalledWith('/private/barbers/{barberId}/reviews/{reviewId}', {
      params: { path: { barberId: 'barber-1', reviewId: review.id } },
      body: { hidden: true },
    })
    expect(outcome).toEqual({
      kind: 'success',
      review: { ...review, hiddenAt: '2027-03-16T10:00:00Z' },
    })
  })

  it.each([
    [404, 'not-found'],
    [422, 'unexpected-error'],
    [500, 'unexpected-error'],
  ])('maps a %s to %s', async (status, kind) => {
    patchMock.mockResolvedValueOnce(problem(status))

    expect(await setReviewHidden('barber-1', review.id, false)).toEqual({ kind })
  })
})
