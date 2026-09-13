// Único punto del módulo `public-booking` que llama a
// `GET /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability`
// (HU-094). Traduce la respuesta real a `ListPublicAvailabilityOutcome`: la
// página coordinadora nunca ve `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ListPublicAvailabilityOutcome } from '../model/publicAvailabilityOutcome'

/**
 * Lista los inicios públicos válidos para `barberId` en el servicio
 * `serviceId` de la barbería `slug`, contra el API real (HU-094). Mapea
 * por `status` (nunca por `detail`): 404 cubre forma de slug inválida,
 * desconocido y no publicable, indistinguibles a propósito (CA-090-02,
 * reutilizada aquí); un `serviceId`/`barberId` ajeno, inexistente,
 * inactivo o sin asignación vigente nunca llega como 404 -el contrato lo
 * resuelve como 200 con `slots: []` (mismo criterio que CA-092-03), así
 * que esta función jamás necesita distinguir esa causa. Cualquier otro
 * estado no-2xx es inesperado. Un fallo de `fetch` en sí (sin respuesta
 * HTTP) es error de red, distinto de un 5xx del servidor.
 */
export async function listPublicAvailability(
  slug: string,
  serviceId: string,
  barberId: string,
): Promise<ListPublicAvailabilityOutcome> {
  try {
    const { data, error, response } = await httpClient.GET(
      '/public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability',
      { params: { path: { slug, serviceId, barberId } } },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        availability: {
          slots: data.slots.map((slot) => ({ startsAt: slot.startsAt })),
          durationMinutes: data.durationMinutes,
          timezone: data.timezone,
          slotGridMinutes: data.slotGridMinutes,
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
