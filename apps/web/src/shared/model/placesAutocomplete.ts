// Autocompletado de direcciones con Google Places (DEC-131, HU-103). Funciones con el
// entorno y `fetch` inyectables para probarlas sin red ni `import.meta.env`.
//
// Usa la API REST de Places (New) desde el navegador con la MISMA clave de navegador del
// mapa (`VITE_GOOGLE_MAPS_BROWSER_KEY`, restringida por dominio): no se carga ningún script
// externo. Sin clave, solo en desarrollo, un proveedor simulado y determinista deja ver y
// probar la pantalla; en producción sin clave no hay forma de elegir una dirección, así
// que la reserva pública no ofrece «a domicilio» (ver `isPlacesAvailable`).
//
// La dirección es dato personal (RN-DAT-02): nada de lo que se escribe o se elige se
// registra en consola ni se guarda fuera del formulario.
import { readMapEnv, type MapEnv } from './mapEnv'

export interface PlaceSuggestion {
  placeId: string
  /** Texto completo de la sugerencia, tal como Google lo muestra. */
  text: string
}

/** Una dirección elegida: texto y coordenadas siempre juntos. */
export interface PlaceSelection {
  address: string
  latitude: number
  longitude: number
}

export type FetchFn = typeof fetch

/** `true` si hay forma de sugerir y resolver direcciones en este entorno. */
export function isPlacesAvailable(env: MapEnv = readMapEnv()): boolean {
  return Boolean(env.key) || env.dev
}

/** Mínimo de letras antes de pedir sugerencias: evita una consulta por cada tecla inicial. */
export const MIN_QUERY_LENGTH = 3
const MAX_SUGGESTIONS = 5
const PLACES_URL = 'https://places.googleapis.com/v1'

export interface PlacesClient {
  suggest(query: string, signal?: AbortSignal): Promise<PlaceSuggestion[]>
  resolve(placeId: string, signal?: AbortSignal): Promise<PlaceSelection | null>
}

/** Cliente real (con clave) o simulado (desarrollo sin clave); `null` si ninguno aplica. */
export function createPlacesClient(
  env: MapEnv = readMapEnv(),
  fetchFn: FetchFn = (...args) => fetch(...args),
): PlacesClient | null {
  if (env.key) return googleClient(env.key, fetchFn)
  if (env.dev) return simulatedClient()
  return null
}

function googleClient(key: string, fetchFn: FetchFn): PlacesClient {
  return {
    async suggest(query, signal) {
      const input = query.trim()
      if ([...input].length < MIN_QUERY_LENGTH) return []
      const response = await fetchFn(`${PLACES_URL}/places:autocomplete`, {
        method: 'POST',
        signal,
        headers: { 'Content-Type': 'application/json', 'X-Goog-Api-Key': key },
        body: JSON.stringify({ input, languageCode: 'es' }),
      })
      if (!response.ok) throw new Error('places-autocomplete-failed')
      const data = (await response.json()) as {
        suggestions?: { placePrediction?: { placeId?: string; text?: { text?: string } } }[]
      }
      const out: PlaceSuggestion[] = []
      for (const s of data.suggestions ?? []) {
        const p = s.placePrediction
        if (p?.placeId && p.text?.text) out.push({ placeId: p.placeId, text: p.text.text })
        if (out.length === MAX_SUGGESTIONS) break
      }
      return out
    },
    async resolve(placeId, signal) {
      const response = await fetchFn(`${PLACES_URL}/places/${encodeURIComponent(placeId)}`, {
        signal,
        headers: {
          'X-Goog-Api-Key': key,
          'X-Goog-FieldMask': 'formattedAddress,location',
          'Accept-Language': 'es',
        },
      })
      if (!response.ok) throw new Error('places-details-failed')
      const data = (await response.json()) as {
        formattedAddress?: string
        location?: { latitude?: number; longitude?: number }
      }
      return toSelection(data.formattedAddress, data.location?.latitude, data.location?.longitude)
    },
  }
}

/** Valida y normaliza lo que devuelve el proveedor: sin coordenadas reales no hay selección. */
export function toSelection(
  address: string | undefined,
  latitude: number | undefined,
  longitude: number | undefined,
): PlaceSelection | null {
  const text = (address ?? '').trim().replace(/\s+/g, ' ')
  if (
    text === '' ||
    typeof latitude !== 'number' ||
    typeof longitude !== 'number' ||
    !Number.isFinite(latitude) ||
    !Number.isFinite(longitude) ||
    Math.abs(latitude) > 90 ||
    Math.abs(longitude) > 180
  ) {
    return null
  }
  return { address: text, latitude, longitude }
}

// Proveedor simulado (solo desarrollo): tres sugerencias derivadas del texto, con
// coordenadas fijas alrededor de Bogotá. No habla con ninguna red.
function simulatedClient(): PlacesClient {
  const points = [
    { latitude: 4.6486, longitude: -74.0628 },
    { latitude: 4.6097, longitude: -74.0817 },
    { latitude: 4.711, longitude: -74.0721 },
  ]
  return {
    async suggest(query) {
      const input = query.trim().replace(/\s+/g, ' ')
      if ([...input].length < MIN_QUERY_LENGTH) return []
      return points.map((_, i) => ({
        placeId: `sim-${i}|${input}`,
        text: `${input}, Bogotá (${i + 1})`,
      }))
    },
    async resolve(placeId) {
      const [head = '', ...rest] = placeId.split('|')
      const index = Number(head.replace('sim-', ''))
      const point = points[index]
      if (!point || rest.length === 0) return null
      return toSelection(
        `${rest.join('|')}, Bogotá (${index + 1})`,
        point.latitude,
        point.longitude,
      )
    },
  }
}
