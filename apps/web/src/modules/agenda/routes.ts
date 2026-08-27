// Ruta hija privada de `agenda`, cargada de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// la combina con las hijas privadas de otros módulos antes de pasarlas a
// `auth.privateShellRoute`; este módulo no monta su propio cascarón ni
// repite el guard de sesión (HU-061, mismo patrón que HU-040/HU-042).
import type { RouteRecordRaw } from 'vue-router'

export const agendaPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: 'turnos/nuevo',
    name: 'agenda-nuevo-turno',
    component: () => import('./pages/NewAppointmentPage.vue'),
  },
]
