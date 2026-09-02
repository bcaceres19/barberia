/**
 * Pruebas de NewAppointmentPage (HU-061): carga (vacío, error), selección
 * de barbero → carga de servicios asignados (DEC-072), validación de
 * forma, éxito con resumen honesto, conflicto (agenda/bloqueo, DEC-073),
 * conflicto de idempotencia, error de red, doble envío bloqueado y
 * accesibilidad. appointmentsApi se sustituye por un doble de prueba; el
 * recorrido real contra el API vive en el E2E de HU-061.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchAssignedServicesMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const createManualAppointmentMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/appointmentsApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchAssignedServices: fetchAssignedServicesMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
  createManualAppointment: createManualAppointmentMock,
}))

const { default: NewAppointmentPage } = await import('../NewAppointmentPage.vue')

const oneBarber = [{ id: 'b-1', fullName: 'Carlos Ramírez' }]
const twoServices = [
  { id: 's-1', name: 'Corte clásico' },
  { id: 's-2', name: 'Barba' },
]

function mountPage() {
  return mount(NewAppointmentPage)
}

async function mountReady(barbers = oneBarber) {
  fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: barbers })
  fetchBarbershopTimezoneMock.mockResolvedValueOnce({
    kind: 'success',
    timezone: 'America/Bogota',
  })
  const wrapper = mountPage()
  await flushPromises()
  return wrapper
}

async function fillValidForm(wrapper: Awaited<ReturnType<typeof mountReady>>) {
  fetchAssignedServicesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
  const barberSelect = wrapper.get<HTMLSelectElement>('#new-appointment-barber')
  await barberSelect.setValue('b-1')
  await flushPromises()

  const serviceSelect = wrapper.get<HTMLSelectElement>('#new-appointment-service')
  await serviceSelect.setValue('s-1')

  const inputs = wrapper.findAll('input')
  const byLabel = (text: string) =>
    inputs.find((i) => i.element.labels?.[0]?.textContent?.includes(text))!
  await byLabel('Persona atendida').setValue('Juan Pérez')
  await byLabel('Nombre del cliente').setValue('Juan Pérez')
  await byLabel('Fecha del turno').setValue('2026-09-03')
  await byLabel('Hora del turno').setValue('14:30')
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

describe('NewAppointmentPage', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchAssignedServicesMock.mockReset()
    fetchBarbershopTimezoneMock.mockReset()
    createManualAppointmentMock.mockReset()
  })

  it('shows a non-blank loading state, then the barber picker', async () => {
    let resolveBarbers: (value: unknown) => void = () => {}
    fetchBarberSummariesMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveBarbers = resolve
      }),
    )
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const wrapper = mountPage()
    expect(wrapper.text()).toContain('Cargando barberos')

    resolveBarbers({ kind: 'success', items: oneBarber })
    await flushPromises()
    expect(wrapper.find('#new-appointment-barber').exists()).toBe(true)
  })

  it('shows an explicit empty state when there are no barbers', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
  })

  it('shows a recoverable error when barbers fail to load', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'network-error' })
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos cargar esta sección')
  })

  it('loads only the services assigned to the selected barber (DEC-072)', async () => {
    const wrapper = await mountReady()
    fetchAssignedServicesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    await wrapper.get('#new-appointment-barber').setValue('b-1')
    await flushPromises()
    expect(fetchAssignedServicesMock).toHaveBeenCalledWith('b-1')
    const options = wrapper.findAll('#new-appointment-service option')
    expect(options.map((o) => o.text())).toContain('Corte clásico')
  })

  it('blocks submit and shows field errors on an empty form', async () => {
    const wrapper = await mountReady()
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()
    expect(createManualAppointmentMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Elige un barbero')
  })

  it('does not show a "Resumen" section on an untouched form', async () => {
    const wrapper = await mountReady()
    expect(wrapper.find('.new-appointment-page__resumen').exists()).toBe(false)
  })

  it('shows a live "Resumen" with only the fields filled so far (Fase 4b, adopción NAVA)', async () => {
    const wrapper = await mountReady()
    fetchAssignedServicesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    await wrapper.get('#new-appointment-barber').setValue('b-1')
    await flushPromises()

    const resumen = wrapper.get('.new-appointment-page__resumen')
    expect(resumen.text()).toContain('Carlos Ramírez')
    expect(resumen.text()).not.toContain('Servicio')
  })

  it('reflects every filled field in "Resumen", including the readable date', async () => {
    const wrapper = await mountReady()
    await fillValidForm(wrapper)

    const resumen = wrapper.get('.new-appointment-page__resumen')
    expect(resumen.text()).toContain('Carlos Ramírez')
    expect(resumen.text()).toContain('Corte clásico')
    expect(resumen.text()).toContain('Juan Pérez')
    expect(resumen.text()).toContain('14:30')
    expect(resumen.text()).toMatch(/jueves.*3.*septiembre/)
  })

  it('submits successfully and shows an honest summary, no link to a non-existent agenda screen', async () => {
    const wrapper = await mountReady()
    await fillValidForm(wrapper)
    createManualAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: {
        id: 'appt-1',
        barberId: 'b-1',
        serviceId: 's-1',
        attendeeName: 'Juan Pérez',
        startsAt: '2026-09-03T14:30:00-05:00',
        endsAt: '2026-09-03T15:00:00-05:00',
        serviceName: 'Corte clásico',
        durationMinutes: 30,
        priceAmount: '20000.00',
        currency: 'COP',
      },
    })

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Turno registrado')
    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('shows the server conflict detail on a schedule/block conflict (DEC-073)', async () => {
    const wrapper = await mountReady()
    await fillValidForm(wrapper)
    createManualAppointmentMock.mockResolvedValueOnce({
      kind: 'conflict',
      detail: 'el barbero tiene un bloqueo vigente en ese intervalo',
    })
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.text()).toContain('el barbero tiene un bloqueo vigente en ese intervalo')
  })

  it('shows a recoverable message on a network error and keeps the form data', async () => {
    const wrapper = await mountReady()
    await fillValidForm(wrapper)
    createManualAppointmentMock.mockResolvedValueOnce({ kind: 'network-error' })
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos registrar el turno')
    const attendeeInput = wrapper
      .findAll('input')
      .find((i) => i.element.labels?.[0]?.textContent?.includes('Persona atendida'))!
    expect(attendeeInput.element.value).toBe('Juan Pérez')
  })

  it('blocks a second submit while the first one is still in flight', async () => {
    const wrapper = await mountReady()
    await fillValidForm(wrapper)
    let resolveCreate: (value: unknown) => void = () => {}
    createManualAppointmentMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveCreate = resolve
      }),
    )

    const form = wrapper.get('form')
    await form.trigger('submit.prevent')
    await form.trigger('submit.prevent')
    expect(createManualAppointmentMock).toHaveBeenCalledTimes(1)

    resolveCreate({
      kind: 'success',
      appointment: {
        id: 'appt-1',
        barberId: 'b-1',
        serviceId: 's-1',
        attendeeName: 'Juan Pérez',
        startsAt: '2026-09-03T14:30:00-05:00',
        endsAt: '2026-09-03T15:00:00-05:00',
        serviceName: 'Corte clásico',
        durationMinutes: 30,
        priceAmount: '20000.00',
        currency: 'COP',
      },
    })
    await flushPromises()
  })

  it('has no detectable axe violations in the ready state', async () => {
    const wrapper = await mountReady()
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
