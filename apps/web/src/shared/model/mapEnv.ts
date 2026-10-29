// Entorno del navegador para Google Maps/Places (DEC-130, DEC-131): la clave de NAVEGADOR
// (`VITE_GOOGLE_MAPS_BROWSER_KEY`, restringida por dominio en Google Cloud) y si se corre
// en el servidor de desarrollo de Vite. Aparte de `mapEmbed.ts` para que el mapa y el
// autocompletado lo compartan sin depender el uno del otro.

export interface MapEnv {
  /** Clave de navegador de Maps Embed y Places API, restringida por dominio. */
  key?: string
  /** `true` solo en el servidor de desarrollo de Vite. */
  dev: boolean
}

/** Entorno real de la compilación. */
export function readMapEnv(): MapEnv {
  const key = import.meta.env.VITE_GOOGLE_MAPS_BROWSER_KEY
  return {
    key: typeof key === 'string' && key.trim() !== '' ? key.trim() : undefined,
    dev: import.meta.env.DEV,
  }
}
