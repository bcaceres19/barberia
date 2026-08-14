// Guard mínimo de `/panel` (DEC-056): sin sesión válida redirige al
// acceso; con sesión válida deja pasar. HU-012 generaliza este guard a
// todas las rutas privadas y agrega conservar el destino pretendido — esta
// historia no lo hace todavía (ver la fila de HU-012 en
// docs/02-requisitos/historias-usuario.md).
import type { NavigationGuardWithThis } from 'vue-router'
import { hasRememberedSession } from '../model/sessionMarker'

export const requireSession: NavigationGuardWithThis<undefined> = () => {
  if (hasRememberedSession()) return true
  return { name: 'acceso' }
}
