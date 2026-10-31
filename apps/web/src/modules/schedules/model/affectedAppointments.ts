// Turnos que quedan bajo un bloqueo recién creado (HU-069, RN-BLQ-03). El bloqueo
// nunca se rechaza por ellos (CA-042-06); esta lista solo se le muestra al
// barbero para que decida uno por uno. Se calcula en el cliente con la agenda
// diaria que el API ya expone, sin endpoint nuevo.
import { getCivilDateInTimezone, shiftCivilDate } from '@/shared/time/civilDate'

/** Más allá de este número de días civiles no se consulta: se avisa de revisar la agenda. */
export const MAX_AFFECTED_DAYS = 7

export type AffectedAppointment = {
  id: string
  attendeeName: string
  serviceName: string
  startsAt: string
  endsAt: string
}

/** Días civiles (en la zona de la barbería) que toca el intervalo semiabierto [startsAt, endsAt). */
export function civilDaysTouched(startsAt: string, endsAt: string, timeZone: string): string[] {
  const first = getCivilDateInTimezone(timeZone, new Date(startsAt))
  // El fin es exclusivo: un bloqueo que termina justo a medianoche no toca el día siguiente.
  const last = getCivilDateInTimezone(timeZone, new Date(new Date(endsAt).getTime() - 1))
  const days: string[] = []
  for (let day = first; day <= last; day = shiftCivilDate(day, 1)) days.push(day)
  return days
}

/** Solo turnos `confirmed` que se solapan con [startsAt, endsAt); sin duplicados y en orden cronológico. */
export function affectedBetween(
  entries: ReadonlyArray<AffectedAppointment & { status: string }>,
  startsAt: string,
  endsAt: string,
): AffectedAppointment[] {
  const from = new Date(startsAt).getTime()
  const to = new Date(endsAt).getTime()
  const seen = new Set<string>()
  const result: AffectedAppointment[] = []
  for (const entry of entries) {
    if (entry.status !== 'confirmed' || seen.has(entry.id)) continue
    if (new Date(entry.startsAt).getTime() < to && new Date(entry.endsAt).getTime() > from) {
      seen.add(entry.id)
      result.push({
        id: entry.id,
        attendeeName: entry.attendeeName,
        serviceName: entry.serviceName,
        startsAt: entry.startsAt,
        endsAt: entry.endsAt,
      })
    }
  }
  return result.sort((a, b) => a.startsAt.localeCompare(b.startsAt) || a.id.localeCompare(b.id))
}
