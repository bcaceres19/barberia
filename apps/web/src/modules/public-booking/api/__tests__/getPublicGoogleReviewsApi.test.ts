/**
 * getPublicGoogleReviews (HU-101, DEC-129): la sección es opcional, así que todo lo que
 * no sea una calificación con datos —estructura vacía, 404, error del servidor, red— se
 * traduce en `empty`, nunca en un error visible (CA-101-01, CA-101-03).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: getMock } }))

const { getPublicGoogleReviews } = await import('../getPublicGoogleReviewsApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: { ok: true, status: 200 } as Response }
}
function fail(status: number) {
  return { data: undefined, error: {}, response: { ok: false, status } as Response }
}

const full = {
  rating: 4.7,
  ratingCount: 128,
  mapsUrl: 'https://maps.google.com/?cid=1',
  reviews: [
    {
      authorName: 'Camila R.',
      authorUrl: 'https://www.google.com/maps/contrib/1',
      rating: 5,
      text: 'Excelente atención.',
      relativeTime: 'hace un mes',
    },
    { authorName: 'Andrés M.', authorUrl: null, rating: 4, text: null, relativeTime: null },
  ],
}

beforeEach(() => getMock.mockReset())

describe('getPublicGoogleReviews', () => {
  it('maps a listing with data and asks the right path', async () => {
    getMock.mockResolvedValueOnce(ok(full))

    expect(await getPublicGoogleReviews('barberia-ejemplo')).toEqual({
      kind: 'success',
      reviews: full,
    })
    expect(getMock).toHaveBeenCalledWith('/public/barbershops/{slug}/reviews', {
      params: { path: { slug: 'barberia-ejemplo' } },
    })
  })

  it('treats the explicit empty structure as nothing to show', async () => {
    getMock.mockResolvedValueOnce(ok({ rating: null, ratingCount: 0, mapsUrl: null, reviews: [] }))

    expect(await getPublicGoogleReviews('barberia-ejemplo')).toEqual({ kind: 'empty' })
  })

  it('treats a rating without any count as nothing to show', async () => {
    getMock.mockResolvedValueOnce(ok({ ...full, ratingCount: 0 }))

    expect(await getPublicGoogleReviews('barberia-ejemplo')).toEqual({ kind: 'empty' })
  })

  it.each([404, 500])('hides the section on a %s instead of surfacing an error', async (status) => {
    getMock.mockResolvedValueOnce(fail(status))

    expect(await getPublicGoogleReviews('barberia-ejemplo')).toEqual({ kind: 'empty' })
  })

  it('hides the section on a network failure', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    expect(await getPublicGoogleReviews('barberia-ejemplo')).toEqual({ kind: 'empty' })
  })
})
