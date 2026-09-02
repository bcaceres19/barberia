// Fecha civil "AAAA-MM-DD" (HU-063, RN-DIS-07): estas funciones nunca leen
// la zona del dispositivo. getCivilDateInTimezone es la única que necesita
// una zona explícita (para saber qué día es "hoy" en la barbería);
// shiftCivilDate y formatCivilDateFull operan sobre el triplete
// año/mes/día ya resuelto y por eso son estables ante DST, fin de mes y fin
// de año: un día civil siempre avanza "+1", sin importar cuántas horas de
// reloj tenga ese día en la zona de la barbería.
const CIVIL_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/

function pad2(value: number): string {
  return String(value).padStart(2, '0')
}

// isCivilDateString valida formato AAAA-MM-DD y que la fecha exista en el
// calendario (rechaza "2026-02-30"): mismo criterio estricto que
// dailyAgendaDateLayout en el backend (booking.agenda.go).
export function isCivilDateString(value: string): boolean {
  if (!CIVIL_DATE_PATTERN.test(value)) return false
  const [year, month, day] = value.split('-').map(Number)
  const date = new Date(Date.UTC(year!, month! - 1, day!))
  return (
    date.getUTCFullYear() === year && date.getUTCMonth() === month! - 1 && date.getUTCDate() === day
  )
}

// getCivilDateInTimezone (CA-063-*): lee el día civil vigente AHORA en
// `timezone` (la de la barbería), nunca en la zona del dispositivo.
// Intl.DateTimeFormat resuelve el día civil real incluso en un día local de
// 23/25 horas por DST.
export function getCivilDateInTimezone(timezone: string, at: Date = new Date()): string {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(at)
  const year = parts.find((p) => p.type === 'year')!.value
  const month = parts.find((p) => p.type === 'month')!.value
  const day = parts.find((p) => p.type === 'day')!.value
  return `${year}-${month}-${day}`
}

// shiftCivilDate: aritmética de calendario pura en UTC (Date.UTC), nunca
// suma 24h de reloj. Cruza mes/año correctamente y es indiferente a DST:
// la zona de la barbería no participa en esta cuenta porque un día civil
// siempre son "+1", sin importar su duración real en horas ahí.
export function shiftCivilDate(civilDate: string, deltaDays: number): string {
  const [year, month, day] = civilDate.split('-').map(Number)
  const shifted = new Date(Date.UTC(year!, month! - 1, day! + deltaDays))
  return `${shifted.getUTCFullYear()}-${pad2(shifted.getUTCMonth() + 1)}-${pad2(shifted.getUTCDate())}`
}

// formatCivilDateFull (CA-063-*, estandar-diseno-visual.md §13) muestra
// "jueves, 31 de agosto de 2026" para una fecha civil ya resuelta. Se
// ancla a mediodía UTC y formatea EN UTC (nunca en la zona de la barbería
// ni del dispositivo): una fecha civil ya no necesita reinterpretarse por
// zona para mostrarse, y hacerlo arriesgaría cruzar el día en una zona con
// desfase extremo (p. ej. UTC+14).
export function formatCivilDateFull(civilDate: string): string {
  const [year, month, day] = civilDate.split('-').map(Number)
  const at = new Date(Date.UTC(year!, month! - 1, day!, 12))
  return new Intl.DateTimeFormat('es-CO', { timeZone: 'UTC', dateStyle: 'full' }).format(at)
}

// minutesIntoCivilDate (Fase 4a de la adopción NAVA, issue de la línea
// temporal de escritorio, especificacion-frontend-nava.md §7.2): minutos
// transcurridos desde la medianoche de `civilDate` (en `timezone`) hasta
// `isoInstant`, recortados a [0, 1440]. Un turno nocturno que interseca
// `civilDate` desde el día anterior se recorta a 0 (aparece "desde el
// inicio" de la línea de este día); uno que termina el día siguiente se
// recorta a 1440 ("hasta el final"). Puramente visual: no decide
// disponibilidad ni reinterpreta el instante que el servidor ya confirmó.
export function minutesIntoCivilDate(
  isoInstant: string,
  civilDate: string,
  timezone: string,
): number {
  const at = new Date(isoInstant)
  const instantCivilDate = getCivilDateInTimezone(timezone, at)
  if (instantCivilDate < civilDate) return 0
  if (instantCivilDate > civilDate) return 24 * 60

  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: timezone,
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(at)
  const hour = Number(parts.find((p) => p.type === 'hour')!.value)
  const minute = Number(parts.find((p) => p.type === 'minute')!.value)
  return hour * 60 + minute
}
