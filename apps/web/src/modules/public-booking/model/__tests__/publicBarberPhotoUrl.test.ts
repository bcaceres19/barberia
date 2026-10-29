import { describe, it, expect } from 'vitest'
import { publicBarberPhotoUrl } from '../publicBarberPhotoUrl'

describe('publicBarberPhotoUrl (DEC-127)', () => {
  it('returns null while the barber has no photo', () => {
    expect(publicBarberPhotoUrl('mi-barberia', { id: 'a', photoUpdatedAt: null })).toBeNull()
  })

  it('builds the public photo URL versioned with photoUpdatedAt', () => {
    expect(
      publicBarberPhotoUrl('mi-barberia', { id: 'a', photoUpdatedAt: '2026-09-30T12:00:00Z' }),
    ).toBe('/api/v1/public/barbershops/mi-barberia/barbers/a/photo?v=2026-09-30T12%3A00%3A00Z')
  })

  it('encodes the slug and the barber id', () => {
    expect(
      publicBarberPhotoUrl('a/b?c', { id: 'x y', photoUpdatedAt: '2026-09-30T12:00:00Z' }),
    ).toContain('/barbershops/a%2Fb%3Fc/barbers/x%20y/photo')
  })
})
