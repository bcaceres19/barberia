// Único punto del módulo `public-booking` que llama a
// `GET /public/barbershops/{slug}/reviews` (HU-101, DEC-129).
import { httpClient } from '@/shared/api/httpClient'
import type { GetPublicGoogleReviewsOutcome } from '../model/publicGoogleReviewsOutcome'

/**
 * Lee la calificación y las reseñas de Google del negocio. La sección es opcional y
 * nunca debe mostrar un error: la estructura vacía del servidor, un 404, un fallo
 * inesperado y un fallo de red son todos `empty` (se oculta sin ruido).
 */
export async function getPublicGoogleReviews(slug: string): Promise<GetPublicGoogleReviewsOutcome> {
  try {
    const { data, response } = await httpClient.GET('/public/barbershops/{slug}/reviews', {
      params: { path: { slug } },
    })
    if (!response.ok || !data) return { kind: 'empty' }
    // Sin calificación no hay nada que mostrar, aunque el servidor devolviera reseñas sueltas.
    if (data.rating === null || data.rating === undefined || data.ratingCount <= 0) {
      return { kind: 'empty' }
    }
    return {
      kind: 'success',
      reviews: {
        rating: data.rating,
        ratingCount: data.ratingCount,
        mapsUrl: data.mapsUrl ?? null,
        reviews: data.reviews.map((review) => ({
          authorName: review.authorName,
          authorUrl: review.authorUrl ?? null,
          rating: review.rating,
          text: review.text ?? null,
          relativeTime: review.relativeTime ?? null,
        })),
      },
    }
  } catch {
    return { kind: 'empty' }
  }
}
