import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import { publicBookingRoutes } from '@/modules/public-booking/routes'
import { buildPublicLinkUrl } from '../publicLink'

describe('buildPublicLinkUrl', () => {
  it('joins the origin, the public booking path and the slug', () => {
    expect(buildPublicLinkUrl('https://nava.example', 'corte-fino-k7x2m9')).toBe(
      'https://nava.example/reservar/corte-fino-k7x2m9',
    )
  })

  it('does not duplicate a trailing slash in the origin', () => {
    expect(buildPublicLinkUrl('https://nava.example/', 'corte-fino-k7x2m9')).toBe(
      'https://nava.example/reservar/corte-fino-k7x2m9',
    )
  })

  it('is exactly the URL the public entry route resolves to', () => {
    const router = createRouter({ history: createMemoryHistory(), routes: publicBookingRoutes })
    const resolved = router.resolve({
      name: 'reserva-publica-entrada',
      params: { slug: 'corte-fino-k7x2m9' },
    })

    expect(buildPublicLinkUrl('http://localhost', 'corte-fino-k7x2m9')).toBe(
      `http://localhost${resolved.href}`,
    )
  })
})
