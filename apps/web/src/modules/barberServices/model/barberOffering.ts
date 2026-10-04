// Servicios que ofrece el único barbero de una barbería individual (DEC-115).
// Es la misma asignación de HU-023 (`barber_service`), pero vista desde el
// servicio: en vez de una matriz barbero × servicio, cada servicio dice «lo
// ofrezco» o no. No inventa nada: usa exactamente las operaciones de
// asignación que ya existen. Solo aplica con UN barbero; con cero o varios no
// hay a quién asignar sin preguntar, y `status` queda en `unavailable`.
import { ref } from 'vue'
import {
  assignService,
  fetchAssignments,
  fetchBarberSummaries,
  unassignService,
} from '../api/barberServicesApi'

export type OfferingStatus = 'idle' | 'loading' | 'ready' | 'unavailable' | 'error'

/** Resultado de ofrecer o dejar de ofrecer un servicio. */
export type OfferingChangeOutcome = 'changed' | 'unchanged' | 'network-error' | 'error'

// Una barbería con cientos de servicios no es el caso de una persona sola, pero el
// cursor de la API no tiene tope: este evita un bucle si el servidor no termina.
const MAX_PAGES = 20

export function useBarberOffering() {
  const status = ref<OfferingStatus>('idle')
  const barberId = ref<string | null>(null)
  const offered = ref<ReadonlySet<string>>(new Set())
  const busy = ref<ReadonlySet<string>>(new Set())

  async function load(): Promise<void> {
    status.value = 'loading'
    const barbers = await fetchBarberSummaries()
    if (barbers.kind !== 'success') {
      status.value = 'error'
      return
    }
    if (barbers.items.length !== 1) {
      barberId.value = null
      offered.value = new Set()
      status.value = 'unavailable'
      return
    }

    const id = barbers.items[0]!.id
    const ids = new Set<string>()
    let cursor: string | undefined
    for (let page = 0; page < MAX_PAGES; page++) {
      const outcome = await fetchAssignments(id, cursor)
      if (outcome.kind !== 'success') {
        status.value = 'error'
        return
      }
      for (const assignment of outcome.page.items) ids.add(assignment.serviceId)
      if (!outcome.page.nextCursor) break
      cursor = outcome.page.nextCursor
    }

    barberId.value = id
    offered.value = ids
    status.value = 'ready'
  }

  function isOffered(serviceId: string): boolean {
    return offered.value.has(serviceId)
  }

  function isBusy(serviceId: string): boolean {
    return busy.value.has(serviceId)
  }

  function mark(
    set: ReadonlySet<string>,
    serviceId: string,
    present: boolean,
  ): ReadonlySet<string> {
    const next = new Set(set)
    if (present) next.add(serviceId)
    else next.delete(serviceId)
    return next
  }

  /**
   * Ofrece o deja de ofrecer `serviceId`. Solo cambia `offered` cuando el servidor lo
   * confirma (nunca de forma optimista: el estado que se ve es el guardado). Una
   * operación en vuelo sobre el mismo servicio no se duplica.
   */
  async function setOffered(serviceId: string, value: boolean): Promise<OfferingChangeOutcome> {
    const id = barberId.value
    if (status.value !== 'ready' || !id) return 'error'
    if (isBusy(serviceId)) return 'unchanged'
    if (isOffered(serviceId) === value) return 'unchanged'

    busy.value = mark(busy.value, serviceId, true)
    try {
      if (value) {
        const outcome = await assignService(id, serviceId)
        if (outcome.kind === 'success') {
          offered.value = mark(offered.value, serviceId, true)
          return 'changed'
        }
        return outcome.kind === 'network-error' ? 'network-error' : 'error'
      }
      const outcome = await unassignService(id, serviceId)
      // `not-found` al retirar: ya no estaba asignado; el resultado que se quería es el real.
      if (outcome.kind === 'success' || outcome.kind === 'not-found') {
        offered.value = mark(offered.value, serviceId, false)
        return 'changed'
      }
      return outcome.kind === 'network-error' ? 'network-error' : 'error'
    } finally {
      busy.value = mark(busy.value, serviceId, false)
    }
  }

  return { status, isOffered, isBusy, load, setOffered }
}

export type BarberOffering = ReturnType<typeof useBarberOffering>
