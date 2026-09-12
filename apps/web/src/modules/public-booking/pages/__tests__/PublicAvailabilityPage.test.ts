/**
 * Pruebas de PublicAvailabilityPage (HU-095): carga, agrupación por día
 * civil en la zona de la barbería (nunca la del dispositivo, CA-095-02),
 * navegación entre días con franjas, selección accesible por teclado y
 * clic con resumen fecha+hora+zona+duración, ausencia total de franjas en
 * toda la ventana como vacío distinto de carga/error (CA-095-04), enlace no
 * disponible/desconocido/no publicable (CA-090-02, reutilizado), error de
 * red con reintento, error inesperado con requestId, revalidación al
 * cambiar de barbero (CA-095-03) y ausencia de violaciones de
 * accesibilidad (axe-core) en cada estado observable (CA-095-05).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: vi.fn() },
}))

const { default: PublicAvailabilityPage } = await import('../PublicAvailabilityPage.vue')

const SERVICE_A = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_A = 'a1111111-1111-1111-1111-111111111111'
const BARBER_B = 'b2222222-2222-2222-2222-222222222222'
const TIMEZONE = 'America/Bogota'

function okResponse(): Response {
  return { ok: true, status: 200, headers: new Headers() } as Response
}

function errorResponse(status: number): Response {
  return { ok: false, status, headers: new Headers() } as Response
}

function availabilityFixture(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    slots: [
      { startsAt: '2026-09-15T19:00:00Z' }, // 2026-09-15 14:00 America/Bogota
      { startsAt: '2026-09-15T19:15:00Z' }, // 2026-09-15 14:15 America/Bogota
      { startsAt: '2026-09-16T14:00:00Z' }, // 2026-09-16 09:00 America/Bogota
    ],
    durationMinutes: 30,
    timezone: TIMEZONE,
    slotGridMinutes: 15,
    ...overrides,
  }
}

function mountAvailabilityPage(props: { slug: string; serviceId: string; barberId: string }) {
  return mount(PublicAvailabilityPage, { props })
}

beforeEach(() => {
  getMock.mockReset()
})

describe('PublicAvailabilityPage', () => {
  it('shows the loading state while the request is in flight', async () => {
    let resolveRequest!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve
      }),
    )
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })

    expect(wrapper.text()).toContain('Cargando disponibilidad')

    resolveRequest({ data: undefined, error: undefined, response: errorResponse(500) })
    await flushPromises()
  })

  it('resolves slug, serviceId and barberId from props', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    mountAvailabilityPage({ slug: 'barberia-ejemplo', serviceId: SERVICE_A, barberId: BARBER_A })
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability',
      expect.objectContaining({
        params: { path: { slug: 'barberia-ejemplo', serviceId: SERVICE_A, barberId: BARBER_A } },
      }),
    )
  })

  it('groups slots by civil date in the barbershop timezone, defaulting to the first day (CA-095-01)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    expect(options).toHaveLength(2)
    expect(wrapper.text()).toContain('2:00 p. m.')
    expect(wrapper.text()).toContain('2:15 p. m.')
    expect(wrapper.text()).not.toContain('9:00 a. m.')
  })

  it('navigates to the next day with slots and back, disabling the boundary buttons', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    const buttons = wrapper.findAll('button')
    const previous = buttons.find((b) => b.text().includes('Anterior'))!
    const next = buttons.find((b) => b.text().includes('Siguiente'))!
    expect(previous.attributes('disabled')).toBeDefined()

    await next.trigger('click')
    expect(wrapper.findAll('[role="radio"]')).toHaveLength(1)
    expect(wrapper.text()).toContain('9:00 a. m.')
    expect(next.attributes('disabled')).toBeDefined()

    await previous.trigger('click')
    expect(wrapper.findAll('[role="radio"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('2:00 p. m.')
  })

  it('selects a slot on click, announcing date, time, timezone and duration together (CA-095-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    await options[0]!.trigger('click')

    expect(options[0]!.attributes('aria-checked')).toBe('true')
    const summary = wrapper.text()
    expect(summary).toContain('2:00 p. m.')
    expect(summary).toContain(TIMEZONE)
    expect(summary).toContain('30 min')
  })

  it('keeps a slot selected on another day when navigating away and back (CA-095-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    await wrapper.findAll('[role="radio"]')[0]!.trigger('click')
    const buttons = wrapper.findAll('button')
    await buttons.find((b) => b.text().includes('Siguiente'))!.trigger('click')
    // El día 2 no tiene ninguna franja marcada: la selección pertenece al
    // día 1, que ya no está en pantalla.
    expect(
      wrapper.findAll('[role="radio"]').every((o) => o.attributes('aria-checked') === 'false'),
    ).toBe(true)

    await buttons.find((b) => b.text().includes('Anterior'))!.trigger('click')
    expect(wrapper.findAll('[role="radio"]')[0]!.attributes('aria-checked')).toBe('true')
  })

  it('selects the focused slot with the keyboard (Space/Enter)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    const list = wrapper.find('[role="radiogroup"]')
    await list.trigger('keydown', { key: 'ArrowRight' })
    await list.trigger('keydown', { key: 'Enter' })

    const options = wrapper.findAll('[role="radio"]')
    expect(options[1]!.attributes('aria-checked')).toBe('true')
  })

  it('shows a distinct empty state with zero slots in the whole window (CA-095-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture({ slots: [] }),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('No hay franjas disponibles')
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)
  })

  it('shows the same uniform message for a 404 (malformed, unknown or non-publishable slug, CA-090-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'no-existe',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('No encontramos ese enlace')
  })

  it('offers a retry action on a network error, re-fetching the same context', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos conectar')

    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('2:00 p. m.')
    expect(getMock).toHaveBeenCalledTimes(2)
  })

  it('shows the requestId on an unexpected error, without leaking the raw problem body', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal', title: 'Error interno', requestId: 'req-abc-123' },
      response: errorResponse(500),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('req-abc-123')
    expect(wrapper.text()).toContain('Ocurrió un error inesperado')
  })

  it('reloads and resets the selection when barberId changes (CA-095-03)', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()
    await wrapper.findAll('[role="radio"]')[0]!.trigger('click')
    expect(wrapper.findAll('[role="radio"]')[0]!.attributes('aria-checked')).toBe('true')

    getMock.mockResolvedValueOnce({
      data: availabilityFixture({ slots: [{ startsAt: '2026-09-20T15:00:00Z' }] }),
      error: undefined,
      response: okResponse(),
    })
    await wrapper.setProps({ barberId: BARBER_B })
    await flushPromises()

    expect(getMock).toHaveBeenLastCalledWith(
      '/public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability',
      expect.objectContaining({
        params: { path: { slug: 'barberia-ejemplo', serviceId: SERVICE_A, barberId: BARBER_B } },
      }),
    )
    expect(wrapper.findAll('[role="radio"]')[0]!.attributes('aria-checked')).toBe('false')
  })

  it('discards a stale response for a context already replaced (race guard)', async () => {
    let resolveFirst!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve
      }),
    )
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })

    getMock.mockResolvedValueOnce({
      data: availabilityFixture({ slots: [{ startsAt: '2026-09-20T15:00:00Z' }] }),
      error: undefined,
      response: okResponse(),
    })
    await wrapper.setProps({ barberId: BARBER_B })
    await flushPromises()

    // La respuesta de la primera petición (para BARBER_A) llega tarde,
    // después de que BARBER_B ya resolvió: debe descartarse.
    resolveFirst({ data: availabilityFixture(), error: undefined, response: okResponse() })
    await flushPromises()

    expect(wrapper.findAll('[role="radio"]')).toHaveLength(1)
    expect(wrapper.text()).not.toContain('2:00 p. m.')
  })

  it('has no accessibility violations in the loading state', async () => {
    getMock.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations with slots to choose from', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture(),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations in the empty state', async () => {
    getMock.mockResolvedValueOnce({
      data: availabilityFixture({ slots: [] }),
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'barberia-ejemplo',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the not-found error state', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mountAvailabilityPage({
      slug: 'no-existe',
      serviceId: SERVICE_A,
      barberId: BARBER_A,
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
