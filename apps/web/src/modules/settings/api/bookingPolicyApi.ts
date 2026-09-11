// Único punto del módulo `settings` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de GET/PUT /private/settings/booking-policy a los outcomes
// discriminados que la página consume; ningún componente ve `Problem`,
// `status` HTTP crudo ni cabeceras (salvo `versionToken`, que SÍ es un dato
// de dominio expuesto a propósito, no una cabecera cruda).
import { httpClient } from '@/shared/api/httpClient'
import type { BookingPolicy, BookingPolicyFormValues } from '../model/bookingPolicy'
import type {
  FetchBookingPolicyOutcome,
  SaveBookingPolicyOutcome,
} from '../model/bookingPolicyOutcome'

export async function fetchBookingPolicy(): Promise<FetchBookingPolicyOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/settings/booking-policy')

    if (response.ok && data) {
      return { kind: 'success', policy: toPolicy(data) }
    }

    // 401 lo intercepta la coordinación única de installSessionHandling
    // (redirige a acceso); 404 es defensivo (nunca debería ocurrir para un
    // principal autenticado real). Ambos, junto con cualquier otro estado
    // no-2xx, se tratan aquí como un error genérico recuperable.
    return { kind: 'unexpected-error' }
  } catch {
    // `fetch` en sí lanzó (red caída, DNS, CORS bloqueado): no hubo
    // respuesta HTTP que traducir (mismo criterio que settingsApi.ts).
    return { kind: 'network-error' }
  }
}

/**
 * Guarda un reemplazo completo de la política (PUT, contrato cerrado). El
 * token de la última lectura viaja como precondición `If-Match`
 * (CA-093-02): si la representación vigente ya cambió, el servidor
 * responde 409 y esta función lo traduce a `version-conflict`, distinto de
 * `validation-error` (un campo fuera de rango o una combinación
 * incoherente).
 */
export async function saveBookingPolicy(
  values: BookingPolicyFormValues,
  versionToken: string,
): Promise<SaveBookingPolicyOutcome> {
  try {
    const { data, response } = await httpClient.PUT('/private/settings/booking-policy', {
      params: { header: { 'If-Match': versionToken } },
      body: {
        minAdvanceMinutes: values.minAdvanceMinutes,
        maxAdvanceDays: values.maxAdvanceDays,
        // El contrato declara slotGridMinutes como enum [5,10,15,20,30,60]
        // (DEC-083); el formulario ya valida esa pertenencia antes de
        // llegar aquí (validateSlotGridMinutes), así que el aserto es
        // seguro.
        slotGridMinutes: values.slotGridMinutes as 5 | 10 | 15 | 20 | 30 | 60,
        cancellationDeadlineMinutes: values.cancellationDeadlineMinutes,
        lateCancellationClientAllowed: values.lateCancellationClientAllowed,
        lateCancellationReasonRequired: values.lateCancellationReasonRequired,
      },
    })

    if (response.ok && data) {
      return { kind: 'success', policy: toPolicy(data) }
    }

    switch (response.status) {
      case 409:
        return { kind: 'version-conflict' }
      case 400:
      case 422:
        return { kind: 'validation-error' }
      default:
        return { kind: 'unexpected-error' }
    }
  } catch {
    return { kind: 'network-error' }
  }
}

function toPolicy(data: {
  minAdvanceMinutes: number
  maxAdvanceDays: number
  slotGridMinutes: number
  cancellationDeadlineMinutes: number
  lateCancellationClientAllowed: boolean
  lateCancellationReasonRequired: boolean
  versionToken: string
}): BookingPolicy {
  return {
    minAdvanceMinutes: data.minAdvanceMinutes,
    maxAdvanceDays: data.maxAdvanceDays,
    slotGridMinutes: data.slotGridMinutes,
    cancellationDeadlineMinutes: data.cancellationDeadlineMinutes,
    lateCancellationClientAllowed: data.lateCancellationClientAllowed,
    lateCancellationReasonRequired: data.lateCancellationReasonRequired,
    versionToken: data.versionToken,
  }
}
