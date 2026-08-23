// Resultados discriminados de las dos operaciones de HU-020, tal como los
// recibe la página coordinadora. Ninguno expone `Problem`, `status` HTTP
// crudo ni cabeceras (docs/03-desarrollo/estandar-frontend-vue.md §6). Un
// 401 real de estas rutas ya lo intercepta la coordinación única de
// `installSessionHandling` (redirige a acceso); ambos tipos igual declaran
// `unexpected-error` como frontera defensiva por si esa solicitud concreta
// llegara a resolverse antes de que la redirección navegue.
import type { BarbershopSettings } from './barbershopSettings'

export type FetchBarbershopSettingsOutcome =
  | { kind: 'success'; settings: BarbershopSettings }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type SaveBarbershopSettingsOutcome =
  | { kind: 'success'; settings: BarbershopSettings }
  | { kind: 'validation-error' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
