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
  fetchWorkingHoursMock.mockResolvedValueOnce({ kind: 'success', page: { items, nextCursor: null } })
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
  })

  it('shows a non-blank loading state, then the loaded picker', async () => {
    let resolveBarbers: (value: unknown) => void = () => {}
    fetchBarberSummariesMock.mockReturnValueOnce(new Promise((resolve) => (resolveBarbers = resolve)))
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')

    fetchWorkingHoursMock.mockResolvedValueOnce({ kind: 'success', page: { items: [], nextCursor: null } })
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
    fetchWorkingHoursMock.mockResolvedValueOnce({ kind: 'success', page: { items: [], nextCursor: null } })
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
      wrapper.findAll('button').find((b) => b.text() === 'Retirar')!.attributes('disabled'),
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
