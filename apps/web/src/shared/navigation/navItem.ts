// Forma mínima y compartida de una entrada de navegación privada. Vive en
// `shared` (no en `auth`, que es dueña de `AppNav`) para que un módulo
// hermano como `settings` pueda producir una entrada compatible sin
// importar internos de `auth` (dirección app → modules → shared, HU-020).
export interface NavItem {
  to: { name: string }
  label: string
  /** Visible siempre en el dock. Sin ella, la entrada solo aparece dentro
   * de "Más" (estandar-diseno-visual.md §6.5: cuatro destinos visibles y
   * "Más" para el resto). */
  primary?: boolean
}
