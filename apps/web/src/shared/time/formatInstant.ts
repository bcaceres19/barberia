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
