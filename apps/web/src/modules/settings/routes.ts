// Ruta hija privada de `settings`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de `auth` (y de cualquier otro módulo)
// antes de pasarlas a `auth.privateShellRoute`; este módulo no monta su
// propio cascarón ni repite el guard de sesión (HU-020, trabajo requerido
// §4.2).
import type { RouteRecordRaw } from 'vue-router'

export const settingsPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'barberia',
    name: 'configuracion-barberia',
    component: () => import('./pages/SettingsPage.vue'),
  },
]
