// Estado de sesión compartido entre TODAS las rutas privadas (DEC-033,
// docs/03-desarrollo/estandar-frontend-vue.md §113: "Pinia solo se
// incorpora con estado compartido real entre rutas"). Aquí sí hay estado
// compartido real -la cabecera, el guard y cualquier llamada privada futura
// necesitan el mismo contexto-, pero una sola pieza de estado no justifica
// añadir una dependencia nueva (DEC-035): un singleton reactivo exportado
// desde un módulo cumple lo mismo con menos superficie. Si en el futuro
// aparece más estado global real, esta es la pieza que se migraría a Pinia,
// no un patrón nuevo.
import { reactive, readonly, type DeepReadonly } from 'vue'
import type { SessionContext } from './sessionContext'
import { fetchSessionContext } from '../api/sessionContextApi'

/**
 * Estado discriminado y explícito del bootstrap de la aplicación (trabajo
 * requerido §2): comprobando, autenticado con contexto mínimo, no
 * autenticado y fallo recuperable. Ningún estado por sí solo concede
 * permisos: la cookie HttpOnly sigue siendo la única autoridad real en cada
 * solicitud (fuera de alcance de HU-012, "tratar un marcador del navegador
 * como prueba de autorización").
 */
export type BootstrapState =
  | { status: 'checking' }
  | ({ status: 'authenticated' } & SessionContext)
  | { status: 'unauthenticated' }
  | { status: 'connection-lost' }

const state = reactive<{ bootstrap: BootstrapState }>({ bootstrap: { status: 'checking' } })

/** Vista de solo lectura para componentes/guards; solo las funciones de
 * este módulo pueden mutar el estado. */
export const sessionState: DeepReadonly<{ bootstrap: BootstrapState }> = readonly(state)

function setChecking(): void {
  state.bootstrap = { status: 'checking' }
}

function setAuthenticated(ctx: SessionContext): void {
  state.bootstrap = { status: 'authenticated', ...ctx }
}

function setUnauthenticated(): void {
  state.bootstrap = { status: 'unauthenticated' }
}

function setConnectionLost(): void {
  state.bootstrap = { status: 'connection-lost' }
}

async function refreshFromServer(): Promise<void> {
  const outcome = await fetchSessionContext()
  switch (outcome.kind) {
    case 'authenticated':
      setAuthenticated(outcome)
      return
    case 'unauthenticated':
      setUnauthenticated()
      return
    case 'network-error':
    case 'unexpected-error':
      setConnectionLost()
  }
}

let inFlight: Promise<void> | null = null

/**
 * Consulta el servidor solo si el estado actual es `checking` (arranque de
 * la app, o después de resetChecking()); llamadas concurrentes comparten la
 * misma solicitud en curso en vez de disparar una por cada ruta/componente
 * que la invoque. El guard de rutas privadas la espera antes de decidir.
 */
export async function ensureBootstrapped(): Promise<void> {
  if (state.bootstrap.status !== 'checking') return
  if (!inFlight) {
    inFlight = refreshFromServer().finally(() => {
      inFlight = null
    })
  }
  await inFlight
}

/**
 * Actualiza únicamente el nombre de la barbería del contexto ya
 * autenticado (HU-020, CA-020-02): la cabecera refleja el nuevo nombre sin
 * recargar la aplicación. Se llama SOLO después de que el servidor
 * confirma el guardado (respuesta 200 de
 * `PATCH /private/settings/barbershop`); un fallo de guardado nunca invoca
 * esta función (sin optimismo). No hace nada si el estado no está
 * `authenticated` (defecto de orden de llamada: la sección de
 * configuración solo es alcanzable ya autenticado). La próxima
 * rehidratación (recarga, `retryBootstrap`) sigue usando
 * `GET /private/auth/session` como autoridad; esta función solo evita una
 * espera visible mientras esa autoridad no se ha vuelto a consultar.
 */
export function updateBarbershopName(name: string): void {
  if (state.bootstrap.status !== 'authenticated') return
  state.bootstrap = { ...state.bootstrap, barbershopName: name }
}

/** Usada por LoginPage tras un acceso exitoso (DEC-060 no amplía
 * LoginResponse con el contexto): fuerza a que la próxima ruta privada
 * vuelva a consultar el servidor en vez de reutilizar un `unauthenticated`
 * previo a este mismo inicio de sesión. */
export function resetForFreshLogin(): void {
  setChecking()
}

/** Usada por el estado "conexión perdida" del cascarón (CA-012-05):
 * reintenta la consulta real, nunca reutiliza un resultado en caché. */
export async function retryBootstrap(): Promise<void> {
  setChecking()
  await ensureBootstrapped()
}

// --- Coordinación única de 401 (CA-012-03) --------------------------------

type UnauthorizedListener = () => void
const unauthorizedListeners = new Set<UnauthorizedListener>()

/** Registra un oyente para la transición a "no autorizado" causada por un
 * 401 real de una ruta privada (nunca por el bootstrap en sí, que ya
 * resuelve su propio resultado arriba). Devuelve una función para
 * desuscribirse. Pensada para que `app/bootstrap` la use para redirigir con
 * el router real, sin que este módulo importe el router (app → modules →
 * shared). */
export function onUnauthorized(listener: UnauthorizedListener): () => void {
  unauthorizedListeners.add(listener)
  return () => unauthorizedListeners.delete(listener)
}

/**
 * Punto único de entrada para "esta solicitud privada recibió 401" (trabajo
 * requerido §5). Idempotente: si el estado ya es `unauthenticated`, no
 * vuelve a limpiar ni a notificar. Varias solicitudes concurrentes que
 * fallan casi al mismo tiempo (CA-012-03) solo producen UNA transición y
 * UNA notificación, porque JavaScript ejecuta esta función de forma
 * síncrona y la segunda llamada ya encuentra el estado actualizado.
 */
export function reportUnauthorized(): void {
  if (state.bootstrap.status === 'unauthenticated') return
  setUnauthenticated()
  unauthorizedListeners.forEach((listener) => listener())
}
