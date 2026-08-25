/**
 * Pruebas de CatalogPage (HU-022): carga (uno, varios, vacío), error
 * recuperable, alta (éxito, validación, conflicto de nombre, conflicto de
 * idempotencia, error de red, doble envío bloqueado), edición (éxito, no
 * encontrado, conflicto de nombre), paginación ("Cargar más"), foco y
 * ausencia de campos fuera de alcance (asignaciones, activación, citas).
 * catalogApi se sustituye por un doble de prueba; el recorrido real contra
 * el API vive en el E2E de HU-022.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchMock = vi.hoisted(() => vi.fn())
const createMock = vi.hoisted(() => vi.fn())
const updateMock = vi.hoisted(() => vi.fn())
const previewDeactivationMock = vi.hoisted(() => vi.fn())
const deactivateMock = vi.hoisted(() => vi.fn())
const reactivateMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/catalogApi', () => ({
  fetchServices: fetchMock,
  createService: createMock,
  updateService: updateMock,
  previewDeactivation: previewDeactivationMock,
  deactivateService: deactivateMock,
  reactivateService: reactivateMock,
}))

const { default: CatalogPage } = await import('../CatalogPage.vue')

function service(id: string, name: string, overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id,
    name,
    description: null,
    durationMinutes: 30,
    price: '45000.00',
    currency: 'COP',
    isActive: true,
    deactivatedAt: null,
    createdAt: '2026-08-24T15:04:05Z',
    updatedAt: '2026-08-24T15:04:05Z',
    ...overrides,
  }
}

const oneService = [service('s-1', 'Corte clásico')]
const fourServices = [
  service('s-1', 'Corte clásico'),
  service('s-2', 'Corte + barba'),
  service('s-3', 'Barba'),
  service('s-4', 'Cejas'),
]

// stubs.teleport hace que Vue Test Utils renderice el contenido de
// <Teleport to="body"> EN EL LUGAR (mismo criterio que StaffPage.test.ts).
function mountPage() {
  return mount(CatalogPage, { global: { stubs: { teleport: true } } })
}

async function mountReady(items = oneService, nextCursor: string | null = null) {
  fetchMock.mockResolvedValueOnce({ kind: 'success', page: { items: [...items], nextCursor } })
  const wrapper = mountPage()
  await flushPromises()
  return wrapper
}

function openDialogElement(wrapper: VueWrapper): HTMLElement {
  const el = wrapper.element.querySelector('.base-dialog--open')
  if (!el) throw new Error('no open dialog found')
  return el as HTMLElement
}

function openDialogInput(wrapper: VueWrapper, name: string): HTMLInputElement {
  return openDialogElement(wrapper).querySelector(`input[name="${name}"]`) as HTMLInputElement
}

function openDialogForm(wrapper: VueWrapper): HTMLFormElement {
  return openDialogElement(wrapper).querySelector('form') as HTMLFormElement
}

function setInputValue(input: HTMLInputElement, value: string) {
  input.value = value
  input.dispatchEvent(new Event('input'))
}

function submitOpenDialog(wrapper: VueWrapper) {
  openDialogForm(wrapper).dispatchEvent(new Event('submit', { cancelable: true }))
}

function fillCreateForm(wrapper: VueWrapper, name: string, price = '45000.00', duration = '30') {
  setInputValue(openDialogInput(wrapper, 'name'), name)
  setInputValue(openDialogInput(wrapper, 'durationMinutes'), duration)
  setInputValue(openDialogInput(wrapper, 'price'), price)
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

function findButtonByText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((b) => b.text() === text)
  if (!button) throw new Error(`button ${JSON.stringify(text)} not found`)
  return button
}

// clickDialogButton apunta al botón DENTRO del diálogo abierto: el mismo
// texto ("Desactivar"/"Reactivar") también nombra el botón de la lista que
// ABRE el diálogo, así que findButtonByText (que toma el primer match del
// documento) no sirve para el botón de confirmación.
function clickDialogButton(wrapper: VueWrapper, text: string) {
  const button = Array.from(openDialogElement(wrapper).querySelectorAll('button')).find(
    (b) => b.textContent?.trim() === text,
  )
  if (!button) throw new Error(`dialog button ${JSON.stringify(text)} not found`)
  button.click()
}

describe('CatalogPage', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    createMock.mockReset()
    updateMock.mockReset()
    previewDeactivationMock.mockReset()
    deactivateMock.mockReset()
    reactivateMock.mockReset()
  })

  it('shows a non-blank loading state, then the loaded catalog', async () => {
    let resolveFetch: (value: unknown) => void = () => {}
    fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')

    resolveFetch({ kind: 'success', page: { items: [...oneService], nextCursor: null } })
    await flushPromises()

    expect(wrapper.text()).toContain('Corte clásico')
  })

  it('a catalog with one service and one with four use the same list component (CA-022-01)', async () => {
    const oneWrapper = await mountReady(oneService)
    expect(oneWrapper.findAll('li').length).toBe(1)

    const fourWrapper = await mountReady(fourServices)
    expect(fourWrapper.findAll('li').length).toBe(4)
    expect(fourWrapper.text()).toContain('Corte + barba')
    expect(fourWrapper.text()).toContain('Cejas')
  })

  it('shows an empty state with no special data shape when there are no services', async () => {
    const wrapper = await mountReady([])
    expect(wrapper.text()).toContain('Aún no tienes servicios registrados')
    expect(wrapper.findAll('li').length).toBe(0)
  })

  it('shows duration and price for each service', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { durationMinutes: 45, price: '65000.00' }),
    ])
    expect(wrapper.text()).toContain('45 min')
    expect(wrapper.text()).toContain('65000.00 COP')
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar el catálogo')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [...oneService], nextCursor: null },
    })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Corte clásico')
  })

  it('shows a "Cargar más" button when there is a next page, and appends items without duplicating', async () => {
    const wrapper = await mountReady(oneService, 'opaque-cursor')
    expect(wrapper.text()).toContain('Cargar más')

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [service('s-2', 'Corte + barba')], nextCursor: null },
    })
    await findButtonByText(wrapper, 'Cargar más').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('li').length).toBe(2)
    expect(wrapper.text()).toContain('Corte + barba')
    expect(wrapper.text()).not.toContain('Cargar más')
  })

  // --- Alta -----------------------------------------------------------

  it('opens the create dialog and blocks submission for an empty name', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    expect(openDialogInput(wrapper, 'name')).toBeTruthy()
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe el nombre del servicio.')
  })

  it('blocks submission for a price of zero (DEC-067)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    fillCreateForm(wrapper, 'Servicio Gratis', '0')
    await flushPromises()
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('El precio debe ser mayor que cero.')
  })

  it('on success, prepends the confirmed service once and closes the dialog', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    fillCreateForm(wrapper, 'Nuevo Servicio')
    await flushPromises()

    createMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-new', 'Nuevo Servicio'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('li').length).toBe(2)
    expect(wrapper.text()).toContain('Nuevo Servicio')
  })

  it('sends the same idempotency key across a submit and a network-error retry of the same logical attempt', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    fillCreateForm(wrapper, 'Reintentado')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitOpenDialog(wrapper)
    await flushPromises()

    createMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-retry', 'Reintentado'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(2)
    const firstKey = createMock.mock.calls[0]![1]
    const secondKey = createMock.mock.calls[1]![1]
    expect(secondKey).toBe(firstKey)
  })

  it('blocks a second submit while the first is still in flight (no double POST)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    fillCreateForm(wrapper, 'Doble Envío')
    await flushPromises()

    let resolveCreate: (value: unknown) => void = () => {}
    createMock.mockReturnValueOnce(new Promise((resolve) => (resolveCreate = resolve)))
    submitOpenDialog(wrapper)
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    resolveCreate({ kind: 'success', service: service('s-x', 'Doble Envío') })
    await flushPromises()
  })

  it('on name-conflict, shows a recoverable message and keeps the typed value (DEC-067)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    const nameInput = openDialogInput(wrapper, 'name')
    fillCreateForm(wrapper, 'Ya Existe')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'name-conflict' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('Ese nombre ya está en uso')
    expect(nameInput.value).toBe('Ya Existe')
  })

  it('on idempotency-conflict, shows a recoverable message and keeps the typed value', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    const nameInput = openDialogInput(wrapper, 'name')
    fillCreateForm(wrapper, 'Conflicto')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos completar el intento anterior')
    expect(nameInput.value).toBe('Conflicto')
  })

  it('never shows placeholders for barber assignment, availability, or appointments (HU-024 activation status is real, not a placeholder)', async () => {
    const wrapper = await mountReady(fourServices)
    const text = wrapper.text().toLowerCase()
    for (const forbidden of ['barbero', 'asignar', 'disponib', 'agenda', 'cita', 'horario']) {
      expect(text).not.toContain(forbidden)
    }
  })

  // --- Edición ----------------------------------------------------------

  it('opens the edit dialog prefilled with the service fields and edits on success', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', {
        description: 'Con lavado',
        durationMinutes: 30,
        price: '45000.00',
      }),
    ])
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    expect(openDialogInput(wrapper, 'name').value).toBe('Corte clásico')
    expect(openDialogInput(wrapper, 'description').value).toBe('Con lavado')
    expect(openDialogInput(wrapper, 'durationMinutes').value).toBe('30')
    expect(openDialogInput(wrapper, 'price').value).toBe('45000.00')

    setInputValue(openDialogInput(wrapper, 'price'), '50000.00')
    await flushPromises()

    updateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', { price: '50000.00' }),
    })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(updateMock).toHaveBeenCalledWith(
      's-1',
      expect.objectContaining({ name: 'Corte clásico', price: '50000.00' }),
    )
    expect(wrapper.findAll('li').length).toBe(1)
    expect(wrapper.text()).toContain('50000.00 COP')
  })

  it('replaces the item by id without duplicating or reordering unstably', async () => {
    const wrapper = await mountReady(fourServices)
    const editButtons = wrapper.findAll('button').filter((b) => b.text() === 'Editar')
    await editButtons[2]!.trigger('click') // "Barba", third item
    await flushPromises()

    updateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-3', 'Barba Premium'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()

    const names = wrapper.findAll('.catalog-page__item-name').map((n) => n.text())
    expect(names).toEqual(['Corte clásico', 'Corte + barba', 'Barba Premium', 'Cejas'])
  })

  it('on not-found, shows a recoverable message', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    updateMock.mockResolvedValueOnce({ kind: 'not-found' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('ya no está disponible')
  })

  it('on name-conflict while editing, shows a recoverable message', async () => {
    const wrapper = await mountReady(fourServices)
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper, 'name'), 'Corte + barba')
    await flushPromises()

    updateMock.mockResolvedValueOnce({ kind: 'name-conflict' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('Ese nombre ya está en uso')
  })

  // --- HU-024: desactivar/reactivar --------------------------------------

  it('shows Activo/Inactivo badges reflecting each service, and the matching action button', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { isActive: true }),
      service('s-2', 'Corte + barba', { isActive: false, deactivatedAt: '2026-08-25T10:00:00Z' }),
    ])

    expect(wrapper.text()).toContain('Activo')
    expect(wrapper.text()).toContain('Inactivo')
    expect(() => findButtonByText(wrapper, 'Desactivar')).not.toThrow()
    expect(() => findButtonByText(wrapper, 'Reactivar')).not.toThrow()
  })

  it('opens the deactivate dialog, queries the real impact, and confirms with a fresh idempotency key (CA-024-01/02/04)', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })

    await findButtonByText(wrapper, 'Desactivar').trigger('click')
    await flushPromises()

    expect(previewDeactivationMock).toHaveBeenCalledWith('s-1')
    expect(wrapper.text()).toContain('No hay citas futuras')

    deactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', { isActive: false, deactivatedAt: '2026-08-25T12:00:00Z' }),
      affectedAppointments: 0,
    })
    clickDialogButton(wrapper, 'Desactivar')
    await flushPromises()

    expect(deactivateMock).toHaveBeenCalledTimes(1)
    const [serviceId, key] = deactivateMock.mock.calls[0] as [string, string]
    expect(serviceId).toBe('s-1')
    expect(typeof key).toBe('string')
    expect(key.length).toBeGreaterThan(0)
    // El diálogo se cierra y la lista refleja el nuevo estado.
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
    expect(wrapper.text()).toContain('Inactivo')
  })

  it('cancelling the deactivate dialog never calls deactivateService', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })

    await findButtonByText(wrapper, 'Desactivar').trigger('click')
    await flushPromises()
    clickDialogButton(wrapper, 'Cancelar')
    await flushPromises()

    expect(deactivateMock).not.toHaveBeenCalled()
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
  })

  it('on a transition-conflict, reloads the real state instead of assuming success', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })
    await findButtonByText(wrapper, 'Desactivar').trigger('click')
    await flushPromises()

    deactivateMock.mockResolvedValueOnce({ kind: 'transition-conflict' })
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [service('s-1', 'Corte clásico', { isActive: false, deactivatedAt: '2026-08-25T12:00:00Z' })], nextCursor: null },
    })
    clickDialogButton(wrapper, 'Desactivar')
    await flushPromises()

    expect(wrapper.text()).toContain('El estado de este servicio cambió')
    expect(fetchMock).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Inactivo')
  })

  it('reactivates a service without a duration/price/name field in the confirmation (CA-024-05)', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { isActive: false, deactivatedAt: '2026-08-25T10:00:00Z' }),
    ])

    await findButtonByText(wrapper, 'Reactivar').trigger('click')
    await flushPromises()
    expect(openDialogElement(wrapper).querySelector('input')).toBeNull()

    reactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', { isActive: true, deactivatedAt: null }),
    })
    clickDialogButton(wrapper, 'Reactivar')
    await flushPromises()

    expect(reactivateMock).toHaveBeenCalledWith('s-1', expect.any(String))
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
    expect(wrapper.text()).toContain('Activo')
  })

  // --- Accesibilidad ------------------------------------------------------

  it('has no axe violations in the list view', async () => {
    const wrapper = await mountReady(fourServices)
    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })

  it('has no axe violations with the create dialog open', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })

  it('has no axe violations with the deactivate dialog open', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })
    await findButtonByText(wrapper, 'Desactivar').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
