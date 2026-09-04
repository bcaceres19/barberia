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
import { shiftCivilDate } from '@/shared/time/civilDate'

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

// 23:30 (2026-08-28, America/Bogota) → 00:30 del día siguiente: cruza
// medianoche, para probar la marca "Cambio de día" (issue #189, evento 11).
const nightEntry = [
  {
    id: 'a-2',
    attendeeName: 'Laura Méndez',
    startsAt: '2026-08-29T04:30:00Z',
    endsAt: '2026-08-29T05:30:00Z',
    status: 'confirmed',
    origin: 'manual',
    serviceName: 'Corte clásico',
    durationMinutes: 60,
    priceAmount: '25000.00',
    currency: 'COP',
  },
]

const terminalEntry = [
  {
    id: 'a-3',
    attendeeName: 'Carlos Ruiz',
    startsAt: '2026-08-28T18:00:00Z',
    endsAt: '2026-08-28T18:30:00Z',
    status: 'completed',
    origin: 'manual',
    serviceName: 'Barba',
    durationMinutes: 30,
    priceAmount: '15000.00',
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
      {
        path: '/panel/turnos/:appointmentId',
        name: 'agenda-detalle-turno',
        component: { template: '<div>detalle</div>' },
      },
      {
        path: '/barberos',
        name: 'staff-barberos',
        component: { template: '<div>barberos</div>' },
      },
    ],
  })
}

// El listbox propio de BarberSelect no tiene un <select> nativo que
// setValue() pueda usar (issue #189): abre el disparador y hace click en la
// opción con el nombre buscado, mismo resultado observable que antes.
async function selectBarber(
  wrapper: Awaited<ReturnType<typeof mountPage>>['wrapper'],
  fullName: string,
) {
  await wrapper.get('#daily-agenda-barber-select').trigger('click')
  const option = wrapper.findAll('[role="option"]').find((o) => o.text().includes(fullName))
  if (!option) throw new Error(`No option found for barber "${fullName}"`)
  await option.trigger('click')
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
const civilDateMatcher = expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/)

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

  it('auto-selects the first barber and loads its agenda on open, with an explicit civil date (CA-062-01, CA-063)', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
    const { wrapper, router } = await mountReady()

    expect(fetchDailyAgendaMock).toHaveBeenCalledWith('b-1', civilDateMatcher)
    expect(wrapper.text()).toContain('Juan Pérez')
    expect(wrapper.text()).toContain('Corte clásico')
    expect(wrapper.text()).toContain('Confirmado')
    // La URL queda normalizada con barbero y fecha explícitos.
    expect(router.currentRoute.value.query.barberId).toBe('b-1')
    expect(router.currentRoute.value.query.date).toMatch(/^\d{4}-\d{2}-\d{2}$/)
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
    await selectBarber(wrapper, 'Ana Gómez')
    await flushPromises()

    // La respuesta lenta del primer barbero (b-1) llega tarde: no debe
    // pisar la agenda ya cargada de b-2.
    resolveFirst({ kind: 'success', items: [] })
    await flushPromises()

    expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(1, 'b-1', civilDateMatcher)
    expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(2, 'b-2', civilDateMatcher)
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

  it('has no detectable axe violations in the initial loading state (PageState)', async () => {
    fetchBarberSummariesMock.mockReturnValueOnce(new Promise(() => {}))
    fetchBarbershopTimezoneMock.mockReturnValueOnce(new Promise(() => {}))
    const { wrapper } = await mountPage()
    await flushPromises()

    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })

  it('has no detectable axe violations with a terminal entry present (outline badge on parchment)', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: terminalEntry })
    const { wrapper } = await mountReady()

    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })

  it('has no detectable axe violations in the empty-day-today composition', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const { wrapper } = await mountReady()

    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })

  it('gives every status badge an outline treatment (issue #189: no light fills on parchment/ink)', async () => {
    fetchDailyAgendaMock.mockResolvedValueOnce({
      kind: 'success',
      items: [...oneEntry, ...terminalEntry],
    })
    const { wrapper } = await mountReady()

    const badges = wrapper.findAll('.base-badge')
    expect(badges.length).toBe(2)
    for (const badge of badges) {
      expect(badge.classes()).toContain('base-badge--outline')
    }
  })

  describe('navegación por fecha (HU-063)', () => {
    function anteriorButton(wrapper: Awaited<ReturnType<typeof mountReady>>['wrapper']) {
      return wrapper.findAll('button').find((b) => b.text() === 'Anterior')!
    }

    function siguienteButton(wrapper: Awaited<ReturnType<typeof mountReady>>['wrapper']) {
      return wrapper.findAll('button').find((b) => b.text() === 'Siguiente')!
    }

    it('"Anterior" navigates one civil day back, keeps the same barber and updates the URL (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper, router } = await mountReady()
      const firstDate = router.currentRoute.value.query.date as string

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      await anteriorButton(wrapper).trigger('click')
      await flushPromises()

      const expectedPrevious = shiftCivilDate(firstDate, -1)

      expect(router.currentRoute.value.query.date).toBe(expectedPrevious)
      expect(router.currentRoute.value.query.barberId).toBe('b-1')
      expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(2, 'b-1', expectedPrevious)
    })

    it('"Siguiente" navigates one civil day forward, keeps the same barber (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper, router } = await mountReady()
      const firstDate = router.currentRoute.value.query.date as string

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      await siguienteButton(wrapper).trigger('click')
      await flushPromises()

      const expectedNext = shiftCivilDate(firstDate, 1)

      expect(router.currentRoute.value.query.date).toBe(expectedNext)
      expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(2, 'b-1', expectedNext)
    })

    it('picking a date in the date selector navigates directly to it (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper, router } = await mountReady()

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      const dateInput = wrapper.get<HTMLInputElement>('input[type="date"]')
      await dateInput.setValue('2026-01-15')
      await flushPromises()

      expect(router.currentRoute.value.query.date).toBe('2026-01-15')
      expect(fetchDailyAgendaMock).toHaveBeenNthCalledWith(2, 'b-1', '2026-01-15')
    })

    it('opening the URL with an explicit date/barber (reload) loads exactly that selection, without an extra navigation', async () => {
      fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoBarbers })
      fetchBarbershopTimezoneMock.mockResolvedValueOnce({
        kind: 'success',
        timezone: 'America/Bogota',
      })
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })

      const router = buildRouter()
      await router.push({ name: 'panel', query: { date: '2026-01-15', barberId: 'b-2' } })
      await router.isReady()
      const wrapper = mount(DailyAgendaPage, { global: { plugins: [router] } })
      await flushPromises()

      expect(fetchDailyAgendaMock).toHaveBeenCalledTimes(1)
      expect(fetchDailyAgendaMock).toHaveBeenCalledWith('b-2', '2026-01-15')
      expect(router.currentRoute.value.query.date).toBe('2026-01-15')
      expect(router.currentRoute.value.query.barberId).toBe('b-2')
      expect(wrapper.text()).toContain('Ana Gómez')
    })

    it('the browser back button restores the previous date and its agenda (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper, router } = await mountReady()
      const firstDate = router.currentRoute.value.query.date as string

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      await anteriorButton(wrapper).trigger('click')
      await flushPromises()
      expect(router.currentRoute.value.query.date).not.toBe(firstDate)

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      router.back()
      await flushPromises()

      expect(router.currentRoute.value.query.date).toBe(firstDate)
      expect(wrapper.text()).toContain('Juan Pérez')
    })

    it('discards an out-of-order response from a previous date after navigating again (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper, router } = await mountReady()
      const firstDate = router.currentRoute.value.query.date as string

      let resolveFirstNav: (value: unknown) => void = () => {}
      fetchDailyAgendaMock.mockReturnValueOnce(
        new Promise((resolve) => {
          resolveFirstNav = resolve
        }),
      )
      await anteriorButton(wrapper).trigger('click')
      await flushPromises()

      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      await anteriorButton(wrapper).trigger('click')
      await flushPromises()

      // La primera navegación ("Anterior" una vez) llega tarde, después de
      // que el usuario ya pidió un segundo "Anterior": no debe pisar la
      // fecha vigente.
      resolveFirstNav({ kind: 'success', items: oneEntry })
      await flushPromises()

      const expectedDate = shiftCivilDate(firstDate, -2)
      expect(router.currentRoute.value.query.date).toBe(expectedDate)
      expect(wrapper.text()).toContain('No hay turnos para Carlos Ramírez el')
    })

    it('keeps the previous agenda visible with an "Actualizando…" indicator while navigating to a new date (CA-063)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: oneEntry })
      const { wrapper } = await mountReady()
      expect(wrapper.text()).toContain('Juan Pérez')

      let resolveNext: (value: unknown) => void = () => {}
      fetchDailyAgendaMock.mockReturnValueOnce(
        new Promise((resolve) => {
          resolveNext = resolve
        }),
      )
      await siguienteButton(wrapper).trigger('click')
      await flushPromises()

      // La agenda anterior sigue visible (no una pantalla en blanco) con un
      // indicador local de actualización, mientras la nueva sigue en vuelo.
      expect(wrapper.text()).toContain('Actualizando…')
      expect(wrapper.text()).toContain('Juan Pérez')

      resolveNext({ kind: 'success', items: [] })
      await flushPromises()
      expect(wrapper.text()).not.toContain('Actualizando…')
    })

    it('disables date navigation when the barbershop timezone is unavailable (RN-DIS-07)', async () => {
      fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoBarbers })
      fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      const { wrapper } = await mountPage()
      await flushPromises()

      // Sin zona conocida, el modelo no puede calcular "hoy" en la
      // barbería (nunca la sustituye por la del dispositivo): la agenda
      // sigue siendo legible (comportamiento degradado de HU-062), pero la
      // navegación por fecha queda deshabilitada en vez de adivinar.
      expect(fetchDailyAgendaMock).toHaveBeenCalledWith('b-1', undefined)
      expect(anteriorButton(wrapper).attributes('disabled')).toBeDefined()
      expect(siguienteButton(wrapper).attributes('disabled')).toBeDefined()
      expect(wrapper.get<HTMLInputElement>('input[type="date"]').element.disabled).toBe(true)
    })
  })

  describe('línea temporal de escritorio (Fase 4a, adopción NAVA)', () => {
    async function mountWithFixedDate(items = oneEntry) {
      fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoBarbers })
      fetchBarbershopTimezoneMock.mockResolvedValueOnce({
        kind: 'success',
        timezone: 'America/Bogota',
      })
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items })

      const router = buildRouter()
      // Fecha fija en el pasado: nunca coincide con "hoy" real al correr la
      // prueba, así que el marcador "Ahora" tiene un resultado determinista
      // (no debe aparecer).
      await router.push({ name: 'panel', query: { date: '2026-08-28', barberId: 'b-1' } })
      await router.isReady()
      const wrapper = mount(DailyAgendaPage, { global: { plugins: [router] } })
      await flushPromises()
      return { wrapper, router }
    }

    it('renders one hidden, non-focusable timeline slip per entry, alongside the accessible list', async () => {
      const { wrapper } = await mountWithFixedDate()

      const timeline = wrapper.get('.daily-agenda-page__timeline')
      expect(timeline.attributes('aria-hidden')).toBe('true')

      const slips = wrapper.findAll('.daily-agenda-page__timeline-slip')
      expect(slips.length).toBe(1)
      expect(slips[0]!.attributes('tabindex')).toBe('-1')
      expect(slips[0]!.text()).toContain('Juan Pérez')

      // La lista accesible sigue existiendo con la misma información.
      expect(wrapper.get('.daily-agenda-page__list').text()).toContain('Juan Pérez')
    })

    it('does not render the timeline when there are no entries for the day', async () => {
      const { wrapper } = await mountWithFixedDate([])
      expect(wrapper.find('.daily-agenda-page__timeline').exists()).toBe(false)
    })

    it('does not show "Ahora" for a date other than today in the barbershop timezone', async () => {
      const { wrapper } = await mountWithFixedDate()
      expect(wrapper.find('.daily-agenda-page__timeline-now').exists()).toBe(false)
    })

    it('positions and sizes a slip from its own startsAt/endsAt, never an arbitrary spread', async () => {
      const { wrapper } = await mountWithFixedDate()

      // 2026-08-28T19:30-20:00Z = 14:30-15:00 America/Bogota (UTC-05).
      // bounds: [14:00, 18:00) redondeado a hora completa con el piso de
      // 4h (MIN_TIMELINE_SPAN_MINUTES) → left=(870-840)/240=12.5%,
      // width=(900-870)/240=12.5%.
      const slip = wrapper.get('.daily-agenda-page__timeline-slip')
      const style = slip.attributes('style') ?? ''
      expect(style).toContain('left: 12.5%')
      expect(style).toContain('width: 12.5%')
    })

    it('shows "Cambio de día" exactly at midnight when a shift ends the next civil day (evento 11)', async () => {
      const { wrapper } = await mountWithFixedDate(nightEntry)

      const mark = wrapper.get('.daily-agenda-page__timeline-mark--day-change')
      // bounds: start 23:00 (1380), turno termina 00:30 del día siguiente
      // (1470) sin recortar → span mínimo de 4h → end=1620.
      // left de medianoche (1440): (1440-1380)/(1620-1380)=25%.
      expect(mark.attributes('style')).toContain('left: 25%')
      expect(mark.text()).toBe('Cambio de día')

      // La ficha ocupa su duración real (60min) más allá de medianoche, no
      // recortada a la medianoche del día que se está viendo.
      const slip = wrapper.get('.daily-agenda-page__timeline-slip')
      const style = slip.attributes('style') ?? ''
      expect(style).toContain('width: 25%') // (1470-1410)/240
    })

    it('keeps the timeline (with "Ahora") visible for a valid empty day when viewing today (evento 09)', async () => {
      fetchDailyAgendaMock.mockResolvedValueOnce({ kind: 'success', items: [] })
      const { wrapper } = await mountReady()

      expect(wrapper.find('.daily-agenda-page__timeline').exists()).toBe(true)
      expect(wrapper.find('.daily-agenda-page__timeline-mark--now').exists()).toBe(true)
    })
  })
})
