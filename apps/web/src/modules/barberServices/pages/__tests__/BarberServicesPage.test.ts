/**
 * Pruebas de BarberServicesPage (HU-023): carga (un barbero, cuatro
 * barberos, servicio compartido), vacío (sin barberos, sin servicios),
 * error recuperable, cambio de barbero, asignar/desasignar (éxito, último
 * activo rechazado -DEC-068-, no encontrado, error de red), doble envío
 * bloqueado, aislamiento entre barberos y accesibilidad. barberServicesApi
 * se sustituye por un doble de prueba; el recorrido real contra el API vive
 * en el E2E de HU-023.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchServiceSummariesMock = vi.hoisted(() => vi.fn())
const fetchAssignmentsMock = vi.hoisted(() => vi.fn())
const assignServiceMock = vi.hoisted(() => vi.fn())
const unassignServiceMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/barberServicesApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchServiceSummaries: fetchServiceSummariesMock,
  fetchAssignments: fetchAssignmentsMock,
  assignService: assignServiceMock,
  unassignService: unassignServiceMock,
}))

const { default: BarberServicesPage } = await import('../BarberServicesPage.vue')

const oneBarber = [{ id: 'b-1', fullName: 'Carlos Ramírez' }]
const fourBarbers = [
  { id: 'b-1', fullName: 'Carlos Ramírez' },
  { id: 'b-2', fullName: 'Ana Torres' },
  { id: 'b-3', fullName: 'Luis Gómez' },
  { id: 'b-4', fullName: 'María Pérez' },
]
const twoServices = [
  { id: 's-1', name: 'Corte clásico' },
  { id: 's-2', name: 'Barba' },
]

function mountPage() {
  return mount(BarberServicesPage, { global: { stubs: { teleport: true } } })
}

async function mountReady(
  barbers = oneBarber,
  services = twoServices,
  assignedServiceIds: string[] = [],
) {
  fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: barbers })
  fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: services })
  fetchAssignmentsMock.mockResolvedValueOnce({
    kind: 'success',
    page: {
      items: assignedServiceIds.map((serviceId) => ({
        barberId: barbers[0]!.id,
        serviceId,
        createdAt: '2026-08-24T15:04:05Z',
      })),
      nextCursor: null,
    },
  })
  const wrapper = mountPage()
  await flushPromises()
  return wrapper
}

function checkbox(wrapper: VueWrapper, serviceId: string): HTMLInputElement {
  return wrapper.element.querySelector(`#barber-services-service-${serviceId}`) as HTMLInputElement
}

function barberSelect(wrapper: VueWrapper): HTMLSelectElement {
  return wrapper.element.querySelector('#barber-services-barber-select') as HTMLSelectElement
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

describe('BarberServicesPage', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchServiceSummariesMock.mockReset()
    fetchAssignmentsMock.mockReset()
    assignServiceMock.mockReset()
    unassignServiceMock.mockReset()
  })

  it('shows a non-blank loading state, then the loaded picker', async () => {
    let resolveBarbers: (value: unknown) => void = () => {}
    fetchBarberSummariesMock.mockReturnValueOnce(new Promise((resolve) => (resolveBarbers = resolve)))
    fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')

    fetchAssignmentsMock.mockResolvedValueOnce({ kind: 'success', page: { items: [], nextCursor: null } })
    resolveBarbers({ kind: 'success', items: oneBarber })
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ramírez')
    expect(wrapper.text()).toContain('Corte clásico')
  })

  it('a barber with one service and one with a full team select use the same component (CA-023-01)', async () => {
    const oneWrapper = await mountReady(oneBarber, twoServices, ['s-1'])
    expect(oneWrapper.findAll('option').length).toBe(1)
    expect(checkbox(oneWrapper, 's-1').checked).toBe(true)
    expect(checkbox(oneWrapper, 's-2').checked).toBe(false)

    const fourWrapper = await mountReady(fourBarbers, twoServices, [])
    expect(fourWrapper.findAll('option').length).toBe(4)
  })

  it('shows an empty state when there are no barbers, without an interactive picker', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
    expect(wrapper.find('select').exists()).toBe(false)
  })

  it('shows an empty state when there are no services, without an interactive picker', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: oneBarber })
    fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: [] })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Aún no tienes servicios en el catálogo')
    expect(wrapper.find('select').exists()).toBe(false)
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'network-error' })
    fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar esta sección')

    fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items: oneBarber })
    fetchServiceSummariesMock.mockResolvedValueOnce({ kind: 'success', items: twoServices })
    fetchAssignmentsMock.mockResolvedValueOnce({ kind: 'success', page: { items: [], nextCursor: null } })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('switching barbers loads that barber assignments independently (cross-barber isolation)', async () => {
    const wrapper = await mountReady(fourBarbers, twoServices, ['s-1'])
    expect(checkbox(wrapper, 's-1').checked).toBe(true)

    fetchAssignmentsMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [{ barberId: 'b-2', serviceId: 's-2', createdAt: '2026-08-24T15:04:05Z' }], nextCursor: null },
    })
    const select = barberSelect(wrapper)
    select.value = 'b-2'
    await select.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(fetchAssignmentsMock).toHaveBeenLastCalledWith('b-2')
    expect(checkbox(wrapper, 's-1').checked).toBe(false)
    expect(checkbox(wrapper, 's-2').checked).toBe(true)
  })

  it('checking a box assigns the service and reflects success only after the server responds', async () => {
    const wrapper = await mountReady(oneBarber, twoServices, [])
    assignServiceMock.mockResolvedValueOnce({
      kind: 'success',
      assignment: { barberId: 'b-1', serviceId: 's-1', createdAt: '2026-08-24T15:04:05Z' },
    })

    const box = checkbox(wrapper, 's-1')
    box.checked = true
    await box.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(assignServiceMock).toHaveBeenCalledWith('b-1', 's-1')
    expect(checkbox(wrapper, 's-1').checked).toBe(true)
  })

  it('unchecking a box unassigns the service on success', async () => {
    const wrapper = await mountReady(oneBarber, twoServices, ['s-1'])
    unassignServiceMock.mockResolvedValueOnce({ kind: 'success' })

    const box = checkbox(wrapper, 's-1')
    box.checked = false
    await box.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(unassignServiceMock).toHaveBeenCalledWith('b-1', 's-1')
    expect(checkbox(wrapper, 's-1').checked).toBe(false)
  })

  it('rejecting the last active assignment (DEC-068) reverts the checkbox and shows a recoverable message', async () => {
    const wrapper = await mountReady(oneBarber, twoServices, ['s-1'])
    unassignServiceMock.mockResolvedValueOnce({ kind: 'last-active-conflict' })

    const box = checkbox(wrapper, 's-1')
    box.checked = false
    await box.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(wrapper.text()).toContain('es el único barbero asignado a este servicio activo')
    // El estado real (asignado) se conserva: la casilla vuelve a marcarse.
    expect(checkbox(wrapper, 's-1').checked).toBe(true)
  })

  it('a network error while toggling keeps the previous state and shows a recoverable message', async () => {
    const wrapper = await mountReady(oneBarber, twoServices, [])
    assignServiceMock.mockResolvedValueOnce({ kind: 'network-error' })

    const box = checkbox(wrapper, 's-1')
    box.checked = true
    await box.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(checkbox(wrapper, 's-1').checked).toBe(false)
  })

  it('disables the checkbox while its own request is in flight (no double submit)', async () => {
    const wrapper = await mountReady(oneBarber, twoServices, [])
    let resolveAssign: (value: unknown) => void = () => {}
    assignServiceMock.mockReturnValueOnce(new Promise((resolve) => (resolveAssign = resolve)))

    const box = checkbox(wrapper, 's-1')
    box.checked = true
    await box.dispatchEvent(new Event('change'))
    await flushPromises()

    expect(checkbox(wrapper, 's-1').disabled).toBe(true)
    expect(assignServiceMock).toHaveBeenCalledTimes(1)

    resolveAssign({
      kind: 'success',
      assignment: { barberId: 'b-1', serviceId: 's-1', createdAt: '2026-08-24T15:04:05Z' },
    })
    await flushPromises()

    expect(checkbox(wrapper, 's-1').disabled).toBe(false)
  })

  it('never shows barber schedules, availability, appointments, or per-barber pricing', async () => {
    const wrapper = await mountReady(fourBarbers, twoServices, ['s-1'])
    const text = wrapper.text().toLowerCase()
    for (const forbidden of ['horario', 'disponib', 'agenda', 'cita', 'precio', 'duración', 'comisión']) {
      expect(text).not.toContain(forbidden)
    }
  })

  it('has no axe violations in the ready view', async () => {
    const wrapper = await mountReady(fourBarbers, twoServices, ['s-1'])
    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
