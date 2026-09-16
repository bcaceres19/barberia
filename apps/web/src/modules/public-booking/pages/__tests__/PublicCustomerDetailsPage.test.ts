/**
 * Pruebas de PublicCustomerDetailsPage (HU-096/HU-097): formulario
 * condicional para sí mismo/otra persona (CA-096-01), campos obligatorios
 * y opcionales (CA-096-02), validación con conservación de datos ante
 * error (CA-096-03), resumen local antes de confirmar, confirmación
 * pública real (éxito, conflicto con alternativas, red, doble toque
 * bloqueado, CA-097-01 a CA-097-07) y ausencia de violaciones de
 * accesibilidad (CA-096-06). confirmPublicAppointmentApi se sustituye por
 * un doble de prueba; la persistencia real vive en las pruebas de backend
 * (publicbooking/confirm_test.go, booking/postgres/public_repository_test.go).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const confirmPublicAppointmentMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/confirmPublicAppointmentApi', () => ({
  confirmPublicAppointment: confirmPublicAppointmentMock,
}))

const { default: PublicCustomerDetailsPage } = await import('../PublicCustomerDetailsPage.vue')

const BASE_PROPS = {
  slug: 'barberia-ejemplo',
  serviceId: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  barberId: 'a1111111-1111-1111-1111-111111111111',
  startsAt: '2026-09-20T14:30:00Z',
}

const CONFIRMED_APPOINTMENT = {
  attendeeName: 'Ana Ríos',
  barbershopName: 'Barbería Ejemplo',
  serviceName: 'Corte clásico',
  durationMinutes: 30,
  priceAmount: '20000.00',
  currency: 'COP',
  startsAt: '2026-09-20T14:30:00Z',
  endsAt: '2026-09-20T15:00:00Z',
  timezone: 'America/Bogota',
  accessToken: 'abc123',
  customerNote: null,
}

function mountPage() {
  return mount(PublicCustomerDetailsPage, { props: BASE_PROPS })
}

beforeEach(() => {
  confirmPublicAppointmentMock.mockReset()
})

async function fillValidForm(wrapper: ReturnType<typeof mountPage>) {
  await wrapper.find('input[type="text"]').setValue('Ana Ríos')
  await wrapper.find('input[type="tel"]').setValue('+573001234567')
  await wrapper.find('input[type="email"]').setValue('ana@example.com')
}

describe('PublicCustomerDetailsPage', () => {
  it('shows required-field errors and preserves entered data on an invalid attempt (CA-096-02, CA-096-03)', async () => {
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('El teléfono es obligatorio.')
    expect(wrapper.text()).toContain('El correo es obligatorio.')
    expect((wrapper.find('input[type="text"]').element as HTMLInputElement).value).toBe('Ana Ríos')
  })

  it('derives the attendee name from the customer without asking twice when booking for self (CA-096-01)', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    // "Para mí" nunca pide un segundo nombre: un solo campo de texto en el
    // formulario (el nombre del cliente).
    expect(wrapper.findAll('input[type="text"]')).toHaveLength(1)

    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('Ana Ríos')
    const rows = wrapper.findAll('.customer-details__summary-row')
    const attendedRow = rows.find((row) => row.text().includes('Atiende a'))!
    expect(attendedRow.text()).toContain('Ana Ríos')
  })

  it('requires a non-empty attendee name when booking for someone else (CA-096-01)', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.findAll('input[type="radio"]')[1]!.setValue()

    expect(wrapper.findAll('input[type="text"]')).toHaveLength(2)

    await wrapper.find('form').trigger('submit')
    expect(wrapper.text()).toContain('El nombre de la persona atendida es obligatorio.')

    const attendeeInput = wrapper.findAll('input[type="text"]')[1]!
    await attendeeInput.setValue('Mateo Ruiz')
    await wrapper.find('form').trigger('submit')

    const rows = wrapper.findAll('.customer-details__summary-row')
    const attendedRow = rows.find((row) => row.text().includes('Atiende a'))!
    expect(attendedRow.text()).toContain('Mateo Ruiz')
    const clientRow = rows.find((row) => row.text().includes('Cliente'))!
    expect(clientRow.text()).toContain('Ana Ríos')
  })

  it('lets the customer go back to editing without losing the data already entered', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.find('button[type="button"]').trigger('click')
    expect(wrapper.find('form').exists()).toBe(true)
    expect((wrapper.find('input[type="tel"]').element as HTMLInputElement).value).toBe(
      '+573001234567',
    )
  })

  it('rejects an invalid phone/email format', async () => {
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('input[type="tel"]').setValue('3001234567')
    await wrapper.find('input[type="email"]').setValue('no-arroba')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('indicativo de país')
    expect(wrapper.text()).toContain('formato inválido')
  })

  it('has no accessibility violations on the form', async () => {
    const wrapper = mountPage()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the form with attempted errors', async () => {
    const wrapper = mountPage()
    await wrapper.find('form').trigger('submit')
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the summary', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('confirms the appointment and shows the server confirmation (CA-097-01, CA-097-07)', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await flushPromises()

    expect(confirmPublicAppointmentMock).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('¡Tu turno quedó confirmado!')
    expect(wrapper.text()).toContain('Barbería Ejemplo')
    expect(wrapper.find('form').exists()).toBe(false)
    // El token en claro nunca se muestra en pantalla (DEC-089): solo viaja
    // por el correo de confirmación.
    expect(wrapper.text()).not.toContain('abc123')
  })

  it('blocks a second tap while the first confirmation is in flight (CA-097-07)', async () => {
    let resolveFirst: (value: unknown) => void = () => {}
    confirmPublicAppointmentMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveFirst = resolve
        }),
    )
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await confirmButton.trigger('click')
    await confirmButton.trigger('click')

    expect(confirmPublicAppointmentMock).toHaveBeenCalledTimes(1)
    resolveFirst({ kind: 'success', appointment: CONFIRMED_APPOINTMENT })
    await flushPromises()
  })

  it('shows nearby alternatives and lets the customer retry with one (RN-CON-05, DEC-090)', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [{ startsAt: '2026-09-20T15:00:00Z' }, { startsAt: '2026-09-20T15:30:00Z' }],
    })
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Ese horario se acaba de ocupar')
    // Los datos del formulario nunca se pierden ante el conflicto (RN-CON-05).
    expect(wrapper.text()).toContain('+573001234567')

    const alternativeButtons = wrapper.findAll('.customer-details__alternatives button')
    expect(alternativeButtons.length).toBe(2)
    await alternativeButtons[0]!.trigger('click')

    const retryButton = wrapper.findAll('button').find((b) => b.text().includes('Confirmar turno'))!
    await retryButton.trigger('click')
    await flushPromises()

    expect(confirmPublicAppointmentMock).toHaveBeenCalledTimes(2)
    // El reintento con una alternativa distinta usa una clave de
    // idempotencia nueva (RN-IDE-01): el contenido cambió (otro startsAt).
    const firstKey = confirmPublicAppointmentMock.mock.calls[0]![4]
    const secondKey = confirmPublicAppointmentMock.mock.calls[1]![4]
    expect(secondKey).not.toBe(firstKey)
    expect(wrapper.text()).toContain('¡Tu turno quedó confirmado!')
  })

  it('shows a network error and lets the customer retry with the same idempotency key', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({ kind: 'network-error' })
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar con el servidor')
    expect(wrapper.find('form').exists()).toBe(false)

    await confirmButton.trigger('click')
    await flushPromises()

    expect(confirmPublicAppointmentMock).toHaveBeenCalledTimes(2)
    const firstKey = confirmPublicAppointmentMock.mock.calls[0]![4]
    const secondKey = confirmPublicAppointmentMock.mock.calls[1]![4]
    expect(secondKey).toBe(firstKey)
    expect(wrapper.text()).toContain('¡Tu turno quedó confirmado!')
  })

  it('has no accessibility violations on the schedule-conflict state', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [{ startsAt: '2026-09-20T15:00:00Z' }],
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the confirmed state', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    const confirmButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
    await confirmButton.trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
