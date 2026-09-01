/**
 * Pruebas de AppointmentDetailPage (HU-064): carga del detalle (listo, no
 * encontrado, error/reintento), historial paginado (listo, error/reintento,
 * "Cargar más"), regreso a la agenda conservando fecha/barbero, y
 * accesibilidad. appointmentsApi se sustituye por un doble de prueba; el
 * recorrido real contra el API vive en el E2E de HU-064.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'

const fetchAppointmentDetailMock = vi.hoisted(() => vi.fn())
const fetchAppointmentHistoryMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/appointmentsApi', () => ({
  fetchAppointmentDetail: fetchAppointmentDetailMock,
  fetchAppointmentHistory: fetchAppointmentHistoryMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
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
  const wrapper = mount(AppointmentDetailPage, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  fetchAppointmentDetailMock.mockReset()
  fetchAppointmentHistoryMock.mockReset()
  fetchBarbershopTimezoneMock.mockReset()
  fetchBarbershopTimezoneMock.mockResolvedValue({ kind: 'success', timezone: 'America/Bogota' })
  fetchAppointmentHistoryMock.mockResolvedValue({
    kind: 'success',
    items: [createdEvent],
    nextCursor: null,
  })
})

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
})
