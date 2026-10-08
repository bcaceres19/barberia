// Perfil público de la barbería del enlace, compartido por el cascarón y la
// pantalla de entrada (HU-090, DEC-119). Las dos lo necesitan al abrir un enlace:
// la entrada para mostrar nombre y contacto, el cascarón para el vocabulario.
// Pedirlo dos veces sería un gasto y, si la red falla justo ahora, el cascarón
// «se llevaría» el fallo y la entrada nunca mostraría su error ni su botón de
// reintento. Por eso una petición en vuelo para el mismo enlace se comparte, y
// cualquier lectura exitosa —también la de un reintento de la entrada— deja el
// vocabulario disponible para todas las pantallas.
//
// No es estado de negocio: solo las palabras ya públicas del último perfil leído
// por enlace, que el cascarón vuelve a pedir cada vez que se abre.
import { reactive } from 'vue'
import type { VocabularyTerms } from '@/shared/model'
import { resolveBarbershop } from '../api/resolveBarbershopApi'
import type { ResolveBarbershopOutcome } from './barbershopProfileOutcome'

const termsBySlug = reactive<Record<string, VocabularyTerms>>({})
const inFlight = new Map<string, Promise<ResolveBarbershopOutcome>>()

/** Resuelve el enlace; llamadas simultáneas para el mismo enlace comparten una sola petición. */
export function loadPublicProfile(slug: string): Promise<ResolveBarbershopOutcome> {
  const pending = inFlight.get(slug)
  if (pending) return pending
  const request = resolveBarbershop(slug)
    .then((outcome) => {
      if (outcome.kind === 'success') termsBySlug[slug] = outcome.profile.vocabulary
      return outcome
    })
    .finally(() => inFlight.delete(slug))
  inFlight.set(slug, request)
  return request
}

/** Palabras de la barbería del enlace, o `null` si aún no se leyeron. */
export function publicTermsFor(slug: string): VocabularyTerms | null {
  return termsBySlug[slug] ?? null
}

/** Solo para pruebas: olvida lo leído. */
export function resetPublicProfileCache(): void {
  for (const key of Object.keys(termsBySlug)) delete termsBySlug[key]
  inFlight.clear()
}
