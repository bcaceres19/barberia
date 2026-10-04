// Geometría del tablero semanal de Horarios: convierte tramos "HH:MM + minutos"
// en posiciones sobre una regla horaria común a los siete días. Es puramente
// visual: no decide disponibilidad ni reinterpreta nada que el servidor ya
// confirmó (la hora civil de un tramo sigue siendo la que escribió el barbero,
// CA-040-06). Un tramo que cruza la medianoche se recorta al fin del día en la
// regla y se rotula "+1" en su etiqueta.
import { getCivilDateInTimezone, isCivilDateString } from '@/shared/time/civilDate'

export const MINUTES_PER_DAY = 1440

// La regla nunca es más estrecha que la jornada típica de una barbería; solo
// crece si algún tramo (o el que se está editando) se sale de ella.
const DEFAULT_SCALE_START = 7 * 60
const DEFAULT_SCALE_END = 21 * 60

export interface BoardScale {
  startMinute: number
  endMinute: number
}

export interface BoardTick {
  minute: number
  label: string
  /** Posición en % sobre la regla. */
  percent: number
}

export function timeToMinutes(time: string): number {
  const [hours, minutes] = time.split(':').map(Number)
  return (hours ?? 0) * 60 + (minutes ?? 0)
}

export function minutesToTime(minuteOfDay: number): string {
  const wrapped = ((minuteOfDay % MINUTES_PER_DAY) + MINUTES_PER_DAY) % MINUTES_PER_DAY
  const hours = Math.floor(wrapped / 60)
  const minutes = wrapped % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

interface TimedSegment {
  startsTime: string
  durationMinutes: number
}

// computeScale ajusta la regla a las horas enteras que contienen todos los
// tramos dados. Los tramos con hora vacía o inválida (formulario a medio
// llenar) se ignoran.
export function computeScale(segments: readonly TimedSegment[]): BoardScale {
  let start = DEFAULT_SCALE_START
  let end = DEFAULT_SCALE_END
  for (const segment of segments) {
    if (!/^\d{2}:\d{2}$/.test(segment.startsTime)) continue
    if (!Number.isFinite(segment.durationMinutes)) continue
    const from = timeToMinutes(segment.startsTime)
    const to = Math.min(MINUTES_PER_DAY, from + Math.max(0, segment.durationMinutes))
    start = Math.min(start, Math.floor(from / 60) * 60)
    end = Math.max(end, Math.ceil(to / 60) * 60)
  }
  return { startMinute: start, endMinute: Math.min(MINUTES_PER_DAY, end) }
}

export function scaleTicks(scale: BoardScale): BoardTick[] {
  const span = scale.endMinute - scale.startMinute
  // Etiquetas cada 2 h en una jornada normal; cada 3 h o 4 h si la regla crece.
  const stepHours = span <= 14 * 60 ? 2 : span <= 18 * 60 ? 3 : 4
  const ticks: BoardTick[] = []
  for (let minute = scale.startMinute; minute <= scale.endMinute; minute += stepHours * 60) {
    ticks.push({
      minute,
      label: String(Math.floor(minute / 60) % 24).padStart(2, '0'),
      percent: ((minute - scale.startMinute) / span) * 100,
    })
  }
  return ticks
}

export interface SegmentGeometry {
  /** Inicio en % sobre la regla. */
  from: number
  /** Ancho en % sobre la regla. */
  span: number
  /** Termina después de medianoche: la barra se recorta al borde derecho. */
  crossesMidnight: boolean
}

export function segmentGeometry(
  startsTime: string,
  durationMinutes: number,
  scale: BoardScale,
): SegmentGeometry {
  const total = scale.endMinute - scale.startMinute
  const from = timeToMinutes(startsTime)
  const rawEnd = from + durationMinutes
  const clippedFrom = Math.max(scale.startMinute, from)
  const clippedEnd = Math.min(scale.endMinute, rawEnd)
  return {
    from: ((clippedFrom - scale.startMinute) / total) * 100,
    span: (Math.max(0, clippedEnd - clippedFrom) / total) * 100,
    crossesMidnight: rawEnd > MINUTES_PER_DAY,
  }
}

// endLabel es la hora de fin de un tramo ("13:00", o "02:00 (+1)" si termina
// el día siguiente).
export function endLabel(startsTime: string, durationMinutes: number): string {
  const end = timeToMinutes(startsTime) + durationMinutes
  const time = minutesToTime(end)
  return end > MINUTES_PER_DAY ? `${time} (+1)` : time
}

// formatHours: 480 → "8 h", 450 → "7 h 30 min", 45 → "45 min".
export function formatHours(minutes: number): string {
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  if (hours === 0) return `${rest} min`
  return rest === 0 ? `${hours} h` : `${hours} h ${rest} min`
}

export interface BoardNow {
  isoWeekday: number
  minute: number
}

// nowInTimezone: día ISO (1 = lunes … 7 = domingo) y minuto del día vigentes
// AHORA en la zona de la barbería (nunca la del dispositivo, CA-040-06).
export function nowInTimezone(timeZone: string, at: Date = new Date()): BoardNow {
  const civilDate = getCivilDateInTimezone(timeZone, at)
  const [year, month, day] = civilDate.split('-').map(Number)
  const weekday = new Date(Date.UTC(year!, month! - 1, day!)).getUTCDay()
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone,
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(at)
  const get = (type: string) => Number(parts.find((p) => p.type === type)?.value ?? 0)
  return { isoWeekday: weekday === 0 ? 7 : weekday, minute: get('hour') * 60 + get('minute') }
}

// civilDateTile parte una fecha civil AAAA-MM-DD en el día ("08") y el mes
// abreviado ("DIC") de la ficha de fecha. Formatea en UTC: una fecha civil no
// se reinterpreta por zona (RN-DIS-07).
export function civilDateTile(civilDate: string): { day: string; month: string } {
  if (!isCivilDateString(civilDate)) return { day: '', month: '' }
  const [year, month, day] = civilDate.split('-').map(Number)
  const at = new Date(Date.UTC(year!, month! - 1, day!, 12))
  const monthLabel = new Intl.DateTimeFormat('es-CO', { timeZone: 'UTC', month: 'short' })
    .format(at)
    .replace('.', '')
    .toUpperCase()
  return { day: String(day).padStart(2, '0'), month: monthLabel }
}
