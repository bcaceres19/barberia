// Ruta hija privada de `barberServices`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de `auth`/`settings`/`staff`/`catalog`
// (y de cualquier otro módulo) antes de pasarlas a `auth.privateShellRoute`;
// este módulo no monta su propio cascarón ni repite el guard de sesión
// (HU-023, mismo patrón que HU-020/HU-021/HU-022).
import type { RouteRecordRaw } from 'vue-router'

export const barberServicesPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'servicios-por-barbero',
    name: 'barber-services',
    component: () => import('./pages/BarberServicesPage.vue'),
  },
]
