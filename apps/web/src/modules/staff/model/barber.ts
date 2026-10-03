// Espejo tipado del cuerpo real del contrato (CA-021-01/02/07, DEC-104):
// id, fullName, createdAt, updatedAt y photoUpdatedAt. `photoUpdatedAt` es
// null mientras el barbero no tiene fotografía (la pantalla muestra su
// monograma) y, cuando existe, la versión de la imagen: se añade a la URL para
// que un cambio de fotografía nunca se sirva desde una caché vieja.
export interface Barber {
  id: string
  fullName: string
  createdAt: string
  updatedAt: string
  photoUpdatedAt: string | null
}

/** Página paginada por cursor (CA-021-02): mismo `items`/`nextCursor` que
 * BarberListResponse. */
export interface BarberPage {
  items: Barber[]
  nextCursor: string | null
}

export interface NumberedBarberPage {
  items: Barber[]
  page: number
  pageSize: number
  total: number
  totalPages: number
}
