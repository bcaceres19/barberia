// Ruta hija privada de `staff`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de `auth`/`settings` (y de cualquier
// otro módulo) antes de pasarlas a `auth.privateShellRoute`; este módulo no
// monta su propio cascarón ni repite el guard de sesión (HU-021, trabajo
// requerido §4.2, mismo patrón que HU-020).
import type { RouteRecordRaw } from 'vue-router'

export const staffPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'barberos',
    name: 'staff-barberos',
    component: () => import('./pages/StaffPage.vue'),
  },
]
