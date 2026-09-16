// Único punto del módulo `public-booking` que llama a
// POST /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/appointments
// (HU-097). Traduce la respuesta real a ConfirmPublicAppointmentOutcome: la
// página coordinadora nunca ve `Problem`, `status` HTTP crudo ni cabeceras.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type { ConfirmPublicAppointmentOutcome } from '../model/confirmPublicAppointmentOutcome'

export interface ConfirmPublicAppointmentInput {
  startsAt: string
  fullName: string
  phone: string
  email: string
  note: string | null
  forSomeoneElse: boolean
  attendeeName: string | null
}

/**
 * Confirma el turno público elegido (HU-097). Mapea por `status`/`code`
 * (nunca por `detail`): `404` cubre la misma familia indistinguible de
 * causas que el resto del módulo (CA-090-02, CA-092-03); `409` distingue
 * `slot-conflict` (RN-CON-05/DEC-090, con `alternatives`) de los
 * conflictos de idempotencia ya conocidos (RN-IDE-01); `422` es un dato de
 * identidad inválido que el servidor revalidó. Un fallo de `fetch` en sí
 * (sin respuesta HTTP) es error de red, distinto de un 5xx del servidor.
 */
export async function confirmPublicAppointment(
  slug: string,
  serviceId: string,
  barberId: string,
  input: ConfirmPublicAppointmentInput,
  idempotencyKey: string,
): Promise<ConfirmPublicAppointmentOutcome> {
  try {
    const { data, response, error } = await httpClient.POST(
      '/public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/appointments',
      {
        params: {
          path: { slug, serviceId, barberId },
          header: { 'Idempotency-Key': idempotencyKey },
        },
        body: {
          startsAt: input.startsAt,
          fullName: input.fullName,
          phone: input.phone,
          email: input.email,
          note: input.note,
          forSomeoneElse: input.forSomeoneElse,
          attendeeName: input.attendeeName,
        },
      },
    )

    if (response.ok && data) {
      return {
        kind: 'success',
        appointment: {
          attendeeName: data.attendeeName,
          barbershopName: data.barbershopName,
          serviceName: data.serviceName,
          durationMinutes: data.durationMinutes,
          priceAmount: data.priceAmount,
          currency: data.currency,
          startsAt: data.startsAt,
          endsAt: data.endsAt,
          timezone: data.timezone,
          accessToken: data.accessToken,
          customerNote: data.customerNote,
        },
      }
    }

    const requestId = isProblem(error) ? error.requestId : undefined

    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        if (isProblemCode(error, 'slot-conflict')) {
          const alternatives =
            error && typeof error === 'object' && 'alternatives' in error
              ? ((error as { alternatives?: { startsAt: string }[] }).alternatives ?? [])
              : []
          return {
            kind: 'schedule-conflict',
            alternatives: alternatives.map((a) => ({ startsAt: a.startsAt })),
          }
        }
        return { kind: 'idempotency-conflict' }
      case 422:
        return { kind: 'validation-error', detail: problemDetail(error) }
      default:
        return { kind: 'unexpected-error', requestId }
    }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir, así que nunca llega a la rama anterior.
    return { kind: 'network-error' }
  }
}

function isProblemCode(error: unknown, code: string): boolean {
  return !!error && typeof error === 'object' && (error as { code?: string }).code === code
}

function problemDetail(error: unknown): string {
  if (
    error &&
    typeof error === 'object' &&
    typeof (error as { detail?: string }).detail === 'string'
  ) {
    return (error as { detail: string }).detail
  }
  return 'No pudimos confirmar tu turno.'
}
