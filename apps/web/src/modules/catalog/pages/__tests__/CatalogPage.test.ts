/**
 * Pruebas de CatalogPage (HU-022): carga (uno, varios, vacío), error
 * recuperable, alta (éxito, validación, conflicto de nombre, conflicto de
 * idempotencia, error de red, doble envío bloqueado), edición (éxito, no
 * encontrado, conflicto de nombre), paginador numerado con total (DEC-103)
 * y buscador con debounce, foco y ausencia de campos fuera de alcance
 * (asignaciones, activación, citas).
 * catalogApi se sustituye por un doble de prueba; el recorrido real contra
 * el API vive en el E2E de HU-022.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import { DEFAULT_MIN_HOLD_MS, PAGE_MIN_HOLD_MS } from '@/shared/composables'

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
// stubs.transition: jsdom no tiene motor CSS real (nunca dispara
// transitionend); el stub integrado de VTU sustituye el <Transition
// mode="out-in"> que envuelve carga/error/listo (issue 2026-09-28) por uno
// que cambia de hijo al instante, sin animación ni espera.
function mountPage() {
  return mount(CatalogPage, { global: { stubs: { teleport: true, transition: true } } })
}

// useMinHoldLoading (mismo issue) mantiene el rombo/esqueleto al menos
// DEFAULT_MIN_HOLD_MS con un setTimeout REAL antes de dejar pasar el
// contenido: flushPromises() (solo microtareas) no alcanza a esperarlo.
// Con margen sobre el valor exacto para no quedar al borde por jitter del
// entorno de pruebas.
function waitOutInitialLoadHold() {
  return new Promise((resolve) => setTimeout(resolve, DEFAULT_MIN_HOLD_MS + 50))
}

// Buscar o cambiar de página (issue 2026-09-29) mantiene los esqueletos al
// menos PAGE_MIN_HOLD_MS antes de mostrar el resultado.
function waitOutPageLoadHold() {
  return new Promise((resolve) => setTimeout(resolve, PAGE_MIN_HOLD_MS + 50))
}

// listPage arma la forma real de ServicePage (CA-022-01, DEC-103): page/
// pageSize/total/totalPages, ya no nextCursor. total/totalPages por
// defecto asumen que items es la única página (el caso común de estas
// pruebas); las pruebas de paginación los sobrescriben explícitamente.
function listPage(items: unknown[], overrides: Partial<Record<string, unknown>> = {}) {
  return {
    items: [...items],
    page: 1,
    pageSize: 20,
    total: items.length,
    totalPages: 1,
    ...overrides,
  }
}

async function mountReady(
  items = oneService,
  pageOverrides: Partial<Record<string, unknown>> = {},
) {
  fetchMock.mockResolvedValueOnce({ kind: 'success', page: listPage(items, pageOverrides) })
  const wrapper = mountPage()
  await flushPromises()
  await waitOutInitialLoadHold()
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

// El interruptor del diálogo de estado es la acción (no hay botón de
// confirmar): tocarlo desactiva o reactiva.
function clickDialogSwitch(wrapper: VueWrapper) {
  const control = openDialogElement(wrapper).querySelector<HTMLButtonElement>('[role="switch"]')
  if (!control) throw new Error('dialog switch not found')
  control.click()
}

// clickDialogButton apunta al botón DENTRO del diálogo abierto (confirmar
// "Cancelar"/"Salir"), no a los de la lista.
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
    resetToasts()
  })

  afterEach(() => {
    resetToasts()
  })

  it('shows a non-blank loading state from the very first instant', () => {
    fetchMock.mockReturnValueOnce(new Promise(() => {})) // nunca resuelve en esta prueba
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')
  })

  // useMinHoldLoading (issue reportado 2026-09-28, "se ve como se genera
  // el objeto... quiero el tema de la carga para tapar ese evento"): el
  // rombo/esqueleto no cede el paso al contenido real antes de
  // DEFAULT_MIN_HOLD_MS, aunque la respuesta ya haya llegado — así una
  // carga instantánea nunca deja ver la tabla "armándose" bajo el rombo.
  it('holds the loading state for a minimum duration even when the response is instant', async () => {
    vi.useFakeTimers()
    fetchMock.mockResolvedValueOnce({ kind: 'success', page: listPage(oneService) })
    const wrapper = mountPage()

    await flushPromises()
    expect(wrapper.text()).toContain('Cargando')
    expect(wrapper.text()).not.toContain('Corte clásico')

    await vi.advanceTimersByTimeAsync(DEFAULT_MIN_HOLD_MS - 50)
    expect(wrapper.text()).not.toContain('Corte clásico')

    await vi.advanceTimersByTimeAsync(100)
    vi.useRealTimers()

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

  it('never concatenates a literal "$" to the price (estandar-diseno-visual.md §10.3, Fase 5)', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { durationMinutes: 45, price: '65000.00' }),
    ])
    expect(wrapper.text()).not.toContain('$')
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()
    await waitOutInitialLoadHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar el catálogo')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({ kind: 'success', page: listPage(oneService) })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    await waitOutInitialLoadHold()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Corte clásico')
  })

  // --- Paginador (issue 2026-09-28, "colócalos en una tabla con paginador
  //     y con su buscador... con scroll infinito está feo") -------------

  it('shows the total count but no page-number controls when there is only one page', async () => {
    const wrapper = await mountReady(oneService, { total: 1, totalPages: 1 })

    expect(wrapper.text()).toContain('1 servicio')
    expect(wrapper.find('nav[aria-label="Paginación de servicios"]').exists()).toBe(false)
  })

  it('shows numbered page controls and fetches the requested page on click (DEC-103)', async () => {
    const wrapper = await mountReady(oneService, { total: 34, totalPages: 2 })
    expect(wrapper.text()).toContain('34 servicios')

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: listPage([service('s-2', 'Corte + barba')], { page: 2, total: 34, totalPages: 2 }),
    })
    await findButtonByText(wrapper, '2').trigger('click')
    await flushPromises()
    await waitOutPageLoadHold()
    await flushPromises()

    expect(fetchMock).toHaveBeenLastCalledWith({ page: 2, pageSize: 20, search: undefined })
    expect(wrapper.findAll('li').length).toBe(1)
    expect(wrapper.text()).toContain('Corte + barba')
    expect(wrapper.text()).not.toContain('Corte clásico')
  })

  it('a page-fetch failure keeps the current rows visible and offers Reintentar', async () => {
    const wrapper = await mountReady(oneService, { total: 34, totalPages: 2 })

    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    await findButtonByText(wrapper, '2').trigger('click')
    await flushPromises()
    await waitOutPageLoadHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar esta página')
    // Las filas de la página anterior NUNCA desaparecen por un fallo de
    // página (issue 2026-09-28: cambiar de página o buscar no debe
    // "parpadear" toda la pantalla, a diferencia de la carga inicial).
    expect(wrapper.text()).toContain('Corte clásico')

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: listPage([service('s-2', 'Corte + barba')], { page: 2, total: 34, totalPages: 2 }),
    })
    await findButtonByText(wrapper, 'Reintentar').trigger('click')
    await flushPromises()
    await waitOutPageLoadHold()
    await flushPromises()

    expect(fetchMock).toHaveBeenLastCalledWith({ page: 2, pageSize: 20, search: undefined })
    expect(wrapper.text()).toContain('Corte + barba')
  })

  it('fits the first page to the viewport without refetching or showing the oversized table', async () => {
    // jsdom no tiene layout: se simula un visor con scroll propio de 600px, la
    // lista a 200px del borde y un paginador de 60px -> caben 3 filas de 80px.
    const scroller = document.createElement('div')
    scroller.style.overflowY = 'auto'
    Object.defineProperty(scroller, 'clientHeight', { value: 600 })
    document.body.appendChild(scroller)
    const rect = (top: number, height: number) => ({ top, bottom: top + height, height }) as DOMRect
    const rectSpy = vi
      .spyOn(HTMLElement.prototype, 'getBoundingClientRect')
      .mockImplementation(function (this: HTMLElement) {
        if (this === scroller) return rect(0, 600)
        if (this.classList.contains('catalog-page__list')) return rect(200, 320)
        if (this.classList.contains('catalog-page__pagination')) return rect(536, 60)
        return rect(0, 0)
      })

    try {
      fetchMock.mockResolvedValueOnce({
        kind: 'success',
        page: listPage(fourServices, { total: 34, totalPages: 2 }),
      })
      const wrapper = mount(CatalogPage, {
        attachTo: scroller,
        global: { stubs: { teleport: true, transition: true } },
      })
      await flushPromises()
      await waitOutInitialLoadHold()
      await flushPromises()

      expect(fetchMock).toHaveBeenCalledTimes(1)
      expect(wrapper.findAll('li').length).toBe(3)
      expect(wrapper.text()).toContain('34 servicios')
      // 34 servicios a 3 por página = 12 páginas.
      expect(findButtonByText(wrapper, '12').exists()).toBe(true)
      expect(wrapper.find('.catalog-page__ready--fitting').exists()).toBe(false)
      wrapper.unmount()
    } finally {
      rectSpy.mockRestore()
      scroller.remove()
    }
  })

  // --- Buscador -----------------------------------------------------------

  it('searches after a debounce, resets to page 1, and sends the trimmed term', async () => {
    const wrapper = await mountReady(oneService, { total: 34, totalPages: 2 })

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: listPage([service('s-2', 'Corte + barba')], { total: 1, totalPages: 1 }),
    })
    const search = wrapper.get('input[name="serviceSearch"]')
    setInputValue(search.element as HTMLInputElement, '  barba  ')
    await flushPromises()
    // Antes del debounce, ninguna solicitud nueva.
    expect(fetchMock).toHaveBeenCalledTimes(1)

    await new Promise((resolve) => setTimeout(resolve, 350))
    await flushPromises()

    // Issue 2026-09-29: el resultado no aparece de golpe. Con la respuesta
    // ya recibida, la tabla muestra esqueletos (sin las filas viejas ni las
    // nuevas) hasta cumplir PAGE_MIN_HOLD_MS, y solo entonces entran las filas.
    expect(wrapper.findAll('.catalog-page__skeleton-row--table').length).toBeGreaterThan(0)
    expect(wrapper.text()).not.toContain('Corte clásico')
    expect(wrapper.text()).not.toContain('Corte + barba')

    await waitOutPageLoadHold()
    await flushPromises()

    expect(wrapper.find('.catalog-page__skeleton-row--table').exists()).toBe(false)
    expect(fetchMock).toHaveBeenLastCalledWith({ page: 1, pageSize: 20, search: 'barba' })
    expect(wrapper.text()).toContain('Corte + barba')
    expect(wrapper.text()).not.toContain('Corte clásico')
    expect(wrapper.text()).toContain('1 resultado')
  })

  it('shows a search-specific empty state distinct from a genuinely empty catalog', async () => {
    const wrapper = await mountReady(oneService, { total: 34, totalPages: 2 })

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: listPage([], { total: 0, totalPages: 1 }),
    })
    const search = wrapper.get('input[name="serviceSearch"]')
    setInputValue(search.element as HTMLInputElement, 'zzz-sin-coincidencias')
    await new Promise((resolve) => setTimeout(resolve, 350))
    await flushPromises()
    await waitOutPageLoadHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No encontramos servicios que coincidan con')
    expect(wrapper.text()).not.toContain('Aún no tienes servicios registrados')
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
    // DEC-095: la confirmación es un aviso emergente; la lista sigue siendo
    // el resultado persistente.
    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['success', 'Servicio creado'],
    ])
    expect(toastState.items[0].detail).toContain('Nuevo Servicio')
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

  it('never raises a toast for a failed save: the error stays inline in the dialog (DEC-095)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar servicio').trigger('click')
    await flushPromises()
    fillCreateForm(wrapper, 'Nuevo Servicio')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(toastState.items).toHaveLength(0)
    expect(openDialogElement(wrapper).textContent).toContain('Revisa tu conexión')
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

  // --- Detalle ----------------------------------------------------------

  const longDescription = 'Corte a tijera con lavado, masaje capilar y perfilado. '.repeat(8).trim()

  it('keeps the description out of the list rows', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { description: longDescription }),
    ])

    expect(wrapper.find('.catalog-page__list').text()).not.toContain('Corte a tijera')
  })

  it('opens a detail dialog with the full description, duration, price and status when the name is pressed', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { description: longDescription, durationMinutes: 45 }),
    ])
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)

    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    const dialog = openDialogElement(wrapper)
    expect(dialog.classList.contains('catalog-page__detail-dialog')).toBe(true)
    expect(dialog.textContent).toContain('Corte clásico')
    expect(dialog.textContent).toContain(longDescription)
    expect(dialog.textContent).toContain('45')
    expect(dialog.textContent).toContain('45000.00')
    expect(dialog.textContent).toContain('Activo')
    expect(dialog.textContent).toContain('24/08/2026')
  })

  it('says so when the service has no description', async () => {
    const wrapper = await mountReady()
    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    expect(openDialogElement(wrapper).textContent).toContain('no tiene descripción')
  })

  it('shows since when an inactive service was deactivated', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', {
        isActive: false,
        deactivatedAt: '2026-09-01T15:00:00Z',
      }),
    ])
    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    const dialog = openDialogElement(wrapper)
    expect(dialog.textContent).toContain('Inactivo')
    expect(dialog.textContent).toContain('01/09/2026')
  })

  it('closes the detail dialog with Cerrar without calling the API', async () => {
    const wrapper = await mountReady()
    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    clickDialogButton(wrapper, 'Cerrar')
    await flushPromises()

    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
    expect(updateMock).not.toHaveBeenCalled()
  })

  it('goes from the detail to the edit dialog prefilled with the same service', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { description: 'Con lavado' }),
    ])
    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    clickDialogButton(wrapper, 'Editar')
    await flushPromises()

    const open = wrapper.element.querySelectorAll('.base-dialog--open')
    expect(open.length).toBe(1)
    expect(open[0].classList.contains('catalog-page__edit-dialog')).toBe(true)
    expect(openDialogInput(wrapper, 'name').value).toBe('Corte clásico')
    expect(openDialogInput(wrapper, 'description').value).toBe('Con lavado')
  })

  it('does not open the detail when a row action is pressed', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    const open = wrapper.element.querySelectorAll('.base-dialog--open')
    expect(open.length).toBe(1)
    expect(open[0].classList.contains('catalog-page__detail-dialog')).toBe(false)
  })

  it('has no axe violations with the detail dialog open', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { description: longDescription }),
    ])
    await wrapper.find('.catalog-page__item-trigger').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
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
    expect(toastState.items.map((item) => item.title)).toEqual(['Servicio actualizado'])
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
    // Un único botón "Cambiar estado" por fila, para ambos sentidos.
    expect(wrapper.findAll('button').filter((b) => b.text() === 'Cambiar estado')).toHaveLength(2)
    expect(wrapper.find('button[aria-label^="Desactivar"]').exists()).toBe(false)
    expect(wrapper.find('button[aria-label^="Reactivar"]').exists()).toBe(false)
  })

  it('opens the deactivate dialog, queries the real impact, and confirms with a fresh idempotency key (CA-024-01/02/04)', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })

    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()

    expect(previewDeactivationMock).toHaveBeenCalledWith('s-1')
    expect(wrapper.text()).toContain('No hay citas futuras')
    // Mismo diálogo "Estado del servicio", con el interruptor hacia Inactivo.
    const dialog = openDialogElement(wrapper)
    expect(dialog.textContent).toContain('Estado del servicio')
    // El interruptor arranca en "activo" (estado actual) y no hay botón
    // aparte de confirmar.
    expect(dialog.querySelector('[role="switch"]')?.getAttribute('aria-checked')).toBe('true')
    expect(
      Array.from(dialog.querySelectorAll('button')).some(
        (b) => b.textContent?.trim() === 'Desactivar',
      ),
    ).toBe(false)

    deactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', {
        isActive: false,
        deactivatedAt: '2026-08-25T12:00:00Z',
      }),
      affectedAppointments: 0,
    })
    clickDialogSwitch(wrapper)
    await flushPromises()

    expect(deactivateMock).toHaveBeenCalledTimes(1)
    const [serviceId, key] = deactivateMock.mock.calls[0] as [string, string]
    expect(serviceId).toBe('s-1')
    expect(typeof key).toBe('string')
    expect(key.length).toBeGreaterThan(0)
    // El diálogo NO se cierra: muestra el estado nuevo (interruptor en
    // "inactivo") hasta que se pulsa "Salir"; la lista ya refleja el cambio.
    expect(wrapper.find('.base-dialog--open').exists()).toBe(true)
    expect(
      openDialogElement(wrapper).querySelector('[role="switch"]')?.getAttribute('aria-checked'),
    ).toBe('false')
    expect(wrapper.text()).toContain('Inactivo')
    expect(toastState.items.map((item) => item.title)).toEqual(['Servicio desactivado'])

    clickDialogButton(wrapper, 'Salir')
    await flushPromises()
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
  })

  it('lets the same open dialog revert the change with the switch, with a fresh idempotency key', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValue({ kind: 'success', affectedAppointments: 0 })
    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()

    deactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', {
        isActive: false,
        deactivatedAt: '2026-08-25T12:00:00Z',
      }),
      affectedAppointments: 0,
    })
    clickDialogSwitch(wrapper)
    await flushPromises()

    reactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', { isActive: true, deactivatedAt: null }),
    })
    clickDialogSwitch(wrapper)
    await flushPromises()

    const deactivateKey = (deactivateMock.mock.calls[0] as [string, string])[1]
    const reactivateKey = (reactivateMock.mock.calls[0] as [string, string])[1]
    expect(reactivateKey).not.toBe(deactivateKey)
    // Sigue abierto y otra vez en "activo", con el impacto consultado de nuevo.
    expect(wrapper.find('.base-dialog--open').exists()).toBe(true)
    expect(
      openDialogElement(wrapper).querySelector('[role="switch"]')?.getAttribute('aria-checked'),
    ).toBe('true')
    expect(previewDeactivationMock).toHaveBeenCalledTimes(2)
  })

  it('cancelling the deactivate dialog never calls deactivateService', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })

    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()
    clickDialogButton(wrapper, 'Salir')
    await flushPromises()

    expect(deactivateMock).not.toHaveBeenCalled()
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
  })

  it('on a transition-conflict, reloads the real state instead of assuming success', async () => {
    const wrapper = await mountReady()
    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })
    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()

    deactivateMock.mockResolvedValueOnce({ kind: 'transition-conflict' })
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: listPage([
        service('s-1', 'Corte clásico', {
          isActive: false,
          deactivatedAt: '2026-08-25T12:00:00Z',
        }),
      ]),
    })
    clickDialogSwitch(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('El estado de este servicio cambió')
    expect(fetchMock).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Inactivo')
  })

  it('previews the reactivated service (duration, price, inactive since) and what is kept', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', {
        isActive: false,
        deactivatedAt: '2026-08-25T15:00:00Z',
        durationMinutes: 45,
        price: '22000.00',
      }),
    ])

    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()

    const dialog = openDialogElement(wrapper)
    expect(dialog.querySelector('[role="switch"]')?.getAttribute('aria-checked')).toBe('false')
    const facts = dialog.querySelector('dl')?.textContent ?? ''
    expect(facts).toContain('45')
    expect(facts).toContain('22000.00')
    expect(facts).toContain('25/08/2026')
    expect(dialog.querySelectorAll('ul li').length).toBe(3)
  })

  it('reactivates a service without a duration/price/name field in the confirmation (CA-024-05)', async () => {
    const wrapper = await mountReady([
      service('s-1', 'Corte clásico', { isActive: false, deactivatedAt: '2026-08-25T10:00:00Z' }),
    ])

    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()
    expect(openDialogElement(wrapper).querySelector('input')).toBeNull()

    previewDeactivationMock.mockResolvedValueOnce({ kind: 'success', affectedAppointments: 0 })
    reactivateMock.mockResolvedValueOnce({
      kind: 'success',
      service: service('s-1', 'Corte clásico', { isActive: true, deactivatedAt: null }),
    })
    clickDialogSwitch(wrapper)
    await flushPromises()

    expect(reactivateMock).toHaveBeenCalledWith('s-1', expect.any(String))
    // Tampoco se cierra al reactivar: queda listo para volver a desactivar.
    expect(wrapper.find('.base-dialog--open').exists()).toBe(true)
    expect(wrapper.text()).toContain('Activo')
    expect(toastState.items.map((item) => item.title)).toEqual(['Servicio reactivado'])
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
    await findButtonByText(wrapper, 'Cambiar estado').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })

  // heading-order: BaseAlert.vue (HU-009) fija el título de una alerta como
  // `<h4>` porque es la etiqueta de un widget transitorio (`role="alert"`),
  // no un encabezado del esquema del documento; junto al `<h1>` real de esta
  // página, axe interpreta ese salto como un esquema de encabezados roto.
  // Mismo criterio documentado en LoginPage.test.ts/SettingsPage.test.ts: no
  // es un defecto de esta pantalla ni de BaseAlert.
  it('has no axe violations in the load-error state', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()
    await waitOutInitialLoadHold()
    await flushPromises()

    const results = await axe(wrapper.element, {
      rules: { ...axeOptions.rules, 'heading-order': { enabled: false } },
    })
    expect(results).toHaveNoViolations()
  })
})
