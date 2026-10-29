// URL del mapa incrustado de Google (DEC-130, HU-102) y del enlace para abrirlo.
// Funciones puras con el entorno inyectable, para probarlas sin tocar
// `import.meta.env`.
//
// La clave de NAVEGADOR (`VITE_GOOGLE_MAPS_BROWSER_KEY`) es distinta de la de
// servidor y se restringe por dominio en Google Cloud: aquí solo viaja la de
// navegador. Sin ella, la sección de mapa queda oculta en producción (DEC-130);
// en desarrollo se usa el incrustado sin clave para poder ver la pantalla.

import { readMapEnv, type MapEnv } from './mapEnv'

export type { MapEnv }
export { readMapEnv }

/** `true` si hay forma de dibujar el mapa en este entorno. */
export function isMapAvailable(env: MapEnv = readMapEnv()): boolean {
  return Boolean(env.key) || env.dev
}

function validPoint(latitude: number, longitude: number): boolean {
  return (
    Number.isFinite(latitude) &&
    Number.isFinite(longitude) &&
    Math.abs(latitude) <= 90 &&
    Math.abs(longitude) <= 180
  )
}

/**
 * URL del `iframe` centrado en el punto, con marcador y zoom de calle; `null`
 * si el punto no es válido o no hay forma de mostrar el mapa en este entorno.
 */
export function mapEmbedUrl(
  latitude: number,
  longitude: number,
  env: MapEnv = readMapEnv(),
): string | null {
  if (!validPoint(latitude, longitude)) return null
  const query = encodeURIComponent(`${latitude},${longitude}`)
  if (env.key) {
    return `https://www.google.com/maps/embed/v1/place?key=${encodeURIComponent(env.key)}&q=${query}&zoom=16`
  }
  if (env.dev) return `https://maps.google.com/maps?q=${query}&z=16&output=embed`
  return null
}

/** Enlace para abrir el punto en Google Maps (no necesita clave). */
export function mapsLinkUrl(latitude: number, longitude: number): string | null {
  if (!validPoint(latitude, longitude)) return null
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(`${latitude},${longitude}`)}`
}
