// Mapa interactivo para que el cliente marque su ubicación a domicilio (DEC-138, HU-103).
// Dos motores detrás de la misma interfaz, para que la pantalla y las pruebas no dependan de
// Google:
//  - Google Maps JavaScript API (con clave de navegador): se carga con un `<script>` y se usa
//    `importLibrary`; no se añade ninguna dependencia. La clave es la MISMA de navegador del
//    mapa incrustado y del autocompletado, restringida por dominio en Google Cloud.
//  - Simulado (solo desarrollo, sin clave): una superficie con botones para mover el punto,
//    determinista y sin red.
// Sin clave y fuera de desarrollo no hay motor (`null`) y la reserva pública no ofrece
// «a domicilio» (`isPlacesAvailable`).
//
// La ubicación es dato personal (RN-DAT-02): nada de lo que se marca se registra en consola
// ni se guarda fuera del formulario.
import { readMapEnv, type MapEnv } from './mapEnv'

export interface LatLng {
  latitude: number
  longitude: number
}

export interface MapPickerHandle {
  /** Centro actual del mapa: el punto que la persona marcó. */
  getCenter(): LatLng
  /** Mueve el mapa por código (no cuenta como una acción de la persona). */
  setCenter(point: LatLng, zoom?: number): void
  destroy(): void
}

export interface MapPickerEngine {
  /**
   * Dibuja el mapa en `el`. `onUserMove` se invoca solo cuando la PERSONA mueve el mapa
   * (arrastrar, tocar, teclado), nunca por `setCenter`.
   */
  mount(
    el: HTMLElement,
    start: LatLng,
    zoom: number,
    onUserMove: (center: LatLng) => void,
  ): Promise<MapPickerHandle>
  /** Dirección aproximada del punto, o `null` si no hay una. */
  describe(point: LatLng, signal?: AbortSignal): Promise<string | null>
}

/** Cercanía (en grados, ~1 m) bajo la cual dos centros son el mismo punto. */
const SAME_POINT_EPSILON = 1e-5

export function samePoint(a: LatLng, b: LatLng): boolean {
  return (
    Math.abs(a.latitude - b.latitude) < SAME_POINT_EPSILON &&
    Math.abs(a.longitude - b.longitude) < SAME_POINT_EPSILON
  )
}

/** Redondea a 6 decimales (~11 cm): más precisión no aporta y delata ruido del mapa. */
export function roundPoint(point: LatLng): LatLng {
  const round = (n: number) => Math.round(n * 1e6) / 1e6
  return { latitude: round(point.latitude), longitude: round(point.longitude) }
}

export function isValidPoint(point: LatLng): boolean {
  return (
    Number.isFinite(point.latitude) &&
    Number.isFinite(point.longitude) &&
    Math.abs(point.latitude) <= 90 &&
    Math.abs(point.longitude) <= 180
  )
}

/** Punto de partida cuando la barbería no tiene ubicación: Bogotá, a vista de ciudad. */
export const DEFAULT_START: LatLng = { latitude: 4.711, longitude: -74.0721 }
export const DEFAULT_START_ZOOM = 11
/** Texto de la ubicación cuando no hay una dirección aproximada para el punto. */
export const FALLBACK_ADDRESS = 'Ubicación marcada en el mapa'
/** Zoom con el que se acerca un punto ya conocido (sugerencia elegida o posición detectada). */
export const FOCUS_ZOOM = 17

/** Motor real (con clave) o simulado (desarrollo sin clave); `null` si ninguno aplica. */
export function createMapPickerEngine(env: MapEnv = readMapEnv()): MapPickerEngine | null {
  if (env.key) return googleEngine(env.key)
  if (env.dev) return simulatedEngine()
  return null
}

// --- Google Maps JavaScript API ------------------------------------------------------------

interface GoogleLatLng {
  lat(): number
  lng(): number
}
interface GoogleMapInstance {
  getCenter(): GoogleLatLng | undefined
  setCenter(point: { lat: number; lng: number }): void
  setZoom(zoom: number): void
  panTo(point: { lat: number; lng: number }): void
  addListener(
    event: string,
    handler: (event?: { latLng?: GoogleLatLng | null }) => void,
  ): { remove(): void }
}
interface GoogleMapsNamespace {
  importLibrary(name: 'maps'): Promise<{
    Map: new (el: HTMLElement, options: Record<string, unknown>) => GoogleMapInstance
  }>
  importLibrary(name: 'geocoding'): Promise<{
    Geocoder: new () => {
      geocode(request: {
        location: { lat: number; lng: number }
        language?: string
      }): Promise<{ results: { formatted_address?: string; types?: string[] }[] }>
    }
  }>
}
type GoogleWindow = Window & { google?: { maps?: Partial<GoogleMapsNamespace> } }

/** Nombre global que la API de Google invoca al quedar lista (parámetro `callback`). */
const READY_CALLBACK = '__navaGoogleMapsReady'

let loading: Promise<GoogleMapsNamespace> | null = null

/** Carga la API una sola vez; un fallo se recuerda solo hasta el siguiente intento. */
export function loadGoogleMaps(key: string): Promise<GoogleMapsNamespace> {
  const existing = (window as GoogleWindow).google?.maps
  if (existing?.importLibrary) return Promise.resolve(existing as GoogleMapsNamespace)
  if (loading) return loading
  loading = new Promise<GoogleMapsNamespace>((resolve, reject) => {
    const script = document.createElement('script')
    // Con `loading=async` el evento `load` del script llega ANTES de que exista
    // `google.maps.importLibrary`; la API avisa que está lista llamando a `callback`.
    const w = window as unknown as GoogleWindow & Record<string, unknown>
    w[READY_CALLBACK] = () => {
      delete w[READY_CALLBACK]
      const maps = w.google?.maps
      if (maps?.importLibrary) resolve(maps as GoogleMapsNamespace)
      else reject(new Error('maps-api-unavailable'))
    }
    const params = new URLSearchParams({
      key,
      v: 'weekly',
      language: 'es',
      loading: 'async',
      callback: READY_CALLBACK,
    })
    script.src = `https://maps.googleapis.com/maps/api/js?${params.toString()}`
    script.async = true
    script.onerror = () => {
      delete w[READY_CALLBACK]
      reject(new Error('maps-api-load-failed'))
    }
    document.head.appendChild(script)
  }).catch((error: unknown) => {
    loading = null
    throw error
  })
  return loading
}

function googleEngine(key: string): MapPickerEngine {
  return {
    async mount(el, start, zoom, onUserMove) {
      const maps = await loadGoogleMaps(key)
      const { Map } = await maps.importLibrary('maps')
      const map = new Map(el, {
        center: { lat: start.latitude, lng: start.longitude },
        zoom,
        // El pin va fijo al centro (lo dibuja la pantalla): sin marcadores ni controles que
        // compitan con él, y sin vista de calle, que no aporta para marcar un punto.
        disableDefaultUI: true,
        zoomControl: true,
        gestureHandling: 'greedy',
        keyboardShortcuts: true,
        clickableIcons: false,
      })
      let known: LatLng = start
      const center = (): LatLng => {
        const c = map.getCenter()
        return c ? { latitude: c.lat(), longitude: c.lng() } : known
      }
      // `idle` llega tras cualquier movimiento (arrastre, tacto, teclado, zoom con rueda).
      // Si el centro cambió respecto del último conocido, lo movió la persona; los movimientos
      // por código fijan `known` antes, así que no se confunden con uno de la persona.
      const idle = map.addListener('idle', () => {
        const now = center()
        if (samePoint(now, known)) return
        known = now
        onUserMove(now)
      })
      const click = map.addListener('click', (event) => {
        const at = event?.latLng
        if (!at) return
        map.panTo({ lat: at.lat(), lng: at.lng() })
      })
      return {
        getCenter: center,
        setCenter(point, nextZoom) {
          known = point
          map.setCenter({ lat: point.latitude, lng: point.longitude })
          if (nextZoom !== undefined) map.setZoom(nextZoom)
        },
        destroy() {
          idle.remove()
          click.remove()
        },
      }
    },
    async describe(point, signal) {
      const maps = await loadGoogleMaps(key)
      const { Geocoder } = await maps.importLibrary('geocoding')
      const { results } = await new Geocoder().geocode({
        location: { lat: point.latitude, lng: point.longitude },
        language: 'es',
      })
      if (signal?.aborted) return null
      // Para un punto sin dirección postal Google devuelve primero un código Plus
      // («67P7JW7R+G8»): no le sirve al barbero, así que se salta y, si no hay otra, queda el texto de respaldo.
      const address = results.find((r) => !r.types?.includes('plus_code'))
      const text = (address?.formatted_address ?? '').trim().replace(/\s+/g, ' ')
      return text === '' ? null : text
    },
  }
}

// --- Simulado (solo desarrollo) -----------------------------------------------------------

/** Paso del punto simulado por botón: ~110 m. */
const SIMULATED_STEP = 0.001

function simulatedEngine(): MapPickerEngine {
  return {
    async mount(el, start, _zoom, onUserMove) {
      let center = start
      el.classList.add('map-picker-sim')
      const surface = document.createElement('div')
      surface.className = 'map-picker-sim__surface'
      surface.textContent = 'Mapa simulado (desarrollo)'
      el.appendChild(surface)
      const buttons: HTMLButtonElement[] = []
      const moves: [string, string, number, number][] = [
        ['Mover el mapa al norte', '↑', SIMULATED_STEP, 0],
        ['Mover el mapa al sur', '↓', -SIMULATED_STEP, 0],
        ['Mover el mapa al oeste', '←', 0, -SIMULATED_STEP],
        ['Mover el mapa al este', '→', 0, SIMULATED_STEP],
      ]
      for (const [label, glyph, dLat, dLng] of moves) {
        const button = document.createElement('button')
        button.type = 'button'
        button.className = 'map-picker-sim__button'
        button.setAttribute('aria-label', label)
        button.textContent = glyph
        button.addEventListener('click', () => {
          center = { latitude: center.latitude + dLat, longitude: center.longitude + dLng }
          onUserMove(center)
        })
        el.appendChild(button)
        buttons.push(button)
      }
      return {
        getCenter: () => center,
        setCenter(point) {
          center = point
        },
        destroy() {
          surface.remove()
          buttons.forEach((b) => b.remove())
          el.classList.remove('map-picker-sim')
        },
      }
    },
    async describe(point) {
      return `Punto marcado cerca de ${point.latitude.toFixed(4)}, ${point.longitude.toFixed(4)}`
    },
  }
}
