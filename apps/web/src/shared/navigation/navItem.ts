// Forma mínima y compartida de una entrada de navegación privada. Vive en
// `shared` (no en `auth`, que es dueña de `AppNav`) para que un módulo
// hermano como `settings` pueda producir una entrada compatible sin
// importar internos de `auth` (dirección app → modules → shared, HU-020).
import type { Vocabulary } from '@/shared/model'

export interface NavItem {
  to: { name: string }
  /** Rótulo por defecto; también identifica la entrada. */
  label: string
  /** Rótulo según el vocabulario de la barbería (DEC-110): "Barberos" se
   * lee "Estilistas" cuando la barbería así lo configura. */
  labelFor?: (vocabulary: Vocabulary) => string
  /** Visible siempre en el dock. Sin ella, la entrada solo aparece dentro
   * de "Más" (estandar-diseno-visual.md §6.5: cuatro destinos visibles y
   * "Más" para el resto). */
  primary?: boolean
}
