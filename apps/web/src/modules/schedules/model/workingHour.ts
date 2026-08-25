// Espejo tipado del cuerpo real del contrato (HU-040, CA-040-01/02/03):
// exactamente id, isoWeekday, startsTime, durationMinutes, createdAt,
// updatedAt. barbershopId/barberId nunca aparecen: el tenant se deriva de
// la sesión y el barbero de la ruta.
export interface WorkingHour {
  id: string
  isoWeekday: number
  startsTime: string
  durationMinutes: number
  createdAt: string
  updatedAt: string
}

/** Página paginada por cursor (CA-040-01): mismo `items`/`nextCursor` que
 * WorkingHourListResponse. */
export interface WorkingHourPage {
  items: WorkingHour[]
  nextCursor: string | null
}

// ISO_WEEKDAYS es la lista fija de los siete días (1 = lunes … 7 = domingo,
// DEC-020): la pantalla siempre los muestra en este orden, con o sin
// tramos configurados (CA-040-01).
export const ISO_WEEKDAYS: { value: number; label: string }[] = [
  { value: 1, label: 'Lunes' },
  { value: 2, label: 'Martes' },
  { value: 3, label: 'Miércoles' },
  { value: 4, label: 'Jueves' },
  { value: 5, label: 'Viernes' },
  { value: 6, label: 'Sábado' },
  { value: 7, label: 'Domingo' },
]

export function weekdayLabel(isoWeekday: number): string {
  return ISO_WEEKDAYS.find((d) => d.value === isoWeekday)?.label ?? `Día ${isoWeekday}`
}

// BarberSummary es un subconjunto deliberadamente mínimo del recurso real
// de `staff` (solo id + el nombre visible necesario para el selector),
// obtenido por este módulo directamente del cliente HTTP compartido -nunca
// importando staffApi.ts, privado de su propio módulo
// (docs/03-desarrollo/estandar-frontend-vue.md §3), mismo criterio que
// barberServices.BarberSummary.
export interface BarberSummary {
  id: string
  fullName: string
}
