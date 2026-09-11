import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { agendaNavItems, agendaPrivateShellChildRoutes } from '@/modules/agenda'
import { authRoutes, privateShellChildRoutes, privateShellRoute } from '@/modules/auth'
import {
  barberServicesNavItems,
  barberServicesPrivateShellChildRoutes,
} from '@/modules/barberServices'
import { catalogNavItems, catalogPrivateShellChildRoutes } from '@/modules/catalog'
import { publicBookingRoutes } from '@/modules/public-booking'
import { schedulesNavItems, schedulesPrivateShellChildRoutes } from '@/modules/schedules'
import { settingsNavItems, settingsPrivateShellChildRoutes } from '@/modules/settings'
import { staffNavItems, staffPrivateShellChildRoutes } from '@/modules/staff'

// Cada ruta se carga de forma diferida (docs/04-arquitectura/frontend.md):
// un módulo futuro no aumenta el bundle inicial sin necesidad. `app/router`
// compone las rutas que cada módulo expone en su índice público; no conoce
// los componentes internos de ningún módulo (estandar-frontend-vue.md §3).
//
// El único cascarón privado (HU-012) se ensambla aquí (HU-020 en
// adelante): `privateShellRoute` recibe las hijas y las entradas de
// navegación combinadas de `auth` y de cada módulo con pantalla privada,
// sin que ningún módulo importe a otro (app → modules → shared).
const routes: RouteRecordRaw[] = [
  {
    // HU-010 entrega la primera pantalla real: la raíz redirige al acceso
    // en vez de conservar la pantalla provisional de arranque.
    path: '/',
    redirect: { name: 'acceso' },
  },
  ...authRoutes,
  // HU-090: entrada pública de reservas, sin sesión del barbero y aislada
  // del cascarón privado (CA-090-03: nunca pasa por requireSession).
  ...publicBookingRoutes,
  privateShellRoute(
    [
      ...privateShellChildRoutes,
      ...settingsPrivateShellChildRoutes,
      ...staffPrivateShellChildRoutes,
      ...catalogPrivateShellChildRoutes,
      ...barberServicesPrivateShellChildRoutes,
      ...schedulesPrivateShellChildRoutes,
      ...agendaPrivateShellChildRoutes,
    ],
    [
      // Orden de los 5 destinos P0 del dock (especificacion-frontend-nava.md
      // §5.1): Agenda (entrada base de AppNav) · Servicios · Barberos ·
      // Horarios · Configuración. barberServicesNavItems aporta un sexto
      // ítem temporal ("Servicios por barbero", ver su propio módulo) hasta
      // que la Fase 5 lo fusione dentro de "Servicios"; agendaNavItems
      // aporta un arreglo vacío desde la Fase 2 (issue #135, "Nuevo turno"
      // es la acción global del header, no un destino) y se conserva en el
      // spread para no crear un caso especial si vuelve a aportar algo.
      ...catalogNavItems,
      ...staffNavItems,
      ...schedulesNavItems,
      ...settingsNavItems,
      ...barberServicesNavItems,
      ...agendaNavItems,
    ],
  ),
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})
