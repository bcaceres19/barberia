// Calificación de un barbero (DEC-128): promedio de 1 decimal sobre 1–5 y total de
// reseñas visibles. Funciones puras compartidas por la reserva pública, el acceso al
// turno y el panel; el servidor calcula ambos valores, aquí solo se presentan.

/** Calificación mínima y máxima que acepta el servidor. */
export const MIN_RATING = 1
export const MAX_RATING = 5
/** Largo máximo del comentario, contado en caracteres como el servidor. */
export const REVIEW_COMMENT_MAX_LENGTH = 500

/** `4.5` → `4,5` (coma decimal es-CO). `null` → cadena vacía. */
export function formatRating(rating: number | null): string {
  return rating === null ? '' : rating.toFixed(1).replace('.', ',')
}

/** «reseña» o «reseñas» según el total. */
export function reviewCountLabel(count: number): string {
  return count === 1 ? '1 reseña' : `${count} reseñas`
}

/**
 * Texto accesible de un promedio: «4,5 de 5 · 12 reseñas». Sin reseñas visibles
 * (`rating` nulo o total 0) dice «Sin reseñas todavía», nunca «0 de 5».
 */
export function ratingSummaryLabel(rating: number | null, count: number): string {
  if (rating === null || count === 0) return 'Sin reseñas todavía'
  return `${formatRating(rating)} de 5 · ${reviewCountLabel(count)}`
}

/** Cuántos caracteres cuenta el servidor en un comentario (puntos de código). */
export function commentLength(value: string): number {
  return Array.from(value.trim()).length
}
