// Calificar el turno del cliente (HU-100, DEC-128): único punto que llama a
// `POST /customer/appointments/{token}/review`. Traduce la respuesta a
// `CreateReviewOutcome`; la página nunca ve `Problem` ni el `status` crudo.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { CreateReviewOutcome, ReviewInput } from '../model/customerAppointmentOutcome'

/**
 * Mapea por `status` (nunca por `detail`): 201 reseña guardada, 422 entrada
 * inválida, 409 turno que no admite reseña (sin distinguir la causa), 404 enlace
 * inválido. Repetir la llamada no crea otra reseña (una por turno), así que no lleva
 * `Idempotency-Key`.
 */
export async function createReview(
  token: string,
  input: ReviewInput,
): Promise<CreateReviewOutcome> {
  try {
    const { data, error, response } = await httpClient.POST(
      '/customer/appointments/{token}/review',
      {
        params: { path: { token } },
        body: { rating: input.rating, comment: input.comment },
      },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        review: { rating: data.rating, comment: data.comment, createdAt: data.createdAt },
      }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'not-reviewable' }
      case 400:
      case 422:
        return { kind: 'validation-error' }
      default:
        return {
          kind: 'unexpected-error',
          requestId: isProblem(error) ? error.requestId : undefined,
        }
    }
  } catch {
    return { kind: 'network-error' }
  }
}
