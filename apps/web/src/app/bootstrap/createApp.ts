import { createApp as createVueApp, type App as VueApp } from 'vue'
import App from '@/App.vue'
import { router } from '@/app/router'
import { installSessionHandling } from '@/modules/auth'

/**
 * Construye la instancia de Vue con sus proveedores globales (router, y los
 * que se agreguen con una necesidad real). No contiene reglas de negocio:
 * eso vive en los módulos, según docs/04-arquitectura/frontend.md.
 */
export function createApplication(): VueApp {
  const app = createVueApp(App)
  app.use(router)
  // HU-012: coordinación única de 401 sobre rutas privadas. Se conecta aquí
  // (no dentro de `modules/auth`) porque es el único punto que posee la
  // instancia real del router.
  installSessionHandling(router)
  return app
}
