/**
 * Pruebas de PublicServiceCatalogPage (HU-091): carga, éxito (lista de
 * servicios con nombre/duración/precio), catálogo vacío, enlace no
 * disponible/desconocido/no publicable (CA-090-02, reutilizado), error de
 * red con reintento, error inesperado con requestId, selección accesible
 * por teclado y clic, y ausencia de violaciones de accesibilidad (axe-core)
 * en cada estado observable (CA-091-04, CA-091-05).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: vi.fn() },
}))

const { default: PublicServiceCatalogPage } = await import('../PublicServiceCatalogPage.vue')

// RouterLink stubbed en vez de un router real (mismo patrón que
// PublicBarbershopEntryPage.test.ts): esta suite verifica el componente en
// aislamiento, no el enrutamiento, y el CTA "Continuar" (HU-092) navega a
// la ruta de selección de barbero del servicio elegido.
const RouterLinkStub = { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' }

function mountCatalogPage(props: { slug: string }) {
  return mount(PublicServiceCatalogPage, {
    props,
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
}

function okResponse(): Response {
  return { ok: true, status: 200, headers: new Headers() } as Response
}

function errorResponse(status: number): Response {
  return { ok: false, status, headers: new Headers() } as Response
}

function serviceFixture(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
    name: 'Corte clásico',
    description: 'Corte con máquina y tijera, incluye lavado.',
    durationMinutes: 30,
    price: '45000.00',
    currency: 'COP',
    ...overrides,
  }
}

beforeEach(() => {
  getMock.mockReset()
})

describe('PublicServiceCatalogPage', () => {
  it('shows the loading state while the request is in flight', async () => {
    let resolveRequest!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve
      }),
    )
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })

    expect(wrapper.text()).toContain('Cargando servicios')

    resolveRequest({ data: undefined, error: undefined, response: errorResponse(500) })
    await flushPromises()
  })

  it('resolves the slug from props (CA-090-03, reused)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [serviceFixture()], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/public/barbershops/{slug}/services',
      expect.objectContaining({
        params: { path: { slug: 'barberia-ejemplo' }, query: { limit: 50 } },
      }),
    )
  })

  it('renders name, duration and price of each active+assigned service (CA-091-01)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          serviceFixture({
            id: 'a',
            name: 'Corte clásico',
            durationMinutes: 30,
            price: '45000.00',
          }),
          serviceFixture({
            id: 'b',
            name: 'Barba',
            description: null,
            durationMinutes: 20,
            price: '20000.00',
          }),
        ],
        nextCursor: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.text()).toContain('30 min')
    expect(wrapper.text()).toContain('45000.00 COP')
    expect(wrapper.text()).toContain('Barba')
    expect(wrapper.text()).toContain('20 min')
    expect(wrapper.text()).toContain('20000.00 COP')
  })

  it('shows an empty state when the barbershop has no public services yet (CA-091-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    expect(wrapper.text()).toContain('todavía no tiene servicios disponibles')
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)
  })

  it('shows the same uniform message for a 404 (malformed, unknown or non-publishable, CA-090-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mountCatalogPage({ slug: 'no-existe' })
    await flushPromises()

    expect(wrapper.text()).toContain('No encontramos ese enlace')
  })

  it('offers a retry action on a network error, re-fetching the same slug', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos conectar')

    getMock.mockResolvedValueOnce({
      data: { items: [serviceFixture()], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Corte clásico')
    expect(getMock).toHaveBeenCalledTimes(2)
  })

  it('shows the requestId on an unexpected error, without leaking the raw problem body', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal', title: 'Error interno', requestId: 'req-abc-123' },
      response: errorResponse(500),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    expect(wrapper.text()).toContain('req-abc-123')
    expect(wrapper.text()).toContain('Ocurrió un error inesperado')
  })

  it('selects a service on click, marking it with aria-checked (CA-091-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          serviceFixture({ id: 'a', name: 'Corte clásico' }),
          serviceFixture({ id: 'b', name: 'Barba' }),
        ],
        nextCursor: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    expect(options).toHaveLength(2)
    expect(options[0]!.attributes('aria-checked')).toBe('false')
    expect(options[1]!.attributes('aria-checked')).toBe('false')

    await options[1]!.trigger('click')

    expect(options[0]!.attributes('aria-checked')).toBe('false')
    expect(options[1]!.attributes('aria-checked')).toBe('true')
  })

  it('selects the focused option with the keyboard (Space/Enter, CA-091-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          serviceFixture({ id: 'a', name: 'Corte clásico' }),
          serviceFixture({ id: 'b', name: 'Barba' }),
        ],
        nextCursor: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    const list = wrapper.find('[role="radiogroup"]')
    await list.trigger('keydown', { key: 'ArrowDown' })
    await list.trigger('keydown', { key: 'Enter' })

    const options = wrapper.findAll('[role="radio"]')
    expect(options[1]!.attributes('aria-checked')).toBe('true')
  })

  it('has no accessibility violations in the loading state', async () => {
    getMock.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations with a populated list', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [serviceFixture({ id: 'a' }), serviceFixture({ id: 'b', name: 'Barba' })],
        nextCursor: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations in the empty state', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
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
    const wrapper = mountCatalogPage({ slug: 'no-existe' })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('hides the "Continuar" CTA until a service is selected (HU-092)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [serviceFixture({ id: 'a' })], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    expect(wrapper.text()).not.toContain('Continuar')
  })

  it('shows "Continuar" navigating to the barber selection route of the selected service (HU-092)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [serviceFixture({ id: 'a' })], nextCursor: null },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountCatalogPage({ slug: 'barberia-ejemplo' })
    await flushPromises()

    await wrapper.find('[role="radio"]').trigger('click')

    const cta = wrapper.findComponent(RouterLinkStub)
    expect(cta.text()).toContain('Continuar')
    expect(cta.props('to')).toEqual({
      name: 'reserva-publica-barbero',
      params: { slug: 'barberia-ejemplo', serviceId: 'a' },
    })
  })
})
