// Único punto del módulo `catalog` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/POST /private/services y GET/PATCH
// /private/services/{serviceId} a los outcomes discriminados que la página
// consume; ningún componente ve `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import type { Service, ServicePage } from '../model/service'
import type {
  CreateServiceOutcome,
  DeactivateServiceOutcome,
  FetchServicesOutcome,
  PreviewDeactivationOutcome,
  ReactivateServiceOutcome,
  UpdateServiceOutcome,
} from '../model/catalogOutcome'

export async function fetchServices(cursor?: string): Promise<FetchServicesOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/services', {
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
    // respuesta HTTP que traducir (mismo criterio que staffApi.ts).
    return { kind: 'network-error' }
  }
}

export interface ServiceInput {
  name: string
  description: string
  durationMinutes: number
  price: string
}

export async function createService(
  input: ServiceInput,
  idempotencyKey: string,
): Promise<CreateServiceOutcome> {
  try {
    const trimmedDescription = input.description.trim()
    const { data, error, response } = await httpClient.POST('/private/services', {
      params: { header: { 'Idempotency-Key': idempotencyKey } },
      body: {
        name: input.name,
        durationMinutes: input.durationMinutes,
        price: input.price,
        ...(trimmedDescription ? { description: trimmedDescription } : {}),
      },
    })

    if (response.ok && data) {
      return { kind: 'success', service: toService(data) }
    }

    switch (response.status) {
      case 409:
        // DEC-067 (code: "conflict") frente a RN-IDE-01 (code:
        // "idempotency-conflict"/"idempotency-locked"): ambos comparten
        // status pero el cliente decide por `code`, nunca por `detail`
        // (docs/06-api/estandar-openapi.md §5).
        return isProblemCode(error, 'conflict')
          ? { kind: 'name-conflict' }
          : { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function updateService(
  serviceId: string,
  input: ServiceInput,
): Promise<UpdateServiceOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/services/{serviceId}', {
      params: { path: { serviceId } },
      body: {
        name: input.name,
        description: input.description.trim(),
        durationMinutes: input.durationMinutes,
        price: input.price,
      },
    })

    if (response.ok && data) {
      return { kind: 'success', service: toService(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return { kind: 'name-conflict' }
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// isProblemCode informa si el cuerpo de error ya decodificado por
// openapi-fetch trae el `code` esperado (Problem.code, RFC 9457).
function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function toService(data: {
  id: string
  name: string
  description: string | null
  durationMinutes: number
  price: string
  currency: string
  isActive: boolean
  deactivatedAt: string | null
  createdAt: string
  updatedAt: string
}): Service {
  return {
    id: data.id,
    name: data.name,
    description: data.description,
    durationMinutes: data.durationMinutes,
    price: data.price,
    currency: data.currency,
    isActive: data.isActive,
    deactivatedAt: data.deactivatedAt,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
  }
}

function toPage(data: {
  items: {
    id: string
    name: string
    description: string | null
    durationMinutes: number
    price: string
    currency: string
    isActive: boolean
    deactivatedAt: string | null
    createdAt: string
    updatedAt: string
  }[]
  nextCursor: string | null
}): ServicePage {
  return { items: data.items.map(toService), nextCursor: data.nextCursor }
}

// --- HU-024: ciclo de vida --------------------------------------------------

// previewDeactivation consulta el impacto real de desactivar serviceId
// (CA-024-01): siempre 0 en B1 (DEC-069), pero SIEMPRE mediante esta
// solicitud real -nunca un valor por defecto asumido en el cliente.
export async function previewDeactivation(serviceId: string): Promise<PreviewDeactivationOutcome> {
  try {
    const { data, response } = await httpClient.GET(
      '/private/services/{serviceId}/deactivation-impact',
      { params: { path: { serviceId } } },
    )

    if (response.ok && data) {
      return { kind: 'success', affectedAppointments: data.affectedAppointments }
    }
    if (response.status === 404) return { kind: 'not-found' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

// deactivateService confirma la desactivación (CA-024-02, CA-024-04),
// protegida por Idempotency-Key (RN-IDE-01): idempotencyKey es la clave del
// intento lógico vigente (misma disciplina que createService).
export async function deactivateService(
  serviceId: string,
  idempotencyKey: string,
): Promise<DeactivateServiceOutcome> {
  try {
    const { data, error, response } = await httpClient.POST(
      '/private/services/{serviceId}/deactivate',
      {
        params: { path: { serviceId }, header: { 'Idempotency-Key': idempotencyKey } },
      },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        service: toService(data.service),
        affectedAppointments: data.affectedAppointments,
      }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        // "conflict" (transición inválida, el servicio ya estaba inactivo)
        // frente a "idempotency-conflict"/"idempotency-locked" (RN-IDE-01):
        // el cliente decide por `code`, nunca por `detail`.
        return isProblemCode(error, 'conflict')
          ? { kind: 'transition-conflict' }
          : { kind: 'idempotency-conflict' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

// reactivateService confirma la reactivación (CA-024-05), mismo protocolo
// de idempotencia que deactivateService, en sentido inverso.
export async function reactivateService(
  serviceId: string,
  idempotencyKey: string,
): Promise<ReactivateServiceOutcome> {
  try {
    const { data, error, response } = await httpClient.POST(
      '/private/services/{serviceId}/reactivate',
      {
        params: { path: { serviceId }, header: { 'Idempotency-Key': idempotencyKey } },
      },
    )

    if (response.ok && data) {
      return { kind: 'success', service: toService(data) }
    }

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return isProblemCode(error, 'conflict')
          ? { kind: 'transition-conflict' }
          : { kind: 'idempotency-conflict' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}
