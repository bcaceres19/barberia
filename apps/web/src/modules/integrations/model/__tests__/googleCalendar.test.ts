import { describe, expect, it } from 'vitest'
import { viewOf, type GoogleCalendarConnection } from '../googleCalendar'

function connection(overrides: Partial<GoogleCalendarConnection> = {}): GoogleCalendarConnection {
  return {
    enabled: true,
    barberLinked: true,
    status: 'connected',
    accountEmail: 'barbero@ejemplo.test',
    reminderMinutes: null,
    connectedAt: null,
    lastSyncedAt: null,
    pendingSyncJobs: 0,
    failedSyncJobs: 0,
    ...overrides,
  }
}

describe('viewOf', () => {
  it.each([
    [{ enabled: false }, 'unavailable'],
    [{ barberLinked: false }, 'needs-barber'],
    [{ status: 'not_connected' as const }, 'not-connected'],
    [{ status: 'disconnected' as const }, 'not-connected'],
    [{ status: 'reauth_required' as const }, 'reauth'],
    [{ status: 'error' as const }, 'sync-error'],
    [{ status: 'connected' as const }, 'connected'],
    [{ pendingSyncJobs: 2 }, 'syncing'],
    [{ failedSyncJobs: 1 }, 'sync-error'],
    [{ failedSyncJobs: 1, pendingSyncJobs: 3 }, 'sync-error'],
  ])('%j → %s', (overrides, expected) => {
    expect(viewOf(connection(overrides))).toBe(expected)
  })

  it('shows the missing configuration before the barber link and the link before the status', () => {
    expect(viewOf(connection({ enabled: false, barberLinked: false, status: 'connected' }))).toBe(
      'unavailable',
    )
    expect(viewOf(connection({ barberLinked: false, status: 'connected' }))).toBe('needs-barber')
  })
})
