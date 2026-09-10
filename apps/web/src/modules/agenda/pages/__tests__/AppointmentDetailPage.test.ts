/**
 * Pruebas de AppointmentDetailPage (HU-064, más reprogramación T2 de
 * HU-065): carga del detalle (listo, no encontrado, error/reintento),
 * historial paginado (listo, error/reintento, "Cargar más"), regreso a la
 * agenda conservando fecha/barbero, reprogramación (visibilidad por
 * estado, resumen/confirmación, doble toque bloqueado, conflictos de
 * agenda/versión/estado con conservación de datos, éxito que refresca
 * detalle+historial y el destino de "Volver"), y accesibilidad.
 * appointmentsApi se sustituye por un doble de prueba; el recorrido real
 * contra el API vive en el E2E de HU-064/HU-065.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'

const fetchAppointmentDetailMock = vi.hoisted(() => vi.fn())
const fetchAppointmentHistoryMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const rescheduleAppointmentMock = vi.hoisted(() => vi.fn())
const cancelAppointmentByBarberMock = vi.hoisted(() => vi.fn())
const completeAppointmentMock = vi.hoisted(() => vi.fn())
const markAppointmentNoShowMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/appointmentsApi', () => ({
  fetchAppointmentDetail: fetchAppointmentDetailMock,
  fetchAppointmentHistory: fetchAppointmentHistoryMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
  rescheduleAppointment: rescheduleAppointmentMock,
  cancelAppointmentByBarber: cancelAppointmentByBarberMock,
  completeAppointment: completeAppointmentMock,
  markAppointmentNoShow: markAppointmentNoShowMock,
}))

const { default: AppointmentDetailPage } = await import('../AppointmentDetailPage.vue')

const readyDetail = {
  id: 'a-1',
  barberId: 'b-1',
  barberFullName: 'Ana Gómez',
  attendeeName: 'Juan Pérez',
  customerFullName: 'Juan Pérez',
  customerPhone: '+573001234567',
  customerEmail: null,
  customerNote: 'Prefiere silla junto a la ventana',
  startsAt: '2026-08-28T19:30:00Z',
  endsAt: '2026-08-28T20:00:00Z',
  status: 'confirmed',
  origin: 'manual',
  serviceName: 'Corte clásico',
  durationMinutes: 30,
  priceAmount: '20000.00',
  currency: 'COP',
  versionToken: 'opaque-token',
  createdAt: '2026-08-20T10:00:00Z',
}

const createdEvent = {
  id: 'h-1',
  eventType: 'appointment_created',
  actorType: 'staff',
  actorLabel: 'Ana Gómez',
  reason: null,
  occurredAt: '2026-08-20T10:00:00Z',
  changes: [],
}

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/panel', name: 'panel', component: { template: '<div>panel</div>' } },
      {
        path: '/panel/turnos/:appointmentId',
        name: 'agenda-detalle-turno',
        component: AppointmentDetailPage,
      },
    ],
  })
}

async function mountPage(query: Record<string, string> = { date: '2026-08-28', barberId: 'b-1' }) {
  const router = buildRouter()
  await router.push({ name: 'agenda-detalle-turno', params: { appointmentId: 'a-1' }, query })
  await router.isReady()
  // stubs.teleport hace que Vue Test Utils renderice el contenido de
  // <Teleport to="body"> (el diálogo de reprogramación, HU-065) EN EL
  // LUGAR, dentro del árbol del wrapper: mismo criterio que
  // StaffPage.test.ts.
  const wrapper = mount(AppointmentDetailPage, {
    global: { plugins: [router], stubs: { teleport: true } },
  })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  fetchAppointmentDetailMock.mockReset()
  fetchAppointmentHistoryMock.mockReset()
  fetchBarbershopTimezoneMock.mockReset()
  rescheduleAppointmentMock.mockReset()
  cancelAppointmentByBarberMock.mockReset()
  completeAppointmentMock.mockReset()
  markAppointmentNoShowMock.mockReset()
  fetchBarbershopTimezoneMock.mockResolvedValue({ kind: 'success', timezone: 'America/Bogota' })
  fetchAppointmentHistoryMock.mockResolvedValue({
    kind: 'success',
    items: [createdEvent],
    nextCursor: null,
  })
})

// --- Ayudantes para el diálogo de reprogramación (HU-065) ----------------
// Mismo criterio que StaffPage.test.ts: BaseDialog usa Teleport EN EL
// LUGAR dentro del wrapper (no al document.body real), así que basta con
// buscar dentro de wrapper.element.
function openDialogElement(wrapper: VueWrapper): HTMLElement {
  const el = wrapper.element.querySelector('.base-dialog--open')
  if (!el) throw new Error('no open dialog found')
  return el as HTMLElement
}

function openDialogForm(wrapper: VueWrapper): HTMLFormElement {
  return openDialogElement(wrapper).querySelector('form') as HTMLFormElement
}

function rescheduleDateInput(wrapper: VueWrapper): HTMLInputElement {
  return openDialogElement(wrapper).querySelector(
    'input[name="rescheduleDate"]',
  ) as HTMLInputElement
}

function rescheduleTimeInput(wrapper: VueWrapper): HTMLInputElement {
  return openDialogElement(wrapper).querySelector(
    'input[name="rescheduleTime"]',
  ) as HTMLInputElement
}

function setInputValue(input: HTMLInputElement, value: string) {
  input.value = value
  input.dispatchEvent(new Event('input'))
}

function submitRescheduleDialog(wrapper: VueWrapper) {
  openDialogForm(wrapper).dispatchEvent(new Event('submit', { cancelable: true }))
}

function findButtonByText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((b) => b.text() === text)
  if (!button) throw new Error(`button ${JSON.stringify(text)} not found`)
  return button
}

async function openRescheduleDialog(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Reprogramar turno').trigger('click')
  await flushPromises()
}

// --- Ayudantes para el diálogo de cancelación (HU-066) -------------------

async function openCancelDialog(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Cancelar turno').trigger('click')
  await flushPromises()
}

async function confirmCancel(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Sí, cancelar turno').trigger('click')
  await flushPromises()
}

// --- Ayudantes para el diálogo de cierre manual (HU-067, T4 manual/T7) ---

async function openCompleteDialog(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Marcar como atendido').trigger('click')
  await flushPromises()
}

async function openNoShowDialog(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Marcar que no asistió').trigger('click')
  await flushPromises()
}

async function confirmComplete(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Sí, marcar como atendido').trigger('click')
  await flushPromises()
}

async function confirmNoShow(wrapper: VueWrapper) {
  await findButtonByText(wrapper, 'Sí, marcar que no asistió').trigger('click')
  await flushPromises()
}

describe('AppointmentDetailPage', () => {
  it('shows the appointment detail once loaded', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('Juan Pérez')
    expect(wrapper.text()).toContain('Ana Gómez')
    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.text()).toContain('20000.00')
    expect(wrapper.text()).toContain('+573001234567')
    expect(wrapper.text()).toContain('Prefiere silla junto a la ventana')
    expect(fetchAppointmentDetailMock).toHaveBeenCalledWith('a-1')
  })

  it('does not render a contact section when phone and email are both absent', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, customerPhone: null, customerEmail: null },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).not.toContain('Contacto')
  })

  it('shows a not-found alert when the appointment does not exist or belongs to another shop', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'not-found' })
    const { wrapper } = await mountPage()

    expect(wrapper.find('[role="alert"]').text()).toContain('ya no está disponible')
  })

  it('shows a load error with retry, and retrying recovers the detail', async () => {
    fetchAppointmentDetailMock
      .mockResolvedValueOnce({ kind: 'network-error' })
      .mockResolvedValueOnce({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    const retry = wrapper.findAll('button').find((b) => b.text().includes('Reintentar'))
    expect(retry).toBeTruthy()
    await retry!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Juan Pérez')
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('loads and shows the history once the detail is ready', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    expect(fetchAppointmentHistoryMock).toHaveBeenCalledWith('a-1', undefined)
    expect(wrapper.text()).toContain('Turno creado')
    expect(wrapper.text()).toContain('Ana Gómez')
  })

  it('shows the reason and field changes of a history entry when present', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    fetchAppointmentHistoryMock.mockResolvedValue({
      kind: 'success',
      items: [
        {
          id: 'h-2',
          eventType: 'appointment_status_corrected',
          actorType: 'system',
          actorLabel: 'Sistema',
          reason: 'corrección administrativa de estado',
          occurredAt: '2026-08-21T09:15:00Z',
          changes: [
            { fieldName: 'status', previousValue: 'confirmed', newValue: 'cancelled_by_barber' },
          ],
        },
      ],
      nextCursor: null,
    })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('Estado corregido')
    expect(wrapper.text()).toContain('corrección administrativa de estado')
    expect(wrapper.text()).toContain('Estado: confirmed → cancelled_by_barber')
  })

  it('shows a "Cargar más" button when the history has a next page, and loads it', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    fetchAppointmentHistoryMock
      .mockResolvedValueOnce({ kind: 'success', items: [createdEvent], nextCursor: 'cursor-1' })
      .mockResolvedValueOnce({
        kind: 'success',
        items: [{ ...createdEvent, id: 'h-2', eventType: 'appointment_completed' }],
        nextCursor: null,
      })
    const { wrapper } = await mountPage()

    const loadMore = wrapper.findAll('button').find((b) => b.text().includes('Cargar más'))
    expect(loadMore).toBeTruthy()
    await loadMore!.trigger('click')
    await flushPromises()

    expect(fetchAppointmentHistoryMock).toHaveBeenNthCalledWith(2, 'a-1', 'cursor-1')
    expect(wrapper.text()).toContain('Turno completado')
    // Ambas páginas conviven: la primera no se reemplaza.
    expect(wrapper.text()).toContain('Turno creado')
  })

  it('shows a history load error with retry', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    fetchAppointmentHistoryMock
      .mockResolvedValueOnce({ kind: 'unexpected-error' })
      .mockResolvedValueOnce({ kind: 'success', items: [createdEvent], nextCursor: null })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('No pudimos cargar el historial')
    const retry = wrapper.findAll('button').find((b) => b.text().includes('Reintentar'))
    await retry!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Turno creado')
  })

  it('the back link preserves the originating date and barberId (HU-063)', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage({ date: '2026-08-30', barberId: 'b-2' })

    const back = wrapper.findAll('a').find((a) => a.text().includes('Volver'))
    expect(back).toBeTruthy()
    expect(back!.attributes('href')).toBe('/panel?date=2026-08-30&barberId=b-2')
  })

  it('has no obvious accessibility violations once ready', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    const results = await axe(wrapper.element)
    expect(results.violations).toEqual([])
  })

  // --- Reprogramación (HU-065, T2) ---------------------------------------

  it('shows the "Reprogramar turno" action only for a confirmed appointment', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Reprogramar turno')).toBe(true)
  })

  it('does not show "Reprogramar turno" for a non-confirmed appointment', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'completed' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Reprogramar turno')).toBe(false)
  })

  it('opens the reschedule dialog prefilled with the current date and time, and shows a preview', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    // readyDetail.startsAt = 2026-08-28T19:30:00Z, America/Bogota (UTC-5) => 14:30 local.
    expect(rescheduleDateInput(wrapper).value).toBe('2026-08-28')
    expect(rescheduleTimeInput(wrapper).value).toBe('14:30')
    expect(openDialogElement(wrapper).textContent).toContain('Horario actual')

    setInputValue(rescheduleTimeInput(wrapper), '15:00')
    await flushPromises()
    expect(openDialogElement(wrapper).textContent).toContain('15:00 – 15:30')
  })

  it('on success, refetches the detail and history, closes the dialog, and updates the back link to the new date', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage({ date: '2026-08-28', barberId: 'b-1' })
    await openRescheduleDialog(wrapper)

    setInputValue(rescheduleDateInput(wrapper), '2026-08-30')
    setInputValue(rescheduleTimeInput(wrapper), '10:00')
    await flushPromises()

    rescheduleAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      startsAt: '2026-08-30T15:00:00Z',
    })
    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, startsAt: '2026-08-30T15:00:00Z', endsAt: '2026-08-30T15:30:00Z' },
    })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    expect(rescheduleAppointmentMock).toHaveBeenCalledWith(
      'a-1',
      { startsAt: '2026-08-30T10:00:00' },
      'opaque-token',
      expect.any(String),
    )
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
    // 2026-08-30T15:00:00Z en America/Bogota (UTC-5) es todavía 2026-08-30.
    const back = wrapper.findAll('a').find((a) => a.text().includes('Volver'))
    expect(back!.attributes('href')).toBe('/panel?date=2026-08-30&barberId=b-1')
  })

  it('blocks a second submit while the first is still in flight (no double POST)', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    let resolveReschedule: (value: unknown) => void = () => {}
    rescheduleAppointmentMock.mockReturnValueOnce(
      new Promise((resolve) => (resolveReschedule = resolve)),
    )
    submitRescheduleDialog(wrapper)
    submitRescheduleDialog(wrapper)
    await flushPromises()

    expect(rescheduleAppointmentMock).toHaveBeenCalledTimes(1)
    resolveReschedule({ kind: 'success', startsAt: readyDetail.startsAt })
    await flushPromises()
  })

  it('sends the same idempotency key across a submit and a network-error retry of the same logical attempt', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    rescheduleAppointmentMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    rescheduleAppointmentMock.mockResolvedValueOnce({
      kind: 'success',
      startsAt: readyDetail.startsAt,
    })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    const firstKey = rescheduleAppointmentMock.mock.calls[0]![3]
    const secondKey = rescheduleAppointmentMock.mock.calls[1]![3]
    expect(secondKey).toBe(firstKey)
  })

  it('on an agenda conflict, shows the server message and keeps the chosen date/time', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    setInputValue(rescheduleDateInput(wrapper), '2026-08-29')
    setInputValue(rescheduleTimeInput(wrapper), '11:00')
    await flushPromises()

    rescheduleAppointmentMock.mockResolvedValueOnce({
      kind: 'conflict',
      detail: 'El barbero ya tiene un turno en ese horario.',
    })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('El barbero ya tiene un turno en ese horario.')
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeTruthy()
    expect(rescheduleDateInput(wrapper).value).toBe('2026-08-29')
    expect(rescheduleTimeInput(wrapper).value).toBe('11:00')
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(1)
  })

  it('on a version conflict, offers a reload that refetches the real detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    rescheduleAppointmentMock.mockResolvedValueOnce({ kind: 'version-conflict' })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('cambió mientras lo editabas')
    const reload = wrapper.findAll('button').find((b) => b.text() === 'Recargar')
    expect(reload).toBeTruthy()

    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, versionToken: 'fresh-token' },
    })
    await reload!.trigger('click')
    await flushPromises()

    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('on an invalid-state conflict, offers a reload with the same recoverable message', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    rescheduleAppointmentMock.mockResolvedValueOnce({ kind: 'invalid-state' })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('ya no se puede reprogramar')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Recargar')).toBe(true)
  })

  // El atrapado/restauración de foco es responsabilidad genérica de
  // BaseDialog (shared/ui/__tests__/BaseDialog.test.ts): esta página solo
  // usa su API estándar (v-model, @close) sin lógica de foco propia. No se
  // duplica aquí una prueba de `document.activeElement` porque jsdom no
  // calcula layout (offsetParent es siempre null), la misma razón por la
  // que BaseDialog.test.ts y StaffPage.test.ts tampoco lo verifican a este
  // nivel; el foco real se confirma en el navegador vía el E2E de HU-065.

  it('has no obvious accessibility violations with the reschedule dialog open', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    const results = await axe(wrapper.element)
    expect(results.violations).toEqual([])
  })

  // heading-order: BaseAlert.vue (HU-009) fija el título de una alerta como
  // `<h4>` porque es la etiqueta de un widget transitorio (`role="alert"`),
  // no un encabezado del esquema del documento; junto al `<h1>`/`<h2>` reales
  // de esta página, axe interpreta ese salto como un esquema de encabezados
  // roto. Mismo criterio documentado en LoginPage.test.ts/SettingsPage.test.ts:
  // no es un defecto de esta pantalla ni de BaseAlert.
  it('has no obvious accessibility violations with the reschedule conflict alert visible', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openRescheduleDialog(wrapper)

    rescheduleAppointmentMock.mockResolvedValueOnce({
      kind: 'conflict',
      detail: 'El barbero ya tiene un turno en ese horario.',
    })
    submitRescheduleDialog(wrapper)
    await flushPromises()

    const results = await axe(wrapper.element, { rules: { 'heading-order': { enabled: false } } })
    expect(results.violations).toEqual([])
  })

  // --- Cancelación (HU-066, T6) -------------------------------------------

  it('shows the "Cancelar turno" action only for a confirmed appointment', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Cancelar turno')).toBe(true)
  })

  it('does not show "Cancelar turno" for a non-confirmed appointment', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'completed' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Cancelar turno')).toBe(false)
  })

  it('opens the cancel dialog with the consequence explained before confirming', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    expect(openDialogElement(wrapper).textContent).toContain('Juan Pérez')
    expect(openDialogElement(wrapper).textContent).toContain('no se puede deshacer')
  })

  it('on success, calls the API with a fresh idempotency key, closes the dialog, and refetches the detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'success' })
    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, status: 'cancelled_by_barber' },
    })
    await confirmCancel(wrapper)

    expect(cancelAppointmentByBarberMock).toHaveBeenCalledWith(
      'a-1',
      'opaque-token',
      expect.any(String),
    )
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('blocks a second confirm while the first is still in flight (no double POST)', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    let resolveCancel: (value: unknown) => void = () => {}
    cancelAppointmentByBarberMock.mockReturnValueOnce(
      new Promise((resolve) => (resolveCancel = resolve)),
    )
    await confirmCancel(wrapper)
    await confirmCancel(wrapper)

    expect(cancelAppointmentByBarberMock).toHaveBeenCalledTimes(1)
    resolveCancel({ kind: 'success' })
    await flushPromises()
  })

  it('renews the idempotency key each time the dialog is reopened', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    await openCancelDialog(wrapper)
    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'network-error' })
    await confirmCancel(wrapper)

    await findButtonByText(wrapper, 'Volver').trigger('click')
    await flushPromises()

    await openCancelDialog(wrapper)
    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'success' })
    fetchAppointmentDetailMock.mockResolvedValueOnce({ kind: 'success', detail: readyDetail })
    await confirmCancel(wrapper)

    const firstKey = cancelAppointmentByBarberMock.mock.calls[0]![2]
    const secondKey = cancelAppointmentByBarberMock.mock.calls[1]![2]
    expect(secondKey).not.toBe(firstKey)
  })

  it('on a version conflict, offers a reload that refetches the real detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'version-conflict' })
    await confirmCancel(wrapper)

    expect(wrapper.text()).toContain('cambió mientras lo revisabas')
    const reload = wrapper.findAll('button').find((b) => b.text() === 'Recargar')
    expect(reload).toBeTruthy()

    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, versionToken: 'fresh-token' },
    })
    await reload!.trigger('click')
    await flushPromises()

    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('on an invalid-state conflict, offers a reload with the same recoverable message', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'invalid-state' })
    await confirmCancel(wrapper)

    expect(wrapper.text()).toContain('ya no se puede cancelar')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Recargar')).toBe(true)
  })

  it('on an idempotency conflict, keeps the dialog open with a recoverable message', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    await confirmCancel(wrapper)

    expect(wrapper.text()).toContain('No pudimos completar el intento anterior')
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeTruthy()
  })

  it('on a network error, keeps the dialog open without losing the intent', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    cancelAppointmentByBarberMock.mockResolvedValueOnce({ kind: 'network-error' })
    await confirmCancel(wrapper)

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeTruthy()
  })

  it('renders a status badge reflecting the terminal cancelled_by_barber state', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'cancelled_by_barber' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('Cancelado por el barbero')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Cancelar turno')).toBe(false)
  })

  it('has no obvious accessibility violations with the cancel dialog open', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCancelDialog(wrapper)

    const results = await axe(wrapper.element)
    expect(results.violations).toEqual([])
  })

  // --- Cierre manual (HU-067, T4 manual y T7) -----------------------------
  // readyDetail.startsAt (2026-08-28) ya pasó respecto al reloj real: sirve
  // tal cual como el caso "ya se puede cerrar" (CA-067-08), sin fixture
  // aparte. futureDetail cubre el caso "confirmed pero aún no empieza".

  const futureDetail = {
    ...readyDetail,
    startsAt: '2999-01-01T10:00:00Z',
    endsAt: '2999-01-01T10:30:00Z',
  }

  it('shows "Marcar como atendido" and "Marcar que no asistió" only once startsAt has passed', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar como atendido')).toBe(true)
    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar que no asistió')).toBe(true)
  })

  it('does not show the close actions for a confirmed appointment that has not started yet', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: futureDetail })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar como atendido')).toBe(false)
    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar que no asistió')).toBe(false)
  })

  it('does not show the close actions for a non-confirmed appointment', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'cancelled_by_barber' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar como atendido')).toBe(false)
    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar que no asistió')).toBe(false)
  })

  it('opens the complete dialog with the consequence explained before confirming', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    expect(openDialogElement(wrapper).textContent).toContain('Juan Pérez')
    expect(openDialogElement(wrapper).textContent).toContain('atendido')
    expect(openDialogElement(wrapper).textContent).toContain('no se puede deshacer')
  })

  it('opens the no-show dialog with the consequence explained before confirming', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openNoShowDialog(wrapper)

    expect(openDialogElement(wrapper).textContent).toContain('Juan Pérez')
    expect(openDialogElement(wrapper).textContent).toContain('no asistió')
  })

  it('on complete success, calls the API with a fresh idempotency key, closes the dialog, and refetches the detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    completeAppointmentMock.mockResolvedValueOnce({ kind: 'success' })
    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, status: 'completed' },
    })
    await confirmComplete(wrapper)

    expect(completeAppointmentMock).toHaveBeenCalledWith('a-1', 'opaque-token', expect.any(String))
    expect(markAppointmentNoShowMock).not.toHaveBeenCalled()
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('on no-show success, calls the API with a fresh idempotency key, closes the dialog, and refetches the detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openNoShowDialog(wrapper)

    markAppointmentNoShowMock.mockResolvedValueOnce({ kind: 'success' })
    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, status: 'no_show' },
    })
    await confirmNoShow(wrapper)

    expect(markAppointmentNoShowMock).toHaveBeenCalledWith(
      'a-1',
      'opaque-token',
      expect.any(String),
    )
    expect(completeAppointmentMock).not.toHaveBeenCalled()
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  it('blocks a second confirm while the first close attempt is still in flight (no double POST)', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    let resolveComplete: (value: unknown) => void = () => {}
    completeAppointmentMock.mockReturnValueOnce(
      new Promise((resolve) => (resolveComplete = resolve)),
    )
    await confirmComplete(wrapper)
    await confirmComplete(wrapper)

    expect(completeAppointmentMock).toHaveBeenCalledTimes(1)
    resolveComplete({ kind: 'success' })
    await flushPromises()
  })

  it('renews the idempotency key each time the close dialog is reopened', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()

    await openCompleteDialog(wrapper)
    completeAppointmentMock.mockResolvedValueOnce({ kind: 'network-error' })
    await confirmComplete(wrapper)

    await findButtonByText(wrapper, 'Volver').trigger('click')
    await flushPromises()

    await openCompleteDialog(wrapper)
    completeAppointmentMock.mockResolvedValueOnce({ kind: 'success' })
    fetchAppointmentDetailMock.mockResolvedValueOnce({ kind: 'success', detail: readyDetail })
    await confirmComplete(wrapper)

    const firstKey = completeAppointmentMock.mock.calls[0]![2]
    const secondKey = completeAppointmentMock.mock.calls[1]![2]
    expect(secondKey).not.toBe(firstKey)
  })

  it('on a version conflict, offers a reload that refetches the real detail', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    completeAppointmentMock.mockResolvedValueOnce({ kind: 'version-conflict' })
    await confirmComplete(wrapper)

    expect(wrapper.text()).toContain('cambió mientras lo revisabas')
    const reload = wrapper.findAll('button').find((b) => b.text() === 'Recargar')
    expect(reload).toBeTruthy()

    fetchAppointmentDetailMock.mockResolvedValueOnce({
      kind: 'success',
      detail: { ...readyDetail, versionToken: 'fresh-token' },
    })
    await reload!.trigger('click')
    await flushPromises()

    expect(wrapper.element.querySelector('.base-dialog--open')).toBeNull()
    expect(fetchAppointmentDetailMock).toHaveBeenCalledTimes(2)
  })

  // CA-067-05: el resultado contrario (o cualquier otro terminal) responde
  // el mismo invalid-state que cualquier operación sobre un turno que ya
  // dejó de estar confirmed.
  it('on an invalid-state conflict (opposite result or other terminal), offers a reload', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    completeAppointmentMock.mockResolvedValueOnce({ kind: 'invalid-state' })
    await confirmComplete(wrapper)

    expect(wrapper.text()).toContain('ya tiene un resultado registrado')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Recargar')).toBe(true)
  })

  // CA-067-03: antes de startsAt, el servidor responde 422 (validation
  // error) en vez de un conflicto -- distinto de cancelar/reprogramar, que
  // no tienen esta frontera temporal.
  it('on a validation error (turn not started yet), offers a reload with an explanatory message', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openNoShowDialog(wrapper)

    markAppointmentNoShowMock.mockResolvedValueOnce({
      kind: 'validation-error',
      detail: 'el turno todavía no comienza; espera hasta su hora de inicio',
    })
    await confirmNoShow(wrapper)

    expect(wrapper.text()).toContain('todavía no comienza')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Recargar')).toBe(true)
  })

  it('on an idempotency conflict, keeps the dialog open with a recoverable message', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    completeAppointmentMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    await confirmComplete(wrapper)

    expect(wrapper.text()).toContain('No pudimos completar el intento anterior')
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeTruthy()
  })

  it('on a network error, keeps the dialog open without losing the intent', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openNoShowDialog(wrapper)

    markAppointmentNoShowMock.mockResolvedValueOnce({ kind: 'network-error' })
    await confirmNoShow(wrapper)

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(wrapper.element.querySelector('.base-dialog--open')).toBeTruthy()
  })

  it('renders a status badge distinguishable by text for completed and no_show', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'completed' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('Completado')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar como atendido')).toBe(false)
  })

  it('renders a status badge distinguishable by text for no_show', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({
      kind: 'success',
      detail: { ...readyDetail, status: 'no_show' },
    })
    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('No se presentó')
    expect(wrapper.findAll('button').some((b) => b.text() === 'Marcar que no asistió')).toBe(false)
  })

  it('has no obvious accessibility violations with the complete dialog open', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openCompleteDialog(wrapper)

    const results = await axe(wrapper.element)
    expect(results.violations).toEqual([])
  })

  it('has no obvious accessibility violations with the no-show dialog open', async () => {
    fetchAppointmentDetailMock.mockResolvedValue({ kind: 'success', detail: readyDetail })
    const { wrapper } = await mountPage()
    await openNoShowDialog(wrapper)

    const results = await axe(wrapper.element)
    expect(results.violations).toEqual([])
  })
})
