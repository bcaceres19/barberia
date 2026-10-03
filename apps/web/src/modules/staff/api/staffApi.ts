// Único punto del módulo `staff` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/POST /private/barbers y GET/PATCH /private/barbers/{id} a
// los outcomes discriminados que la página consume; ningún componente ve
// `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import type { Barber, BarberPage, NumberedBarberPage } from '../model/barber'
import type {
  CreateBarberOutcome,
  FetchBarbersOutcome,
  RemoveBarberPhotoOutcome,
  RenameBarberOutcome,
  UploadBarberPhotoOutcome,
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

export async function fetchBarberPage(
  page: number,
  pageSize: number,
): Promise<
  { kind: 'success'; page: NumberedBarberPage } | { kind: 'network-error' | 'unexpected-error' }
> {
  try {
    const { data, response } = await httpClient.GET('/private/barbers', {
      params: { query: { page, pageSize } },
    })
    if (
      response.ok &&
      data &&
      data.page !== undefined &&
      data.pageSize !== undefined &&
      data.total !== undefined &&
      data.totalPages !== undefined
    ) {
      return {
        kind: 'success',
        page: {
          items: data.items.map(toBarber),
          page: data.page,
          pageSize: data.pageSize,
          total: data.total,
          totalPages: data.totalPages,
        },
      }
    }
    return { kind: 'unexpected-error' }
  } catch {
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

export { barberPhotoUrl } from '@/shared/api/barberPhotoUrl'

export async function uploadBarberPhoto(
  barberId: string,
  photo: Blob,
): Promise<UploadBarberPhotoOutcome> {
  try {
    const { data, response } = await httpClient.PUT('/private/barbers/{barberId}/photo', {
      params: { path: { barberId } },
      // openapi-typescript tipa un cuerpo `format: binary` como `string`, pero
      // fetch envía un Blob tal cual: bodySerializer lo deja pasar sin
      // convertirlo a JSON y la cabecera declara el tipo real de la imagen.
      body: photo as unknown as string,
      bodySerializer: (body: unknown) => body as Blob,
      headers: { 'Content-Type': photo.type },
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

export async function removeBarberPhoto(barberId: string): Promise<RemoveBarberPhotoOutcome> {
  try {
    const { response } = await httpClient.DELETE('/private/barbers/{barberId}/photo', {
      params: { path: { barberId } },
    })

    if (response.ok) return { kind: 'success' }
    if (response.status === 404) return { kind: 'not-found' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

interface BarberWire {
  id: string
  fullName: string
  createdAt: string
  updatedAt: string
  // Opcional a propósito: una respuesta de alta repetida por idempotencia que
  // se guardó antes de DEC-104 no lo trae; ausente equivale a "sin fotografía".
  photoUpdatedAt?: string | null
}

function toBarber(data: BarberWire): Barber {
  return {
    id: data.id,
    fullName: data.fullName,
    createdAt: data.createdAt,
    updatedAt: data.updatedAt,
    photoUpdatedAt: data.photoUpdatedAt ?? null,
  }
}

function toPage(data: { items: BarberWire[]; nextCursor: string | null }): BarberPage {
  return { items: data.items.map(toBarber), nextCursor: data.nextCursor }
}
