import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

// Cada ruta se carga de forma diferida, incluida esta pantalla provisional,
// para que un módulo futuro no aumente el bundle inicial sin necesidad,
// según docs/04-arquitectura/frontend.md.
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'inicio',
    component: () => import('@/app/BaseStatusView.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})
