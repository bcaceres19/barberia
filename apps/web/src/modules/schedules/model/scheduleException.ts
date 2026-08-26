// Espejo tipado del cuerpo real del contrato (HU-041, CA-041-04/05):
// exactamente id, effectiveDate, isClosed, reason, segments, createdAt,
// updatedAt. barbershopId/barberId nunca aparecen: el tenant se deriva de
// la sesión y el barbero de la ruta.
export interface ScheduleExceptionSegment {
  id: string
  startsTime: string
  durationMinutes: number
}

export interface ScheduleException {
  id: string
  effectiveDate: string
  isClosed: boolean
  reason: string | null
  segments: ScheduleExceptionSegment[]
  createdAt: string
  updatedAt: string
}

/** Página paginada por cursor (CA-041-04): mismo `items`/`nextCursor` que
 * ScheduleExceptionListResponse. */
export interface ScheduleExceptionPage {
  items: ScheduleException[]
  nextCursor: string | null
}

// ColombianHoliday es el dato de referencia calendárica de HU-041: exacto,
// fecha civil y nombre, igual para toda barbería y todo barbero.
export interface ColombianHoliday {
  date: string
  name: string
}
