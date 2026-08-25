// Único punto del módulo `barberServices` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/PUT/DELETE /private/barbers/{barberId}/services[/{serviceId}]
// a los outcomes discriminados que la página consume; ningún componente ve
// `Problem`, `status` HTTP crudo ni cabeceras.
//
// fetchBarberSummaries/fetchServiceSummaries llaman DIRECTAMENTE al cliente
// HTTP compartido (nunca a staffApi.ts/catalogApi.ts, privados de sus
// propios módulos, docs/03-desarrollo/estandar-frontend-vue.md §3): esta
// pequeña duplicación de forma es preferible a un acoplamiento cruzado entre
// módulos (mismo criterio que catalog.Cursor/staff.Cursor en el backend).
import { httpClient } from '@/shared/api/httpClient'
import type { Assignment, AssignmentPage, BarberSummary, ServiceSummary } from '../model/assignment'
import type {
  AssignServiceOutcome,
  FetchAssignmentsOutcome,
  FetchBarberSummariesOutcome,
  FetchServiceSummariesOutcome,
  UnassignServiceOutcome,
} from '../model/barberServicesOutcome'

// PICKER_LIMIT es el máximo que el contrato admite por página
// (docs/06-api/estandar-openapi.md §6.10). Una barbería con más de 50
// barberos o 50 servicios queda fuera del alcance de esta pantalla (HU-023
// no exige paginar los selectores, a diferencia de StaffPage/CatalogPage,
// que sí pagina su propia lista completa): un equipo o catálogo tan grande
// es un caso de una versión posterior.
const PICKER_LIMIT = 50

export async function fetchBarberSummaries(): Promise<FetchBarberSummariesOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers', {
      params: { query: { limit: PICKER_LIMIT } },
    })

    if (response.ok && data) {
      return { kind: 'success', items: data.items.map((b) => ({ id: b.id, fullName: b.fullName })) }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function fetchServiceSummaries(): Promise<FetchServiceSummariesOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/services', {
      params: { query: { limit: PICKER_LIMIT } },
    })

    if (response.ok && data) {
      return { kind: 'success', items: data.items.map((s) => ({ id: s.id, name: s.name })) }
    }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function fetchAssignments(
  barberId: string,
  cursor?: string,
): Promise<FetchAssignmentsOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers/{barberId}/services', {
      params: {
        path: { barberId },
        query: cursor ? { cursor } : {},
      },
    })

    if (response.ok && data) {
      return { kind: 'success', page: toPage(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function assignService(
  barberId: string,
  serviceId: string,
): Promise<AssignServiceOutcome> {
  try {
    const { data, response } = await httpClient.PUT(
      '/private/barbers/{barberId}/services/{serviceId}',
      {
        params: { path: { barberId, serviceId } },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', assignment: toAssignment(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function unassignService(
  barberId: string,
  serviceId: string,
): Promise<UnassignServiceOutcome> {
  try {
    const { response, error } = await httpClient.DELETE(
      '/private/barbers/{barberId}/services/{serviceId}',
      {
        params: { path: { barberId, serviceId } },
      },
    )

    if (response.ok) {
      return { kind: 'success' }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        // DEC-068 (code: "conflict") es el único 409 posible de esta
        // operación (no hay Idempotency-Key ni protocolo de idempotencia
        // en este DELETE): el cliente todavía decide por `code`, nunca por
        // `detail` (docs/06-api/estandar-openapi.md §5), por si una versión
        // futura agregara otro problem type sobre este mismo status.
        return isProblemCode(error, 'conflict')
          ? { kind: 'last-active-conflict' }
          : { kind: 'unexpected-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function toAssignment(data: {
  barberId: string
  serviceId: string
  createdAt: string
}): Assignment {
  return { barberId: data.barberId, serviceId: data.serviceId, createdAt: data.createdAt }
}

function toPage(data: {
  items: { barberId: string; serviceId: string; createdAt: string }[]
  nextCursor: string | null
}): AssignmentPage {
  return { items: data.items.map(toAssignment), nextCursor: data.nextCursor }
}

export type { BarberSummary, ServiceSummary }
