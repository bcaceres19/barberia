// Rutas del módulo customer-access, cargadas de forma diferida (docs/03-
// desarrollo/estandar-frontend-vue.md §7.8) y registradas desde
// `app/router`. Este archivo es la API pública de rutas del módulo:
// `app/router/index.ts` las importa y las compone, sin conocer los
// componentes internos del módulo.
//
// `/mi-turno/:token` es la lectura del turno del cliente por la credencial
// del enlace (HU-098): un cascarón propio, separado del panel autenticado
// (`auth.privateShellRoute`) y del cascarón público de reserva
// (`public-booking.routes`), y sin el guard `requireSession` -ninguna ruta
// de esta audiencia exige sesión del barbero (CA-098-01). La ruta coincide
// EXACTAMENTE con el enlace que el backend construye
// (`apps/api/internal/modules/publicbooking/confirm.go`,
// `webBaseURL + "/mi-turno/" + tokenPlain`, DEC-091): un cambio aquí sin el
// mismo cambio allá rompe todo correo de confirmación ya enviado.
// `props: true` entrega `token` como prop del componente, no vía
// `useRoute()`: la página no depende del router para su propio parámetro,
// lo que la hace verificable en aislamiento (component tests).
import type { RouteRecordRaw } from 'vue-router'

export const customerAccessRoutes: RouteRecordRaw[] = [
  {
    path: '/mi-turno/:token',
    name: 'mi-turno',
    component: () => import('./pages/CustomerAppointmentPage.vue'),
    props: true,
  },
]
