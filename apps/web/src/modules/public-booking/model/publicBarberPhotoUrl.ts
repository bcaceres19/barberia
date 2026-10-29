// URL de la fotografía pública de un barbero
// (GET /public/barbershops/{slug}/barbers/{barberId}/photo, DEC-127), lista para
// el `src` de una imagen: no lleva sesión. `photoUpdatedAt` va como parámetro de
// consulta solo para que el navegador no reutilice la copia de una fotografía
// anterior; el servidor lo ignora. Devuelve null cuando el barbero no tiene
// fotografía (se muestra su monograma). Mismo patrón que
// `shared/api/barberPhotoUrl.ts`, que sirve a la fotografía privada del panel.
export function publicBarberPhotoUrl(
  slug: string,
  barber: { id: string; photoUpdatedAt: string | null },
): string | null {
  if (!barber.photoUpdatedAt) return null
  return `/api/v1/public/barbershops/${encodeURIComponent(slug)}/barbers/${encodeURIComponent(barber.id)}/photo?v=${encodeURIComponent(barber.photoUpdatedAt)}`
}
