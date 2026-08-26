// Conversión entre hora civil (lo que el barbero escribe/lee) e instante
// absoluto (lo que time_block.startsAt/endsAt almacena), SIEMPRE en la zona
// IANA de la barbería, nunca la del dispositivo (RN-DIS-05, RN-DIS-07,
// CA-042). Los bloqueos puntuales son la única parte de `schedules` que
// necesita esta conversión: working_hour/schedule-exceptions trabajan con
// hora civil "HH:MM" pura, sin fecha, porque son recurrentes; un bloqueo
// puntual tiene fecha Y hora, así que el servidor exige un instante RFC
// 3339 con offset explícito.

// civilDateTimeToInstant convierte una fecha civil (YYYY-MM-DD) y una hora
// civil (HH:MM) EN timeZone a un instante absoluto ISO 8601 (siempre en
// UTC, sufijo "Z"). Usa el truco estándar de doble formateo con
// Intl.DateTimeFormat: no depende de ninguna tabla de zonas horaria
// embebida, solo de los datos ICU que el runtime ya trae.
export function civilDateTimeToInstant(date: string, time: string, timeZone: string): string {
  const guessUtcMs = Date.parse(`${date}T${time}:00Z`)
  if (Number.isNaN(guessUtcMs)) {
    throw new RangeError(`fecha u hora civil inválida: ${date}T${time}`)
  }

  const shownMs = wallClockMsInZone(new Date(guessUtcMs), timeZone)
  const offsetMs = shownMs - guessUtcMs
  const actualUtcMs = guessUtcMs - offsetMs
  return new Date(actualUtcMs).toISOString()
}

// formatInstantInTimezone es la conversión inversa para MOSTRAR un instante
// ya almacenado: no necesita el truco de doble formateo, Intl ya sabe hacer
// esto directamente.
export function formatInstantInTimezone(
  instant: string,
  timeZone: string,
): { date: string; time: string } {
  const dtf = new Intl.DateTimeFormat('en-CA', {
    timeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  })
  const parts = dtf.formatToParts(new Date(instant))
  const get = (type: string) => parts.find((p) => p.type === type)?.value ?? '00'
  return {
    date: `${get('year')}-${get('month')}-${get('day')}`,
    time: `${get('hour')}:${get('minute')}`,
  }
}

// wallClockMsInZone interpreta instant EN timeZone y devuelve el mismo
// valor de reloj (año/mes/día/hora/minuto/segundo) codificado como si esos
// componentes fueran UTC. Es un valor auxiliar de comparación, no un
// instante real.
function wallClockMsInZone(instant: Date, timeZone: string): number {
  const dtf = new Intl.DateTimeFormat('en-US', {
    timeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  })
  const parts = dtf.formatToParts(instant)
  const get = (type: string) => parts.find((p) => p.type === type)?.value ?? '00'
  return Date.UTC(
    Number(get('year')),
    Number(get('month')) - 1,
    Number(get('day')),
    Number(get('hour')),
    Number(get('minute')),
    Number(get('second')),
  )
}
