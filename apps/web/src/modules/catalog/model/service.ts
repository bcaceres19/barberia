// Espejo tipado del cuerpo real del contrato (CA-022-01/02/07, CA-024-02,
// CA-024-05): id, name, description, durationMinutes, price, currency,
// isActive, deactivatedAt, createdAt, updatedAt. price es un string decimal
// exacto (docs/06-api/estandar-openapi.md §6: "importe exacto: string
// decimal más código de moneda"): nunca se convierte a number en el
// cliente, para no arriesgar coma flotante en un valor que solo se muestra
// tal cual llega. currency es siempre "COP" (DEC-067): informativo, nunca
// editable desde esta pantalla. isActive/deactivatedAt son de solo lectura
// (HU-024): solo cambian mediante deactivateService/reactivateService.
export interface Service {
  id: string
  name: string
  description: string | null
  durationMinutes: number
  price: string
  currency: string
  isActive: boolean
  deactivatedAt: string | null
  createdAt: string
  updatedAt: string
}

/** Página paginada por cursor (CA-022-01): mismo `items`/`nextCursor` que
 * ServiceListResponse. */
export interface ServicePage {
  items: Service[]
  nextCursor: string | null
}
