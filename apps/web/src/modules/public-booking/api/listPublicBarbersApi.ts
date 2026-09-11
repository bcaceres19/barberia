// Único punto del módulo `public-booking` que llama a
// `GET /public/barbershops/{slug}/services/{serviceId}/barbers` (HU-092).
// Traduce la respuesta real a `ListPublicBarbersOutcome`: la página
// coordinadora nunca ve `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ListPublicBarbersOutcome } from '../model/publicBarberListOutcome'

/**
 * Lista los barberos elegibles para `serviceId` en `slug` contra el API
 * real (HU-092). Mapea por `status` (nunca por `detail`): 404 cubre forma
 * de slug inválida, desconocido y no publicable, indistinguibles a
 * propósito (CA-090-02, reutilizada aquí); un `serviceId` ajeno,
 * inexistente o ya no asignado nunca llega como 404 -el contrato lo
 * resuelve como 200 con `items: []` (CA-092-03), así que esta función
 * jamás necesita distinguir esa causa. Cualquier otro estado no-2xx es
 * inesperado. Un fallo de `fetch` en sí (sin respuesta HTTP) es error de
 * red, distinto de un 5xx del servidor.
 */
export async function listPublicBarbers(
  slug: string,
  serviceId: string,
): Promise<ListPublicBarbersOutcome> {
  try {
    const { data, error, response } = await httpClient.GET(
      '/public/barbershops/{slug}/services/{serviceId}/barbers',
      { params: { path: { slug, serviceId } } },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        barbers: data.items.map((item) => ({ id: item.id, fullName: item.fullName })),
      }
    }

    const requestId = isProblem(error) ? error.requestId : undefined

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      default:
        return { kind: 'unexpected-error', requestId }
    }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir, así que nunca llega a la rama anterior.
    return { kind: 'network-error' }
  }
}
