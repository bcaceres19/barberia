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

/** Largo máximo del Place ID, igual que `barbershop_google_place_id_ck`. */
export const GOOGLE_PLACE_ID_MAX_LENGTH = 200

/** Recorta el Place ID; `''` si no queda nada (sin ficha). */
export function normalizeGooglePlaceId(raw: string): string {
  return raw.trim()
}

/**
 * Mensaje de error del Place ID, o `undefined` si es válido (puede ir vacío). Solo
 * letras, números, guion y guion bajo: lo mismo que acepta el servidor, para que quien
 * pegue por error un enlace de Google Maps vea el motivo antes de guardar.
 */
export function validateGooglePlaceId(raw: string): string | undefined {
  const id = normalizeGooglePlaceId(raw)
  if (id === '') return undefined
  if (id.length > GOOGLE_PLACE_ID_MAX_LENGTH || !/^[A-Za-z0-9_-]+$/.test(id)) {
    return 'El identificador solo admite letras, números, guion y guion bajo (no pegues un enlace).'
  }
  return undefined
}

export interface LocationDraft {
  hasPhysicalLocation: boolean
  /** Texto del campo, tal como se escribe. */
  address: string
  offersHomeService: boolean
  /** Texto del campo del Place ID, tal como se escribe. */
  googlePlaceId: string
}

export function draftFromSaved(saved: LocationSettings): LocationDraft {
  return {
    hasPhysicalLocation: saved.hasPhysicalLocation,
    address: saved.address ?? '',
    offersHomeService: saved.offersHomeService,
    googlePlaceId: saved.googlePlaceId ?? '',
  }
}

export function googlePlaceIdChanged(draft: LocationDraft, saved: LocationSettings): boolean {
  return normalizeGooglePlaceId(draft.googlePlaceId) !== (saved.googlePlaceId ?? '')
}

export function addressChanged(draft: LocationDraft, saved: LocationSettings): boolean {
  return normalizeAddress(draft.address) !== (saved.address ?? '')
}

export function sameLocation(draft: LocationDraft, saved: LocationSettings): boolean {
  return (
    draft.hasPhysicalLocation === saved.hasPhysicalLocation &&
    draft.offersHomeService === saved.offersHomeService &&
    !addressChanged(draft, saved) &&
    !googlePlaceIdChanged(draft, saved)
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
  if (googlePlaceIdChanged(draft, saved)) {
    const normalized = normalizeGooglePlaceId(draft.googlePlaceId)
    change.googlePlaceId = normalized === '' ? null : normalized
  }
  return change
}
