// Rutas del módulo public-booking, cargadas de forma diferida (docs/03-
// desarrollo/estandar-frontend-vue.md §7.8) y registradas desde
// `app/router`. Este archivo es la API pública de rutas del módulo:
// `app/router/index.ts` las importa y las compone, sin conocer los
// componentes internos del módulo.
//
// `/reservar/:slug` es la entrada pública (HU-090), `/reservar/:slug/servicios`
// el catálogo público de servicios (HU-091), `/reservar/:slug/servicios/:serviceId/barbero`
// la selección pública de barbero (HU-092),
// `/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario` la
// exploración pública de fechas y horarios (HU-095) y
// `/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente`
// la captura de datos del cliente (HU-096): cinco cascarones propios,
// separados del panel autenticado (`auth.privateShellRoute`) y sin el guard
// `requireSession` -ninguna ruta pública de reserva exige sesión
// (CA-090-03). `props: true` entrega `slug`/`serviceId`/`barberId`/
// `startsAt` como props del componente, no vía `useRoute()`: la página no
// depende del router para sus propios parámetros, lo que la hace
// verificable en aislamiento (component tests). `startsAt` viaja en la URL
// sin interpretarse en esta página (HU-096 no persiste ni revalida la
// franja, HU-097 lo hará): solo se conserva para que la confirmación
// pública, cuando exista, reciba el contexto completo sin pedirlo de nuevo.
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
  {
    path: '/reservar/:slug/servicios/:serviceId/barbero/:barberId/horario/:startsAt/cliente',
    name: 'reserva-publica-cliente',
    component: () => import('./pages/PublicCustomerDetailsPage.vue'),
    props: true,
  },
]
