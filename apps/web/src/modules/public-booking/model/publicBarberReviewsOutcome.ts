// Modelo de las reseñas públicas de un barbero (HU-100, DEC-128). La página
// solo ve estos tipos; la forma RFC 9457 queda dentro de
// `api/listPublicBarberReviewsApi.ts`.

/** Reseña visible: anónima para el visitante (sin turno ni cliente). */
export interface PublicBarberReview {
  rating: number
  comment: string | null
  createdAt: string
}

export interface PublicBarberReviews {
  rating: number | null
  ratingCount: number
  items: PublicBarberReview[]
}

export type ListPublicBarberReviewsOutcome =
  | { kind: 'success'; reviews: PublicBarberReviews }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
