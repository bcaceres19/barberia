// Resumen de «mi día» para quien trabaja solo (DEC-115): cuántos turnos le
// quedan y cuál es el siguiente o el que está en curso. Se deriva de los turnos
// que la agenda ya cargó; no consulta nada nuevo ni cambia ningún dato.
import type { DailyAgendaEntry } from './dailyAgenda'

export interface DaySummary {
  /** Turnos del día, en cualquier estado. */
  total: number
  /** Confirmados que aún no terminaron (el que está en curso incluido). */
  remaining: number
  /** Turno confirmado en curso en `now`, si lo hay. */
  inProgress: DailyAgendaEntry | null
  /** Primer confirmado que todavía no empezó en `now`, si lo hay. */
  next: DailyAgendaEntry | null
}

/**
 * Resume el día de `entries` respecto a `now` (milisegundos). Solo cuentan los
 * turnos `confirmed`: completados, cancelados y no presentados ya no ocupan lo que
 * resta del día. Un confirmado que ya pasó de hora sin cerrarse no suma a lo que
 * queda ni a «siguiente».
 */
export function summarizeDay(entries: readonly DailyAgendaEntry[], now: number): DaySummary {
  const confirmed = entries
    .filter((entry) => entry.status === 'confirmed')
    .sort((a, b) => new Date(a.startsAt).getTime() - new Date(b.startsAt).getTime())

  const remainingEntries = confirmed.filter((entry) => new Date(entry.endsAt).getTime() > now)
  const inProgress =
    remainingEntries.find(
      (entry) =>
        new Date(entry.startsAt).getTime() <= now && now < new Date(entry.endsAt).getTime(),
    ) ?? null
  const next = remainingEntries.find((entry) => new Date(entry.startsAt).getTime() > now) ?? null

  return { total: entries.length, remaining: remainingEntries.length, inProgress, next }
}
