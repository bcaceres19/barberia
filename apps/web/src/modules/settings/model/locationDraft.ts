// Borrador de la ubicación de la barbería en Configuración (DEC-130, HU-102):
// ayudas puras para normalizar la dirección como la guardará el servidor, saber
// si cambió y qué enviar. Sin Vue ni red, para probarlas sin montar nada.
import type { LocationChange, LocationSettings } from './brandOutcome'

/** Largo máximo de la dirección, igual que `barbershop_address_ck`. */
export const ADDRESS_MAX_LENGTH = 200

/** Recorta y colapsa espacios, igual que el servidor; `''` si no queda nada. */
export function normalizeAddress(raw: string): string {
  return raw.trim().replace(/\s+/g, ' ')
}

/** Mensaje de error de la dirección, o `undefined` si es válida (puede ir vacía). */
export function validateAddress(raw: string): string | undefined {
  if ([...normalizeAddress(raw)].length > ADDRESS_MAX_LENGTH) {
    return `La dirección no puede superar ${ADDRESS_MAX_LENGTH} caracteres.`
  }
  return undefined
}

export interface LocationDraft {
  hasPhysicalLocation: boolean
  /** Texto del campo, tal como se escribe. */
  address: string
  offersHomeService: boolean
}

export function draftFromSaved(saved: LocationSettings): LocationDraft {
  return {
    hasPhysicalLocation: saved.hasPhysicalLocation,
    address: saved.address ?? '',
    offersHomeService: saved.offersHomeService,
  }
}

export function addressChanged(draft: LocationDraft, saved: LocationSettings): boolean {
  return normalizeAddress(draft.address) !== (saved.address ?? '')
}

export function sameLocation(draft: LocationDraft, saved: LocationSettings): boolean {
  return (
    draft.hasPhysicalLocation === saved.hasPhysicalLocation &&
    draft.offersHomeService === saved.offersHomeService &&
    !addressChanged(draft, saved)
  )
}

/**
 * Lo que hay que enviar para llevar `saved` al borrador: solo lo que cambió. Una
 * dirección vacía se envía como `null` (la quita junto a sus coordenadas).
 */
export function locationChange(draft: LocationDraft, saved: LocationSettings): LocationChange {
  const change: LocationChange = {}
  if (draft.hasPhysicalLocation !== saved.hasPhysicalLocation) {
    change.hasPhysicalLocation = draft.hasPhysicalLocation
  }
  if (draft.offersHomeService !== saved.offersHomeService) {
    change.offersHomeService = draft.offersHomeService
  }
  if (addressChanged(draft, saved)) {
    const normalized = normalizeAddress(draft.address)
    change.address = normalized === '' ? null : normalized
  }
  return change
}
