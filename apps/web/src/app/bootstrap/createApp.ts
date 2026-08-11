import { createApp as createVueApp, type App as VueApp } from 'vue'
import App from '@/App.vue'
import { router } from '@/app/router'

/**
 * Construye la instancia de Vue con sus proveedores globales (router, y los
 * que se agreguen con una necesidad real). No contiene reglas de negocio:
 * eso vive en los módulos, según docs/04-arquitectura/frontend.md.
 */
export function createApplication(): VueApp {
  const app = createVueApp(App)
  app.use(router)
  return app
}
