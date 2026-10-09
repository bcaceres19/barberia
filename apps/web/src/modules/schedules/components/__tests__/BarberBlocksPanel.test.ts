/**
 * Pruebas de BarberBlocksPanel centradas en el feedback de cada cambio (DEC-095):
 * una confirmación emergente por alta o retiro, y un aviso de error con
 * "Reintentar" cuando un retiro falla, porque no hay formulario donde
 * mostrarlo. Los datos de la API se sustituyen por dobles controlables.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { BaseInput, BaseDatePicker, BaseTimePicker, BaseSelect } from '@/shared/ui'
import { toastState } from '@/shared/model/toastStore'
import { DEFAULT_BRAND, resetBrand, setBrand } from '@/shared/model'
import { displayCivilDate } from '../../model/displayDate'

const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const fetchTimeBlocksMock = vi.hoisted(() => vi.fn())
const fetchTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const createTimeBlockMock = vi.hoisted(() => vi.fn())
const createTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const deleteTimeBlockMock = vi.hoisted(() => vi.fn())
const deleteTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const createDateListSeriesMock = vi.hoisted(() => vi.fn())
const updateTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const addSeriesExceptionMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/schedulesApi', () => ({
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
}))
vi.mock('../../api/timeBlocksApi', () => ({
  fetchTimeBlocks: fetchTimeBlocksMock,
  fetchTimeBlockSeries: fetchTimeBlockSeriesMock,
  createTimeBlock: createTimeBlockMock,
  createTimeBlockSeries: createTimeBlockSeriesMock,
  deleteTimeBlock: deleteTimeBlockMock,
  deleteTimeBlockSeries: deleteTimeBlockSeriesMock,
  createDateListSeries: createDateListSeriesMock,
  updateTimeBlockSeries: updateTimeBlockSeriesMock,
  addSeriesException: addSeriesExceptionMock,
}))

const { default: BarberBlocksPanel } = await import('../../components/BarberBlocksPanel.vue')

const block = {
  id: 'blk-1',
  blockType: 'lunch' as const,
  source: 'manual' as const,
  startsAt: '2099-01-05T17:00:00Z',
  endsAt: '2099-01-05T18:00:00Z',
  reason: null,
  deletedAt: null,
  deletedBy: null,
  createdAt: '2026-09-01T12:00:00Z',
  updatedAt: '2026-09-01T12:00:00Z',
}

const weeklySeries = {
  id: 'ser-1',
  blockType: 'lunch' as const,
  recurrenceKind: 'weekly' as const,
  isoWeekday: 1,
  startsTime: '12:00',
  durationMinutes: 60,
  effectiveFrom: '2026-09-01',
  effectiveUntil: null,
  reason: null,
  deletedAt: null,
  deletedBy: null,
  createdAt: '2026-09-01T12:00:00Z',
  updatedAt: '2026-09-01T12:00:00Z',
  // El contrato siempre responde ambas colecciones, vacías si no hay ninguna.
  dates: [],
  exceptions: [],
}

async function mountReady() {
  fetchBarbershopTimezoneMock.mockResolvedValueOnce({
    kind: 'success',
    timezone: 'America/Bogota',
  })
  fetchTimeBlocksMock.mockResolvedValueOnce({
    kind: 'success',
    page: { items: [block], nextCursor: null },
  })
  fetchTimeBlockSeriesMock.mockResolvedValueOnce({
    kind: 'success',
    page: { items: [weeklySeries], nextCursor: null },
  })
  const wrapper = mount(BarberBlocksPanel, {
    props: { barberId: 'b-1', barberName: 'Barbero de prueba' },
    global: { stubs: { teleport: true, RouterLink: true } },
  })
  await flushPromises()
  return wrapper
}

// Los dos "Retirar" de la pantalla: el primero es el bloqueo puntual y el
// segundo la serie semanal.
function retireButtons(wrapper: VueWrapper) {
  return wrapper.findAll('button').filter((button) => button.text() === 'Retirar')
}

describe('BarberBlocksPanel · avisos emergentes (DEC-095)', () => {
  beforeEach(() => {
    for (const mock of [
      fetchBarbershopTimezoneMock,
      fetchTimeBlocksMock,
      fetchTimeBlockSeriesMock,
      createTimeBlockMock,
      createTimeBlockSeriesMock,
      deleteTimeBlockMock,
      deleteTimeBlockSeriesMock,
      createDateListSeriesMock,
      updateTimeBlockSeriesMock,
      addSeriesExceptionMock,
    ]) {
      mock.mockReset()
    }
  })

  it('confirms a created block with a success toast and lists it', async () => {
    const wrapper = await mountReady()
    createTimeBlockMock.mockResolvedValueOnce({
      kind: 'success',
      block: { ...block, id: 'blk-2', blockType: 'emergency' },
    })

    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Agregar bloqueo')!
      .trigger('click')
    await flushPromises()
    const form = wrapper.get('form.blocks-page__form')
    const dates = form.findAllComponents(BaseDatePicker)
    const times = form.findAllComponents(BaseTimePicker)
    const inputs = [dates[0]!, times[0]!, dates[1]!, times[1]!, form.getComponent(BaseInput)]
    inputs[0]!.vm.$emit('update:modelValue', '2099-01-06')
    inputs[1]!.vm.$emit('update:modelValue', '09:00')
    inputs[2]!.vm.$emit('update:modelValue', '2099-01-06')
    inputs[3]!.vm.$emit('update:modelValue', '10:00')
    await flushPromises()
    await form.trigger('submit')
    await flushPromises()

    expect(createTimeBlockMock).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('Emergencia')
    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['success', 'Bloqueo agregado'],
    ])
  })

  it('confirms a retired block with a success toast', async () => {
    const wrapper = await mountReady()
    deleteTimeBlockMock.mockResolvedValueOnce({ kind: 'success' })

    await retireButtons(wrapper)[0]!.trigger('click')
    await flushPromises()

    expect(deleteTimeBlockMock).toHaveBeenCalledWith('b-1', 'blk-1')
    expect(toastState.items.map((item) => item.title)).toEqual(['Bloqueo retirado'])
  })

  it('confirms a retired series with a success toast', async () => {
    const wrapper = await mountReady()
    deleteTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'success' })

    await retireButtons(wrapper)[1]!.trigger('click')
    await flushPromises()

    expect(deleteTimeBlockSeriesMock).toHaveBeenCalledWith('b-1', 'ser-1')
    expect(toastState.items.map((item) => item.title)).toEqual(['Serie retirada'])
  })

  it('treats an already-gone block as retired, with no toast and no error', async () => {
    const wrapper = await mountReady()
    deleteTimeBlockMock.mockResolvedValueOnce({ kind: 'not-found' })

    await retireButtons(wrapper)[0]!.trigger('click')
    await flushPromises()

    expect(toastState.items).toHaveLength(0)
    expect(wrapper.text()).toContain('Sin bloqueos puntuales')
  })

  it('raises an error toast with a retry when retiring a block fails, keeping the row', async () => {
    const wrapper = await mountReady()
    deleteTimeBlockMock.mockResolvedValueOnce({ kind: 'network-error' })

    await retireButtons(wrapper)[0]!.trigger('click')
    await flushPromises()

    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['danger', 'No pudimos retirar el bloqueo'],
    ])
    expect(retireButtons(wrapper)).toHaveLength(2)

    // "Reintentar" repite el retiro; si ahora sale bien, llega la confirmación.
    deleteTimeBlockMock.mockResolvedValueOnce({ kind: 'success' })
    toastState.items[0]!.action!.run()
    await flushPromises()

    expect(deleteTimeBlockMock).toHaveBeenCalledTimes(2)
    expect(toastState.items.map((item) => item.title)).toContain('Bloqueo retirado')
  })

  it('raises an error toast with a retry when retiring a series fails', async () => {
    const wrapper = await mountReady()
    deleteTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'unexpected-error' })

    await retireButtons(wrapper)[1]!.trigger('click')
    await flushPromises()

    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['danger', 'No pudimos retirar la serie'],
    ])
    expect(toastState.items[0]!.action?.label).toBe('Reintentar')
  })

  it('shows the seven approved block types and validates a reversed interval without a POST', async () => {
    const wrapper = await mountReady()
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Agregar bloqueo')!
      .trigger('click')
    await flushPromises()
    const form = wrapper.findAll('form.blocks-page__form')[0]!
    await form.get('#create-block-type').trigger('click')
    expect(form.findAll('[role="option"]')).toHaveLength(7)
    await form.get('[role="listbox"]').trigger('keydown', { key: 'Escape' })
    const dates = form.findAllComponents(BaseDatePicker)
    const times = form.findAllComponents(BaseTimePicker)
    const inputs = [dates[0]!, times[0]!, dates[1]!, times[1]!, form.getComponent(BaseInput)]
    for (const [index, value] of ['2099-01-06', '10:00', '2099-01-06', '09:00'].entries())
      inputs[index]!.vm.$emit('update:modelValue', value)
    await flushPromises()
    await form.trigger('submit')
    expect(createTimeBlockMock).not.toHaveBeenCalled()
    expect(wrapper.get('form.blocks-page__form').find('[role="alert"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('retains the draft and reuses the idempotency key after a network error', async () => {
    const wrapper = await mountReady()
    createTimeBlockMock.mockResolvedValue({ kind: 'network-error' })
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Agregar bloqueo')!
      .trigger('click')
    await flushPromises()
    const form = wrapper.findAll('form.blocks-page__form')[0]!
    const dates = form.findAllComponents(BaseDatePicker)
    const times = form.findAllComponents(BaseTimePicker)
    const inputs = [dates[0]!, times[0]!, dates[1]!, times[1]!, form.getComponent(BaseInput)]
    for (const [index, value] of [
      '2099-01-06',
      '09:00',
      '2099-01-06',
      '10:00',
      'Pausa de prueba',
    ].entries())
      inputs[index]!.vm.$emit('update:modelValue', value)
    await flushPromises()
    await form.trigger('submit')
    await flushPromises()
    expect(wrapper.get('form.blocks-page__form').text()).toContain('Tus datos siguen aquí')
    expect(wrapper.get('form.blocks-page__form').getComponent(BaseInput).props('modelValue')).toBe(
      'Pausa de prueba',
    )
    await form.trigger('submit')
    await flushPromises()
    expect(createTimeBlockMock.mock.calls[0]).toEqual(createTimeBlockMock.mock.calls[1])
    expect(createTimeBlockMock.mock.calls[0]!.slice(0, 5)).toEqual([
      'b-1',
      'emergency',
      '2099-01-06T14:00:00.000Z',
      '2099-01-06T15:00:00.000Z',
      'Pausa de prueba',
    ])
    wrapper.unmount()
  })

  it('creates a weekly series for the context barber and confirms it in the list', async () => {
    const wrapper = await mountReady()
    createTimeBlockSeriesMock.mockResolvedValueOnce({
      kind: 'success',
      series: { ...weeklySeries, id: 'ser-2' },
    })
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Agregar serie semanal')!
      .trigger('click')
    await flushPromises()
    const form = wrapper.findAll('form.blocks-page__form')[1]!
    form.getComponent(BaseTimePicker).vm.$emit('update:modelValue', '12:00')
    form.findAllComponents(BaseInput)[0]!.vm.$emit('update:modelValue', 45)
    form.getComponent(BaseDatePicker).vm.$emit('update:modelValue', '2099-01-01')
    await flushPromises()
    await form.trigger('submit')
    await flushPromises()
    expect(createTimeBlockSeriesMock.mock.calls[0]!.slice(0, 7)).toEqual([
      'b-1',
      'lunch',
      1,
      '12:00',
      45,
      '2099-01-01',
      null,
    ])
    expect(wrapper.findAll('.blocks-panel__record')).toHaveLength(3)
    expect(toastState.items.map((item) => item.title)).toEqual(['Serie agregada'])
    wrapper.unmount()
  })

  it('disables creation when the timezone is unavailable and allows a safe retry', async () => {
    fetchBarbershopTimezoneMock.mockResolvedValueOnce({ kind: 'unavailable' })
    const wrapper = await mountReady()
    const create = wrapper.findAll('button').find((button) => button.text() === 'Agregar bloqueo')!
    expect(create.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('No pudimos consultar la zona horaria')
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Reintentar zona horaria')!
      .trigger('click')
    await flushPromises()
    expect(create.attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('keeps weekly series visible when loading blocks fails, then retries', async () => {
    fetchTimeBlocksMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = await mountReady()
    expect(wrapper.text()).toContain('No pudimos cargar los bloqueos')
    expect(wrapper.text()).toContain('Se repite cada semana')
    fetchTimeBlockSeriesMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [weeklySeries], nextCursor: null },
    })
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Reintentar bloqueos')!
      .trigger('click')
    await flushPromises()
    expect(wrapper.text()).not.toContain('No pudimos cargar los bloqueos')
    expect(wrapper.findAll('.blocks-panel__record')).toHaveLength(2)
    wrapper.unmount()
  })

  it('loads the next page of blocks without duplicating a known record', async () => {
    fetchTimeBlocksMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [block], nextCursor: 'blocks-next' },
    })
    const wrapper = await mountReady()
    fetchTimeBlocksMock.mockReset()
    fetchTimeBlocksMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [block, { ...block, id: 'blk-2' }], nextCursor: null },
    })
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Cargar más bloqueos')!
      .trigger('click')
    await flushPromises()
    expect(fetchTimeBlocksMock).toHaveBeenCalledWith('b-1', 'blocks-next')
    expect(wrapper.findAll('.blocks-panel__record')).toHaveLength(3)
    wrapper.unmount()
  })

  it('renders a removed block only as absence and identifies a past interval', async () => {
    fetchTimeBlocksMock.mockResolvedValueOnce({
      kind: 'success',
      page: {
        items: [
          { ...block, id: 'removed', deletedAt: '2026-09-01T12:00:00Z' },
          { ...block, startsAt: '2020-01-01T12:00:00Z', endsAt: '2020-01-01T13:00:00Z' },
        ],
        nextCursor: null,
      },
    })
    const wrapper = await mountReady()
    expect(wrapper.findAll('.blocks-panel__record--past')).toHaveLength(1)
    expect(wrapper.findAll('.blocks-panel__record')).toHaveLength(2)
    wrapper.unmount()
  })

  it('does not reinsert a retired block when a previous page request arrives late', async () => {
    fetchTimeBlocksMock.mockResolvedValueOnce({
      kind: 'success',
      page: { items: [block], nextCursor: 'next' },
    })
    const wrapper = await mountReady()
    let release!: (value: {
      kind: 'success'
      page: { items: (typeof block)[]; nextCursor: null }
    }) => void
    fetchTimeBlocksMock.mockReset()
    fetchTimeBlocksMock.mockReturnValueOnce(
      new Promise((resolve) => {
        release = resolve
      }),
    )
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Cargar más bloqueos')!
      .trigger('click')
    deleteTimeBlockMock.mockResolvedValueOnce({ kind: 'success' })
    await retireButtons(wrapper)[0]!.trigger('click')
    await flushPromises()
    release({ kind: 'success', page: { items: [block], nextCursor: null } })
    await flushPromises()
    expect(wrapper.text()).toContain('Sin bloqueos puntuales')
    expect(wrapper.findAll('.blocks-panel__record')).toHaveLength(1)
    wrapper.unmount()
  })
})

describe('BarberBlocksPanel · vocabulario del negocio (DEC-119)', () => {
  beforeEach(() => {
    for (const mock of [
      fetchBarbershopTimezoneMock,
      fetchTimeBlocksMock,
      fetchTimeBlockSeriesMock,
      createTimeBlockMock,
    ]) {
      mock.mockReset()
    }
    resetBrand()
  })

  it('names the business and the professional with the configured words', async () => {
    setBrand({
      ...DEFAULT_BRAND,
      businessTerm: 'estudio',
      businessTermGender: 'masculine',
      professionalTerm: 'manicurista',
      professionalTermPlural: 'manicuristas',
      professionalTermGender: 'feminine',
    })
    const wrapper = await mountReady()

    expect(wrapper.text()).toContain('Horas del estudio · America/Bogota')
    expect(wrapper.text()).not.toContain('de la barbería')
    resetBrand()
    wrapper.unmount()
  })

  it('keeps the original words by default', async () => {
    const wrapper = await mountReady()

    expect(wrapper.text()).toContain('Horas de la barbería · America/Bogota')
    wrapper.unmount()
  })
})

describe('BarberBlocksPanel · series por fechas, edición y excepciones (#100)', () => {
  beforeEach(() => {
    for (const mock of [
      fetchBarbershopTimezoneMock,
      fetchTimeBlocksMock,
      fetchTimeBlockSeriesMock,
      createDateListSeriesMock,
      updateTimeBlockSeriesMock,
      addSeriesExceptionMock,
    ]) {
      mock.mockReset()
    }
  })

  function named(wrapper: VueWrapper, label: string) {
    const found = wrapper.findAll('button').find((b) => b.attributes('aria-label') === label)
    if (!found) throw new Error(`Botón no encontrado: ${label}`)
    return found
  }
  // El contenido de un diálogo se vuelve a montar al abrirse: cada paso lo
  // resuelve de nuevo en lugar de conservar wrappers que pueden quedar viejos.
  function dialogOf(wrapper: VueWrapper, title: string) {
    const found = wrapper.findAll('[role="dialog"]').find((d) => d.text().includes(title))
    if (!found) throw new Error(`Diálogo no encontrado: ${title}`)
    return found
  }
  async function pickDate(wrapper: VueWrapper, title: string, index: number, value: string) {
    const picker = dialogOf(wrapper, title).findAllComponents(BaseDatePicker).at(index)!
    picker.vm.$emit('update:modelValue', value)
    await flushPromises()
  }
  async function press(wrapper: VueWrapper, title: string, label: string) {
    const target = dialogOf(wrapper, title)
      .findAll('button')
      .find((b) => b.text() === label)!
    await target.trigger('click')
    await flushPromises()
  }
  async function submit(wrapper: VueWrapper, title: string) {
    await dialogOf(wrapper, title).get('form').trigger('submit')
    await flushPromises()
  }
  const seriesRecords = (wrapper: VueWrapper) =>
    wrapper.findAll('section[aria-labelledby="weekly-blocks-title"] .blocks-panel__record')

  it('creates a block by dates and lists it with its date count', async () => {
    const wrapper = await mountReady()
    createDateListSeriesMock.mockResolvedValueOnce({
      kind: 'success',
      series: {
        ...weeklySeries,
        id: 'ser-d',
        blockType: 'vacation',
        recurrenceKind: 'date_list',
        isoWeekday: null,
        startsTime: '00:00',
        durationMinutes: 1440,
        dates: [{ blockDate: '2099-12-15' }, { blockDate: '2099-12-16' }],
      },
    })
    const title = 'Agregar bloqueo por fechas'

    await wrapper
      .findAll('button')
      .find((b) => b.text() === title)!
      .trigger('click')
    await flushPromises()
    await pickDate(wrapper, title, 0, '2099-12-01')
    for (const date of ['2099-12-15', '2099-12-16']) {
      await pickDate(wrapper, title, 1, date)
      await press(wrapper, title, 'Añadir a la lista')
    }
    await submit(wrapper, title)

    expect(createDateListSeriesMock).toHaveBeenCalledOnce()
    const records = seriesRecords(wrapper)
    expect(records).toHaveLength(2)
    expect(records[1]!.text()).toContain('Vacaciones')
    expect(records[1]!.text()).toContain('2 fechas')
  })

  it('edits the whole series in place', async () => {
    const wrapper = await mountReady()
    updateTimeBlockSeriesMock.mockResolvedValueOnce({
      kind: 'success',
      series: { ...weeklySeries, startsTime: '13:30', durationMinutes: 45 },
    })

    await named(wrapper, 'Editar serie Almuerzo Lunes').trigger('click')
    await flushPromises()
    await submit(wrapper, 'Editar serie de bloqueo')

    const records = seriesRecords(wrapper)
    expect(records).toHaveLength(1)
    expect(records[0]!.text()).toContain('13:30')
  })

  it('shows both halves of a split series: the original cut the day before and the new one', async () => {
    const wrapper = await mountReady()
    updateTimeBlockSeriesMock.mockResolvedValueOnce({
      kind: 'success',
      series: { ...weeklySeries, id: 'ser-new', startsTime: '14:00', effectiveFrom: '2026-12-01' },
    })
    const title = 'Editar serie de bloqueo'

    await named(wrapper, 'Editar serie Almuerzo Lunes').trigger('click')
    await flushPromises()
    const scope = dialogOf(wrapper, title).findAllComponents(BaseSelect)[1]
    ;(scope as unknown as VueWrapper).vm.$emit('update:modelValue', 'this_and_following')
    await flushPromises()
    await pickDate(wrapper, title, 0, '2026-12-01')
    await submit(wrapper, title)

    const records = seriesRecords(wrapper)
    expect(records).toHaveLength(2)
    expect(records[0]!.text()).toContain(displayCivilDate('2026-11-30'))
    expect(records[1]!.text()).toContain('14:00')
  })

  it('adds an exception from its dialog and counts it on the record', async () => {
    const wrapper = await mountReady()
    addSeriesExceptionMock.mockResolvedValueOnce({ kind: 'success' })
    const title = 'Excepciones de la serie'

    await named(wrapper, 'Excepciones de la serie Almuerzo Lunes').trigger('click')
    await flushPromises()
    await pickDate(wrapper, title, 0, '2099-01-12')
    await press(wrapper, title, 'Agregar excepción')

    expect(addSeriesExceptionMock).toHaveBeenCalledWith('b-1', 'ser-1', '2099-01-12', null)
    expect(seriesRecords(wrapper)[0]!.text()).toContain('1 excepción')
  })
})
