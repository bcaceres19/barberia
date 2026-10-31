// Guarda de la URL de «Tus datos» (F-RQ-02, #436): un `startsAt` que no es una fecha, o que ya
// pasó, viene de un enlace o una pestaña vieja y no puede ser una franja reservable. Solo se
// descarta lo evidentemente inválido; la disponibilidad sigue siendo del servidor (HU-097).
export function isBookableStart(startsAt: unknown, now: number = Date.now()): boolean {
  if (typeof startsAt !== 'string') return false
  const time = Date.parse(startsAt)
  return !Number.isNaN(time) && time > now
}
