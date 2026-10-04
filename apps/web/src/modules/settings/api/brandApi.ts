// Cliente de `GET`/`PATCH /private/settings/brand` (issue #292, DEC-110). Mismo
// criterio que `settingsApi.ts`: traduce la respuesta real a un outcome
// discriminado y mapea por `status`, nunca por el texto del error.
import { httpClient } from '@/shared/api/httpClient'
import { isBrandAccentKey, isPanelProfile, type BrandSettings } from '@/shared/model'
import type { FetchBrandOutcome, SaveBrandOutcome } from '../model/brandOutcome'

export async function fetchBrand(): Promise<FetchBrandOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/brand')
    if (response.ok && data) {
      const brand = toBrand(data)
      return brand ? { kind: 'success', brand } : { kind: 'unexpected-error' }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function saveBrand(brand: BrandSettings): Promise<SaveBrandOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/settings/brand', { body: brand })
    if (response.ok && data) {
      const saved = toBrand(data)
      return saved ? { kind: 'success', brand: saved } : { kind: 'unexpected-error' }
    }
    switch (response.status) {
      case 400:
      case 422:
        return { kind: 'validation-error' }
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
