// Recuperación cuando falla la descarga diferida de una pantalla (CA-090-05, PENTRY-F-01).
// Todas las rutas se cargan con `import()`: si la conexión cae después de abrir la entrada,
// la importación falla, vue-router aborta la navegación y la persona se quedaba con un botón
// muerto, sin aviso. Además el navegador recuerda el fallo de ese módulo, de modo que volver a
// pulsar con la red restablecida tampoco navegaba hasta recargar con F5.
//
// Aquí se avisa con un aviso recuperable y, al reintentar con red, se navega con una carga
// completa de la dirección (`location.assign`), que sí vuelve a pedir el módulo. No hay
// reintento automático: lo decide la persona.
import type { Router } from 'vue-router'
import { pushToast } from '@/shared/model/toastStore'

const CHUNK_ERROR =
  /dynamically imported module|importing a module script failed|unable to preload/i

/** `true` si el error de navegación viene de no poder descargar el módulo de la pantalla. */
export function isChunkLoadError(error: unknown): boolean {
  if (!(error instanceof Error)) return false
  return error.name === 'ChunkLoadError' || CHUNK_ERROR.test(error.message)
}

export interface ChunkLoadRecoveryOptions {
  /** Carga completa de `href`; se inyecta en las pruebas. */
  navigate?: (href: string) => void
  isOnline?: () => boolean
}

export function installChunkLoadRecovery(
  router: Router,
  {
    navigate = (href) => window.location.assign(href),
    isOnline = () => navigator.onLine,
  }: ChunkLoadRecoveryOptions = {},
): void {
  // Destino cuya descarga falló y aún no se recuperó.
  let failedFullPath: string | null = null

  function warn(fullPath: string) {
    failedFullPath = fullPath
    pushToast({
      variant: 'danger',
      title: 'No pudimos abrir esta pantalla',
      detail: isOnline()
        ? 'La descarga se interrumpió. Inténtalo de nuevo.'
        : 'Sin conexión. Revisa tu red e inténtalo de nuevo.',
      action: { label: 'Reintentar', icon: 'retry', run: () => retry(fullPath) },
    })
  }

  function retry(fullPath: string) {
    if (!isOnline()) {
      warn(fullPath)
      return
    }
    failedFullPath = null
    navigate(router.resolve(fullPath).href)
  }

  router.onError((error, to) => {
    if (isChunkLoadError(error)) warn(to.fullPath)
  })

  // Pulsar de nuevo el mismo destino equivale a Reintentar.
  router.beforeEach((to) => {
    if (failedFullPath === null) return true
    if (to.fullPath !== failedFullPath) {
      failedFullPath = null
      return true
    }
    retry(to.fullPath)
    return false
  })
}
