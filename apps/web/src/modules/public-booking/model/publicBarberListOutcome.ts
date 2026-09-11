// Modelo de dominio de la selección pública de barbero (HU-092).
// `PublicBarber` es la forma que la página consume; `ListPublicBarbersOutcome`
// es lo único que `api/listPublicBarbersApi.ts` entrega. Ninguno de los dos
// expone la forma RFC 9457 (`Problem`) del contrato (docs/03-desarrollo/
// estandar-frontend-vue.md §6), mismo criterio que
// model/publicServiceListOutcome.ts.

/** Barbero público mínimo, con asignación vigente al servicio activo
 * elegido (CA-092-02). Nunca foto, biografía ni preferencia (fuera de
 * alcance de HU-092). */
export interface PublicBarber {
  id: string
  fullName: string
}

/** Resultado discriminado de listar los barberos elegibles contra el API
 * real. `not-found` cubre las mismas tres causas indistinguibles de
 * CA-090-02 (slug con forma inválida, desconocido, no publicable) que
 * `ListPublicServicesOutcome`; un `serviceId` ajeno, inexistente, ya no
 * asignado o de un servicio inactivo NUNCA produce `not-found`: llega como
 * `success` con `barbers: []` (CA-092-03, misma respuesta uniforme que
 * "cero barberos asignados hoy"). `network-error` y `unexpected-error` son
 * las mismas fronteras defensivas del resto de la aplicación. */
export type ListPublicBarbersOutcome =
  | { kind: 'success'; barbers: PublicBarber[] }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
