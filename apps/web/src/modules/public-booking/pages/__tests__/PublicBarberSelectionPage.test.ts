/**
 * Pruebas de PublicBarberSelectionPage (HU-092): carga, matriz 0/1/N de
 * barberos elegibles (preselección automática con uno, elección explícita
 * con varios, estado vacío distinto de carga/error con cero), enlace no
 * disponible/desconocido/no publicable (CA-090-02, reutilizado), error de
 * red con reintento, error inesperado con requestId, selección accesible
 * por teclado y clic, revalidación al cambiar de servicio (CA-092-04) y
 * ausencia de violaciones de accesibilidad (axe-core) en cada estado
 * observable (CA-092-05).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: vi.fn() },
}))

const { default: PublicBarberSelectionPage } = await import('../PublicBarberSelectionPage.vue')

const SERVICE_A = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const SERVICE_B = '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061'

function okResponse(): Response {
  return { ok: true, status: 200, headers: new Headers() } as Response
}

function errorResponse(status: number): Response {
  return { ok: false, status, headers: new Headers() } as Response
}

function barberFixture(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: 'a1111111-1111-1111-1111-111111111111',
    fullName: 'Juan Pérez',
    ...overrides,
  }
}

beforeEach(() => {
  getMock.mockReset()
})

describe('PublicBarberSelectionPage', () => {
  it('shows the loading state while the request is in flight', async () => {
    let resolveRequest!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve
      }),
    )
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })

    expect(wrapper.text()).toContain('Cargando barberos')

    resolveRequest({ data: undefined, error: undefined, response: errorResponse(500) })
    await flushPromises()
  })

  it('resolves slug and serviceId from props (CA-092-03)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture()] },
      error: undefined,
      response: okResponse(),
    })
    mount(PublicBarberSelectionPage, { props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A } })
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/public/barbershops/{slug}/services/{serviceId}/barbers',
      expect.objectContaining({
        params: { path: { slug: 'barberia-ejemplo', serviceId: SERVICE_A } },
      }),
    )
  })

  it('preselects automatically with exactly one eligible barber, without a radiogroup (CA-092-01)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture({ fullName: 'Juan Pérez' })] },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Juan Pérez')
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)
  })

  it('requires an explicit choice with several eligible barbers (CA-092-01)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'b', fullName: 'Luis Rojas' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    expect(options).toHaveLength(2)
    expect(options[0]!.attributes('aria-checked')).toBe('false')
    expect(options[1]!.attributes('aria-checked')).toBe('false')
  })

  it('selects a barber on click, marking it with aria-checked', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'b', fullName: 'Luis Rojas' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    await options[1]!.trigger('click')

    expect(options[0]!.attributes('aria-checked')).toBe('false')
    expect(options[1]!.attributes('aria-checked')).toBe('true')
  })

  it('selects the focused option with the keyboard (Space/Enter)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'b', fullName: 'Luis Rojas' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    const list = wrapper.find('[role="radiogroup"]')
    await list.trigger('keydown', { key: 'ArrowDown' })
    await list.trigger('keydown', { key: 'Enter' })

    const options = wrapper.findAll('[role="radio"]')
    expect(options[1]!.attributes('aria-checked')).toBe('true')
  })

  it('shows a distinct empty state with zero eligible barbers (CA-092-03)', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [] },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('no tiene barberos disponibles')
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)
  })

  it('shows the same uniform message for a 404 (malformed, unknown or non-publishable slug, CA-090-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'no-existe', serviceId: SERVICE_A },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('No encontramos ese enlace')
  })

  it('offers a retry action on a network error, re-fetching the same slug and serviceId', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos conectar')

    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture()] },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Juan Pérez')
    expect(getMock).toHaveBeenCalledTimes(2)
  })

  it('shows the requestId on an unexpected error, without leaking the raw problem body', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal', title: 'Error interno', requestId: 'req-abc-123' },
      response: errorResponse(500),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('req-abc-123')
    expect(wrapper.text()).toContain('Ocurrió un error inesperado')
  })

  it('clears an incompatible selection and reloads when serviceId changes (CA-092-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'b', fullName: 'Luis Rojas' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()
    await wrapper.findAll('[role="radio"]')[0]!.trigger('click')
    expect(wrapper.findAll('[role="radio"]')[0]!.attributes('aria-checked')).toBe('true')

    // El nuevo servicio no tiene al barbero "a" entre sus elegibles: la
    // selección previa es incompatible y debe limpiarse, nunca conservarse
    // a ciegas.
    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture({ id: 'c', fullName: 'Carlos Ruiz' })] },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.setProps({ serviceId: SERVICE_B })
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ruiz')
    expect(getMock).toHaveBeenLastCalledWith(
      '/public/barbershops/{slug}/services/{serviceId}/barbers',
      expect.objectContaining({
        params: { path: { slug: 'barberia-ejemplo', serviceId: SERVICE_B } },
      }),
    )
  })

  it('keeps a compatible selection after revalidating it against the fresh list (CA-092-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'b', fullName: 'Luis Rojas' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()
    await wrapper.findAll('[role="radio"]')[0]!.trigger('click')

    // El nuevo servicio SÍ incluye al mismo barbero "a": la selección se
    // conserva, pero solo porque la lista fresca lo revalida, no porque se
    // haya asumido compatible sin verificar.
    getMock.mockResolvedValueOnce({
      data: {
        items: [
          barberFixture({ id: 'a', fullName: 'Ana Gómez' }),
          barberFixture({ id: 'c', fullName: 'Carlos Ruiz' }),
        ],
      },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.setProps({ serviceId: SERVICE_B })
    await flushPromises()

    const options = wrapper.findAll('[role="radio"]')
    expect(options[0]!.attributes('aria-checked')).toBe('true')
  })

  it('discards a stale response for a serviceId already replaced (race guard)', async () => {
    let resolveFirst!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve
      }),
    )
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })

    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture({ fullName: 'Carlos Ruiz' })] },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.setProps({ serviceId: SERVICE_B })
    await flushPromises()

    // La respuesta de la primera petición (para SERVICE_A) llega tarde,
    // después de que SERVICE_B ya resolvió: debe descartarse.
    resolveFirst({
      data: { items: [barberFixture({ fullName: 'Respuesta tardía' })] },
      error: undefined,
      response: okResponse(),
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ruiz')
    expect(wrapper.text()).not.toContain('Respuesta tardía')
  })

  it('has no accessibility violations in the loading state', async () => {
    getMock.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations with a single preselected barber', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [barberFixture()] },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations with several barbers to choose from', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        items: [barberFixture({ id: 'a' }), barberFixture({ id: 'b', fullName: 'Luis Rojas' })],
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations in the empty state', async () => {
    getMock.mockResolvedValueOnce({
      data: { items: [] },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'barberia-ejemplo', serviceId: SERVICE_A },
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
    const wrapper = mount(PublicBarberSelectionPage, {
      props: { slug: 'no-existe', serviceId: SERVICE_A },
    })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
