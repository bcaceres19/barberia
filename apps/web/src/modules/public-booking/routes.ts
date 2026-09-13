// Rutas del módulo public-booking, cargadas de forma diferida (docs/03-
// desarrollo/estandar-frontend-vue.md §7.8) y registradas desde
// `app/router`. Este archivo es la API pública de rutas del módulo:
// `app/router/index.ts` las importa y las compone, sin conocer los
// componentes internos del módulo.
//
// `/reservar/:slug` es la entrada pública (HU-090), `/reservar/:slug/servicios`
// el catálogo público de servicios (HU-091), `/reservar/:slug/servicios/:serviceId/barbero`
// la selección pública de barbero (HU-092) y
// `/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario` la
// exploración pública de fechas y horarios (HU-095): cuatro cascarones
// propios, separados del panel autenticado (`auth.privateShellRoute`) y sin
// el guard `requireSession` -ninguna ruta pública de reserva exige sesión
// (CA-090-03). `props: true` entrega `slug`/`serviceId`/`barberId` como
// props del componente, no vía `useRoute()`: la página no depende del
// router para sus propios parámetros, lo que la hace verificable en
// aislamiento (component tests).
import type { RouteRecordRaw } from 'vue-router'

export const publicBookingRoutes: RouteRecordRaw[] = [
  {
    path: '/reservar/:slug',
    name: 'reserva-publica-entrada',
    component: () => import('./pages/PublicBarbershopEntryPage.vue'),
    props: true,
  },
  {
    path: '/reservar/:slug/servicios',
    name: 'reserva-publica-servicios',
    component: () => import('./pages/PublicServiceCatalogPage.vue'),
    props: true,
  },
  {
    path: '/reservar/:slug/servicios/:serviceId/barbero',
    name: 'reserva-publica-barbero',
    component: () => import('./pages/PublicBarberSelectionPage.vue'),
    props: true,
  },
  {
    path: '/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario',
    name: 'reserva-publica-horario',
    component: () => import('./pages/PublicAvailabilityPage.vue'),
    props: true,
  },
]
