// Modelo de dominio del catálogo público de servicios (HU-091).
// `PublicService` es la forma que la página consume; `ListPublicServicesOutcome`
// es lo único que `api/listPublicServicesApi.ts` entrega. Ninguno de los dos
// expone la forma RFC 9457 (`Problem`) del contrato (docs/03-desarrollo/
// estandar-frontend-vue.md §6), mismo criterio que
// model/barbershopProfileOutcome.ts.

/** Servicio público mínimo, ya activo y con al menos una asignación
 * vigente a un barbero (CA-091-01). `price` viaja como el string decimal
 * exacto del contrato: nunca se convierte a `number` en el cliente
 * (docs/06-api/estandar-openapi.md §6). */
export interface PublicService {
  id: string
  name: string
  description: string | null
  durationMinutes: number
  price: string
  currency: string
}

/** Resultado discriminado de listar el catálogo público contra el API
 * real. `not-found` cubre las mismas tres causas indistinguibles de
 * CA-090-02 (forma inválida, desconocido, no publicable) que
 * `ResolveBarbershopOutcome`; `network-error` y `unexpected-error` son las
 * mismas fronteras defensivas del resto de la aplicación. */
export type ListPublicServicesOutcome =
  | { kind: 'success'; services: PublicService[] }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
