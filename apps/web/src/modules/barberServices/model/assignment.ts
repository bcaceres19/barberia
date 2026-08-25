// Espejo tipado del cuerpo real del contrato (HU-023, CA-023-07): exactamente
// barberId, serviceId, createdAt. Nunca nombre, duración, precio ni estado
// del barbero o del servicio: esos campos siguen siendo responsabilidad
// exclusiva de staff/catalog; esta pantalla los cruza por identificador
// contra BarberSummary/ServiceSummary, obtenidos por separado.
export interface Assignment {
  barberId: string
  serviceId: string
  createdAt: string
}

/** Página paginada por cursor (CA-023-01): mismo `items`/`nextCursor` que
 * AssignmentListResponse. */
export interface AssignmentPage {
  items: Assignment[]
  nextCursor: string | null
}

// BarberSummary/ServiceSummary son subconjuntos deliberadamente mínimos de
// los recursos reales de `staff`/`catalog` (solo id + el nombre visible
// necesario para el selector y la lista de servicios), obtenidos por este
// módulo directamente del cliente HTTP compartido -nunca importando
// staffApi.ts ni catalogApi.ts, que son privados de sus propios módulos
// (docs/03-desarrollo/estandar-frontend-vue.md §3: "un módulo no importa
// archivos internos de otro").
export interface BarberSummary {
  id: string
  fullName: string
}

export interface ServiceSummary {
  id: string
  name: string
}
