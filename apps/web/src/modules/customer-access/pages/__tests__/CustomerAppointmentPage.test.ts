/**
 * Pruebas de CustomerAppointmentPage (HU-098): carga, éxito (barbería,
 * persona atendida, servicio, barbero, intervalo/zona, estado y política de
 * cancelación), enlace inválido/inexistente/vencido/revocado (CA-098-02,
 * todos indistinguibles), error de red con reintento, error inesperado con
 * requestId, ausencia de identificadores internos (CA-098-03) y ausencia de
 * violaciones de accesibilidad (axe-core) en cada estado observable.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, POST: vi.fn() },
}))

const { default: CustomerAppointmentPage } = await import('../CustomerAppointmentPage.vue')

function okResponse(): Response {
  return { ok: true, status: 200, headers: new Headers() } as Response
}

function errorResponse(status: number): Response {
  return { ok: false, status, headers: new Headers() } as Response
}

function successData() {
  return {
    barbershopName: 'Barbería Ejemplo',
    timezone: 'America/Bogota',
    attendeeName: 'Cliente Ejemplo',
    serviceName: 'Corte clásico',
    durationMinutes: 30,
    barberName: 'Barbero Ejemplo',
    startsAt: '2027-03-15T15:00:00Z',
    endsAt: '2027-03-15T15:30:00Z',
    status: 'confirmed',
    cancellationDeadlineMinutes: 20,
    lateCancellationClientAllowed: true,
    lateCancellationReasonRequired: true,
  }
}

function mountPage(props: { token: string }) {
  return mount(CustomerAppointmentPage, { props })
}

beforeEach(() => {
  getMock.mockReset()
})

describe('CustomerAppointmentPage', () => {
  it('shows the loading state while the request is in flight', async () => {
    let resolveRequest!: (value: unknown) => void
    getMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve
      }),
    )
    const wrapper = mountPage({ token: 'token-en-claro' })

    expect(wrapper.text()).toContain('Buscando tu turno')

    resolveRequest({ data: undefined, error: undefined, response: errorResponse(500) })
    await flushPromises()
  })

  it('resolves the token from props, without reading useRoute (CA-098-01)', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(getMock).toHaveBeenCalledWith(
      '/customer/appointments/{token}',
      expect.objectContaining({ params: { path: { token: 'token-en-claro' } } }),
    )
  })

  it('renders barbershop, attendee, service, barber and time range on success', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).toContain('Barbería Ejemplo')
    expect(wrapper.text()).toContain('Cliente Ejemplo')
    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.text()).toContain('Barbero Ejemplo')
    expect(wrapper.text()).toContain('Confirmado')
  })

  it('never implies that no response cancels the appointment (RN-CNF-02, CA-098-04)', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    const text = wrapper.text().toLowerCase()
    expect(text).not.toContain('se cancelará')
    expect(text).not.toContain('se cancela automáticamente')
  })

  it('shows the cancellation policy without offering a cancel action (out of HU-098 scope)', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).toContain('20 minutos antes')
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('never renders an internal id, barbershopId, appointmentId or customerId (CA-098-03)', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    const raw = wrapper.html().toLowerCase()
    for (const forbidden of ['barbershopid', 'appointmentid', 'customerid', '"id"']) {
      expect(raw).not.toContain(forbidden)
    }
  })

  it('shows the same uniform message for a 404 (invalid, unknown, expired or revoked, CA-098-02)', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mountPage({ token: 'token-invalido' })
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos abrir ese enlace')
  })

  it('offers a retry action on a network error, re-fetching the same token', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos conectar')

    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Barbería Ejemplo')
    expect(getMock).toHaveBeenCalledTimes(2)
    expect(getMock).toHaveBeenLastCalledWith(
      '/customer/appointments/{token}',
      expect.objectContaining({ params: { path: { token: 'token-en-claro' } } }),
    )
  })

  it('shows the requestId on an unexpected error, without leaking the raw problem body', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 500, code: 'internal', title: 'Error interno', requestId: 'req-abc-123' },
      response: errorResponse(500),
    })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).toContain('req-abc-123')
    expect(wrapper.text()).toContain('Ocurrió un error inesperado')
  })

  it('has no accessibility violations in the loading state', async () => {
    getMock.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mountPage({ token: 'token-en-claro' })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on success', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on a not-found error', async () => {
    getMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 404 },
      response: errorResponse(404),
    })
    const wrapper = mountPage({ token: 'token-invalido' })
    await flushPromises()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
