// URL de la fotografía de un barbero (GET /private/barbers/{id}/photo, DEC-104),
// lista para el `src` de una imagen: la cookie de sesión viaja sola. La usan
// varios módulos (equipo, agenda), por eso vive en shared/api y no en uno solo.
// `photoUpdatedAt` va como parámetro de consulta solo para que el navegador no
// reutilice la copia de una fotografía anterior; el servidor lo ignora. Devuelve
// null cuando el barbero no tiene fotografía (se muestra su monograma).
export function barberPhotoUrl(barber: {
  id: string
  photoUpdatedAt?: string | null
}): string | null {
  if (!barber.photoUpdatedAt) return null
  return `/api/v1/private/barbers/${barber.id}/photo?v=${encodeURIComponent(barber.photoUpdatedAt)}`
}
