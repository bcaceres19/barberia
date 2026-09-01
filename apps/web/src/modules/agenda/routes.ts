// Rutas hijas privadas de `agenda`, cargadas de forma diferida
// (docs/03-desarrollo/estandar-frontend-vue.md §7.8). `app/router/index.ts`
// las combina con las hijas privadas de otros módulos antes de pasarlas a
// `auth.privateShellRoute`; este módulo no monta su propio cascarón ni
// repite el guard de sesión (HU-061, mismo patrón que HU-040/HU-042).
//
// `path: ''` (HU-062) es la raíz del panel privado (`/panel`, name
// 'panel'): reemplaza el marcador de posición de HU-012
// (`auth/pages/PanelPage.vue`, retirado) con la agenda diaria real. `auth`
// ya no registra ninguna hija propia (`privateShellChildRoutes` vacío).
import type { RouteRecordRaw } from 'vue-router'

export const agendaPrivateShellChildRoutes: RouteRecordRaw[] = [
  {
    path: '',
    name: 'panel',
    component: () => import('./pages/DailyAgendaPage.vue'),
  },
  {
    path: 'turnos/nuevo',
    name: 'agenda-nuevo-turno',
    component: () => import('./pages/NewAppointmentPage.vue'),
  },
  {
    path: 'turnos/:appointmentId',
    name: 'agenda-detalle-turno',
    component: () => import('./pages/AppointmentDetailPage.vue'),
  },
]
