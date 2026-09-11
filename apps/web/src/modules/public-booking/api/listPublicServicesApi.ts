// Único punto del módulo `public-booking` que llama a
// `GET /public/barbershops/{slug}/services` (HU-091). Traduce la respuesta
// real a `ListPublicServicesOutcome`: la página coordinadora nunca ve
// `Problem`, `status` HTTP crudo ni cabeceras. No pagina: pide una sola
// página con el límite máximo del contrato (50) porque HU-091 no incluye
// "cargar más" (fuera de alcance del prompt persistente).
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ListPublicServicesOutcome } from '../model/publicServiceListOutcome'

const MAX_PAGE_LIMIT = 50

/**
 * Lista el catálogo público de servicios de `slug` contra el API real
 * (HU-091). Mapea por `status` (nunca por `detail`): 404 cubre forma
 * inválida, desconocido y no publicable, indistinguibles a propósito
 * (CA-090-02, reutilizada aquí); cualquier otro estado no-2xx es
 * inesperado. Un fallo de `fetch` en sí (sin respuesta HTTP) es error de
 * red, distinto de un 5xx del servidor.
 */
export async function listPublicServices(slug: string): Promise<ListPublicServicesOutcome> {
  try {
    const { data, error, response } = await httpClient.GET('/public/barbershops/{slug}/services', {
      params: { path: { slug }, query: { limit: MAX_PAGE_LIMIT } },
    })

    if (response.ok && data) {
      return {
        kind: 'success',
        services: data.items.map((item) => ({
          id: item.id,
          name: item.name,
          description: item.description,
          durationMinutes: item.durationMinutes,
          price: item.price,
          currency: item.currency,
        })),
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
