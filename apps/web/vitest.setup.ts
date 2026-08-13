// Registra el matcher toHaveNoViolations (vitest-axe) para toda la suite
// de componentes. Ver docs/03-desarrollo/estrategia-pruebas.md §5.4 y
// docs/03-desarrollo/estandar-diseno-visual.md §16: verificación accesible
// automatizada, herramienta elegida y justificada en la auditoría HU-009.
//
// vitest-axe@0.1.0 publica dist/extend-expect.js vacío (build incompleto de
// esa versión): su registro automático por efecto secundario no funciona.
// Se registra el matcher a mano contra el mismo expect que importan las
// pruebas, que es exactamente lo que ese archivo debería hacer.
import { expect } from 'vitest'
import { toHaveNoViolations } from 'vitest-axe/matchers'

expect.extend({ toHaveNoViolations })
