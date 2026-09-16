// Único punto del módulo `customer-access` que llama al cliente HTTP
// tipado (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce la
// respuesta real de `GET /customer/appointments/{token}` a
// `GetCustomerAppointmentOutcome`: la página coordinadora nunca ve
// `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type {
  CustomerAppointment,
  GetCustomerAppointmentOutcome,
} from '../model/customerAppointmentOutcome'

/**
 * Resuelve el token del enlace de acceso contra el API real (HU-098). Mapea
 * por `status` (nunca por `detail`, trabajo requerido §8): 404 cubre forma
 * inválida, inexistente, vencido y revocado, indistinguibles a propósito
 * (CA-098-02); cualquier otro estado no-2xx es inesperado. Un fallo de
 * `fetch` en sí (sin respuesta HTTP) es error de red, distinto de un 5xx
 * del servidor.
 */
export async function getCustomerAppointment(
  token: string,
): Promise<GetCustomerAppointmentOutcome> {
  try {
    const { data, error, response } = await httpClient.GET('/customer/appointments/{token}', {
      params: { path: { token } },
    })

    if (response.ok && data) {
      const appointment: CustomerAppointment = {
        barbershopName: data.barbershopName,
        timezone: data.timezone,
        attendeeName: data.attendeeName,
        serviceName: data.serviceName,
        durationMinutes: data.durationMinutes,
        barberName: data.barberName,
        startsAt: data.startsAt,
        endsAt: data.endsAt,
        status: data.status as CustomerAppointment['status'],
        cancellationDeadlineMinutes: data.cancellationDeadlineMinutes,
        lateCancellationClientAllowed: data.lateCancellationClientAllowed,
        lateCancellationReasonRequired: data.lateCancellationReasonRequired,
      }
      return { kind: 'success', appointment }
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
