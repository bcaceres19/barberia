// Punto de instalación único, llamado una sola vez desde
// app/bootstrap/createApp.ts (trabajo requerido §5, "coordinación única de
// 401"). Vive en `auth` -no en `shared`- porque decide sobre el dominio de
// sesión; recibe el router ya construido por `app` en vez de importar una
// instancia propia, para no invertir la dirección app → modules → shared.
import type { Router } from 'vue-router'
import { httpClient } from '@/shared/api/httpClient'
import { onUnauthorized, reportUnauthorized } from '../model/sessionStore'

// El bootstrap de sesión (GET /private/auth/session) ya interpreta su
// propio 401 en sessionStore.ensureBootstrapped; excluirlo aquí evita una
// segunda redirección compitiendo con la que el guard de navegación ya está
// resolviendo para esa misma solicitud.
const SESSION_CONTEXT_PATH = '/private/auth/session'

export function installSessionHandling(router: Router): void {
  httpClient.use({
    onResponse({ request, response }) {
      const isPrivateRoute = new URL(request.url).pathname.includes('/api/v1/private/')
      const isBootstrapCall = request.url.includes(SESSION_CONTEXT_PATH)
      if (response.status === 401 && isPrivateRoute && !isBootstrapCall) {
        reportUnauthorized()
      }
      return undefined
    },
  })

  onUnauthorized(() => {
    if (router.currentRoute.value.name === 'acceso') return
    void router.push({ name: 'acceso', query: { motivo: 'sesion-expirada' } })
  })
}
