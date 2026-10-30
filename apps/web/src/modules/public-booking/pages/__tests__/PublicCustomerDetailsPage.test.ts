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
import { ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { bookingChromeKey } from '../../model/bookingChrome'

const confirmPublicAppointmentMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/confirmPublicAppointmentApi', () => ({
  confirmPublicAppointment: confirmPublicAppointmentMock,
}))

const routerPushMock = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: routerPushMock }),
}))

const listPublicBarbersMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/listPublicBarbersApi', () => ({
  listPublicBarbers: listPublicBarbersMock,
}))

const loadPublicProfileMock = vi.hoisted(() => vi.fn())
vi.mock('../../model/publicProfile', () => ({
  loadPublicProfile: loadPublicProfileMock,
  publicTermsFor: () => null,
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

function profileOutcome(
  offersHomeService: boolean,
  location: { address: string; latitude: number; longitude: number } | null = null,
) {
  return {
    kind: 'success',
    profile: {
      name: 'Barbería Ejemplo',
      timezone: 'America/Bogota',
      contactEmail: null,
      contactPhone: null,
      vocabulary: { barber: 'Barbero', service: 'Servicio', appointment: 'Turno' },
      location,
      offersHomeService,
    },
  }
}

function mountPage() {
  return mount(PublicCustomerDetailsPage, { props: BASE_PROPS })
}

beforeEach(() => {
  // El diálogo de conflicto se teletransporta al <body>: sin limpiarlo, los de
  // pruebas anteriores se acumulan.
  document.body.innerHTML = ''
  confirmPublicAppointmentMock.mockReset()
  routerPushMock.mockReset()
  listPublicBarbersMock.mockReset()
  loadPublicProfileMock.mockReset()
  loadPublicProfileMock.mockResolvedValue(profileOutcome(false))
  listPublicBarbersMock.mockResolvedValue({
    kind: 'success',
    barbers: [
      {
        id: BASE_PROPS.barberId,
        fullName: 'Mateo Cruz',
        description: null,
        photoUpdatedAt: '2026-10-01T10:00:00Z',
      },
    ],
  })
})

async function fillValidForm(wrapper: ReturnType<typeof mountPage>) {
  // Por autocomplete y no por tipo: con el paso a domicilio abierto, el combobox de la
  // dirección también es type="text".
  await wrapper.find('input[autocomplete="name"]').setValue('Ana Ríos')
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
    // Para uno mismo el resumen no repite el nombre como «Turno para».
    const labels = wrapper.findAll('.customer-details__summary-row dt').map((dt) => dt.text())
    expect(labels).toEqual(['Cliente', 'Teléfono', 'Correo'])
    expect(wrapper.text()).not.toContain('Atiende a')
  })

  it('shows the appointment date and time as the headline of the summary', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const when = wrapper.get('.customer-details__when')
    expect(when.get('.customer-details__when-time').text()).toMatch(/\d{1,2}:\d{2}/)
    expect(when.get('.customer-details__when-day').text()).toContain('2026')
  })

  it('shows the chosen barber with their photo in the summary', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    const who = wrapper.get('.customer-details__who')
    expect(who.text()).toContain('Mateo Cruz')
    expect(who.get('img').attributes('src')).toContain(
      `/public/barbershops/barberia-ejemplo/barbers/${BASE_PROPS.barberId}/photo`,
    )
  })

  it('still shows a complete summary when the barber cannot be loaded', async () => {
    listPublicBarbersMock.mockResolvedValue({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')

    expect(wrapper.find('.customer-details__who').exists()).toBe(false)
    expect(wrapper.text()).toContain('Ana Ríos')
  })

  it('tells the customer, before confirming, that Google may invite them at their email (DEC-122)', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    expect(wrapper.find('.customer-details__invite-note').exists()).toBe(false)

    await wrapper.find('form').trigger('submit')

    const note = wrapper.get('.customer-details__invite-note')
    expect(note.text()).toContain('Si el negocio usa Google Calendar')
    expect(note.text()).toContain('recibirás una invitación')
    // Condicional: nunca promete una invitación que el negocio quizá no envíe.
    expect(note.text()).not.toMatch(/^Recibirás/)
    expect(note.get('strong').text()).not.toBe('')
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
    const attendedRow = rows.find((row) => row.text().includes('Turno para'))!
    expect(attendedRow.text()).toContain('Mateo Ruiz')
    const clientRow = rows.find((row) => row.text().includes('Cliente'))!
    expect(clientRow.text()).toContain('Ana Ríos')
  })

  it('sends the E.164 number built from the prefix box and the national number', async () => {
    confirmPublicAppointmentMock.mockResolvedValue({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('select').setValue('MX')
    await wrapper.find('input[type="tel"]').setValue('55 1234 5678')
    await wrapper.find('input[type="email"]').setValue('ana@example.com')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('+52 5512345678')

    await wrapper.findAll('button[type="button"]')[1]!.trigger('click')
    await flushPromises()

    expect(confirmPublicAppointmentMock.mock.calls[0]![3].phone).toBe('+525512345678')
  })

  it('keeps the note box at a fixed size', () => {
    const wrapper = mountPage()
    const note = wrapper.get('textarea')

    expect(note.attributes('rows')).toBe('3')
  })

  it('lets the customer go back to editing without losing the data already entered', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.find('button[type="button"]').trigger('click')
    expect(wrapper.find('form').exists()).toBe(true)
    // El prefijo vive en su casilla: al volver a editar, el teléfono sigue
    // partido en país y número nacional.
    expect((wrapper.find('input[type="tel"]').element as HTMLInputElement).value).toBe('3001234567')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('CO')
  })

  it('rejects an invalid phone/email format', async () => {
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('input[type="tel"]').setValue('123')
    await wrapper.find('input[type="email"]').setValue('no-arroba')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('no parece completo')
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

  it('tells the shell the booking is complete only once the server confirms it (DEC-111)', async () => {
    const completed = ref(false)
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mount(PublicCustomerDetailsPage, {
      props: BASE_PROPS,
      global: { provide: { [bookingChromeKey as symbol]: { completed } } },
    })
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    // Revisar el resumen todavía no es una reserva: el progreso sigue en el paso.
    expect(completed.value).toBe(false)

    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
      .trigger('click')
    await flushPromises()

    expect(completed.value).toBe(true)
    expect(wrapper.find('h1').text()).toBe('Todo listo')
  })

  it('does not complete the shell on a schedule conflict (DEC-111)', async () => {
    const completed = ref(false)
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [{ startsAt: '2026-09-20T15:00:00Z' }],
    })
    const wrapper = mount(PublicCustomerDetailsPage, {
      props: BASE_PROPS,
      global: { provide: { [bookingChromeKey as symbol]: { completed } } },
    })
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
      .trigger('click')
    await flushPromises()

    expect(completed.value).toBe(false)
    expect(wrapper.find('h1').text()).toBe('Tus datos')
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

  // El diálogo de conflicto se teletransporta al <body>: se consulta ahí.
  const conflictDialog = () => document.body.querySelector<HTMLElement>('.pb-conflict-dialog')
  const isConflictOpen = () => !!conflictDialog()?.closest('.base-dialog__overlay--open')

  async function reachConflict(wrapper: ReturnType<typeof mountPage>) {
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
      .trigger('click')
    await flushPromises()
  }

  it('opens a blocking dialog on a schedule conflict and lets the customer retry with an alternative (RN-CON-05, DEC-090)', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [{ startsAt: '2026-09-20T15:00:00Z' }, { startsAt: '2026-09-20T15:30:00Z' }],
    })
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await reachConflict(wrapper)

    expect(isConflictOpen()).toBe(true)
    expect(conflictDialog()!.getAttribute('role')).toBe('dialog')
    expect(conflictDialog()!.textContent).toContain('Ese horario se acaba de ocupar')
    // Sin salida lateral: ni botón de cerrar, y el aviso ya no vive en la hoja.
    expect(conflictDialog()!.querySelector('.base-dialog__close')).toBeNull()
    expect(wrapper.find('.base-alert').exists()).toBe(false)
    // Los datos del formulario nunca se pierden ante el conflicto (RN-CON-05).
    expect(wrapper.text()).toContain('+57 3001234567')

    const alternativeButtons =
      conflictDialog()!.querySelectorAll<HTMLButtonElement>('.pb-conflict__alt')
    expect(alternativeButtons.length).toBe(2)
    alternativeButtons[0]!.click()
    await flushPromises()

    // Elegir una alternativa cierra el diálogo y exige confirmar de nuevo.
    expect(isConflictOpen()).toBe(false)
    expect(confirmPublicAppointmentMock).toHaveBeenCalledTimes(1)

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
    wrapper.unmount()
  })

  it('sends the customer back to the calendar from the conflict dialog', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [],
    })
    const wrapper = mountPage()
    await reachConflict(wrapper)

    expect(conflictDialog()!.textContent).toContain('No quedan horarios libres')
    expect(conflictDialog()!.querySelector('.pb-conflict__alt')).toBeNull()
    conflictDialog()!.querySelector<HTMLButtonElement>('.base-dialog__footer button')!.click()

    expect(routerPushMock).toHaveBeenCalledWith({
      name: 'reserva-publica-horario',
      params: {
        slug: BASE_PROPS.slug,
        serviceId: BASE_PROPS.serviceId,
        barberId: BASE_PROPS.barberId,
      },
    })
    wrapper.unmount()
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

  it('has no accessibility violations on the schedule-conflict dialog', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'schedule-conflict',
      alternatives: [{ startsAt: '2026-09-20T15:00:00Z' }],
    })
    const wrapper = mountPage()
    await reachConflict(wrapper)

    const results = await axe(conflictDialog()!)
    expect(results).toHaveNoViolations()
    wrapper.unmount()
  })

  it('presents the confirmation as a ticket with service, price, date, barber and no access token', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: CONFIRMED_APPOINTMENT,
    })
    const wrapper = mountPage()
    await flushPromises()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
      .trigger('click')
    await flushPromises()

    const ticket = wrapper.get('.ticket__card')
    expect(ticket.text()).toContain('Corte clásico')
    expect(ticket.text()).toContain('30 min')
    expect(ticket.text()).toMatch(/20\.000/)
    expect(ticket.get('.ticket__leaf-day').text()).toBe('20')
    expect(ticket.get('.ticket__who-name').text()).toBe('Mateo Cruz')
    // Para uno mismo no hay fila «Turno para»; la credencial nunca se dibuja (DEC-089).
    expect(ticket.find('.ticket__for').exists()).toBe(false)
    expect(wrapper.html()).not.toContain(CONFIRMED_APPOINTMENT.accessToken)
    expect(wrapper.text()).toContain('ana@example.com')
  })

  it('names the attendee on the ticket only when the turn is for someone else', async () => {
    confirmPublicAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      appointment: { ...CONFIRMED_APPOINTMENT, attendeeName: 'Mateo Ruiz' },
    })
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.findAll('input[type="radio"]')[1]!.setValue()
    await wrapper.findAll('input[type="text"]')[1]!.setValue('Mateo Ruiz')
    await wrapper.find('form').trigger('submit')
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Confirmar turno'))!
      .trigger('click')
    await flushPromises()

    expect(wrapper.get('.ticket__for').text()).toContain('Mateo Ruiz')
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
  describe('servicio a domicilio (DEC-131, HU-103)', () => {
    async function mountOffering() {
      loadPublicProfileMock.mockResolvedValue(profileOutcome(true))
      const wrapper = mountPage()
      await flushPromises()
      return wrapper
    }

    async function chooseAddress(wrapper: Awaited<ReturnType<typeof mountOffering>>) {
      await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
      await wrapper.get('input[role="combobox"]').setValue('Carrera 7 71')
      await new Promise((r) => setTimeout(r, 320))
      await flushPromises()
      await wrapper.get('[role="option"]').trigger('click')
      await flushPromises()
    }

    it('does not offer the modality when the barbershop does not serve at home (CA-103-01)', async () => {
      const wrapper = mountPage()
      await flushPromises()
      expect(wrapper.find('input[name="service-mode"]').exists()).toBe(false)
      expect(wrapper.text()).not.toContain('A domicilio')
    })

    it('offers local/home when enabled and defaults to the shop (CA-103-01)', async () => {
      const wrapper = await mountOffering()
      const local = wrapper.get('input[name="service-mode"][value="local"]')
      expect((local.element as HTMLInputElement).checked).toBe(true)
      expect(wrapper.find('input[role="combobox"]').exists()).toBe(false)
    })

    it('requires a marked location before reviewing: typed text alone is not one (CA-103-02, CA-138-03)', async () => {
      const wrapper = await mountOffering()
      await fillValidForm(wrapper)
      await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
      await wrapper.get('input[role="combobox"]').setValue('Carrera 7 71')
      await wrapper.find('form').trigger('submit')

      expect(wrapper.text()).toContain('Marca tu ubicación en el mapa')
      expect(wrapper.text()).not.toContain('Confirmar turno')
    })

    it('lets the customer mark the location by moving the map, with no address typed (CA-138-01)', async () => {
      confirmPublicAppointmentMock.mockResolvedValueOnce({
        kind: 'success',
        appointment: CONFIRMED_APPOINTMENT,
      })
      const wrapper = await mountOffering()
      await fillValidForm(wrapper)
      await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
      await flushPromises()
      await wrapper.get('button[aria-label="Mover el mapa al norte"]').trigger('click')
      await flushPromises()
      expect(wrapper.text()).toContain('Ubicación marcada:')

      await wrapper.find('form').trigger('submit')
      expect(wrapper.text()).toContain('Te atendemos en')
      await wrapper
        .findAll('button')
        .find((b) => b.text().includes('Confirmar turno'))!
        .trigger('click')
      await flushPromises()

      const input = confirmPublicAppointmentMock.mock.calls[0]![3]
      expect(input.homeService).toMatchObject({ latitude: 4.712, longitude: -74.0721 })
      expect(input.homeService.address).toContain('Punto marcado cerca de')
    })

    it('opens the map on the barbershop when it published a location', async () => {
      loadPublicProfileMock.mockResolvedValue(
        profileOutcome(true, { address: 'Calle 1', latitude: 6.2442, longitude: -75.5812 }),
      )
      confirmPublicAppointmentMock.mockResolvedValueOnce({
        kind: 'success',
        appointment: CONFIRMED_APPOINTMENT,
      })
      const wrapper = mountPage()
      await flushPromises()
      await fillValidForm(wrapper)
      await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
      await flushPromises()
      await wrapper.get('button[aria-label="Mover el mapa al este"]').trigger('click')
      await flushPromises()
      await wrapper.find('form').trigger('submit')
      await wrapper
        .findAll('button')
        .find((b) => b.text().includes('Confirmar turno'))!
        .trigger('click')
      await flushPromises()

      expect(confirmPublicAppointmentMock.mock.calls[0]![3].homeService).toMatchObject({
        latitude: 6.2442,
        longitude: -75.5802,
      })
    })

    it('detects the device location with the button and sends it (CA-138-02)', async () => {
      confirmPublicAppointmentMock.mockResolvedValueOnce({
        kind: 'success',
        appointment: CONFIRMED_APPOINTMENT,
      })
      const getCurrentPosition = vi.fn((ok: (p: unknown) => void) =>
        ok({ coords: { latitude: 4.65, longitude: -74.06, accuracy: 20 } }),
      )
      Object.defineProperty(navigator, 'geolocation', {
        value: { getCurrentPosition },
        configurable: true,
      })
      try {
        const wrapper = await mountOffering()
        await fillValidForm(wrapper)
        await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
        await flushPromises()
        await wrapper
          .findAll('button')
          .find((b) => b.text().includes('Usar mi ubicación'))!
          .trigger('click')
        await flushPromises()
        expect(getCurrentPosition).toHaveBeenCalledTimes(1)

        await wrapper.find('form').trigger('submit')
        await wrapper
          .findAll('button')
          .find((b) => b.text().includes('Confirmar turno'))!
          .trigger('click')
        await flushPromises()

        expect(confirmPublicAppointmentMock.mock.calls[0]![3].homeService).toMatchObject({
          latitude: 4.65,
          longitude: -74.06,
        })
      } finally {
        Reflect.deleteProperty(navigator, 'geolocation')
      }
    })

    it('never asks the browser for the location until the button is pressed (CA-138-05)', async () => {
      const getCurrentPosition = vi.fn()
      Object.defineProperty(navigator, 'geolocation', {
        value: { getCurrentPosition },
        configurable: true,
      })
      try {
        const wrapper = await mountOffering()
        await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
        await flushPromises()
        expect(getCurrentPosition).not.toHaveBeenCalled()
      } finally {
        Reflect.deleteProperty(navigator, 'geolocation')
      }
    })

    it('sends the chosen address with coordinates on confirm and shows it in the summary (CA-103-03)', async () => {
      confirmPublicAppointmentMock.mockResolvedValueOnce({
        kind: 'success',
        appointment: CONFIRMED_APPOINTMENT,
      })
      const wrapper = await mountOffering()
      await chooseAddress(wrapper)
      await fillValidForm(wrapper)
      await wrapper.find('form').trigger('submit')

      expect(wrapper.text()).toContain('Te atendemos en')
      expect(wrapper.text()).toContain('Carrera 7 71, Bogotá (1)')

      await wrapper
        .findAll('button')
        .find((b) => b.text().includes('Confirmar turno'))!
        .trigger('click')
      await flushPromises()

      const input = confirmPublicAppointmentMock.mock.calls[0]![3]
      expect(input.homeService).toEqual({
        address: 'Carrera 7 71, Bogotá (1)',
        latitude: 4.6486,
        longitude: -74.0628,
      })
    })

    it('sends no address when the customer stays with the shop modality', async () => {
      confirmPublicAppointmentMock.mockResolvedValueOnce({
        kind: 'success',
        appointment: CONFIRMED_APPOINTMENT,
      })
      const wrapper = await mountOffering()
      await chooseAddress(wrapper)
      await wrapper.find('input[name="service-mode"][value="local"]').setValue(true)
      await fillValidForm(wrapper)
      await wrapper.find('form').trigger('submit')
      expect(wrapper.text()).not.toContain('Te atendemos en')

      await wrapper
        .findAll('button')
        .find((b) => b.text().includes('Confirmar turno'))!
        .trigger('click')
      await flushPromises()

      expect(confirmPublicAppointmentMock.mock.calls[0]![3].homeService).toBeNull()
    })

    it('has no accessibility violations with the address step open', async () => {
      const wrapper = await mountOffering()
      await wrapper.find('input[name="service-mode"][value="home"]').setValue(true)
      const results = await axe(wrapper.element)
      expect(results).toHaveNoViolations()
    })
  })
})
