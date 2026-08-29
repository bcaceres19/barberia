// Desenlaces discriminados de las llamadas de red de `agenda` (HU-061):
// ningún componente ve `Problem`, `status` HTTP crudo ni cabeceras, mismo
// criterio que schedules/model/scheduleOutcome.ts.
import type { BarberSummary, ServiceSummary } from './appointment'
import type { DailyAgendaEntry } from './dailyAgenda'

export type FetchBarberSummariesOutcome =
  | { kind: 'success'; items: BarberSummary[] }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchAssignedServicesOutcome =
  | { kind: 'success'; items: ServiceSummary[] }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type FetchBarbershopTimezoneOutcome =
  { kind: 'success'; timezone: string } | { kind: 'unavailable' }

export type CreatedManualAppointment = {
  id: string
  barberId: string
  serviceId: string
  attendeeName: string
  startsAt: string
  endsAt: string
  serviceName: string
  durationMinutes: number
  priceAmount: string
  currency: string
}

// 'conflict' cubre tanto un cruce de agenda (RN-CON-01) como un bloqueo
// vigente (DEC-073): el contrato responde el mismo `code: conflict` para
// ambos (el cliente decide por `code`, nunca por `detail`, pero aquí no
// hay un `code` más específico que distinguirlos); detail trae el mensaje
// seguro que el backend ya redactó para mostrar directamente.
// FetchDailyAgendaOutcome (HU-062): 'not-found' cubre un barberId
// inexistente o de otra barbería (RN-TEN-01), verificado ya en el
// selector, pero posible si el barbero se retira mientras la pantalla
// sigue abierta.
export type FetchDailyAgendaOutcome =
  | { kind: 'success'; items: DailyAgendaEntry[] }
  | { kind: 'not-found' }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }

export type CreateManualAppointmentOutcome =
  | { kind: 'success'; appointment: CreatedManualAppointment }
  | { kind: 'not-found' }
  | { kind: 'conflict'; detail: string }
  | { kind: 'idempotency-conflict' }
  | { kind: 'validation-error'; detail: string }
  | { kind: 'network-error' }
  | { kind: 'unexpected-error' }
