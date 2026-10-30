/**
 * Pruebas de CustomerAppointmentPage (HU-098): carga, éxito (barbería,
 * persona atendida, servicio, barbero, intervalo/zona, estado y política de
 * cancelación), enlace inválido/inexistente/vencido/revocado (CA-098-02,
 * todos indistinguibles), error de red con reintento, error inesperado con
 * requestId, ausencia de identificadores internos (CA-098-03) y ausencia de
 * violaciones de accesibilidad (axe-core) en cada estado observable.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
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
    versionToken: 'version-token-1',
    barbershopContactEmail: 'contacto@barberia-ejemplo.test',
    barbershopContactPhone: '+573001234567',
    vocabulary: {
      businessTerm: 'barbería',
      businessTermGender: 'feminine',
      professionalTerm: 'barbero',
      professionalTermPlural: 'barberos',
      professionalTermGender: 'masculine',
    },
    canReview: false,
    review: null,
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

  it('un turno ya cancelado se ve como cancelado al abrir el enlace, sin política ni acciones (DEC-136)', async () => {
    getMock.mockResolvedValueOnce({
      data: { ...successData(), status: 'cancelled_by_customer' },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).toContain('Cancelado por el cliente')
    expect(wrapper.text()).toContain('Este turno está cancelado')
    expect(wrapper.text()).not.toContain('Si cancelas')
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.findAll('button')).toHaveLength(0)
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

  it('shows the cancellation policy and, since HU-099, offers «Cancelar turno» only while confirmed', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).toContain('20 minutos antes')
    expect(wrapper.findAll('button').map((b) => b.text())).toEqual(['Cancelar turno'])

    getMock.mockResolvedValueOnce({
      data: { ...successData(), status: 'completed' },
      error: undefined,
      response: okResponse(),
    })
    const completed = mountPage({ token: 'token-en-claro' })
    await flushPromises()
    expect(completed.find('button').exists()).toBe(false)
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

  it('names the business and the professional with the words the barbershop configured (DEC-119)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        ...successData(),
        barbershopName: 'Estudio Lila',
        barberName: 'Laura Mejía',
        status: 'cancelled_by_barber',
        lateCancellationClientAllowed: false,
        vocabulary: {
          businessTerm: 'estudio',
          businessTermGender: 'masculine',
          professionalTerm: 'tatuadora',
          professionalTermPlural: 'tatuadoras',
          professionalTermGender: 'feminine',
        },
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Tatuadora')
    expect(text).toContain('Cancelado por la tatuadora')
    expect(text).toContain('Este turno está cancelado')
    expect(text).not.toMatch(/barber(o|a|os|as|ía|ías)\b/i)
  })

  it('names the business and the professional with the words the barbershop configured en la política de un turno vigente (DEC-119)', async () => {
    getMock.mockResolvedValueOnce({
      data: {
        ...successData(),
        barbershopName: 'Estudio Lila',
        barberName: 'Laura Mejía',
        status: 'confirmed',
        lateCancellationClientAllowed: false,
        vocabulary: {
          businessTerm: 'estudio',
          businessTermGender: 'masculine',
          professionalTerm: 'tatuadora',
          professionalTermPlural: 'tatuadoras',
          professionalTermGender: 'feminine',
        },
      },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Tatuadora')
    expect(text).toContain('contacta directamente al estudio')
    expect(text).not.toMatch(/barber(o|a|os|as|ía|ías)\b/i)
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

// HU-100 (DEC-128): «Calificar» solo con el turno completado y sin reseña.
describe('CustomerAppointmentPage · reseña', () => {
  const postMock = vi.fn()

  beforeEach(async () => {
    postMock.mockReset()
    const { httpClient } = await import('@/shared/api/httpClient')
    ;(httpClient as unknown as { POST: typeof postMock }).POST = postMock
  })

  function reviewable(overrides: Record<string, unknown> = {}) {
    return { ...successData(), status: 'completed', canReview: true, review: null, ...overrides }
  }

  async function mountReviewable(data = reviewable()) {
    getMock.mockResolvedValueOnce({ data, error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()
    return wrapper
  }

  it('un turno no calificable no ofrece «Calificar» (CA-100-01)', async () => {
    getMock.mockResolvedValueOnce({ data: successData(), error: undefined, response: okResponse() })
    const wrapper = mountPage({ token: 'token-en-claro' })
    await flushPromises()

    expect(wrapper.text()).not.toContain('Califica a')
    expect(wrapper.find('textarea').exists()).toBe(false)
  })

  it('un turno completado sin reseña muestra el formulario con el aviso de privacidad', async () => {
    const wrapper = await mountReviewable()

    expect(wrapper.get('h2').text()).toBe('Califica a Barbero Ejemplo')
    expect(wrapper.findAll('input[type="radio"]')).toHaveLength(5)
    expect(wrapper.text()).toContain(
      'No escribas teléfonos, direcciones ni nombres de otras personas',
    )
    expect(wrapper.text()).toContain('(opcional)')
  })

  it('no envía sin calificación y avisa (CA-100-02)', async () => {
    const wrapper = await mountReviewable()

    await wrapper.get('form').trigger('submit')

    expect(postMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Elige de 1 a 5 estrellas.')
  })

  it('no envía un comentario de más de 500 caracteres', async () => {
    const wrapper = await mountReviewable()
    await wrapper.findAll('input[type="radio"]')[4]!.setValue(true)
    await wrapper.get('textarea').setValue('a'.repeat(501))

    await wrapper.get('form').trigger('submit')

    expect(postMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('El comentario no puede superar 500 caracteres.')
  })

  it('envía calificación y comentario recortado, y pasa a «Gracias por tu reseña»', async () => {
    postMock.mockResolvedValueOnce({
      data: { rating: 5, comment: 'Muy buen corte', createdAt: '2027-03-15T16:10:00Z' },
      error: undefined,
      response: { ok: true, status: 201, headers: new Headers() } as Response,
    })
    const wrapper = await mountReviewable()
    await wrapper.findAll('input[type="radio"]')[4]!.setValue(true)
    await wrapper.get('textarea').setValue('  Muy buen corte  ')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(postMock).toHaveBeenCalledWith('/customer/appointments/{token}/review', {
      params: { path: { token: 'token-en-claro' } },
      body: { rating: 5, comment: 'Muy buen corte' },
    })
    expect(wrapper.text()).toContain('Gracias por tu reseña')
    expect(wrapper.text()).toContain('Muy buen corte')
    expect(wrapper.find('textarea').exists()).toBe(false)
  })

  it('un comentario vacío viaja como null', async () => {
    postMock.mockResolvedValueOnce({
      data: { rating: 3, comment: null, createdAt: '2027-03-15T16:10:00Z' },
      error: undefined,
      response: { ok: true, status: 201, headers: new Headers() } as Response,
    })
    const wrapper = await mountReviewable()
    await wrapper.findAll('input[type="radio"]')[2]!.setValue(true)
    await wrapper.get('textarea').setValue('   ')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(postMock.mock.calls[0]![1].body).toEqual({ rating: 3, comment: null })
  })

  it('un 409 avisa que ya no admite reseña y recarga el turno', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: { status: 409 },
      response: errorResponse(409),
    })
    const wrapper = await mountReviewable()
    getMock.mockResolvedValueOnce({
      data: reviewable({
        canReview: false,
        review: { rating: 4, comment: null, createdAt: '2027-03-15T16:10:00Z' },
      }),
      error: undefined,
      response: okResponse(),
    })
    await wrapper.findAll('input[type="radio"]')[3]!.setValue(true)

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(getMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Gracias por tu reseña')
  })

  it('un error de red conserva lo escrito y permite reintentar', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    const wrapper = await mountReviewable()
    await wrapper.findAll('input[type="radio"]')[3]!.setValue(true)
    await wrapper.get('textarea').setValue('Mi comentario')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect((wrapper.get('textarea').element as HTMLTextAreaElement).value).toBe('Mi comentario')
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('con la reseña ya dejada la muestra de solo lectura, sin formulario', async () => {
    const wrapper = await mountReviewable(
      reviewable({
        canReview: false,
        review: { rating: 4, comment: 'Todo bien', createdAt: '2027-03-15T16:10:00Z' },
      }),
    )

    expect(wrapper.text()).toContain('Gracias por tu reseña')
    expect(wrapper.text()).toContain('Todo bien')
    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('el formulario y la reseña publicada no tienen violaciones de axe', async () => {
    const form = await mountReviewable()
    document.body.appendChild(form.element)
    expect(await axe(form.element)).toHaveNoViolations()
    form.unmount()

    const done = await mountReviewable(
      reviewable({
        canReview: false,
        review: { rating: 4, comment: 'Todo bien', createdAt: '2027-03-15T16:10:00Z' },
      }),
    )
    document.body.appendChild(done.element)
    expect(await axe(done.element)).toHaveNoViolations()
    done.unmount()
  })
})

describe('CustomerAppointmentPage · cancelación (HU-099)', () => {
  const postMock = vi.fn()

  beforeEach(async () => {
    postMock.mockReset()
    const { httpClient } = await import('@/shared/api/httpClient')
    ;(httpClient as unknown as { POST: typeof postMock }).POST = postMock
  })

  const mounted: Array<{ unmount: () => void }> = []

  afterEach(() => {
    vi.useRealTimers()
    mounted.splice(0).forEach((w) => w.unmount())
  })

  const problem = (status: number, code: string) => ({
    type: `/api/v1/problems/${code}`,
    title: 'x',
    status,
    code,
    requestId: 'req-1',
  })

  async function mountConfirmed(data: Record<string, unknown> = successData()) {
    getMock.mockResolvedValueOnce({ data, error: undefined, response: okResponse() })
    const wrapper = mount(CustomerAppointmentPage, {
      props: { token: 'token-en-claro' },
      attachTo: document.body,
    })
    await flushPromises()
    mounted.push(wrapper)
    return wrapper
  }

  async function openForm(wrapper: Awaited<ReturnType<typeof mountConfirmed>>) {
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Cancelar turno')!
      .trigger('click')
  }

  async function confirm(wrapper: Awaited<ReturnType<typeof mountConfirmed>>) {
    await wrapper.get('form').trigger('submit')
    await flushPromises()
  }

  it('cancela dentro de plazo: manda versión, clave y motivo, y pasa a cancelado sin enlace vigente', async () => {
    postMock.mockResolvedValueOnce({
      data: { status: 'cancelled_by_customer' },
      error: undefined,
      response: okResponse(),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    await wrapper.get('textarea').setValue('  Me surgió algo  ')
    await confirm(wrapper)

    const [path, init] = postMock.mock.calls[0]!
    expect(path).toBe('/customer/appointments/{token}/cancel')
    expect(init.params.path).toEqual({ token: 'token-en-claro' })
    expect(init.params.header['If-Match']).toBe('version-token-1')
    expect(init.params.header['Idempotency-Key']).toMatch(/^[0-9a-f-]{36}$/)
    expect(init.body).toEqual({ reason: 'Me surgió algo' })

    expect(wrapper.text()).toContain('Cancelado por el cliente')
    expect(wrapper.text()).toContain('Cancelaste tu turno')
    expect(wrapper.text()).not.toContain('ya no funciona')
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.findAll('button')).toHaveLength(0)
    expect(document.activeElement).toBe(wrapper.get('.customer-appointment__cancelled').element)
  })

  it('un turno cancelado por el cliente ya no ofrece cancelar ni calificar', async () => {
    const wrapper = await mountConfirmed({ ...successData(), status: 'cancelled_by_customer' })
    expect(wrapper.text()).toContain('Cancelado por el cliente')
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('un reintento tras un fallo de red reutiliza la clave; con otro motivo usa una nueva', async () => {
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    postMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    postMock.mockResolvedValueOnce({
      data: { status: 'cancelled_by_customer' },
      response: okResponse(),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)

    await wrapper.get('textarea').setValue('Motivo uno')
    await confirm(wrapper)
    expect(wrapper.text()).toContain('No pudimos conectar')
    await confirm(wrapper)
    await wrapper.get('textarea').setValue('Motivo dos')
    await confirm(wrapper)

    const keys = postMock.mock.calls.map((c) => c[1].params.header['Idempotency-Key'])
    expect(keys[0]).toBe(keys[1])
    expect(keys[2]).not.toBe(keys[0])
  })

  it('un rechazo por plazo vencido explica y muestra el contacto, y deja el turno confirmado', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: problem(422, 'cancellation-window-closed'),
      response: errorResponse(422),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    await confirm(wrapper)

    expect(wrapper.text()).toContain('El plazo para cancelar por tu cuenta ya venció')
    expect(wrapper.find('a[href="tel:+573001234567"]').exists()).toBe(true)
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.text()).toContain('Confirmado')
  })

  it('un rechazo por motivo faltante vuelve obligatorio el campo y conserva el formulario', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: problem(422, 'cancellation-reason-required'),
      response: errorResponse(422),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    await confirm(wrapper)
    await flushPromises()

    expect(wrapper.find('form').exists()).toBe(true)
    expect(wrapper.text()).toContain('(obligatorio)')
    expect(wrapper.text()).toContain('Cuéntanos brevemente el motivo')
  })

  it('fuera de plazo (reloj local) con motivo obligatorio exige el motivo antes de llamar al API', async () => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2027-03-15T14:50:00Z'))
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    expect(wrapper.text()).toContain('(obligatorio)')

    await confirm(wrapper)
    expect(postMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Cuéntanos brevemente el motivo')
  })

  it('fuera de plazo con la política que no lo permite no hay botón y se muestra el contacto', async () => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2027-03-15T14:50:00Z'))
    const wrapper = await mountConfirmed({
      ...successData(),
      lateCancellationClientAllowed: false,
      lateCancellationReasonRequired: false,
    })
    expect(wrapper.text()).toContain('El plazo para cancelar por tu cuenta ya venció')
    expect(wrapper.find('a[href="mailto:contacto@barberia-ejemplo.test"]').exists()).toBe(true)
    expect(wrapper.findAll('button').map((b) => b.text())).not.toContain('Cancelar turno')
  })

  it('un 409 recarga el turno y avisa que cambió', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: problem(409, 'version-conflict'),
      response: errorResponse(409),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    getMock.mockResolvedValueOnce({
      data: { ...successData(), status: 'cancelled_by_barber' },
      error: undefined,
      response: okResponse(),
    })
    await confirm(wrapper)

    expect(getMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Tu turno cambió mientras tanto')
    expect(wrapper.text()).toContain('Cancelado')
    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('un 404 (enlace ya revocado) muestra el estado de enlace no accesible', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: problem(404, 'not-found'),
      response: errorResponse(404),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    await confirm(wrapper)
    expect(wrapper.text()).toContain('No pudimos abrir ese enlace')
  })

  it('un error inesperado muestra el código de soporte y permite reintentar', async () => {
    postMock.mockResolvedValueOnce({
      data: undefined,
      error: problem(500, 'internal-error'),
      response: errorResponse(500),
    })
    const wrapper = await mountConfirmed()
    await openForm(wrapper)
    await confirm(wrapper)
    expect(wrapper.text()).toContain('req-1')
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('no tiene violaciones de axe en el formulario, el rechazo ni el estado cancelado', async () => {
    const form = await mountConfirmed()
    await openForm(form)
    expect(await axe(form.element)).toHaveNoViolations()
    form.unmount()

    postMock.mockResolvedValueOnce({
      data: { status: 'cancelled_by_customer' },
      response: okResponse(),
    })
    const done = await mountConfirmed()
    await openForm(done)
    await confirm(done)
    expect(await axe(done.element)).toHaveNoViolations()
  })
})
