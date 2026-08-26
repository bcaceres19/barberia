// Resultados discriminados de las operaciones de HU-042, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6),
// mismo criterio que scheduleOutcome.ts (HU-040/041).
import type { TimeBlock, TimeBlockPage, TimeBlockSeries, TimeBlockSeriesPage } from './timeBlock'

export type FetchTimeBlocksOutcome =
  | { kind: 'success'; page: TimeBlockPage }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type CreateTimeBlockOutcome =
  | { kind: 'success'; block: TimeBlock }
  | { kind: 'validation-error' }
  | { kind: 'not-found' }
  // RN-IDE-01: la misma clave de idempotencia ya se usó con un contenido
  // distinto, o la operación con esa clave sigue en curso (DEC-043). Un
  // bloqueo puntual nunca produce un conflicto de negocio (RN-BLQ-03).
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type DeleteTimeBlockOutcome =
  | { kind: 'success' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchTimeBlockSeriesOutcome =
  | { kind: 'success'; page: TimeBlockSeriesPage }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type CreateTimeBlockSeriesOutcome =
  | { kind: 'success'; series: TimeBlockSeries }
  | { kind: 'validation-error' }
  | { kind: 'not-found' }
  | { kind: 'idempotency-conflict' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type DeleteTimeBlockSeriesOutcome =
  | { kind: 'success' }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
