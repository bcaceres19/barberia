// Forma mínima y compartida de una entrada de navegación privada. Vive en
// `shared` (no en `auth`, que es dueña de `AppNav`) para que un módulo
// hermano como `settings` pueda producir una entrada compatible sin
// importar internos de `auth` (dirección app → modules → shared, HU-020).
import type { PanelProfile, Vocabulary } from '@/shared/model'

/** Cómo cambia una entrada en un perfil del panel distinto del completo (DEC-115). */
export interface NavItemProfileOverride {
  /** La entrada no aparece en ese perfil (su ruta sigue existiendo). */
  hidden?: boolean
  /** Rótulo propio de ese perfil; prevalece sobre `label` y `labelFor`. */
  label?: string
  /** Sustituye a `primary` en ese perfil (dock frente a "Más"). */
  primary?: boolean
}

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
  /** Ajustes por perfil del panel; sin entrada para un perfil, se usa la declaración base
   * (`shop` es siempre la base: el panel completo de siempre). */
  profiles?: Partial<Record<PanelProfile, NavItemProfileOverride>>
}
