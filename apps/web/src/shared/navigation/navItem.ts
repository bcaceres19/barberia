// Forma mínima y compartida de una entrada de navegación privada. Vive en
// `shared` (no en `auth`, que es dueña de `AppNav`) para que un módulo
// hermano como `settings` pueda producir una entrada compatible sin
// importar internos de `auth` (dirección app → modules → shared, HU-020).
export interface NavItem {
  to: { name: string }
  label: string
}
