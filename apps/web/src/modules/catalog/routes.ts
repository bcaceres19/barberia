// Ruta hija privada de `catalog`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de `auth`/`settings`/`staff` (y de
// cualquier otro módulo) antes de pasarlas a `auth.privateShellRoute`; este
// módulo no monta su propio cascarón ni repite el guard de sesión (HU-022,
// mismo patrón que HU-020/HU-021).
import type { RouteRecordRaw } from 'vue-router'

export const catalogPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'servicios',
    name: 'catalog-servicios',
    component: () => import('./pages/CatalogPage.vue'),
  },
]
