// Reseñas de un barbero vistas desde el panel (HU-100, DEC-128, CA-100-05).
// A diferencia de la lista pública incluyen las ocultas, con la fecha en que se ocultaron.
export interface BarberReview {
  id: string
  rating: number
  comment: string | null
  /** Momento en que se ocultó, o null si está visible. */
  hiddenAt: string | null
  createdAt: string
}

export interface BarberReviewList {
  rating: number | null
  ratingCount: number
  items: BarberReview[]
}

export type ListBarberReviewsOutcome =
  | { kind: 'success'; reviews: BarberReviewList }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SetReviewHiddenOutcome =
  | { kind: 'success'; review: BarberReview }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
