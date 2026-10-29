// Resultados discriminados de las dos operaciones de marca y vocabulario
// (issue #292, DEC-110), igual criterio que `settingsOutcome.ts`: ninguno
// expone `Problem`, `status` HTTP crudo ni cabeceras. Un 401 real lo
// intercepta la coordinación única de `installSessionHandling`.
import type { BrandSettings } from '@/shared/model'

/**
 * Ubicación física de la barbería (DEC-130, HU-102), tal como la guarda el
 * servidor. Las coordenadas son solo lectura: salen de geocodificar `address`
 * al guardarla. Vive aparte de `BrandSettings` porque solo Configuración la usa.
 */
export interface LocationSettings {
  hasPhysicalLocation: boolean
  address: string | null
  latitude: number | null
  longitude: number | null
}

/** Cambio de ubicación a enviar: lo ausente no se toca; `address: null` la quita. */
export interface LocationChange {
  hasPhysicalLocation?: boolean
  address?: string | null
}

export const DEFAULT_LOCATION: LocationSettings = {
  hasPhysicalLocation: true,
  address: null,
  latitude: null,
  longitude: null,
}

export type FetchBrandOutcome =
  | { kind: 'success'; brand: BrandSettings; location: LocationSettings }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SaveBrandOutcome =
  | { kind: 'success'; brand: BrandSettings; location: LocationSettings }
  | { kind: 'validation-error' }
  /** El servidor no pudo ubicar la dirección enviada (422): no se guardó nada. */
  | { kind: 'location-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
