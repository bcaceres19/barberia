// Fecha civil «AAAA-MM-DD» en palabras («5 ene 2099»), sin zona del
// dispositivo (RN-DIS-07): se formatea en UTC a mediodía.
export function displayCivilDate(date: string): string {
  return new Intl.DateTimeFormat('es-CO', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(`${date}T12:00:00Z`))
}
