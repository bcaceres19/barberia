// Valida el destino pretendido que el guard conserva al redirigir a
// /acceso (CA-012-02) y que LoginPage recupera después de autenticarse.
// Solo una ruta interna con forma segura autoriza la redirección: nunca un
// destino externo (protocolo, `//host`) ni un ciclo hacia /acceso o
// /recuperar-acceso (trabajo requerido §3).
const CYCLE_PATHS = new Set(['/acceso', '/recuperar-acceso'])

export function isSafeInternalRedirect(path: unknown): path is string {
  if (typeof path !== 'string' || path === '') return false
  if (!path.startsWith('/') || path.startsWith('//')) return false
  if (path.includes('://')) return false
  const withoutQuery = path.split(/[?#]/)[0]
  if (CYCLE_PATHS.has(withoutQuery)) return false
  return true
}
