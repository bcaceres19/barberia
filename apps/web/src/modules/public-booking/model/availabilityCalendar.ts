// Agrupación por día civil de los inicios públicos de HU-094 (HU-095,
// CA-095-01). El contrato de disponibilidad no acepta una fecha como
// parámetro: devuelve, en una sola respuesta, todos los inicios válidos de
// la ventana pública vigente (anticipación/ventana/rejilla ya resueltas en
// el servidor, HU-093). Por eso la exploración por fecha es enteramente
// una agrupación en el cliente sobre datos ya recibidos, nunca una consulta
// nueva por cada fecha visitada.
import { getCivilDateInTimezone } from '@/shared/time/civilDate'
import type { PublicAvailabilitySlot } from './publicAvailabilityOutcome'

/** Un día civil con al menos un inicio público válido. Un día sin ningún
 * slot (festivo, fuera de jornada, completamente ocupado) simplemente no
 * genera una entrada aquí: la ventana navegable se deriva por completo de
 * los datos reales que devolvió el servidor, nunca de un cálculo propio de
 * anticipación/ventana en el cliente (RN-DIS-04 vive en el servidor). */
export interface AvailabilityDay {
  civilDate: string
  slots: PublicAvailabilitySlot[]
}

/** Agrupa `slots` (ya ordenados cronológicamente por el servidor, HU-094)
 * por día civil en `timezone` (RN-DIS-07: siempre la de la barbería, nunca
 * la del dispositivo). El orden de inserción de `Map` conserva el orden
 * cronológico de los días sin un `sort` adicional, porque `slots` ya
 * llega ordenado. */
export function groupSlotsByCivilDate(
  slots: PublicAvailabilitySlot[],
  timezone: string,
): AvailabilityDay[] {
  const byDay = new Map<string, PublicAvailabilitySlot[]>()
  for (const slot of slots) {
    const civilDate = getCivilDateInTimezone(timezone, new Date(slot.startsAt))
    const bucket = byDay.get(civilDate)
    if (bucket) {
      bucket.push(slot)
    } else {
      byDay.set(civilDate, [slot])
    }
  }
  return Array.from(byDay.entries(), ([civilDate, daySlots]) => ({ civilDate, slots: daySlots }))
}

/** Partes cortas de un día civil para la tira de fechas: «lun», «15» y
 * «sep». Como `formatCivilDateFull`, se ancla a mediodía UTC y se formatea
 * en UTC: una fecha civil ya resuelta nunca se reinterpreta por la zona de
 * la barbería ni por la del dispositivo (RN-DIS-07). */
export interface CivilDateChip {
  weekday: string
  day: string
  month: string
}

export function civilDateChip(civilDate: string): CivilDateChip {
  const [year, month, day] = civilDate.split('-').map(Number)
  const at = new Date(Date.UTC(year!, month! - 1, day!, 12))
  const parts = new Intl.DateTimeFormat('es-CO', {
    timeZone: 'UTC',
    weekday: 'short',
    day: 'numeric',
    month: 'short',
  }).formatToParts(at)
  const pick = (type: Intl.DateTimeFormatPartTypes) =>
    (parts.find((part) => part.type === type)?.value ?? '').replace(/\.$/, '')
  return { weekday: pick('weekday'), day: pick('day'), month: pick('month') }
}
