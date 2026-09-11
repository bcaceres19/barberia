// Rutas del módulo public-booking, cargadas de forma diferida (docs/03-
// desarrollo/estandar-frontend-vue.md §7.8) y registradas desde
// `app/router`. Este archivo es la API pública de rutas del módulo:
// `app/router/index.ts` las importa y las compone, sin conocer los
// componentes internos del módulo.
//
// `/reservar/:slug` es la entrada pública (HU-090): un cascarón propio,
// separado del panel autenticado (`auth.privateShellRoute`) y sin el guard
// `requireSession` -esta ruta nunca exige sesión (CA-090-03). `props: true`
// entrega `slug` como prop del componente, no vía `useRoute()`: la página
// no depende del router para su propio parámetro, lo que la hace
// verificable en aislamiento (component tests).
import type { RouteRecordRaw } from 'vue-router'

export const publicBookingRoutes: RouteRecordRaw[] = [
  {
    path: '/reservar/:slug',
    name: 'reserva-publica-entrada',
    component: () => import('./pages/PublicBarbershopEntryPage.vue'),
    props: true,
  },
]
