import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { authRoutes } from '@/modules/auth'

// Cada ruta se carga de forma diferida (docs/04-arquitectura/frontend.md):
// un módulo futuro no aumenta el bundle inicial sin necesidad. `app/router`
// compone las rutas que cada módulo expone en su índice público; no conoce
// los componentes internos de ningún módulo (estandar-frontend-vue.md §3).
const routes: RouteRecordRaw[] = [
  {
    // HU-010 entrega la primera pantalla real: la raíz redirige al acceso
    // en vez de conservar la pantalla provisional de arranque.
    path: '/',
    redirect: { name: 'acceso' },
  },
  ...authRoutes,
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})
