/**
 * Reseñas públicas del barbero elegido (HU-100, DEC-128, CA-100-04): se piden
 * al abrir el desplegable, se muestran anónimas, admiten reintento y se
 * descartan al cambiar de barbero.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: getMock } }))

const { default: BarberReviews } = await import('../BarberReviews.vue')

const BARBER = 'a1111111-1111-1111-1111-111111111111'

function ok(body: unknown) {
  return { data: body, error: undefined, response: { ok: true, status: 200 } as Response }
}
function fail(status: number) {
  return { data: undefined, error: {}, response: { ok: false, status } as Response }
}

const list = {
  rating: 4.5,
  ratingCount: 2,
  items: [
    { rating: 5, comment: 'Muy buen corte', createdAt: '2027-03-15T16:10:00Z' },
    { rating: 4, comment: null, createdAt: '2027-03-10T12:00:00Z' },
  ],
}

function mountIt(barberId = BARBER) {
  return mount(BarberReviews, {
    attachTo: document.body,
    props: { slug: 'barberia-ejemplo', barberId, barberName: 'Juan Pérez', ratingCount: 2 },
  })
}

beforeEach(() => getMock.mockReset())

describe('BarberReviews', () => {
  it('asks for nothing until the visitor opens the list', () => {
    const wrapper = mountIt()

    expect(getMock).not.toHaveBeenCalled()
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    expect(wrapper.text()).toContain('Ver reseñas (2)')
  })

  it('loads and lists the visible reviews, with comment only when there is one', async () => {
    getMock.mockResolvedValueOnce(ok(list))
    const wrapper = mountIt()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/public/barbershops/{slug}/barbers/{barberId}/reviews',
      expect.objectContaining({ params: { path: { slug: 'barberia-ejemplo', barberId: BARBER } } }),
    )
    const items = wrapper.findAll('li')
    expect(items).toHaveLength(2)
    expect(items[0]!.text()).toContain('Muy buen corte')
    expect(items[1]!.find('.pb-reviews__comment').exists()).toBe(false)
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('true')
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })

  it('does not request the list again when it is closed and reopened', async () => {
    getMock.mockResolvedValueOnce(ok(list))
    const wrapper = mountIt()

    await wrapper.get('button').trigger('click')
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(getMock).toHaveBeenCalledTimes(1)
  })

  it('offers a retry after an error', async () => {
    getMock.mockResolvedValueOnce(fail(500)).mockResolvedValueOnce(ok(list))
    const wrapper = mountIt()

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('No pudimos cargar las reseñas')

    await wrapper.get('.pb-reviews__retry').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('li')).toHaveLength(2)
  })

  it('says so when no review is visible', async () => {
    getMock.mockResolvedValueOnce(ok({ rating: null, ratingCount: 0, items: [] }))
    const wrapper = mountIt()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Todavía no hay reseñas')
  })

  it('closes and forgets the list when the barber changes', async () => {
    getMock.mockResolvedValueOnce(ok(list))
    const wrapper = mountIt()
    await wrapper.get('button').trigger('click')
    await flushPromises()

    await wrapper.setProps({ barberId: 'b2222222-2222-2222-2222-222222222222' })

    expect(wrapper.findAll('li')).toHaveLength(0)
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
  })
})
