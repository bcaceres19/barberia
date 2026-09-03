// Cierre de sesión compartido (HU-012, CA-012-07): extraído de AppHeader en
// la Fase 2 del rediseño NAVA (issue #187) para que el menú "Más" del dock
// pueda ofrecer la misma acción en móvil sin duplicar la lógica de doble
// envío, revocación de sesión y navegación.
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { httpClient } from '@/shared/api/httpClient'
import { resetForFreshLogin } from './sessionStore'

export function useLogout() {
  const router = useRouter()
  const loggingOut = ref(false)

  async function logout() {
    // Guardia de doble envío: la asignación síncrona bloquea un segundo
    // clic mientras la solicitud está en curso.
    if (loggingOut.value) return
    loggingOut.value = true

    try {
      await httpClient.POST('/private/auth/logout')
    } catch {
      // Un fallo de red ambiguo no debe fingir éxito ni bloquear el cierre
      // local: la sesión igual se trata como terminada del lado del
      // cliente, porque no hay una forma segura de reintentar
      // automáticamente una acción destructiva sin saber si el servidor ya
      // la aplicó.
    } finally {
      // Limpia el estado local no sensible (nunca hubo token en JS) y
      // navega a acceso; una próxima entrada a una ruta privada vuelve a
      // consultar el servidor en vez de reutilizar este resultado (mismo
      // mecanismo que tras un login exitoso).
      resetForFreshLogin()
      loggingOut.value = false
      await router.push({ name: 'acceso' })
    }
  }

  return { loggingOut, logout }
}
