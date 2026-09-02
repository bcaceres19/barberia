/**
 * Pruebas de StaffPage (HU-021): carga (una fila, cuatro filas, vacío),
 * error recuperable, alta (éxito, validación, conflicto de idempotencia,
 * error de red, doble envío bloqueado), edición (éxito, no encontrado),
 * paginación ("Cargar más"), foco y ausencia de campos fuera de alcance.
 * staffApi se sustituye por un doble de prueba; el recorrido real contra
 * el API vive en el E2E de HU-021.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchMock = vi.hoisted(() => vi.fn())
const createMock = vi.hoisted(() => vi.fn())
const renameMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/staffApi', () => ({
  fetchBarbers: fetchMock,
  createBarber: createMock,
  renameBarber: renameMock,
}))

const { default: StaffPage } = await import('../StaffPage.vue')

function barber(id: string, fullName: string) {
  return { id, fullName, createdAt: '2026-08-23T15:04:05Z', updatedAt: '2026-08-23T15:04:05Z' }
}

const oneBarber = [barber('b-1', 'Carlos Ramírez')]
const fourBarbers = [
  barber('b-1', 'Carlos Ramírez'),
  barber('b-2', 'Ana Torres'),
  barber('b-3', 'Luis Gómez'),
  barber('b-4', 'María Pérez'),
]

// stubs.teleport hace que Vue Test Utils renderice el contenido de
// <Teleport to="body"> EN EL LUGAR, dentro del árbol del wrapper, en vez de
// moverlo al document.body real: cada mount() queda completamente aislado
// en su propio wrapper.element, sin ningún riesgo de que un diálogo de una
// prueba anterior contamine la siguiente (BaseDialog usa Teleport;
// shared/ui/__tests__/BaseDialog.test.ts monta contra el DOM real y por eso
// sí necesita limpiarlo a mano, un problema que este stub evita aquí desde
// la raíz).
function mountPage() {
  return mount(StaffPage, { global: { stubs: { teleport: true } } })
}

async function mountReady(items = oneBarber, nextCursor: string | null = null) {
  // Clona el arreglo (StaffPage.vue asigna barbers.value = outcome.page.items
  // por referencia, y "Cargar más" hace push directo sobre ese mismo
  // arreglo): sin clonar, dos pruebas que reutilizan la misma constante
  // oneBarber/fourBarbers mutarían el fixture compartido de la otra,
  // exactamente como haría un backend real que SIEMPRE devuelve un arreglo
  // nuevo por respuesta.
  fetchMock.mockResolvedValueOnce({ kind: 'success', page: { items: [...items], nextCursor } })
  const wrapper = mountPage()
  await flushPromises()
  return wrapper
}

// La pila de BaseDialog mantiene ambos diálogos (alta y edición) SIEMPRE en
// el DOM (v-show, no v-if): se selecciona el que esté visualmente abierto
// en vez de "form" a secas, que ambigüaría entre los dos.
function openDialogElement(wrapper: VueWrapper): HTMLElement {
  const el = wrapper.element.querySelector('.base-dialog--open')
  if (!el) throw new Error('no open dialog found')
  return el as HTMLElement
}

function openDialogInput(wrapper: VueWrapper): HTMLInputElement {
  return openDialogElement(wrapper).querySelector('input[name="fullName"]') as HTMLInputElement
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

// color-contrast se desactiva: jsdom no implementa Canvas2D (mismo
// criterio documentado en shared/ui/__tests__/BaseButton.test.ts); el
// contraste ya está verificado en la tabla aprobada de
// estandar-diseno-visual.md §4.3.
const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

function findButtonByText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((b) => b.text() === text)
  if (!button) throw new Error(`button ${JSON.stringify(text)} not found`)
  return button
}

describe('StaffPage', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    createMock.mockReset()
    renameMock.mockReset()
  })

  it('shows a non-blank loading state, then the loaded team', async () => {
    let resolveFetch: (value: unknown) => void = () => {}
    fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')

    resolveFetch({ kind: 'success', page: { items: [...oneBarber], nextCursor: null } })
    await flushPromises()

    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('a one-person shop and a four-person shop use the same list component (CA-021-01/02)', async () => {
    const oneWrapper = await mountReady(oneBarber)
    expect(oneWrapper.findAll('li').length).toBe(1)

    const fourWrapper = await mountReady(fourBarbers)
    expect(fourWrapper.findAll('li').length).toBe(4)
    expect(fourWrapper.text()).toContain('Ana Torres')
    expect(fourWrapper.text()).toContain('María Pérez')
  })

  it('shows a decorative, accessible monogram with the initials of each barber (Fase 5, issue #153)', async () => {
    const wrapper = await mountReady(fourBarbers)
    const avatars = wrapper.findAll('.staff-page__item-avatar')
    expect(avatars.map((a) => a.text())).toEqual(['CR', 'AT', 'LG', 'MP'])
    for (const avatar of avatars) {
      expect(avatar.attributes('aria-hidden')).toBe('true')
    }
  })

  it('uses a single initial for a one-word name', async () => {
    const wrapper = await mountReady([barber('b-1', 'Madonna')])
    expect(wrapper.get('.staff-page__item-avatar').text()).toBe('M')
  })

  it('shows an empty state with no special data shape when there are no barbers', async () => {
    const wrapper = await mountReady([])
    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
    expect(wrapper.findAll('li').length).toBe(0)
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar el equipo')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [...oneBarber], nextCursor: null },
    })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('shows a "Cargar más" button when there is a next page, and appends items without duplicating', async () => {
    const wrapper = await mountReady(oneBarber, 'opaque-cursor')
    expect(wrapper.text()).toContain('Cargar más')

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [barber('b-2', 'Ana Torres')], nextCursor: null },
    })
    await findButtonByText(wrapper, 'Cargar más').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('li').length).toBe(2)
    expect(wrapper.text()).toContain('Ana Torres')
    expect(wrapper.text()).not.toContain('Cargar más')
  })

  // --- Alta -----------------------------------------------------------

  it('opens the create dialog and blocks submission for an empty name', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    expect(openDialogInput(wrapper)).toBeTruthy()
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe el nombre del barbero.')
  })

  it('on success, prepends the confirmed barber once and closes the dialog', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Nuevo Barbero')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-new', 'Nuevo Barbero') })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('li').length).toBe(2)
    expect(wrapper.text()).toContain('Nuevo Barbero')
  })

  it('sends the same idempotency key across a submit and a network-error retry of the same logical attempt', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Reintentado')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitOpenDialog(wrapper)
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-retry', 'Reintentado') })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(2)
    const firstKey = createMock.mock.calls[0]![1]
    const secondKey = createMock.mock.calls[1]![1]
    expect(secondKey).toBe(firstKey)
  })

  it('blocks a second submit while the first is still in flight (no double POST)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Doble Envío')
    await flushPromises()

    let resolveCreate: (value: unknown) => void = () => {}
    createMock.mockReturnValueOnce(new Promise((resolve) => (resolveCreate = resolve)))
    submitOpenDialog(wrapper)
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    resolveCreate({ kind: 'success', barber: barber('b-x', 'Doble Envío') })
    await flushPromises()
  })

  it('on idempotency-conflict, shows a recoverable message and keeps the typed value', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    const input = openDialogInput(wrapper)
    setInputValue(input, 'Conflicto')
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos completar el intento anterior')
    expect(input.value).toBe('Conflicto')
  })

  it('never shows placeholders for service, schedule, active status, or linked user', async () => {
    const wrapper = await mountReady(fourBarbers)
    const text = wrapper.text().toLowerCase()
    for (const forbidden of [
      'servicio',
      'horario',
      'disponib',
      'activo',
      'inactivo',
      'usuario vinculado',
      'agenda',
    ]) {
      expect(text).not.toContain(forbidden)
    }
  })

  // --- Edición ----------------------------------------------------------

  it('opens the edit dialog prefilled with the barber name and renames on success', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    const input = openDialogInput(wrapper)
    expect(input.value).toBe('Carlos Ramírez')

    setInputValue(input, 'Carlos A. Ramírez')
    await flushPromises()

    renameMock.mockResolvedValueOnce({
      kind: 'success',
      barber: barber('b-1', 'Carlos A. Ramírez'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(renameMock).toHaveBeenCalledWith('b-1', 'Carlos A. Ramírez')
    expect(wrapper.findAll('li').length).toBe(1)
    expect(wrapper.text()).toContain('Carlos A. Ramírez')
  })

  it('replaces the item by id without duplicating or reordering unstably', async () => {
    const wrapper = await mountReady(fourBarbers)
    const editButtons = wrapper.findAll('button').filter((b) => b.text() === 'Editar')
    await editButtons[2]!.trigger('click') // "Luis Gómez", third item
    await flushPromises()

    renameMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-3', 'Luis A. Gómez') })
    submitOpenDialog(wrapper)
    await flushPromises()

    const names = wrapper.findAll('.staff-page__item-name').map((n) => n.text())
    expect(names).toEqual(['Carlos Ramírez', 'Ana Torres', 'Luis A. Gómez', 'María Pérez'])
  })

  it('on not-found, shows a recoverable message', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    renameMock.mockResolvedValueOnce({ kind: 'not-found' })
    submitOpenDialog(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('ya no está disponible')
  })

  // --- Accesibilidad ------------------------------------------------------

  it('has no axe violations in the list view', async () => {
    const wrapper = await mountReady(fourBarbers)
    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })

  it('has no axe violations with the create dialog open', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
