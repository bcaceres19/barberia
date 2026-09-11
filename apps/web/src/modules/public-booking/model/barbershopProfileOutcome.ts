// Modelo de dominio de la resolución pública de una barbería (HU-090).
// `PublicBarbershopProfile` es la forma que la página consume;
// `ResolveBarbershopOutcome` es lo único que `api/resolveBarbershopApi.ts`
// entrega. Ninguno de los dos expone la forma RFC 9457 (`Problem`) del
// contrato (docs/03-desarrollo/estandar-frontend-vue.md §6): un slug mal
// formado, desconocido o no publicable llegan TODOS como `not-found`
// (CA-090-02), nunca distinguidos.

/** Contexto público mínimo de una barbería habilitada (CA-090-01,
 * CA-090-04). contactEmail/contactPhone son `null` cuando la barbería no
 * tiene ese contacto configurado. */
export interface PublicBarbershopProfile {
  name: string
  timezone: string
  contactEmail: string | null
  contactPhone: string | null
}

/** Resultado discriminado de resolver un enlace público contra el API
 * real. `not-found` cubre las tres causas indistinguibles de CA-090-02
 * (forma inválida, desconocido, no publicable); `network-error` y
 * `unexpected-error` son las mismas fronteras defensivas que el resto de
 * la aplicación ya usa (ver auth/model/loginOutcome.ts). */
export type ResolveBarbershopOutcome =
  | { kind: 'success'; profile: PublicBarbershopProfile }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error'; requestId?: string }
