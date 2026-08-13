// Amplía los matchers de Vitest con toHaveNoViolations (vitest-axe), que se
// registra en tiempo de ejecución vía vitest.setup.ts. Sin este archivo,
// TypeScript no reconoce el matcher aunque el runtime sí lo tenga.
import type { AxeMatchers } from 'vitest-axe'

declare module 'vitest' {
  // El cuerpo vacío es el patrón estándar de augmentación de interfaz de
  // TypeScript (declaration merging): extends aporta los miembros, no hay
  // nada que agregar aquí. La regla no distingue ese caso del real.
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface Assertion extends AxeMatchers {}
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface AsymmetricMatchersContaining extends AxeMatchers {}
}
