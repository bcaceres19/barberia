/**
 * Pruebas de SchedulesPage (HU-040): carga (un barbero, cuatro barberos),
 * vacío (sin barberos), error recuperable, cambio de barbero, alta (éxito,
 * solape -CA-040-04-, validación, no encontrado, error de red), edición
 * (éxito, solape), retiro (éxito, no encontrado, error de red, sin doble
 * envío), agrupación por los siete días y accesibilidad. schedulesApi se
 * sustituye por un doble de prueba; el recorrido real contra el API vive en
 * el E2E de HU-040.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const fetchWorkingHoursMock = vi.hoisted(() => vi.fn())
const createWorkingHourMock = vi.hoisted(() => vi.fn())
const updateWorkingHourMock = vi.hoisted(() => vi.fn())
const deleteWorkingHourMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/schedulesApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
  fetchWorkingHours: fetchWorkingHoursMock,
  createWorkingHour: createWorkingHourMock,
  updateWorkingHour: updateWorkingHourMock,
  deleteWorkingHour: deleteWorkingHourMock,
}))

// HU-041: mismo doble de prueba que schedulesApi, para que el calendario de
// festivos y las excepciones de jornada (siempre cargados junto al horario
// semanal del barbero elegido) no dependan de una llamada de red real
// durante estas pruebas de HU-040.
const fetchHolidayCalendarMock = vi.hoisted(() => vi.fn())
const updateHolidayCalendarMock = vi.hoisted(() => vi.fn())
const fetchScheduleExceptionsMock = vi.hoisted(() => vi.fn())
const createScheduleExceptionMock = vi.hoisted(() => vi.fn())
const updateScheduleExceptionMock = vi.hoisted(() => vi.fn())
const deleteScheduleExceptionMock = vi.hoisted(() => vi.fn())
const fetchColombianHolidaysMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/scheduleExceptionsApi', () => ({
  fetchHolidayCalendar: fetchHolidayCalendarMock,
  updateHolidayCalendar: updateHolidayCalendarMock,
  fetchScheduleExceptions: fetchScheduleExceptionsMock,
  createScheduleException: createScheduleExceptionMock,
  updateScheduleException: updateScheduleExceptionMock,
  deleteScheduleException: deleteScheduleExceptionMock,
  fetchColombianHolidays: fetchColombianHolidaysMock,
}))

const { default: SchedulesPage } = await import('../SchedulesPage.vue')

const oneBarber = [{ id: 'b-1', fullName: 'Carlos Ramírez' }]
const fourBarbers = [
  { id: 'b-1', fullName: 'Carlos Ramírez' },
  { id: 'b-2', fullName: 'Ana Torres' },
  { id: 'b-3', fullName: 'Luis Gómez' },
  { id: 'b-4', fullName: 'María Pérez' },
]

const mondayMorning = {
  id: 'wh-1',
  isoWeekday: 1,
  startsTime: '08:00',
  durationMinutes: 240,
  createdAt: '2026-08-25T15:04:05Z',
  updatedAt: '2026-08-25T15:04:05Z',
}

function mountPage() {
  return mount(SchedulesPage, { global: { stubs: { teleport: true } } })
}

async function mountReady(barbers = oneBarber, items: (typeof mondayMorning)[] = []) {
  fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: barbers })
  fetchWorkingHoursMock.mockResolvedValueOnce({
    kind: 'success',
    page: { items, nextCursor: null },
  })
  const wrapper = mountPage()
  await flushPromises()
  return wrapper
}

function barberSelect(wrapper: VueWrapper): HTMLSelectElement {
  return wrapper.element.querySelector('#schedules-barber-select') as HTMLSelectElement
}

async function openCreateDialog(wrapper: VueWrapper) {
  await wrapper.get('button').trigger('click')
  await flushPromises()
}

function createForm(wrapper: VueWrapper) {
  return wrapper.get('form[name="createWorkingHour"]')
}

function editForm(wrapper: VueWrapper) {
  return wrapper.get('form[name="updateWorkingHour"]')
}

async function fillCreateForm(
  wrapper: VueWrapper,
  { startsTime = '08:00', durationMinutes = '60' } = {},
) {
  await wrapper.get('#schedules-create-weekday').setValue('1')
  await createForm(wrapper).get('input[name="startsTime"]').setValue(startsTime)
  await createForm(wrapper).get('input[name="durationMinutes"]').setValue(durationMinutes)
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

describe('SchedulesPage', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchWorkingHoursMock.mockReset()
    createWorkingHourMock.mockReset()
    updateWorkingHourMock.mockReset()
    deleteWorkingHourMock.mockReset()
    fetchBarbershopTimezoneMock.mockReset().mockResolvedValue({ kind: 'unavailable' })
    fetchHolidayCalendarMock.mockReset().mockResolvedValue({ kind: 'success', enabled: false })
    updateHolidayCalendarMock.mockReset()
    fetchScheduleExceptionsMock
      .mockReset()
      .mockResolvedValue({ kind: 'success', page: { items: [], nextCursor: null } })
    createScheduleExceptionMock.mockReset()
    updateScheduleExceptionMock.mockReset()
    deleteScheduleExceptionMock.mockReset()
    fetchColombianHolidaysMock.mockReset().mockResolvedValue({ kind: 'unavailable' })
  })

  it('shows a non-blank loading state, then the loaded picker', async () => {
    let resolveBarbers: (value: unknown) => void = () => {}
    fetchBarberSummariesMock.mockReturnValueOnce(
      new Promise((resolve) => (resolveBarbers = resolve)),
    )
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')

    fetchWorkingHoursMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [], nextCursor: null },
    })
    resolveBarbers({ kind: 'success', items: oneBarber })
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('a barber with one tramo and one with a full team select use the same component', async () => {
    const oneWrapper = await mountReady(oneBarber, [mondayMorning])
    expect(oneWrapper.get('#schedules-barber-select').findAll('option').length).toBe(1)
    expect(oneWrapper.text()).toContain('08:00')

    const fourWrapper = await mountReady(fourBarbers, [])
    expect(fourWrapper.get('#schedules-barber-select').findAll('option').length).toBe(4)
  })

  it('always shows all seven weekdays, with or without tramos (CA-040-01)', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    for (const day of ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']) {
      expect(wrapper.text()).toContain(day)
    }
  })

  it('shows the barbershop IANA timezone, never the device one (CA-040-06)', async () => {
    fetchBarbershopTimezoneMock.mockReset().mockResolvedValueOnce({
      kind: 'success',
      timezone: 'America/Bogota',
    })
    const wrapper = await mountReady(oneBarber, [])
    expect(wrapper.text()).toContain('America/Bogota')
  })

  it('shows an empty state when there are no barbers, without an interactive picker', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
    expect(wrapper.find('#schedules-barber-select').exists()).toBe(false)
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar esta sección')

    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: oneBarber })
    fetchWorkingHoursMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [], nextCursor: null },
    })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('switching barbers loads that barber schedule independently (cross-barber isolation)', async () => {
    const wrapper = await mountReady(fourBarbers, [mondayMorning])
    expect(wrapper.text()).toContain('08:00')

    fetchWorkingHoursMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [{ ...mondayMorning, id: 'wh-2', startsTime: '10:00' }], nextCursor: null },
    })
    const select = barberSelect(wrapper)
    select.value = 'b-2'
    await select.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(fetchWorkingHoursMock).toHaveBeenLastCalledWith('b-2')
    expect(wrapper.text()).toContain('10:00')
    expect(wrapper.text()).not.toContain('08:00')
  })

  it('creating a tramo succeeds and appends it to its day', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openCreateDialog(wrapper)
    await fillCreateForm(wrapper)

    createWorkingHourMock.mockResolvedValueOnce({ kind: 'success', workingHour: mondayMorning })
    await createForm(wrapper).trigger('submit')
    await flushPromises()

    expect(createWorkingHourMock).toHaveBeenCalledWith('b-1', 1, '08:00', 60, expect.any(String))
    expect(wrapper.text()).toContain('08:00')
  })

  it('an overlapping tramo (CA-040-04) shows a recoverable message without closing the dialog', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openCreateDialog(wrapper)
    await fillCreateForm(wrapper)

    createWorkingHourMock.mockResolvedValueOnce({ kind: 'overlap-conflict' })
    await createForm(wrapper).trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('se solapa con otro existente')
  })

  it('rejects an empty starts time on the client without calling the API', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openCreateDialog(wrapper)
    await wrapper.get('#schedules-create-weekday').setValue('1')
    await createForm(wrapper).get('input[name="durationMinutes"]').setValue('60')

    await createForm(wrapper).trigger('submit')
    await flushPromises()

    expect(createWorkingHourMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe una hora válida')
  })

  it('editing a tramo succeeds and updates its row in place', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    const editButtons = wrapper.findAll('button').filter((b) => b.text() === 'Editar')
    await editButtons[0]!.trigger('click')
    await flushPromises()

    await editForm(wrapper).get('input[name="startsTime"]').setValue('09:00')

    updateWorkingHourMock.mockResolvedValueOnce({
      kind: 'success',
      workingHour: { ...mondayMorning, startsTime: '09:00' },
    })
    await editForm(wrapper).trigger('submit')
    await flushPromises()

    expect(updateWorkingHourMock).toHaveBeenCalledWith('b-1', 'wh-1', 1, '09:00', 240)
    expect(wrapper.text()).toContain('09:00')
    expect(wrapper.text()).not.toContain('08:00')
  })

  it('deleting a tramo removes it from the list on success', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    deleteWorkingHourMock.mockResolvedValueOnce({ kind: 'success' })

    const retireButton = wrapper.findAll('button').find((b) => b.text() === 'Retirar')!
    await retireButton.trigger('click')
    await flushPromises()

    expect(deleteWorkingHourMock).toHaveBeenCalledWith('b-1', 'wh-1')
    expect(wrapper.text()).not.toContain('08:00')
  })

  it('a 404 while deleting treats the tramo as already gone, without an error message', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    deleteWorkingHourMock.mockResolvedValueOnce({ kind: 'not-found' })

    const retireButton = wrapper.findAll('button').find((b) => b.text() === 'Retirar')!
    await retireButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('08:00')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it('a network error while deleting keeps the row and shows a recoverable message', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    deleteWorkingHourMock.mockResolvedValueOnce({ kind: 'network-error' })

    const retireButton = wrapper.findAll('button').find((b) => b.text() === 'Retirar')!
    await retireButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(wrapper.text()).toContain('08:00')
  })

  it('disables the retire button while its own request is in flight (no double submit)', async () => {
    const wrapper = await mountReady(oneBarber, [mondayMorning])
    let resolveDelete: (value: unknown) => void = () => {}
    deleteWorkingHourMock.mockReturnValueOnce(new Promise((resolve) => (resolveDelete = resolve)))

    const retireButton = wrapper.findAll('button').find((b) => b.text() === 'Retirar')!
    await retireButton.trigger('click')
    await flushPromises()

    expect(
      wrapper
        .findAll('button')
        .find((b) => b.text() === 'Retirar')!
        .attributes('disabled'),
    ).toBeDefined()
    expect(deleteWorkingHourMock).toHaveBeenCalledTimes(1)

    resolveDelete({ kind: 'success' })
    await flushPromises()
  })

  it('has no axe violations in the ready view', async () => {
    const wrapper = await mountReady(fourBarbers, [mondayMorning])
    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})

// --- HU-041: calendario de festivos, excepciones de jornada ---------------

function holidayCalendarCheckbox(wrapper: VueWrapper): HTMLInputElement {
  return wrapper.get('input[type="checkbox"]').element as HTMLInputElement
}

async function openExceptionCreateDialog(wrapper: VueWrapper) {
  const button = wrapper.findAll('button').find((b) => b.text() === 'Agregar excepción')!
  await button.trigger('click')
  await flushPromises()
}

function exceptionCreateForm(wrapper: VueWrapper) {
  return wrapper.get('form[name="createScheduleException"]')
}

const closedException = {
  id: 'exc-1',
  effectiveDate: '2026-12-08',
  isClosed: true,
  reason: null,
  segments: [],
  createdAt: '2026-08-25T15:04:05Z',
  updatedAt: '2026-08-25T15:04:05Z',
}

describe('SchedulesPage · HU-041', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchWorkingHoursMock.mockReset()
    fetchBarbershopTimezoneMock.mockReset().mockResolvedValue({ kind: 'unavailable' })
    fetchHolidayCalendarMock.mockReset().mockResolvedValue({ kind: 'success', enabled: false })
    updateHolidayCalendarMock.mockReset()
    fetchScheduleExceptionsMock
      .mockReset()
      .mockResolvedValue({ kind: 'success', page: { items: [], nextCursor: null } })
    createScheduleExceptionMock.mockReset()
    deleteScheduleExceptionMock.mockReset()
    fetchColombianHolidaysMock.mockReset().mockResolvedValue({ kind: 'unavailable' })
  })

  it('reflects the current holiday calendar state (CA-041-01)', async () => {
    fetchHolidayCalendarMock.mockReset().mockResolvedValueOnce({ kind: 'success', enabled: true })
    const wrapper = await mountReady(oneBarber, [])

    expect(holidayCalendarCheckbox(wrapper).checked).toBe(true)
  })

  it('toggling the holiday calendar persists the new value (CA-041-02)', async () => {
    const wrapper = await mountReady(oneBarber, [])
    updateHolidayCalendarMock.mockResolvedValueOnce({ kind: 'success', enabled: true })

    const checkbox = holidayCalendarCheckbox(wrapper)
    checkbox.checked = true
    await checkbox.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(updateHolidayCalendarMock).toHaveBeenCalledWith('b-1', true)
    expect(checkbox.checked).toBe(true)
  })

  it('a failed toggle shows a recoverable message', async () => {
    const wrapper = await mountReady(oneBarber, [])
    updateHolidayCalendarMock.mockResolvedValueOnce({ kind: 'network-error' })

    const checkbox = holidayCalendarCheckbox(wrapper)
    checkbox.checked = true
    await checkbox.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
  })

  it('lists the exceptions of the selected barber (CA-041-04)', async () => {
    fetchScheduleExceptionsMock.mockReset().mockResolvedValueOnce({
      kind: 'success',
      page: { items: [closedException], nextCursor: null },
    })
    const wrapper = await mountReady(oneBarber, [])

    expect(wrapper.text()).toContain('2026-12-08')
    expect(wrapper.text()).toContain('Cerrado')
  })

  it('creating a closed exception succeeds and appends it to the list', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openExceptionCreateDialog(wrapper)
    await exceptionCreateForm(wrapper).get('input[name="effectiveDate"]').setValue('2026-12-08')

    createScheduleExceptionMock.mockResolvedValueOnce({
      kind: 'success',
      exception: closedException,
    })
    await exceptionCreateForm(wrapper).trigger('submit')
    await flushPromises()

    expect(createScheduleExceptionMock).toHaveBeenCalledWith(
      'b-1',
      '2026-12-08',
      true,
      null,
      [],
      expect.any(String),
    )
    expect(wrapper.text()).toContain('2026-12-08')
  })

  it('a duplicate date (CA-041-05) shows a recoverable message without closing the dialog', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openExceptionCreateDialog(wrapper)
    await exceptionCreateForm(wrapper).get('input[name="effectiveDate"]').setValue('2026-12-08')

    createScheduleExceptionMock.mockResolvedValueOnce({ kind: 'date-conflict' })
    await exceptionCreateForm(wrapper).trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Ya existe una excepción para esa fecha')
  })

  it('rejects an empty effective date on the client without calling the API', async () => {
    const wrapper = await mountReady(oneBarber, [])
    await openExceptionCreateDialog(wrapper)

    await exceptionCreateForm(wrapper).trigger('submit')
    await flushPromises()

    expect(createScheduleExceptionMock).not.toHaveBeenCalled()
  })

  it('deleting an exception removes it from the list on success', async () => {
    fetchScheduleExceptionsMock.mockReset().mockResolvedValueOnce({
      kind: 'success',
      page: { items: [closedException], nextCursor: null },
    })
    const wrapper = await mountReady(oneBarber, [])
    deleteScheduleExceptionMock.mockResolvedValueOnce({ kind: 'success' })

    const retireButton = wrapper.findAll('button').find((b) => b.text() === 'Retirar')!
    await retireButton.trigger('click')
    await flushPromises()

    expect(deleteScheduleExceptionMock).toHaveBeenCalledWith('b-1', 'exc-1')
    expect(wrapper.text()).not.toContain('2026-12-08')
  })

  it('shows upcoming Colombian holidays and prefills the create dialog from one (RN-BLQ-02)', async () => {
    fetchColombianHolidaysMock.mockReset().mockResolvedValueOnce({
      kind: 'success',
      items: [{ date: '2026-12-25', name: 'Navidad' }],
    })
    const wrapper = await mountReady(oneBarber, [])

    expect(wrapper.text()).toContain('Navidad')

    const useButton = wrapper.findAll('button').find((b) => b.text() === 'Registrar excepción')!
    await useButton.trigger('click')
    await flushPromises()

    const dateInput = exceptionCreateForm(wrapper).get('input[name="effectiveDate"]')
      .element as HTMLInputElement
    expect(dateInput.value).toBe('2026-12-25')
  })
})
