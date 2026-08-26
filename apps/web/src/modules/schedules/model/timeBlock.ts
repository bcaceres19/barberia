// Espejo tipado del cuerpo real del contrato (HU-042): exactamente id,
// blockType, source, startsAt, endsAt, reason, deletedAt, deletedBy,
// createdAt, updatedAt. barbershopId/barberId nunca aparecen: el tenant se
// deriva de la sesión y el barbero de la ruta.
export interface TimeBlock {
  id: string
  blockType: BlockType
  source: 'manual' | 'holiday_calendar'
  startsAt: string
  endsAt: string
  reason: string | null
  deletedAt: string | null
  deletedBy: string | null
  createdAt: string
  updatedAt: string
}

export interface TimeBlockPage {
  items: TimeBlock[]
  nextCursor: string | null
}

// BLOCK_TYPES es el vocabulario cerrado de los siete tipos (RN-BLQ-01,
// criterio de salida de B2): la pantalla siempre ofrece esta misma lista,
// nunca un texto libre.
export type BlockType =
  'break' | 'lunch' | 'unavailable' | 'day_off' | 'holiday' | 'vacation' | 'emergency'

export const BLOCK_TYPES: { value: BlockType; label: string }[] = [
  { value: 'break', label: 'Descanso' },
  { value: 'lunch', label: 'Almuerzo' },
  { value: 'unavailable', label: 'No disponible' },
  { value: 'day_off', label: 'Día libre' },
  { value: 'holiday', label: 'Festivo' },
  { value: 'vacation', label: 'Vacaciones' },
  { value: 'emergency', label: 'Emergencia' },
]

export function blockTypeLabel(blockType: BlockType): string {
  return BLOCK_TYPES.find((t) => t.value === blockType)?.label ?? blockType
}

// Definición recurrente (weekly o date_list, DEC-020). ISOWeekday solo
// aparece en weekly; dates solo se completa en date_list.
export interface TimeBlockSeries {
  id: string
  blockType: BlockType
  recurrenceKind: 'weekly' | 'date_list'
  isoWeekday: number | null
  startsTime: string
  durationMinutes: number
  effectiveFrom: string
  effectiveUntil: string | null
  reason: string | null
  deletedAt: string | null
  createdAt: string
  updatedAt: string
  dates: { blockDate: string }[]
  exceptions: { excludedDate: string; reason: string | null; createdAt: string }[]
}

export interface TimeBlockSeriesPage {
  items: TimeBlockSeries[]
  nextCursor: string | null
}
