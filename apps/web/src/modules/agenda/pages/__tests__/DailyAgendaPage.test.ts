/**
 * Pruebas de DailyAgendaPage (HU-062): carga (vacío, error), selector
 * obligatorio de un barbero (DEC-074, sin vista consolidada), estados de la
 * agenda (carga, vacío, error, listo), aislamiento entre selecciones
 * rápidas de barbero, navegación a "Nuevo turno" y accesibilidad.
 * appointmentsApi se sustituye por un doble de prueba; el recorrido real
 * contra el API vive en el E2E de HU-062.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const fetchDailyAgendaMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/appointmentsApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
  fetchDailyAgenda: fetchDailyAgendaMock,
}))

const { default: DailyAgendaPage } = await import('../DailyAgendaPage.vue')

const twoBarbers = [
  { id: 'b-1', fullName: 'Carlos Ramírez' },
  { id: 'b-2', fullName: 'Ana Gómez' },
]

const oneEntry = [
  {
    id: 'a-1',
    attendeeName: 'Juan Pérez',
    startsAt: '2026-08-28T19:30:00Z',
    endsAt: '2026-08-28T20:00:00Z',
    status: 'confirmed',
    origin: 'manual',
    serviceName: 'Corte clásico',
    durationMinutes: 30,
    priceAmount: '20000.00',
    currency: 'COP',
  },
]

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/panel', name: 'panel', component: { template: '<div>panel</div>' } },
      {
        path: '/panel/turnos/nuevo',
        name: 'agenda-nuevo-turno',
        component: { template: '<div>nuevo turno</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = buildRouter()
  await router.push({ name: 'panel' })
  await router.isReady()
  const wrapper = mount(DailyAgendaPage, { global: { plugins: [router] } })
  return { wrapper, router }
}

async function mountReady(barbers = twoBarbers) {
  fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: barbers })
  fetchBarbershopTimezoneMock.mockResolvedValueOnce({
    kind: 'success',
    timezone: 'America/Bogota',
  })
  const { wrapper, router } = await mountPage()
  await flushPromises()
  return { wrapper, router }
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

describe('DailyAgendaPage', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchBarbershopTimezoneMock.mockReset()
    fetchDailyAgendaMock.mockReset()
  })

  it('shows a non-blank loading state, then the barber picker (DEC-074: selector obligatorio)', async () => {
    let resolveBarbers: (value: unknown) => void = () => {}
    fetchBarberSummariesMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveBarbers = resolve
      }),
    )
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const { wrapper } = await mountPage()
    expect(wrapper.text()).toContain('Cargando barberos')

    resolveBarbers({ kind: 'success', items: twoBarbers })
    await flushPromises()
    expect(wrapper.find('#daily-agenda-barber-select').exists()).toBe(true)
  })

  it('shows an empty state with no barber picker when there are no barbers', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const { wrapper } = await mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
    expect(wrapper.find('#daily-agenda-barber-select').exists()).toBe(false)
    expect(fetchDailyAgendaMock).not.toHaveBeenCalled()
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'network-error' })
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const { wrapper } = await mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar esta sección')

    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoBarbers })
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({
      kind: 'success',
      timezone: 'America/Bogota',
    })
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    await wrapper.get('button[type="button"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('#daily-agenda-barber-select').exists()).toBe(true)
  })

  it('auto-selects the first barber and loads its agenda on open (CA-062-01)', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
    const { wrapper } = await mountReady()

    expect(fetchDailyAgendaMock).toHaveBeenCalledWith('b-1')
    expect(wrapper.text()).toContain('Juan Pérez')
    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.text()).toContain('Confirmado')
  })

  it('shows an explicit empty message when the selected barber has no turns today', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const { wrapper } = await mountReady()

    expect(wrapper.text()).toContain('No hay turnos para Carlos Ramírez hoy')
  })

  it('switching barbers loads that barber agenda independently (discards a stale in-flight response)', async () => {
    let resolveFirst: (value: unknown) => void = () => {}
    fetchDailyAgendaMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve
      }),
    )
    const { wrapper } = await mountReady()

    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
    const select = wrapper.get<HTMLSelectElement>('#daily-agenda-barber-select')
    await select.setValue('b-2')
    await flushPromises()

    // La respuesta lenta del primer barbero (b-1) llega tarde: no debe
    // pisar la agenda ya cargada de b-2.
    resolveFirst({ kind: 'success', items: [] })
    await flushPromises()

    expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(1, 'b-1')
    expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(2, 'b-2')
    expect(wrapper.text()).toContain('Juan Pérez')
  })

  it('shows a recoverable error with Reintentar when the agenda fails to load', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'network-error' })
    const { wrapper } = await mountReady()

    expect(wrapper.text()).toContain('No pudimos cargar la agenda de este barbero')

    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
    const retryButtons = wrapper.findAll('button').filter((b) => b.text() === 'Reintentar')
    await retryButtons[0]!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Juan Pérez')
  })

  it('shows a not-found message when the selected barber no longer exists', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'not-found' })
    const { wrapper } = await mountReady()

    expect(wrapper.text()).toContain('Este barbero ya no está disponible')
  })

  it('"Nuevo turno" navigates to the real HU-061 form', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const { wrapper, router } = await mountReady()

    const newButton = wrapper.findAll('button').find((b) => b.text() === 'Nuevo turno')!
    await newButton.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('agenda-nuevo-turno')
  })

  it('has no detectable axe violations in the ready state with entries', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
    const { wrapper } = await mountReady()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
