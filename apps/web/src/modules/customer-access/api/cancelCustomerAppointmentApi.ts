// Cancelar el turno del cliente (HU-099, T5): único punto que llama a
// `POST /customer/appointments/{token}/cancel`. Traduce la respuesta a
// `CancelAppointmentOutcome`; la página nunca ve `Problem` ni el `status` crudo.
import { httpClient } from '@/shared/api/httpClient'
import { isProblem } from '@/shared/api/problem'
import type {
  CancelAppointmentInput,
  CancelAppointmentOutcome,
} from '../model/customerAppointmentOutcome'

/**
 * Mapea por `status` y `code` (nunca por `detail`): 200 cancelado; 422 con
 * `cancellation-window-closed` o `cancellation-reason-required` son los rechazos
 * de política y cualquier otro 422/400 es una entrada inválida; 409 con
 * `idempotency-locked` es la misma clave todavía en curso y cualquier otro 409
 * (`version-conflict`, `invalid-state`, conflicto de idempotencia) obliga a
 * recargar; 404 es el enlace inexistente, vencido o ya revocado.
 */
export async function cancelCustomerAppointment(
  token: string,
  versionToken: string,
  idempotencyKey: string,
  input: CancelAppointmentInput,
): Promise<CancelAppointmentOutcome> {
  try {
    const { error, response } = await httpClient.POST('/customer/appointments/{token}/cancel', {
      params: {
        path: { token },
        header: { 'Idempotency-Key': idempotencyKey, 'If-Match': versionToken },
      },
      body: { reason: input.reason },
    })

    if (response.ok) return { kind: 'success' }

    const problem = isProblem(error) ? error : undefined
    switch (response.status) {
      case 404:
        return { kind: 'not-found' }
      case 409:
        return problem?.code === 'idempotency-locked'
          ? { kind: 'idempotency-in-progress' }
          : { kind: 'changed' }
      case 422:
        if (problem?.code === 'cancellation-window-closed') return { kind: 'window-closed' }
        if (problem?.code === 'cancellation-reason-required') return { kind: 'reason-required' }
        return { kind: 'validation-error' }
      case 400:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error', requestId: problem?.requestId }
    }
  } catch {
    return { kind: 'network-error' }
  }
}
