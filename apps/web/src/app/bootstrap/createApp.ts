import { createApp as createVueApp, watch, type App as VueApp } from 'vue'
import App from '@/App.vue'
import { router } from '@/app/router'
import { installSessionHandling, sessionStatus } from '@/modules/auth'
import { loadWorkspaceBrand } from '@/modules/settings'
import { resetBrand } from '@/shared/model'

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
  // DEC-110: la marca y el vocabulario de la barbería viajan con la sesión. Se
  // piden al autenticarse y se descartan al salir, para que la siguiente
  // barbería nunca herede el acento ni las palabras de la anterior.
  watch(
    sessionStatus,
    (status) => {
      if (status === 'authenticated') void loadWorkspaceBrand()
      else if (status === 'unauthenticated') resetBrand()
    },
    { immediate: true },
  )
  return app
}
