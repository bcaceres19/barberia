// Modelo de dominio de la selección pública de barbero (HU-092).
// `PublicBarber` es la forma que la página consume; `ListPublicBarbersOutcome`
// es lo único que `api/listPublicBarbersApi.ts` entrega. Ninguno de los dos
// expone la forma RFC 9457 (`Problem`) del contrato (docs/03-desarrollo/
// estandar-frontend-vue.md §6), mismo criterio que
// model/publicServiceListOutcome.ts.

/** Barbero público con asignación vigente al servicio activo elegido
 * (CA-092-02). Por DEC-127 incluye su descripción opcional y la versión de su
 * fotografía opcional (los bytes se piden aparte, ver `publicBarberPhotoUrl`);
 * nunca preferencia ni ningún otro dato interno. */
export interface PublicBarber {
  id: string
  fullName: string
  /** Descripción corta que escribió la barbería, o null si no tiene. */
  description: string | null
  /** Versión de la fotografía, o null si no tiene (se muestra el monograma). */
  photoUpdatedAt: string | null
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
