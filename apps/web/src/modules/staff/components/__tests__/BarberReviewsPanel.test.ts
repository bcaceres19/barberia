/**
 * Moderación de reseñas en el panel (HU-100, DEC-128, CA-100-05): lista con las
 * ocultas y su fecha, ocultar/mostrar reversible y promedio recalculado por el servidor.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({ httpClient: { GET: getMock, PATCH: patchMock } }))

const { default: BarberReviewsPanel } = await import('../BarberReviewsPanel.vue')

const BARBER = 'a1111111-1111-1111-1111-111111111111'
const R1 = 'b0000000-0000-0000-0000-000000000001'
const R2 = 'b0000000-0000-0000-0000-000000000002'

function ok(body: unknown) {
  return { data: body, error: undefined, response: { ok: true, status: 200 } as Response }
}
function fail(status: number) {
  return { data: undefined, error: {}, response: { ok: false, status } as Response }
}

const both = {
  rating: 5,
  ratingCount: 1,
  items: [
    { id: R1, rating: 5, comment: 'Excelente', hiddenAt: null, createdAt: '2027-03-15T16:10:00Z' },
    {
      id: R2,
      rating: 1,
      comment: 'Texto inapropiado',
      hiddenAt: '2027-03-16T10:00:00Z',
      createdAt: '2027-03-14T09:00:00Z',
    },
  ],
}

function mountIt() {
  return mount(BarberReviewsPanel, {
    attachTo: document.body,
    props: { barberId: BARBER, rating: 3, ratingCount: 2 },
  })
}

async function openList(wrapper: ReturnType<typeof mountIt>) {
  await wrapper.get('button').trigger('click')
  await flushPromises()
}

beforeEach(() => {
  getMock.mockReset()
  patchMock.mockReset()
})

describe('BarberReviewsPanel', () => {
  it('shows the summary from the barber without requesting the list', () => {
    const wrapper = mountIt()

    expect(getMock).not.toHaveBeenCalled()
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('3,0 de 5 · 2 reseñas')
  })

  it('lists every review, hidden ones included with the date they were hidden', async () => {
    getMock.mockResolvedValueOnce(ok(both))
    const wrapper = mountIt()
    await openList(wrapper)

    expect(getMock).toHaveBeenCalledWith('/private/barbers/{barberId}/reviews', {
      params: { path: { barberId: BARBER } },
    })
    const items = wrapper.findAll('li')
    expect(items).toHaveLength(2)
    expect(items[0]!.text()).not.toContain('Oculta desde')
    expect(items[1]!.text()).toContain('Oculta desde')
    expect(items[1]!.text()).toContain('Texto inapropiado')
    expect(items[0]!.get('button').text()).toBe('Ocultar')
    expect(items[1]!.get('button').text()).toBe('Mostrar')
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })

  it('hides a review, reloads the server average and announces the change', async () => {
    getMock.mockResolvedValueOnce(ok(both)).mockResolvedValueOnce(
      ok({
        rating: null,
        ratingCount: 0,
        items: both.items.map((r) => ({ ...r, hiddenAt: '2027-03-17T08:00:00Z' })),
      }),
    )
    patchMock.mockResolvedValueOnce(ok({ ...both.items[0], hiddenAt: '2027-03-17T08:00:00Z' }))
    const wrapper = mountIt()
    await openList(wrapper)

    await wrapper.findAll('li')[0]!.get('button').trigger('click')
    await flushPromises()

    expect(patchMock).toHaveBeenCalledWith('/private/barbers/{barberId}/reviews/{reviewId}', {
      params: { path: { barberId: BARBER, reviewId: R1 } },
      body: { hidden: true },
    })
    expect(wrapper.emitted('changed')).toEqual([[{ rating: null, ratingCount: 0 }]])
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('Sin reseñas todavía')
  })

  it('shows a hidden review again with hidden: false', async () => {
    getMock.mockResolvedValue(ok(both))
    patchMock.mockResolvedValueOnce(ok({ ...both.items[1], hiddenAt: null }))
    const wrapper = mountIt()
    await openList(wrapper)

    await wrapper.findAll('li')[1]!.get('button').trigger('click')
    await flushPromises()

    expect(patchMock.mock.calls[0]![1].body).toEqual({ hidden: false })
  })

  it('keeps the list and tells the user when the change fails', async () => {
    getMock.mockResolvedValueOnce(ok(both))
    patchMock.mockResolvedValueOnce(fail(500))
    const wrapper = mountIt()
    await openList(wrapper)

    await wrapper.findAll('li')[0]!.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toContain('No pudimos actualizar la reseña')
    expect(wrapper.findAll('li')).toHaveLength(2)
    expect(wrapper.emitted('changed')).toBeUndefined()
  })

  it('offers a retry when the list cannot be loaded', async () => {
    getMock.mockResolvedValueOnce(fail(500)).mockResolvedValueOnce(ok(both))
    const wrapper = mountIt()
    await openList(wrapper)
    expect(wrapper.get('[role="alert"]').text()).toContain('No pudimos cargar las reseñas')

    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('li')).toHaveLength(2)
  })
})
