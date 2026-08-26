// Ruta hija privada de `schedules`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de otros módulos antes de pasarlas a
// `auth.privateShellRoute`; este módulo no monta su propio cascarón ni
// repite el guard de sesión (HU-040, mismo patrón que HU-021/HU-023).
import type { RouteRecordRaw } from 'vue-router'

export const schedulesPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'horarios',
    name: 'schedules-horarios',
    component: () => import('./pages/SchedulesPage.vue'),
  },
  {
    path: 'bloqueos',
    name: 'schedules-bloqueos',
    component: () => import('./pages/BlocksPage.vue'),
  },
]
