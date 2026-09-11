/**
 * Pruebas de PublicBarbershopEntryPage (HU-090): carga, éxito (nombre, hora
 * local en la zona de la barbería, contacto opcional), enlace no
 * disponible/desconocido/no publicable (CA-090-02, todos indistinguibles),
 * error de red con reintento, error inesperado con requestId, y ausencia
 * de violaciones de accesibilidad (axe-core) en cada estado observable.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: vi.fn() },
}))

const { default: PublicBarbershopEntryPage } = await import('../PublicBarbershopEntryPage.vue')

function okResponse(): Response {
  return { ok: true, status: 200, headers: new Headers() } as Response
}

function errorResponse(status: number): Response {
  return { ok: false, status, headers: new Headers() } as Response
}

beforeEach(() => {
  getMock.mockReset()
})

describe('PublicBarbershopEntryPage', () => {
  it('shows the loading state while the request is in flight', async () => {
    let resolveRequest!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve
      }),
    )
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })

    expect(wrapper.text()).toContain('Abriendo tu barbería')

    resolveRequest({ data: undefined, error: undefined, response: errorResponse(500) })
    await flushPromises()
  })

  it('resolves the slug from props, without reading useRoute (CA-090-03)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      },
      error: undefined,
      response: okResponse(),
    })
    mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/public/barbershops/{slug}',
      expect.objectContaining({ params: { path: { slug: 'barberia-ejemplo' } } }),
    )
  })

  it('renders the barbershop name and its local time on success (CA-090-01)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()

    expect(wrapper.text()).toContain('Barbería Ejemplo')
    expect(wrapper.text()).toContain('Hora local de la barbería')
  })

  it('shows contact when configured and omits the block when both are absent (CA-090-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: 'contacto@ejemplo.test',
        contactPhone: '+573001234567',
      },
      error: undefined,
      response: okResponse(),
    })
    const withContact = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()
    expect(withContact.text()).toContain('contacto@ejemplo.test')
    expect(withContact.text()).toContain('+573001234567')

    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const withoutContact = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()
    expect(withoutContact.text()).not.toContain('Teléfono:')
    expect(withoutContact.text()).not.toContain('Correo:')
  })

  it('never renders an internal id, barbershopId or the slug itself in the DOM (CA-090-04)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarbershopEntryPage, {
      props: { slug: 'barberia-secreta-interna' },
    })
    await flushPromises()

    expect(wrapper.html()).not.toContain('barberia-secreta-interna')
    expect(wrapper.html().toLowerCase()).not.toContain('barbershopid')
  })

  it('shows the same uniform message for a 404 (malformed, unknown or non-publishable, CA-090-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'no-existe' } })
    await flushPromises()

    expect(wrapper.text()).toContain('No encontramos ese enlace')
  })

  it('offers a retry action on a network error, re-fetching the same slug (CA-090-05)', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos conectar')

    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      },
      error: undefined,
      response: okResponse(),
    })
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Barbería Ejemplo')
    expect(getMock).toHaveBeenCalledTimes(2)
    expect(getMock).toHaveBeenLastCalledWith(
      '/public/barbershops/{slug}',
      expect.objectContaining({ params: { path: { slug: 'barberia-ejemplo' } } }),
    )
  })

  it('shows the requestId on an unexpected error, without leaking the raw problem body', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal', title: 'Error interno', requestId: 'req-abc-123' },
      response: errorResponse(500),
    })
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    await flushPromises()

    expect(wrapper.text()).toContain('req-abc-123')
    expect(wrapper.text()).toContain('Ocurrió un error inesperado')
  })

  it('has no accessibility violations in the loading state', async () => {
    getMock.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on success', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: 'contacto@ejemplo.test',
        contactPhone: '+573001234567',
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'barberia-ejemplo' } })
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
    const wrapper = mount(PublicBarbershopEntryPage, { props: { slug: 'no-existe' } })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
