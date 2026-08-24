// Único punto del módulo `staff` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/POST /private/barbers y GET/PATCH /private/barbers/{id} a
// los outcomes discriminados que la página consume; ningún componente ve
// `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import type { Barber, BarberPage } from '../model/barber'
import type {
  CreateBarberOutcome,
  FetchBarbersOutcome,
  RenameBarberOutcome,
} from '../model/staffOutcome'

export async function fetchBarbers(cursor?: string): Promise<FetchBarbersOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers', {
      params: { query: cursor ? { cursor } : {} },
    })

    if (response.ok && data) {
      return { kind: 'success', page: toPage(data) }
    }

    // 401 lo intercepta la coordinación única de installSessionHandling
    // (redirige a acceso); cualquier otro estado no-2xx se trata aquí como
    // un error genérico recuperable.
    return { kind: 'unexpected-error' }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir (mismo criterio que settingsApi.ts).
    return { kind: 'network-error' }
  }
}

export async function createBarber(
  fullName: string,
  idempotencyKey: string,
): Promise<CreateBarberOutcome> {
  try {
    const { data, response } = await httpClient.POST('/private/barbers', {
      params: { header: { 'Idempotency-Key': idempotencyKey } },
      body: { fullName },
    })

    if (response.ok && data) {
      return { kind: 'success', barber: toBarber(data) }
    }

    switch (response.status) {
      case 409:
        return { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function renameBarber(
  barberId: string,
  fullName: string,
): Promise<RenameBarberOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/barbers/{barberId}', {
      params: { path: { barberId } },
      body: { fullName },
    })

    if (response.ok && data) {
      return { kind: 'success', barber: toBarber(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toBarber(data: {
  id: string
  fullName: string
  createdAt: string
  updatedAt: string
}): Barber {
  return {
    id: data.id,
    fullName: data.fullName,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
  }
}

function toPage(data: {
  items: { id: string; fullName: string; createdAt: string; updatedAt: string }[]
  nextCursor: string | null
}): BarberPage {
  return { items: data.items.map(toBarber), nextCursor: data.nextCursor }
}
