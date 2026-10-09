// Espejo tipado del recurso `GoogleCalendarConnectionResponse` (issue #323/#324,
// DEC-099): lo que el barbero necesita ver de su integración. Nunca contiene
// tokens, secretos, el código OAuth ni el `state`: el cliente jamás los maneja.
export type GoogleCalendarStatus =
  'not_connected' | 'connected' | 'reauth_required' | 'error' | 'disconnected'

export interface GoogleCalendarConnection {
  /** El servidor tiene la integración configurada. */
  enabled: boolean
  /** El usuario declaró cuál barbero es (DEC-100); sin eso no se puede conectar. */
  barberLinked: boolean
  status: GoogleCalendarStatus
  accountEmail: string | null
  /** Minutos de anticipación del recordatorio, o null para los de Google. */
  reminderMinutes: number | null
  connectedAt: string | null
  lastSyncedAt: string | null
  /** Cambios que esperan publicarse en Google. */
  pendingSyncJobs: number
  /** Cambios que agotaron sus intentos. */
  failedSyncJobs: number
}

/** Desenlace de la vuelta de Google (POST .../callback). */
export type GoogleCalendarCallbackResult = 'connected' | 'denied' | 'failed'

/** Estado que la pantalla muestra, derivado del recurso (no es un campo de la API). */
export type GoogleCalendarView =
  | 'unavailable'
  | 'needs-barber'
  | 'not-connected'
  | 'connected'
  | 'syncing'
  | 'reauth'
  | 'sync-error'

export function viewOf(connection: GoogleCalendarConnection): GoogleCalendarView {
  if (!connection.enabled) return 'unavailable'
  if (!connection.barberLinked) return 'needs-barber'
  switch (connection.status) {
    case 'reauth_required':
      return 'reauth'
    case 'error':
      return 'sync-error'
    case 'connected':
      if (connection.failedSyncJobs > 0) return 'sync-error'
      return connection.pendingSyncJobs > 0 ? 'syncing' : 'connected'
    default:
      return 'not-connected'
  }
}
