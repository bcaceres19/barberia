// Espejo tipado del cuerpo real del contrato (CA-022-01/02/07): exactamente
// id, name, description, durationMinutes, price, currency, createdAt,
// updatedAt. price es un string decimal exacto (docs/06-api/
// estandar-openapi.md §6: "importe exacto: string decimal más código de
// moneda"): nunca se convierte a number en el cliente, para no arriesgar
// coma flotante en un valor que solo se muestra tal cual llega. currency es
// siempre "COP" (DEC-067): informativo, nunca editable desde esta pantalla.
export interface Service {
  id: string
  name: string
  description: string | null
  durationMinutes: number
  price: string
  currency: string
  createdAt: string
  updatedAt: string
}

/** Página paginada por cursor (CA-022-01): mismo `items`/`nextCursor` que
 * ServiceListResponse. */
export interface ServicePage {
  items: Service[]
  nextCursor: string | null
}
