// Guard generalizado de todas las rutas privadas (HU-012, DEC-056): un solo
// camino, reutilizado y no duplicado por cada ruta privada futura. A
// diferencia del guard mínimo de HU-010 (marcador local en sessionStorage,
// DP-SEG-08), este consulta la fuente autoritativa real
// (GET /private/auth/session, DEC-060) antes de decidir.
import type { NavigationGuardWithThis } from 'vue-router'
import { ensureBootstrapped, sessionState } from '../model/sessionStore'
import { isSafeInternalRedirect } from '../model/redirectTarget'

export const requireSession: NavigationGuardWithThis<undefined> = async (to) => {
  await ensureBootstrapped()

  if (sessionState.bootstrap.status === 'authenticated') return true

  if (sessionState.bootstrap.status === 'unauthenticated') {
    // Conserva el destino pretendido (CA-012-02) solo si es una ruta
    // interna segura; nunca un destino externo ni un ciclo hacia
    // acceso/logout (trabajo requerido §3).
    const redirect = isSafeInternalRedirect(to.fullPath) ? to.fullPath : undefined
    return { name: 'acceso', query: redirect ? { redirect } : undefined }
  }

  // `connection-lost`: deja entrar a la ruta privada. El cascarón (HU-012)
  // renderiza el estado recuperable con "Reintentar" ahí mismo, sin bucle
  // de redirección ni pantalla en blanco (CA-012-05).
  return true
}
