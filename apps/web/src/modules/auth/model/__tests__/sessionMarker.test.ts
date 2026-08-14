import { describe, it, expect, beforeEach } from 'vitest'
import { rememberSessionUntil, hasRememberedSession, forgetSession } from '../sessionMarker'

describe('sessionMarker', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
  })

  it('has no remembered session before any login succeeds', () => {
    expect(hasRememberedSession()).toBe(false)
  })

  it('remembers a session until a future expiresAt', () => {
    rememberSessionUntil(new Date(Date.now() + 60_000).toISOString())
    expect(hasRememberedSession()).toBe(true)
  })

  it('treats an already-past expiresAt as no session', () => {
    rememberSessionUntil(new Date(Date.now() - 60_000).toISOString())
    expect(hasRememberedSession()).toBe(false)
  })

  it('treats a malformed value as no session instead of throwing', () => {
    window.sessionStorage.setItem('barberia:session-expires-at', 'no-es-una-fecha')
    expect(hasRememberedSession()).toBe(false)
  })

  it('forgets a remembered session', () => {
    rememberSessionUntil(new Date(Date.now() + 60_000).toISOString())
    expect(hasRememberedSession()).toBe(true)
    forgetSession()
    expect(hasRememberedSession()).toBe(false)
  })

  it('never stores the password or the session token, only expiresAt', () => {
    rememberSessionUntil(new Date(Date.now() + 60_000).toISOString())
    const raw = window.sessionStorage.getItem('barberia:session-expires-at')
    expect(raw).not.toBeNull()
    expect(Number.isNaN(Date.parse(raw as string))).toBe(false)
  })
})
