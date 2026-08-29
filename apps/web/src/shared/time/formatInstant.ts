// Presentación de instantes en una zona IANA SIEMPRE explícita, con la API
// de plataforma `Intl.DateTimeFormat` (sin agregar una librería de fechas
// nueva, docs/10-backlog/prompts/hu/hu-020-configuracion-barberia.md
// "trabajo requerido" §4.6). `timezone` es obligatorio y nunca tiene un
// valor por defecto: la zona del dispositivo NUNCA sustituye la zona
// configurada de la barbería (CA-020-04, RN-DIS-07). Ningún otro punto del
// código debe formatear un instante para el barbero sin pasar por aquí (o
// una función que, como esta, reciba la zona de forma explícita).
export function formatInstantInTimezone(isoInstant: string, timezone: string): string {
  return new Intl.DateTimeFormat('es-CO', {
    timeZone: timezone,
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(isoInstant))
}

// formatTimeInTimezone (HU-062) muestra solo la hora de un instante, en la
// zona explícita recibida: cada fila de la agenda diaria necesita la hora
// sin repetir la fecha completa ya visible en el encabezado.
export function formatTimeInTimezone(isoInstant: string, timezone: string): string {
  return new Intl.DateTimeFormat('es-CO', {
    timeZone: timezone,
    timeStyle: 'short',
  }).format(new Date(isoInstant))
}

// formatFullDateInTimezone (HU-062, CA-062-01) muestra la fecha civil
// completa ("jueves, 28 de agosto de 2026") de `at` en `timezone`: la
// agenda diaria siempre confirma en pantalla en qué día y zona está parada,
// nunca solo una hora suelta. `at` por defecto es el instante real del
// dispositivo SOLO para construir el objeto Date a formatear (el día civil
// resultante lo decide `timeZone`, nunca el reloj/zona del dispositivo).
export function formatFullDateInTimezone(timezone: string, at: Date = new Date()): string {
  return new Intl.DateTimeFormat('es-CO', {
    timeZone: timezone,
    dateStyle: 'full',
  }).format(at)
}
