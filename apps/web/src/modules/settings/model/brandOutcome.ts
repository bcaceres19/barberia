// Resultados discriminados de las dos operaciones de marca y vocabulario
// (issue #292, DEC-110), igual criterio que `settingsOutcome.ts`: ninguno
// expone `Problem`, `status` HTTP crudo ni cabeceras. Un 401 real lo
// intercepta la coordinación única de `installSessionHandling`.
import type { BrandSettings } from '@/shared/model'

export type FetchBrandOutcome =
  | { kind: 'success'; brand: BrandSettings }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SaveBrandOutcome =
  | { kind: 'success'; brand: BrandSettings }
  | { kind: 'validation-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
