// Único punto del módulo `public-booking` que llama a
// `GET /public/barbershops/{slug}/barbers/{barberId}/reviews` (HU-100).
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ListPublicBarberReviewsOutcome } from '../model/publicBarberReviewsOutcome'

/**
 * Lista las reseñas visibles del barbero. Mapea por `status`: 404 cubre slug o
 * barbero inexistente (indistinguibles a propósito); cualquier otro no-2xx es
 * inesperado y un fallo de `fetch` es error de red.
 */
export async function listPublicBarberReviews(
  slug: string,
  barberId: string,
): Promise<ListPublicBarberReviewsOutcome> {
  try {
    const { data, error, response } = await httpClient.GET(
      '/public/barbershops/{slug}/barbers/{barberId}/reviews',
      { params: { path: { slug, barberId } } },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        reviews: {
          rating: data.rating ?? null,
          ratingCount: data.ratingCount,
          items: data.items.map((item) => ({
            rating: item.rating,
            comment: item.comment ?? null,
            createdAt: item.createdAt,
          })),
        },
      }
    }

    if (response.status === 404) return { kind: 'not-found' }
    return { kind: 'unexpected-error', requestId: isProblem(error) ? error.requestId : undefined }
  } catch {
    return { kind: 'network-error' }
  }
}
