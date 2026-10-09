// Único punto del módulo `integrations` que llama al cliente HTTP tipado
// (docs/03-desarrollo/estandar-frontend-vue.md §6). Traduce las respuestas
// reales de /private/integrations/google-calendar a los outcomes discriminados
// que las páginas consumen; ningún componente ve `Problem`, `status` HTTP crudo
// ni cabeceras. Nunca maneja tokens: el flujo OAuth ocurre entre el navegador,
// Google y el API.
import { httpClient } from '@/shared/api/httpClient'
import type {
  GoogleCalendarCallbackResult,
  GoogleCalendarConnection,
} from '../model/googleCalendar'
import type {
  CompleteCallbackOutcome,
  DisconnectOutcome,
  FetchConnectionOutcome,
  StartConnectOutcome,
  SyncNowOutcome,
  UpdateReminderOutcome,
} from '../model/googleCalendarOutcome'

interface ConnectionWire {
  enabled: boolean
  barberLinked: boolean
  status: GoogleCalendarConnection['status']
  accountEmail: string | null
  reminderMinutes: number | null
  connectedAt: string | null
  lastSyncedAt: string | null
  pendingSyncJobs: number
  failedSyncJobs: number
}

function toConnection(data: ConnectionWire): GoogleCalendarConnection {
  return {
    enabled: data.enabled,
    barberLinked: data.barberLinked,
    status: data.status,
    accountEmail: data.accountEmail,
    reminderMinutes: data.reminderMinutes,
    connectedAt: data.connectedAt,
    lastSyncedAt: data.lastSyncedAt,
    pendingSyncJobs: data.pendingSyncJobs,
    failedSyncJobs: data.failedSyncJobs,
  }
}

export async function fetchConnection(): Promise<FetchConnectionOutcome> {
  try {
    const { data, response } = await httpClient.GET('/private/integrations/google-calendar')
    if (response.ok && data) return { kind: 'success', connection: toConnection(data) }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function startConnect(): Promise<StartConnectOutcome> {
  try {
    const { data, response } = await httpClient.POST(
      '/private/integrations/google-calendar/connect',
    )
    if (response.ok && data) return { kind: 'success', authorizationUrl: data.authorizationUrl }
    if (response.status === 409) return { kind: 'unavailable' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

/** Reenvía a la API lo que Google agregó a la URL de retorno. */
export async function completeCallback(params: {
  state: string
  code?: string
  error?: string
}): Promise<CompleteCallbackOutcome> {
  try {
    const { data, response } = await httpClient.POST(
      '/private/integrations/google-calendar/callback',
      {
        body: {
          state: params.state,
          ...(params.code ? { code: params.code } : {}),
          ...(params.error ? { error: params.error } : {}),
        },
      },
    )
    if (response.ok && data) {
      return { kind: 'success', result: data.result as GoogleCalendarCallbackResult }
    }
    if (response.status === 400) return { kind: 'invalid' }
    if (response.status === 409) return { kind: 'unavailable' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function syncNow(): Promise<SyncNowOutcome> {
  try {
    const { data, response } = await httpClient.POST('/private/integrations/google-calendar/sync')
    if (response.ok && data) {
      return {
        kind: 'success',
        pendingSyncJobs: data.pendingSyncJobs,
        failedSyncJobs: data.failedSyncJobs,
      }
    }
    if (response.status === 404) return { kind: 'not-connected' }
    if (response.status === 409) return { kind: 'unavailable' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

export async function disconnect(): Promise<DisconnectOutcome> {
  try {
    const { response } = await httpClient.DELETE('/private/integrations/google-calendar')
    if (response.ok) return { kind: 'success' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}

/** `minutes` null usa los recordatorios predeterminados del calendario. */
export async function updateReminder(minutes: number | null): Promise<UpdateReminderOutcome> {
  try {
    const { data, response } = await httpClient.PATCH('/private/integrations/google-calendar', {
      body: { reminderMinutes: minutes },
    })
    if (response.ok && data) return { kind: 'success', connection: toConnection(data) }
    if (response.status === 422) return { kind: 'validation-error' }
    if (response.status === 404) return { kind: 'not-connected' }
    return { kind: 'unexpected-error' }
  } catch {
    return { kind: 'network-error' }
  }
}
