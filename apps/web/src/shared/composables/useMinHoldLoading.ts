// Cubre el momento en que el contenido real "se arma" detrás de un rombo o
// esqueleto de carga (issue reportado 2026-09-28: al entrar a Agenda o
// Servicios se alcanzaba a ver la construcción cruda de la pantalla real).
// No retrasa cuándo EMPIEZA un indicador de carga (sigue apareciendo de
// inmediato, como corresponde: nunca una pantalla en blanco sin motivo) ni
// sustituye el guardado propio de cada pantalla contra respuestas
// obsoletas (p. ej. `requestToken`/`agendaRequestSeq"): solo retrasa el
// instante en que se le permite CEDER el paso al contenido real, para que
// una respuesta rápida no lo retire a medio parpadeo. Distinto del
// `transition-delay` que ya usa `.daily-agenda-page--updating` (ese evita
// que una recarga posterior MUESTRE un indicador si es rápida; este evita
// que el indicador inicial DESAPAREZCA demasiado rápido).
import { onUnmounted } from 'vue'

// Exportado para que las pruebas que necesiten esperar el retraso real
// (p. ej. CatalogPage.test.ts) lo hagan contra esta constante y no un
// número suelto que se desincroniza si el valor por defecto cambia.
export const DEFAULT_MIN_HOLD_MS = 400

// Retención de los esqueletos al buscar o cambiar de página dentro de una
// pantalla ya cargada (issue 2026-09-29): más corta que la carga inicial
// porque la persona ya está dentro y espera algo ágil.
export const PAGE_MIN_HOLD_MS = 250

export interface MinHoldLoading {
  /** Marca el instante en que el indicador de carga empezó a mostrarse. */
  start: () => void
  /** Ejecuta `reveal` (el cambio de estado que muestra el contenido real)
   * no antes de que se cumplan `minMs` desde el último `start()`. */
  hold: (reveal: () => void) => void
}

export function useMinHoldLoading(minMs = DEFAULT_MIN_HOLD_MS): MinHoldLoading {
  let startedAt = 0
  let releaseTimer: ReturnType<typeof setTimeout> | undefined

  function start() {
    clearTimeout(releaseTimer)
    startedAt = Date.now()
  }

  function hold(reveal: () => void) {
    clearTimeout(releaseTimer)
    const remaining = minMs - (Date.now() - startedAt)
    if (remaining <= 0) {
      reveal()
      return
    }
    releaseTimer = setTimeout(reveal, remaining)
  }

  onUnmounted(() => clearTimeout(releaseTimer))

  return { start, hold }
}
