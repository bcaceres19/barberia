// Espejo tipado del cuerpo real del contrato (CA-021-01/02/07): exactamente
// id, fullName, createdAt, updatedAt. La pantalla administra únicamente
// nombres (trabajo requerido §4.6): createdAt/updatedAt no se muestran,
// pero viajan en el modelo por si un consumidor futuro los necesita, sin
// inventar un campo que el contrato no declare.
export interface Barber {
  id: string
  fullName: string
  createdAt: string
  updatedAt: string
}

/** Página paginada por cursor (CA-021-02): mismo `items`/`nextCursor` que
 * BarberListResponse. */
export interface BarberPage {
  items: Barber[]
  nextCursor: string | null
}
