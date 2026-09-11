// Único punto del módulo `public-booking` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce la respuesta
// real de `GET /public/barbershops/{slug}` a `ResolveBarbershopOutcome`: la
// página coordinadora nunca ve `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ResolveBarbershopOutcome } from '../model/barbershopProfileOutcome'

/**
 * Resuelve el enlace público `slug` contra el API real (HU-090). Mapea por
 * `status` (nunca por `detail`, trabajo requerido §8): 404 cubre forma
 * inválida, desconocido y no publicable, indistinguibles a propósito
 * (CA-090-02); cualquier otro estado no-2xx es inesperado. Un fallo de
 * `fetch` en sí (sin respuesta HTTP) es error de red, distinto de un 5xx
 * del servidor.
 */
export async function resolveBarbershop(slug: string): Promise<ResolveBarbershopOutcome> {
  try {
    const { data, error, response } = await httpClient.GET('/public/barbershops/{slug}', {
      params: { path: { slug } },
    })

    if (response.ok && data) {
      return {
        kind: 'success',
        profile: {
          name: data.name,
          timezone: data.timezone,
          contactEmail: data.contactEmail,
          contactPhone: data.contactPhone,
        },
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
