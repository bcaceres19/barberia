// Catálogo de zonas horarias para el selector de Configuración (DEC-110).
// Sale de `Intl.supportedValuesOf('timeZone')`, la misma base IANA del
// navegador, sin tabla propia ni dependencia nueva. El servidor sigue siendo la
// autoridad (CA-020-03): confirma cada zona contra `pg_timezone_names` y
// responde 422 si no la reconoce; este catálogo solo evita que la persona
// tenga que teclear de memoria un identificador exacto.

/** Respaldo cuando el navegador no expone la lista (muy antiguo o entorno de prueba). */
const FALLBACK_TIMEZONES = [
  'America/Argentina/Buenos_Aires',
  'America/Bogota',
  'America/Caracas',
  'America/Guayaquil',
  'America/Lima',
  'America/Mexico_City',
  'America/Montevideo',
  'America/New_York',
  'America/Panama',
  'America/Santiago',
  'America/Sao_Paulo',
  'Europe/Madrid',
  'UTC',
]

export function listTimezones(extra: readonly string[] = []): string[] {
  let zones: string[] = []
  try {
    const supported = (Intl as unknown as { supportedValuesOf?: (key: string) => string[] })
      .supportedValuesOf
    zones = supported ? supported('timeZone') : []
  } catch {
    zones = []
  }
  if (zones.length === 0) zones = [...FALLBACK_TIMEZONES]
  // La zona vigente de la barbería siempre aparece, aunque el navegador no la liste.
  const all = new Set([...zones, ...extra.filter(Boolean)])
  return [...all].sort((a, b) => a.localeCompare(b, 'en'))
}

/** Zona del dispositivo, solo como sugerencia: nunca sustituye la configurada (RN-DIS-07). */
export function detectDeviceTimezone(): string | null {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || null
  } catch {
    return null
  }
}

/** `true` si el navegador reconoce `timezone` como zona IANA. */
export function isKnownTimezone(timezone: string): boolean {
  try {
    new Intl.DateTimeFormat('es-CO', { timeZone: timezone })
    return true
  } catch {
    return false
  }
}

/** "UTC−5", "UTC+5:30", "UTC" en el instante `at` (respeta el horario de verano). */
export function offsetLabel(timezone: string, at: Date = new Date()): string {
  try {
    const part = new Intl.DateTimeFormat('en-US', {
      timeZone: timezone,
      timeZoneName: 'longOffset',
    })
      .formatToParts(at)
      .find((p) => p.type === 'timeZoneName')?.value
    if (!part || part === 'GMT') return 'UTC'
    // "GMT-05:00" → "UTC−5", "GMT+05:30" → "UTC+5:30".
    const match = /^GMT([+-])(\d{2}):(\d{2})$/.exec(part)
    if (!match) return part.replace('GMT', 'UTC')
    if (Number(match[2]) === 0 && match[3] === '00') return 'UTC'
    const sign = match[1] === '-' ? '−' : '+'
    const hours = String(Number(match[2]))
    return `UTC${sign}${hours}${match[3] === '00' ? '' : `:${match[3]}`}`
  } catch {
    return ''
  }
}

/** Reloj de la zona: hora en 24 h sin segundos para que el ritmo lo ponga la animación. */
export function clockParts(
  timezone: string,
  at: Date = new Date(),
): { hours: string; minutes: string; day: string } | null {
  try {
    const time = new Intl.DateTimeFormat('en-GB', {
      timeZone: timezone,
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
    }).formatToParts(at)
    const rawDay = new Intl.DateTimeFormat('es-CO', {
      timeZone: timezone,
      weekday: 'long',
      day: 'numeric',
      month: 'long',
    }).format(at)
    const day = rawDay.charAt(0).toLocaleUpperCase('es') + rawDay.slice(1)
    return {
      hours: time.find((p) => p.type === 'hour')?.value ?? '--',
      minutes: time.find((p) => p.type === 'minute')?.value ?? '--',
      day,
    }
  } catch {
    return null
  }
}

function fold(text: string): string {
  return text.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/[_/]/g, ' ')
}

/** Filtra por cualquier fragmento del identificador ("bogota", "new york", "america") o del desfase ("utc-5"). */
export function filterTimezones(zones: readonly string[], query: string, at: Date = new Date()) {
  const needle = fold(query.trim()).replace('−', '-')
  if (!needle) return [...zones]
  const terms = needle.split(/\s+/).filter(Boolean)
  return zones.filter((zone) => {
    const haystack = `${fold(zone)} ${offsetLabel(zone, at).toLowerCase().replace('−', '-')}`
    return terms.every((term) => haystack.includes(term))
  })
}

/** "America/Bogota" → "Bogota" para el título de una opción; el identificador va aparte. */
export function cityOf(timezone: string): string {
  const last = timezone.split('/').pop() ?? timezone
  return last.replace(/_/g, ' ')
}
