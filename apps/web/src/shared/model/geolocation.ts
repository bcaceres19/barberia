// Detección de la ubicación del cliente con la API de geolocalización del navegador
// (DEC-138, HU-103). Solo se invoca por una acción explícita de la persona (un botón): el
// navegador pide su propio permiso y nunca se consulta al cargar la página.
//
// La posición es dato personal (RN-DAT-02): no se registra ni se guarda; quien llama la usa
// para centrar el mapa y la persona decide si la confirma.
import { isValidPoint, roundPoint, type LatLng } from './mapPicker'

export type LocateResult =
  | { kind: 'success'; point: LatLng; accuracyMeters: number | null }
  /** La persona (o el navegador) negó el permiso. */
  | { kind: 'denied' }
  /** El dispositivo no pudo determinar la posición. */
  | { kind: 'unavailable' }
  | { kind: 'timeout' }
  /** Este navegador o contexto (p. ej. sin HTTPS) no ofrece geolocalización. */
  | { kind: 'unsupported' }

/** Espera máxima de una detección: lo bastante para un GPS en frío sin dejar la pantalla colgada. */
export const LOCATE_TIMEOUT_MS = 12_000

/** `true` si este navegador ofrece geolocalización. */
export function isGeolocationSupported(
  geo: Pick<Geolocation, 'getCurrentPosition'> | undefined = currentGeolocation(),
): boolean {
  return typeof geo?.getCurrentPosition === 'function'
}

function currentGeolocation(): Geolocation | undefined {
  return typeof navigator === 'undefined' ? undefined : navigator.geolocation
}

/** Detecta la posición actual. Nunca lanza: cada fallo es un resultado explícito. */
export function locateUser(
  geo: Pick<Geolocation, 'getCurrentPosition'> | undefined = currentGeolocation(),
  timeoutMs: number = LOCATE_TIMEOUT_MS,
): Promise<LocateResult> {
  if (!isGeolocationSupported(geo)) return Promise.resolve({ kind: 'unsupported' })
  return new Promise<LocateResult>((resolve) => {
    geo!.getCurrentPosition(
      (position) => {
        const point = roundPoint({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        })
        if (!isValidPoint(point)) {
          resolve({ kind: 'unavailable' })
          return
        }
        const accuracy = position.coords.accuracy
        resolve({
          kind: 'success',
          point,
          accuracyMeters: Number.isFinite(accuracy) ? Math.round(accuracy) : null,
        })
      },
      (error) => {
        // 1 = PERMISSION_DENIED, 2 = POSITION_UNAVAILABLE, 3 = TIMEOUT.
        if (error.code === 1) resolve({ kind: 'denied' })
        else if (error.code === 3) resolve({ kind: 'timeout' })
        else resolve({ kind: 'unavailable' })
      },
      { enableHighAccuracy: true, timeout: timeoutMs, maximumAge: 0 },
    )
  })
}
