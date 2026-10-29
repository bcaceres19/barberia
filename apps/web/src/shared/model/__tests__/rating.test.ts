import { describe, expect, it } from 'vitest'
import { commentLength, formatRating, ratingSummaryLabel, reviewCountLabel } from '../rating'

describe('rating', () => {
  it('formatea el promedio con coma decimal', () => {
    expect(formatRating(4.5)).toBe('4,5')
    expect(formatRating(5)).toBe('5,0')
    expect(formatRating(null)).toBe('')
  })

  it('singular y plural del total', () => {
    expect(reviewCountLabel(1)).toBe('1 reseña')
    expect(reviewCountLabel(0)).toBe('0 reseñas')
    expect(reviewCountLabel(12)).toBe('12 reseñas')
  })

  it('el texto accesible nunca dice «0 de 5»', () => {
    expect(ratingSummaryLabel(4.5, 12)).toBe('4,5 de 5 · 12 reseñas')
    expect(ratingSummaryLabel(null, 0)).toBe('Sin reseñas todavía')
    expect(ratingSummaryLabel(null, 3)).toBe('Sin reseñas todavía')
  })

  it('cuenta caracteres Unicode como el servidor', () => {
    expect(commentLength('  ñandú  ')).toBe(5)
    expect(commentLength('👍👍')).toBe(2)
  })
})
