/**
 * Pruebas de StaffPage (HU-021, DEC-104): carga (una fila, cuatro filas, vacío),
 * error recuperable, alta (éxito, validación, conflicto de idempotencia, error
 * de red, doble envío bloqueado, con y sin foto), edición (nombre, foto, quitar
 * foto, fallos a medias, sin cambios, no encontrado), detalle, paginación
 * (numerada, DEC-107), foco y ausencia de campos fuera de alcance. staffApi se
 * sustituye por un doble de prueba; el recorrido real contra el API vive en el
 * E2E de HU-021.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { DEFAULT_MIN_HOLD_MS, PAGE_MIN_HOLD_MS } from '@/shared/composables'

const fetchMock = vi.hoisted(() => vi.fn())
const createMock = vi.hoisted(() => vi.fn())
const renameMock = vi.hoisted(() => vi.fn())
const uploadPhotoMock = vi.hoisted(() => vi.fn())
const removePhotoMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/staffApi', () => ({
  fetchBarberPage: async (page: number, pageSize: number) => {
    const outcome = await fetchMock(page, pageSize)
    if (outcome.kind !== 'success') return outcome
    return {
      kind: 'success',
      page: {
        page,
        pageSize,
        total: outcome.page.items.length + (outcome.page.nextCursor ? 1 : 0),
        totalPages: outcome.page.nextCursor ? 2 : 1,
        ...outcome.page,
      },
    }
  },
  createBarber: createMock,
  renameBarber: renameMock,
  uploadBarberPhoto: uploadPhotoMock,
  removeBarberPhoto: removePhotoMock,
  // La URL real depende del cliente HTTP; aquí basta una forma reconocible.
  barberPhotoUrl: (b: { id: string; photoUpdatedAt: string | null }) =>
    b.photoUpdatedAt ? `/photo/${b.id}?v=${b.photoUpdatedAt}` : null,
}))

const { default: StaffPage } = await import('../StaffPage.vue')
const { default: BarberPhotoField } = await import('../../components/BarberPhotoField.vue')

function barber(id: string, fullName: string, photoUpdatedAt: string | null = null) {
  return {
    id,
    fullName,
    createdAt: '2026-08-23T15:04:05Z',
    updatedAt: '2026-08-23T15:04:05Z',
    photoUpdatedAt,
  }
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
// la raíz). stubs.transition: jsdom no tiene motor CSS (nunca dispara
// transitionend), así que el <Transition mode="out-in"> de carga/error/listo
// cambia de hijo al instante.
function mountPage() {
  return mount(StaffPage, { global: { stubs: { teleport: true, transition: true } } })
}

// useMinHoldLoading mantiene el rombo/esqueleto al menos DEFAULT_MIN_HOLD_MS con
// un setTimeout REAL antes de dejar pasar el contenido: flushPromises() (solo
// microtareas) no alcanza a esperarlo. Con margen sobre el valor exacto.
function waitOutInitialLoadHold() {
  return new Promise((resolve) => setTimeout(resolve, DEFAULT_MIN_HOLD_MS + 50))
}

// "Cargar más" mantiene los esqueletos al menos PAGE_MIN_HOLD_MS.
function waitOutMoreHold() {
  return new Promise((resolve) => setTimeout(resolve, PAGE_MIN_HOLD_MS + 50))
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
  await waitOutInitialLoadHold()
  await flushPromises()
  return wrapper
}

// La pila de BaseDialog mantiene los diálogos (detalle, alta y edición) SIEMPRE
// en el DOM (v-show, no v-if): se selecciona el que esté visualmente abierto en
// vez de "form" a secas, que ambigüaría entre ellos.
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

// Elige una foto en el campo del diálogo abierto. El campo emite `update:draft`
// con la imagen ya preparada (el recorte y el input de archivo se prueban en
// BarberPhotoField.test.ts y preparePhoto.test.ts): aquí se ejercita lo que la
// página hace con ese borrador. Se emite desde el componente en vez de
// disparar `change` en el DOM porque el stub de Teleport de Vue Test Utils
// vuelve a montar el contenido de los diálogos en cada re-render, y una
// preparación asíncrona en curso emitiría desde una instancia ya descartada.
async function pickPhoto(wrapper: VueWrapper, previewUrl = 'blob:preview') {
  const blob = new Blob(['prepared'], { type: 'image/jpeg' })
  const open = openDialogElement(wrapper)
  const field = wrapper
    .findAllComponents(BarberPhotoField)
    .find((candidate) => open.contains(candidate.element))
  if (!field) throw new Error('no photo field in the open dialog')
  field.vm.$emit('update:draft', { kind: 'new', blob, previewUrl })
  await flushPromises()
  return blob
}

function dialogButton(wrapper: VueWrapper, text: string) {
  const button = Array.from(openDialogElement(wrapper).querySelectorAll('button')).find(
    (b) => b.textContent?.trim() === text,
  )
  if (!button) throw new Error(`dialog button ${JSON.stringify(text)} not found`)
  return button as HTMLButtonElement
}

describe('StaffPage', () => {
  beforeEach(() => {
    resetToasts()
    fetchMock.mockReset()
    fetchMock.mockImplementation(async () => {
      const created = await createMock.mock.results.at(-1)?.value
      const photo = await uploadPhotoMock.mock.results.at(-1)?.value
      const added = photo?.kind === 'success' ? photo.barber : created?.barber
      return {
        kind: 'success',
        page: { items: added ? [...oneBarber, added] : [...oneBarber], nextCursor: null },
      }
    })
    createMock.mockReset()
    renameMock.mockReset()
    uploadPhotoMock.mockReset()
    removePhotoMock.mockReset()
    URL.revokeObjectURL = vi.fn()
  })

  it('shows a non-blank loading state, then the loaded team', async () => {
    let resolveFetch: (value: unknown) => void = () => {}
    fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
    const wrapper = mountPage()

    expect(wrapper.text()).toContain('Cargando')
    // La carga muestra el rombo y la forma de las filas por llegar, no una
    // pantalla en blanco.
    expect(wrapper.find('.diamond-loader').exists()).toBe(true)
    expect(wrapper.findAll('.staff-page__skeleton-row').length).toBe(3)

    resolveFetch({ kind: 'success', page: { items: [...oneBarber], nextCursor: null } })
    await flushPromises()
    // Una respuesta rápida no retira el indicador antes del mínimo.
    expect(wrapper.text()).toContain('Cargando')

    await waitOutInitialLoadHold()
    await flushPromises()

    expect(wrapper.text()).not.toContain('Cargando')
    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('a one-person shop and a four-person shop use the same list component (CA-021-01/02)', async () => {
    const oneWrapper = await mountReady(oneBarber)
    expect(oneWrapper.findAll('.staff-page__row').length).toBe(1)

    const fourWrapper = await mountReady(fourBarbers)
    expect(fourWrapper.findAll('.staff-page__row').length).toBe(4)
    expect(fourWrapper.text()).toContain('Ana Torres')
    expect(fourWrapper.text()).toContain('María Pérez')
  })

  it('shows a decorative, accessible monogram with the initials of each barber without a photo (Fase 5, issue #153)', async () => {
    const wrapper = await mountReady(fourBarbers)
    const avatars = wrapper.findAll('.staff-page__row .barber-avatar')
    expect(avatars.map((a) => a.text())).toEqual(['CR', 'AT', 'LG', 'MP'])
    for (const avatar of avatars) {
      expect(avatar.attributes('aria-hidden')).toBe('true')
    }
  })

  it('uses a single initial for a one-word name', async () => {
    const wrapper = await mountReady([barber('b-1', 'Madonna')])
    expect(wrapper.get('.staff-page__row .barber-avatar').text()).toBe('M')
  })

  it('shows the real photo of a barber who has one and the monogram of one who does not (DEC-104)', async () => {
    const wrapper = await mountReady([
      barber('b-1', 'Carlos Ramírez', '2026-09-30T12:00:00Z'),
      barber('b-2', 'Ana Torres'),
    ])
    const rows = wrapper.findAll('.staff-page__row')

    const photo = rows[0]!.get('img.barber-avatar__photo')
    expect(photo.attributes('src')).toBe('/photo/b-1?v=2026-09-30T12:00:00Z')
    expect(photo.attributes('alt')).toBe('')
    expect(rows[0]!.text()).toContain('Con foto')

    expect(rows[1]!.find('img').exists()).toBe(false)
    expect(rows[1]!.get('.barber-avatar__monogram').text()).toBe('AT')
    expect(rows[1]!.text()).toContain('Sin foto')
  })

  it('shows the enrolment date of each barber in the shop timezone', async () => {
    const wrapper = await mountReady()
    expect(wrapper.get('.staff-page__item-since').text()).toBe('23/08/2026')
  })

  it('shows an empty state with no special data shape when there are no barbers', async () => {
    const wrapper = await mountReady([])
    expect(wrapper.text()).toContain('Aún no tienes barberos registrados')
    expect(wrapper.find('.staff-page__ready--fitting').exists()).toBe(false)
    expect(wrapper.findAll('li').length).toBe(0)
    expect(wrapper.find('.staff-page__footer').exists()).toBe(false)
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountPage()
    await flushPromises()
    await waitOutInitialLoadHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar el equipo')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [...oneBarber], nextCursor: null },
    })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    await waitOutInitialLoadHold()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Carlos Ramírez')
  })

  it('replaces the page, shows exact total and returns to the previous page', async () => {
    const wrapper = await mountReady(oneBarber, 'opaque-cursor')
    expect(wrapper.find('nav[aria-label="Paginación de barberos"]').exists()).toBe(true)
    expect(wrapper.get('.staff-page__count').text()).toBe('2 barberos')
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: {
        items: [barber('b-2', 'Ana Torres')],
        page: 2,
        pageSize: 20,
        total: 2,
        totalPages: 2,
      },
    })
    await wrapper.get('button[aria-label="Página siguiente"]').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.staff-page__skeleton-row--page').length).toBe(1)
    await waitOutMoreHold()
    await flushPromises()
    expect(wrapper.findAll('.staff-page__row').length).toBe(1)
    expect(wrapper.text()).toContain('Ana Torres')
    expect(wrapper.text()).not.toContain('Carlos Ramírez')
    expect(wrapper.get('button[aria-label="Página 2"]').attributes('aria-current')).toBe('page')
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: oneBarber, page: 1, pageSize: 20, total: 2, totalPages: 2 },
    })
    await wrapper.get('button[aria-label="Página anterior"]').trigger('click')
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()
    expect(wrapper.text()).toContain('Carlos Ramírez')
    expect(fetchMock).toHaveBeenLastCalledWith(1, 20)
  })

  it('keeps the loaded page on error and retries the failed target page', async () => {
    const wrapper = await mountReady(oneBarber, 'cursor')
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    await wrapper.get('button[aria-label="Página siguiente"]').trigger('click')
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos cargar esta página')
    expect(wrapper.text()).toContain('Carlos Ramírez')
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      page: {
        items: [barber('b-2', 'Ana Torres')],
        page: 2,
        pageSize: 20,
        total: 2,
        totalPages: 2,
      },
    })
    await findButtonByText(wrapper, 'Reintentar página').trigger('click')
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()
    expect(fetchMock).toHaveBeenLastCalledWith(2, 20)
    expect(wrapper.text()).toContain('Ana Torres')
  })

  it('fits the initial page and ignores an old page response after a smaller resize (#288)', async () => {
    const scroller = document.createElement('div')
    scroller.style.overflowY = 'auto'
    let height = 600
    Object.defineProperty(scroller, 'clientHeight', { get: () => height })
    document.body.appendChild(scroller)
    const style = document.createElement('style')
    style.textContent = '.staff-page { padding-bottom:48px }'
    document.head.appendChild(style)
    const rect = (top: number, height: number) => ({ top, bottom: top + height, height }) as DOMRect
    const spy = vi
      .spyOn(HTMLElement.prototype, 'getBoundingClientRect')
      .mockImplementation(function (this: HTMLElement) {
        if (this === scroller) return rect(0, height)
        if (
          this.classList.contains('staff-page__row') ||
          this.classList.contains('staff-page__skeleton-row')
        )
          return rect(0, 84)
        const rows = scroller.querySelectorAll('.staff-page__row,.staff-page__skeleton-row').length
        if (this.classList.contains('staff-page__list')) return rect(200, rows * 84 + 2)
        if (this.classList.contains('staff-page__footer')) return rect(218 + rows * 84, 34)
        return rect(0, 0)
      })
    let wrapper: VueWrapper | undefined
    try {
      fetchMock.mockResolvedValueOnce({
        kind: 'success',
        page: { items: fourBarbers, page: 1, pageSize: 20, total: 17, totalPages: 1 },
      })
      wrapper = mount(StaffPage, {
        attachTo: scroller,
        global: { stubs: { teleport: true, transition: true } },
      })
      await flushPromises()
      await waitOutInitialLoadHold()
      await flushPromises()
      expect(wrapper.findAll('.staff-page__row')).toHaveLength(3)
      expect(fetchMock).toHaveBeenCalledTimes(1)
      let resolveOld: (value: unknown) => void = () => {}
      fetchMock.mockReturnValueOnce(
        new Promise((resolve) => {
          resolveOld = resolve
        }),
      )
      await wrapper.get('button[aria-label="Página siguiente"]').trigger('click')
      await flushPromises()
      fetchMock.mockResolvedValueOnce({
        kind: 'success',
        page: { items: oneBarber, page: 1, pageSize: 1, total: 17, totalPages: 17 },
      })
      height = 420
      window.dispatchEvent(new Event('resize'))
      await new Promise((resolve) => setTimeout(resolve, 160))
      await flushPromises()
      await waitOutMoreHold()
      await flushPromises()
      expect(fetchMock).toHaveBeenLastCalledWith(1, 1)
      resolveOld({
        kind: 'success',
        page: { items: fourBarbers, page: 2, pageSize: 3, total: 17, totalPages: 6 },
      })
      await flushPromises()
      await waitOutMoreHold()
      await flushPromises()
      expect(wrapper.findAll('.staff-page__row')).toHaveLength(1)
      expect(wrapper.get('button[aria-label="Página 1"]').attributes('aria-current')).toBe('page')
    } finally {
      wrapper?.unmount()
      spy.mockRestore()
      scroller.remove()
      style.remove()
    }
  })

  // --- Detalle ------------------------------------------------------------

  it('opens the barber detail with the portrait, and its Editar button hands over to the edit dialog', async () => {
    const wrapper = await mountReady([barber('b-1', 'Carlos Ramírez', '2026-09-30T12:00:00Z')])

    await wrapper.get('.staff-page__item-trigger').trigger('click')
    await flushPromises()

    const dialog = openDialogElement(wrapper)
    expect(dialog.textContent).toContain('Carlos Ramírez')
    expect(dialog.querySelector('img.barber-avatar__photo')).toBeTruthy()
    expect(dialog.textContent).toContain('23/08/2026')
    expect(dialog.textContent).toContain('Con foto')

    dialogButton(wrapper, 'Editar').click()
    await flushPromises()

    expect(openDialogInput(wrapper).value).toBe('Carlos Ramírez')
  })

  // --- Alta -----------------------------------------------------------

  it('opens the create dialog and blocks submission for an empty name', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    expect(openDialogInput(wrapper)).toBeTruthy()
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
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
    await waitOutMoreHold()
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    expect(uploadPhotoMock).not.toHaveBeenCalled()
    expect(wrapper.findAll('.staff-page__row').length).toBe(2)
    expect(wrapper.text()).toContain('Nuevo Barbero')
    expect(toastState.items.map((item) => item.title)).toEqual(['Barbero agregado'])
    // La fila recién agregada se ilumina un instante.
    expect(wrapper.get('.staff-page__row--new').text()).toContain('Nuevo Barbero')
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
  })

  it('creates the barber and then uploads the chosen photo, showing it in the new row (DEC-104)', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Con Foto')
    const blob = await pickPhoto(wrapper)
    createMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-new', 'Con Foto') })
    uploadPhotoMock.mockResolvedValueOnce({
      kind: 'success',
      barber: barber('b-new', 'Con Foto', '2026-09-30T13:00:00Z'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(uploadPhotoMock).toHaveBeenCalledWith('b-new', blob)
    const row = wrapper.findAll('.staff-page__row').find((row) => row.text().includes('Con Foto'))!
    expect(row.get('img.barber-avatar__photo').attributes('src')).toBe(
      '/photo/b-new?v=2026-09-30T13:00:00Z',
    )
    expect(toastState.items.map((item) => item.title)).toEqual(['Barbero agregado'])
  })

  it('keeps the barber and warns when only the photo fails after a successful create', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Foto Fallida')
    await pickPhoto(wrapper)
    createMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-new', 'Foto Fallida') })
    uploadPhotoMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    // El alta ya ocurrió: el barbero está en la lista y el diálogo se cerró.
    expect(wrapper.text()).toContain('Foto Fallida')
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['warning', 'Barbero agregado, pero no pudimos guardar la foto'],
    ])
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
    await waitOutMoreHold()
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-retry', 'Reintentado') })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
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
    await waitOutMoreHold()
    await flushPromises()

    expect(createMock).toHaveBeenCalledTimes(1)
    resolveCreate({ kind: 'success', barber: barber('b-x', 'Doble Envío') })
    await flushPromises()
  })

  it('on idempotency-conflict, shows a recoverable message and keeps the typed value and the chosen photo', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Agregar barbero').trigger('click')
    await flushPromises()

    const input = openDialogInput(wrapper)
    setInputValue(input, 'Conflicto')
    await pickPhoto(wrapper)
    await flushPromises()

    createMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos completar el intento anterior')
    expect(input.value).toBe('Conflicto')
    expect(openDialogElement(wrapper).querySelector('img.barber-avatar__photo')).toBeTruthy()
    expect(uploadPhotoMock).not.toHaveBeenCalled()
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
    await waitOutMoreHold()
    await flushPromises()

    expect(renameMock).toHaveBeenCalledWith('b-1', 'Carlos A. Ramírez')
    expect(uploadPhotoMock).not.toHaveBeenCalled()
    expect(removePhotoMock).not.toHaveBeenCalled()
    expect(wrapper.findAll('.staff-page__row').length).toBe(1)
    expect(wrapper.text()).toContain('Carlos A. Ramírez')
    expect(toastState.items.map((item) => item.title)).toEqual(['Nombre actualizado'])
  })

  it('replaces the item by id without duplicating or reordering unstably', async () => {
    const wrapper = await mountReady(fourBarbers)
    const editButtons = wrapper.findAll('button').filter((b) => b.text() === 'Editar')
    await editButtons[2]!.trigger('click') // "Luis Gómez", third item
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Luis A. Gómez')
    renameMock.mockResolvedValueOnce({ kind: 'success', barber: barber('b-3', 'Luis A. Gómez') })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    const names = wrapper.findAll('.staff-page__item-name').map((n) => n.text())
    expect(names).toEqual(['Carlos Ramírez', 'Ana Torres', 'Luis A. Gómez', 'María Pérez'])
  })

  it('closes without sending anything when nothing changed', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(renameMock).not.toHaveBeenCalled()
    expect(uploadPhotoMock).not.toHaveBeenCalled()
    expect(removePhotoMock).not.toHaveBeenCalled()
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
    expect(toastState.items).toEqual([])
  })

  it('uploads a new photo without renaming when only the photo changed', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    const blob = await pickPhoto(wrapper)
    uploadPhotoMock.mockResolvedValueOnce({
      kind: 'success',
      barber: barber('b-1', 'Carlos Ramírez', '2026-09-30T13:00:00Z'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(renameMock).not.toHaveBeenCalled()
    expect(uploadPhotoMock).toHaveBeenCalledWith('b-1', blob)
    expect(wrapper.get('.staff-page__row img.barber-avatar__photo').attributes('src')).toBe(
      '/photo/b-1?v=2026-09-30T13:00:00Z',
    )
    expect(toastState.items.map((item) => item.title)).toEqual(['Foto actualizada'])
  })

  it('removes the photo and shows the monogram again', async () => {
    const wrapper = await mountReady([barber('b-1', 'Carlos Ramírez', '2026-09-30T12:00:00Z')])
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    dialogButton(wrapper, 'Quitar foto').click()
    await flushPromises()
    removePhotoMock.mockResolvedValueOnce({ kind: 'success' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(removePhotoMock).toHaveBeenCalledWith('b-1')
    expect(renameMock).not.toHaveBeenCalled()
    const row = wrapper.get('.staff-page__row')
    expect(row.find('img').exists()).toBe(false)
    expect(row.get('.barber-avatar__monogram').text()).toBe('CR')
    expect(row.text()).toContain('Sin foto')
    expect(toastState.items.map((item) => item.title)).toEqual(['Foto quitada'])
  })

  it('renames and replaces the photo in one save, in that order', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Carlos A. Ramírez')
    await pickPhoto(wrapper)
    const order: string[] = []
    renameMock.mockImplementationOnce(async () => {
      order.push('rename')
      return { kind: 'success', barber: barber('b-1', 'Carlos A. Ramírez') }
    })
    uploadPhotoMock.mockImplementationOnce(async () => {
      order.push('photo')
      return {
        kind: 'success',
        barber: barber('b-1', 'Carlos A. Ramírez', '2026-09-30T13:00:00Z'),
      }
    })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(order).toEqual(['rename', 'photo'])
    expect(toastState.items.map((item) => item.title)).toEqual(['Barbero actualizado'])
  })

  it('after a partial failure (name saved, photo failed) a retry does not rename again', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Carlos A. Ramírez')
    const blob = await pickPhoto(wrapper)
    renameMock.mockResolvedValueOnce({
      kind: 'success',
      barber: barber('b-1', 'Carlos A. Ramírez'),
    })
    uploadPhotoMock.mockResolvedValueOnce({ kind: 'validation-error' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    // El nombre ya quedó en la lista; la foto se avisa aparte y el diálogo sigue abierto.
    expect(wrapper.findAll('.staff-page__item-name').map((n) => n.text())).toEqual([
      'Carlos A. Ramírez',
    ])
    expect(wrapper.text()).toContain('No pudimos guardar la foto')
    expect(openDialogInput(wrapper).value).toBe('Carlos A. Ramírez')

    uploadPhotoMock.mockResolvedValueOnce({
      kind: 'success',
      barber: barber('b-1', 'Carlos A. Ramírez', '2026-09-30T13:00:00Z'),
    })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(renameMock).toHaveBeenCalledTimes(1)
    expect(uploadPhotoMock).toHaveBeenCalledTimes(2)
    expect(uploadPhotoMock).toHaveBeenLastCalledWith('b-1', blob)
    expect(wrapper.find('.base-dialog--open').exists()).toBe(false)
  })

  it('on a network error while saving the photo, keeps the dialog open with the chosen photo', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    await pickPhoto(wrapper)
    uploadPhotoMock.mockResolvedValueOnce({ kind: 'network-error' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(openDialogElement(wrapper).querySelector('img.barber-avatar__photo')).toBeTruthy()
  })

  it('on not-found, shows a recoverable message', async () => {
    const wrapper = await mountReady()
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    setInputValue(openDialogInput(wrapper), 'Otro Nombre')
    renameMock.mockResolvedValueOnce({ kind: 'not-found' })
    submitOpenDialog(wrapper)
    await flushPromises()
    await waitOutMoreHold()
    await flushPromises()

    expect(wrapper.text()).toContain('ya no está disponible')
  })

  // --- Accesibilidad ------------------------------------------------------

  it('has no axe violations in the list view', async () => {
    const wrapper = await mountReady(fourBarbers)
    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })

  it('has no axe violations in the empty state', async () => {
    const wrapper = await mountReady([])
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

  it('has no axe violations with the edit dialog open on a barber with a photo', async () => {
    const wrapper = await mountReady([barber('b-1', 'Carlos Ramírez', '2026-09-30T12:00:00Z')])
    await findButtonByText(wrapper, 'Editar').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
