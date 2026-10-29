// Cliente de `GET`/`PATCH /private/settings/brand` (issue #292, DEC-110). Mismo
// criterio que `settingsApi.ts`: traduce la respuesta real a un outcome
// discriminado y mapea por `status`, nunca por el texto del error.
import { httpClient } from '@/shared/api/httpClient'
import { isBrandAccentKey, isPanelProfile, type BrandSettings } from '@/shared/model'
import type {
  FetchBrandOutcome,
  LocationChange,
  LocationSettings,
  SaveBrandOutcome,
} from '../model/brandOutcome'

export async function fetchBrand(): Promise<FetchBrandOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/brand')
    if (response.ok && data) {
      const brand = toBrand(data)
      return brand
        ? { kind: 'success', brand, location: toLocation(data) }
        : { kind: 'unexpected-error' }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

/**
 * Guarda la marca y, si se pide, un cambio de ubicación. Los campos de
 * `location` ausentes no se envían (el servidor conserva lo guardado); una
 * dirección nueva se geocodifica en el servidor y, si no se puede ubicar,
 * responde 422 como `location-error` sin haber guardado nada.
 */
export async function saveBrand(
  brand: BrandSettings,
  location: LocationChange = {},
): Promise<SaveBrandOutcome> {
  try {
    // El panel lo deciden los roles del usuario (DEC-134): el PATCH ya no lo admite.
    const { data, response } = await httpClient.PATCH('/private/settings/brand', {
      body: {
        accent: brand.accent,
        businessTerm: brand.businessTerm,
        businessTermGender: brand.businessTermGender,
        professionalTerm: brand.professionalTerm,
        professionalTermPlural: brand.professionalTermPlural,
        professionalTermGender: brand.professionalTermGender,
        ...(location.hasPhysicalLocation !== undefined
          ? { hasPhysicalLocation: location.hasPhysicalLocation }
          : {}),
        ...(location.address !== undefined ? { address: location.address } : {}),
        ...(location.offersHomeService !== undefined
          ? { offersHomeService: location.offersHomeService }
          : {}),
        ...(location.googlePlaceId !== undefined ? { googlePlaceId: location.googlePlaceId } : {}),
      },
    })
    if (response.ok && data) {
      const saved = toBrand(data)
      return saved
        ? { kind: 'success', brand: saved, location: toLocation(data) }
        : { kind: 'unexpected-error' }
    }
    switch (response.status) {
      case 400:
        return { kind: 'validation-error' }
      case 422:
        // Con una dirección en la petición, el 422 es casi seguro de ella: los
        // términos ya se validaron en el formulario antes de enviar.
        return location.address ? { kind: 'location-error' } : { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

/** Una clave de acento que esta versión del cliente no conoce se descarta en vez de pintarse. */
function toBrand(data: BrandSettings): BrandSettings | null {
  if (!isBrandAccentKey(data.accent)) return null
  return {
    accent: data.accent,
    businessTerm: data.businessTerm,
    businessTermGender: data.businessTermGender,
    professionalTerm: data.professionalTerm,
    professionalTermPlural: data.professionalTermPlural,
    professionalTermGender: data.professionalTermGender,
    // Un perfil que esta versión no conoce cae al panel completo, nunca a uno recortado.
    panelProfile: isPanelProfile(data.panelProfile) ? data.panelProfile : 'shop',
  }
}

function toLocation(data: LocationSettings): LocationSettings {
  return {
    hasPhysicalLocation: data.hasPhysicalLocation,
    address: data.address,
    latitude: data.latitude,
    longitude: data.longitude,
    offersHomeService: data.offersHomeService,
    googlePlaceId: data.googlePlaceId,
  }
}
