// Modelo de las reseñas de Google del negocio en la reserva pública (HU-101, DEC-129).
// La página solo ve estos tipos; la forma RFC 9457 queda dentro de
// `api/getPublicGoogleReviewsApi.ts`.

/** Reseña de Google: la atribución al autor es obligatoria al mostrarla. */
export interface PublicGoogleReview {
  authorName: string
  /** Perfil del autor en Google Maps, o `null`. */
  authorUrl: string | null
  rating: number
  text: string | null
  /** «hace un mes», ya redactado por Google. */
  relativeTime: string | null
}

export interface PublicGoogleReviews {
  rating: number
  ratingCount: number
  /** Ficha real en Google Maps, destino de la atribución. */
  mapsUrl: string | null
  reviews: PublicGoogleReview[]
}

/**
 * `empty` cubre todo lo que no se muestra: sin ficha, sin credencial, consulta fallida
 * (el servidor responde la estructura vacía) y también cualquier error de este cliente:
 * la sección se oculta sin error visible (CA-101-01, CA-101-03).
 */
export type GetPublicGoogleReviewsOutcome =
  { kind: 'success'; reviews: PublicGoogleReviews } | { kind: 'empty' }
