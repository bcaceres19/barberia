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
    // Cascarón privado (HU-012, DEC-056): un único guard generalizado sobre
    // esta ruta padre cubre toda ruta privada hija, sin repetirlo por
    // pantalla. `/panel` es hoy la única hija real; las historias
    // siguientes agregan las suyas aquí mismo, no un árbol paralelo.
    path: '/panel',
    component: () => import('./layouts/PrivateShell.vue'),
    beforeEnter: requireSession,
    children: [
      {
        path: '',
        name: 'panel',
        component: () => import('./pages/PanelPage.vue'),
      },
    ],
  },
  {
    // Destino provisional del enlace de recuperación de CA-010-08,
    // documentado en `DP-UX-06` (docs/00-control/dudas-pendientes.md).
    path: '/recuperar-acceso',
    name: 'recuperar-acceso',
    component: () => import('./pages/RecoveryPendingPage.vue'),
  },
]
