// photoUrl: fotografía del barbero (DEC-104) o null/ausente si no tiene; sin ella
// se muestra su monograma.
export type BarberSummary = { id: string; fullName: string; photoUrl?: string | null }
export type ServiceSummary = { id: string; name: string }
