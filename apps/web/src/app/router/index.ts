import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { agendaNavItems, agendaPrivateShellChildRoutes } from '@/modules/agenda'
import { authRoutes, privateShellChildRoutes, privateShellRoute } from '@/modules/auth'
import {
  barberServicesNavItems,
  barberServicesPrivateShellChildRoutes,
} from '@/modules/barberServices'
import { catalogNavItems, catalogPrivateShellChildRoutes } from '@/modules/catalog'
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
      ...settingsNavItems,
      ...staffNavItems,
      ...catalogNavItems,
      ...barberServicesNavItems,
      ...schedulesNavItems,
      ...agendaNavItems,
    ],
  ),
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})
