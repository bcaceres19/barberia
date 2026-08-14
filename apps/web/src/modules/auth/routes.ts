// Rutas del módulo auth, cargadas de forma diferida (docs/03-desarrollo/
// estandar-frontend-vue.md §7.8) y registradas desde `app/router`
// (trabajo requerido §1). Este archivo es la API pública de rutas del
// módulo: `app/router/index.ts` las importa y las compone, sin conocer
// los componentes internos de `auth`.
import type { RouteRecordRaw } from 'vue-router'
import { requireSession } from './guards/requireSession'

export const authRoutes: RouteRecordRaw[] = [
  {
    path: '/acceso',
    name: 'acceso',
    component: () => import('./pages/LoginPage.vue'),
  },
  {
    // Ruta privada real de CA-010-01/DEC-056. El guard es deliberadamente
    // mínimo: HU-012 lo reutiliza y lo generaliza, no crea uno paralelo.
    path: '/panel',
    name: 'panel',
    component: () => import('./pages/PanelPlaceholderPage.vue'),
    beforeEnter: requireSession,
  },
  {
    // Destino provisional del enlace de recuperación de CA-010-08,
    // documentado en `DP-UX-06` (docs/00-control/dudas-pendientes.md).
    path: '/recuperar-acceso',
    name: 'recuperar-acceso',
    component: () => import('./pages/RecoveryPendingPage.vue'),
  },
]
