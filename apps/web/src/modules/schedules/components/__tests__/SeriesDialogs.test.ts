/**
 * Pruebas de los diálogos de series de bloqueo de HU-042 (#100): alta por
 * fechas explícitas, edición por alcance y fechas/excepciones. La API se
 * sustituye por dobles controlables; el recorrido real vive en el E2E de
 * bloqueos dentro de Barberos.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { BaseDatePicker, BaseInput, BaseSelect } from '@/shared/ui'
import { toastState } from '@/shared/model/toastStore'
import { displayCivilDate as d } from '../../model/displayDate'
import type { TimeBlockSeries } from '../../model/timeBlock'

const createDateListSeriesMock = vi.hoisted(() => vi.fn())
const updateTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const addSeriesDateMock = vi.hoisted(() => vi.fn())
const removeSeriesDateMock = vi.hoisted(() => vi.fn())
const addSeriesExceptionMock = vi.hoisted(() => vi.fn())
const removeSeriesExceptionMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/timeBlocksApi', () => ({
  createDateListSeries: createDateListSeriesMock,
  updateTimeBlockSeries: updateTimeBlockSeriesMock,
  addSeriesDate: addSeriesDateMock,
  removeSeriesDate: removeSeriesDateMock,
  addSeriesException: addSeriesExceptionMock,
  removeSeriesException: removeSeriesExceptionMock,
}))

const { default: SeriesFormDialog } = await import('../SeriesFormDialog.vue')
const { default: SeriesInstancesDialog } = await import('../SeriesInstancesDialog.vue')

const weekly: TimeBlockSeries = {
  id: 'ser-w',
  blockType: 'lunch',
  recurrenceKind: 'weekly',
  isoWeekday: 1,
  startsTime: '13:00',
  durationMinutes: 60,
  effectiveFrom: '2099-01-05',
  effectiveUntil: null,
  reason: null,
  deletedAt: null,
  createdAt: '2026-10-01T12:00:00Z',
  updatedAt: '2026-10-01T12:00:00Z',
  dates: [],
  exceptions: [],
}
const dateList: TimeBlockSeries = {
  ...weekly,
  id: 'ser-d',
  blockType: 'vacation',
  recurrenceKind: 'date_list',
  isoWeekday: null,
  startsTime: '00:00',
  durationMinutes: 1440,
  effectiveFrom: '2099-12-01',
  effectiveUntil: '2099-12-31',
  dates: [{ blockDate: '2099-12-16' }, { blockDate: '2099-12-15' }],
  exceptions: [{ excludedDate: '2099-12-20', reason: 'Libró', createdAt: '2026-10-01T12:00:00Z' }],
}

const global = { stubs: { teleport: true } }

function button(wrapper: VueWrapper, name: string | RegExp) {
  const found = wrapper
    .findAll('button')
    .find((item) => (typeof name === 'string' ? item.text() === name : name.test(item.text())))
  if (!found) throw new Error(`Botón no encontrado: ${String(name)}`)
  return found
}
function byLabel(wrapper: VueWrapper, label: string) {
  const found = wrapper.findAll('button').find((item) => item.attributes('aria-label') === label)
  if (!found) throw new Error(`Botón no encontrado: ${label}`)
  return found
}
// BaseSelect es genérico: su wrapper se tipa como DOMWrapper, de ahí el cast.
function chooseSelect(wrapper: VueWrapper, index: number, value: string) {
  ;(wrapper.findAllComponents(BaseSelect)[index] as unknown as VueWrapper).vm.$emit(
    'update:modelValue',
    value,
  )
}
function pickers(wrapper: VueWrapper) {
  return wrapper.findAllComponents(BaseDatePicker)
}

beforeEach(() => {
  for (const mock of [
    createDateListSeriesMock,
    updateTimeBlockSeriesMock,
    addSeriesDateMock,
    removeSeriesDateMock,
    addSeriesExceptionMock,
    removeSeriesExceptionMock,
  ]) {
    mock.mockReset()
  }
})

describe('SeriesFormDialog · alta por fechas', () => {
  function mountCreate() {
    return mount(SeriesFormDialog, {
      props: {
        modelValue: true,
        mode: 'create-dates',
        barberId: 'b-1',
        barberName: 'Alex',
        today: '2099-01-01',
      },
      global,
    })
  }
  // Orden de los selectores de fecha: «Vigente desde», (fin opcional) y el
  // selector de la lista de fechas, que es el último.
  async function fillRange(wrapper: VueWrapper, from: string) {
    pickers(wrapper)[0]!.vm.$emit('update:modelValue', from)
    await flushPromises()
  }
  async function addDate(wrapper: VueWrapper, date: string) {
    const all = pickers(wrapper)
    all[all.length - 1]!.vm.$emit('update:modelValue', date)
    await flushPromises()
    await button(wrapper, 'Añadir a la lista').trigger('click')
    await flushPromises()
  }

  it('asks for at least one date before calling the API', async () => {
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Agrega al menos una fecha.')
    expect(createDateListSeriesMock).not.toHaveBeenCalled()
  })

  it('lists the chosen dates in order and rejects a repeated one', async () => {
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-16')
    await addDate(wrapper, '2099-12-15')
    await addDate(wrapper, '2099-12-15')

    const items = wrapper.findAll('ul[aria-label="Fechas elegidas"] li > span:first-child')
    expect(items.map((item) => item.text())).toEqual([d('2099-12-15'), d('2099-12-16')])
    expect(wrapper.text()).toContain('ya está en la lista')
  })

  it('removes a date from the list before saving', async () => {
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-15')

    await byLabel(wrapper, `Quitar la fecha ${d('2099-12-15')}`).trigger('click')

    expect(wrapper.text()).toContain('Aún no elegiste ninguna fecha.')
  })

  it('rejects a date that falls outside the validity range', async () => {
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-10')
    await addDate(wrapper, '2099-12-05')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('queda fuera de la vigencia')
    expect(createDateListSeriesMock).not.toHaveBeenCalled()
  })

  it('requires an end date when the validity is bounded', async () => {
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-15')
    chooseSelect(wrapper, 1, 'until')
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Elige la fecha de fin de la vigencia.')
    expect(createDateListSeriesMock).not.toHaveBeenCalled()
  })

  it('saves the series with all its dates, emits it and confirms with a toast', async () => {
    createDateListSeriesMock.mockResolvedValueOnce({ kind: 'success', series: dateList })
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-16')
    await addDate(wrapper, '2099-12-15')
    wrapper.findAllComponents(BaseInput).at(-1)!.vm.$emit('update:modelValue', ' Cierre de año ')
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(createDateListSeriesMock).toHaveBeenCalledWith(
      'b-1',
      {
        blockType: 'vacation',
        startsTime: '00:00',
        durationMinutes: 1440,
        effectiveFrom: '2099-12-01',
        effectiveUntil: null,
        reason: 'Cierre de año',
        explicitDates: ['2099-12-15', '2099-12-16'],
      },
      expect.any(String),
    )
    expect(wrapper.emitted('created')?.[0]).toEqual([dateList])
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([false])
    expect(toastState.items.map((item) => [item.variant, item.title])).toEqual([
      ['success', 'Bloqueo por fechas agregado'],
    ])
  })

  it('keeps the data and explains an idempotency conflict', async () => {
    createDateListSeriesMock.mockResolvedValueOnce({ kind: 'idempotency-conflict' })
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-15')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Este intento ya se envió con otros datos')
    expect(wrapper.emitted('created')).toBeUndefined()
    expect(wrapper.text()).toContain(d('2099-12-15'))
  })

  it('shows a recoverable message when the network fails', async () => {
    createDateListSeriesMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountCreate()
    await fillRange(wrapper, '2099-12-01')
    await addDate(wrapper, '2099-12-15')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos guardar los cambios')
    expect(wrapper.emitted('created')).toBeUndefined()
  })
})

describe('SeriesFormDialog · edición', () => {
  function mountEdit(series: TimeBlockSeries) {
    return mount(SeriesFormDialog, {
      props: {
        modelValue: true,
        mode: 'edit',
        barberId: 'b-1',
        barberName: 'Alex',
        series,
        today: '2099-01-01',
      },
      global,
    })
  }

  it('preloads the series and edits the whole series', async () => {
    const updated = { ...weekly, startsTime: '13:30', durationMinutes: 45 }
    updateTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'success', series: updated })
    const wrapper = mountEdit(weekly)

    expect(wrapper.text()).toContain('Editar serie de bloqueo')
    wrapper.get('input[inputmode="numeric"]').setValue(45)
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateTimeBlockSeriesMock).toHaveBeenCalledWith(
      'b-1',
      'ser-w',
      {
        blockType: 'lunch',
        startsTime: '13:00',
        durationMinutes: 45,
        effectiveFrom: '2099-01-05',
        effectiveUntil: null,
        reason: null,
      },
      null,
    )
    expect(wrapper.emitted('updated')?.[0]).toEqual([
      { originalId: 'ser-w', series: updated, splitFrom: null },
    ])
    expect(toastState.items.map((item) => item.title)).toEqual(['Serie actualizada'])
  })

  it('splits a weekly series from a cut-off date and reports the cut', async () => {
    const created = { ...weekly, id: 'ser-new', effectiveFrom: '2099-06-01' }
    updateTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'success', series: created })
    const wrapper = mountEdit(weekly)

    chooseSelect(wrapper, 1, 'this_and_following')
    await flushPromises()
    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-06-01')
    await flushPromises()

    expect(wrapper.text()).toContain(`La serie actual termina el ${d('2099-05-31')}`)
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateTimeBlockSeriesMock).toHaveBeenCalledWith(
      'b-1',
      'ser-w',
      expect.objectContaining({ effectiveFrom: '2099-06-01' }),
      { effectiveDate: '2099-06-01' },
    )
    expect(wrapper.emitted('updated')?.[0]).toEqual([
      { originalId: 'ser-w', series: created, splitFrom: '2099-06-01' },
    ])
    expect(toastState.items.map((item) => item.title)).toEqual(['Serie dividida'])
  })

  it.each([
    ['', 'Elige desde qué fecha aplica el cambio.'],
    ['2099-01-05', 'El corte debe ser posterior al inicio de la serie'],
  ])('rejects the cut-off date "%s" before calling the API', async (cut, message) => {
    const wrapper = mountEdit(weekly)
    chooseSelect(wrapper, 1, 'this_and_following')
    await flushPromises()
    if (cut) pickers(wrapper)[0]!.vm.$emit('update:modelValue', cut)
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain(message)
    expect(updateTimeBlockSeriesMock).not.toHaveBeenCalled()
  })

  it('rejects a cut-off date after the end of the validity', async () => {
    const wrapper = mountEdit({ ...weekly, effectiveUntil: '2099-03-01' })
    chooseSelect(wrapper, 1, 'this_and_following')
    await flushPromises()
    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-04-01')
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('El corte debe caer dentro de la vigencia')
    expect(updateTimeBlockSeriesMock).not.toHaveBeenCalled()
  })

  it('offers no scope choice for a date list, which only edits as a whole', () => {
    const wrapper = mountEdit(dateList)

    expect(wrapper.text()).not.toContain('Qué cambia')
    expect(wrapper.text()).toContain('«Fechas y excepciones»')
  })

  it('explains a rejected edit and keeps the form open', async () => {
    updateTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'validation-error' })
    const wrapper = mountEdit(weekly)

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Revisa los datos')
    expect(wrapper.emitted('updated')).toBeUndefined()
  })

  it('warns when the series disappeared meanwhile', async () => {
    updateTimeBlockSeriesMock.mockResolvedValueOnce({ kind: 'not-found' })
    const wrapper = mountEdit(weekly)

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Esta serie ya no está disponible')
  })
})

describe('SeriesInstancesDialog', () => {
  function mountInstances(series: TimeBlockSeries) {
    return mount(SeriesInstancesDialog, {
      props: { modelValue: true, barberId: 'b-1', series, today: '2099-01-01' },
      global,
    })
  }

  it('lists the dates in order and the exceptions with their reason', () => {
    const wrapper = mountInstances(dateList)

    expect(
      wrapper
        .findAll('ul[aria-label="Fechas de la serie"] li > span:first-child')
        .map((i) => i.text()),
    ).toEqual([d('2099-12-15'), d('2099-12-16')])
    expect(wrapper.get('ul[aria-label="Excepciones de la serie"]').text()).toContain('Libró')
  })

  it('hides the date list for a weekly series', () => {
    const wrapper = mountInstances(weekly)

    expect(wrapper.text()).toContain('Excepciones de la serie')
    expect(wrapper.text()).not.toContain('Fechas del bloqueo')
    expect(wrapper.text()).toContain('Ninguna instancia está exceptuada.')
  })

  it('removes an explicit date and returns the updated series', async () => {
    removeSeriesDateMock.mockResolvedValueOnce({ kind: 'success' })
    const wrapper = mountInstances(dateList)

    await byLabel(wrapper, `Quitar la fecha ${d('2099-12-15')}`).trigger('click')
    await flushPromises()

    expect(removeSeriesDateMock).toHaveBeenCalledWith('b-1', 'ser-d', '2099-12-15')
    expect(wrapper.emitted('update:series')?.[0]).toEqual([
      { ...dateList, dates: [{ blockDate: '2099-12-16' }] },
    ])
    expect(toastState.items.map((item) => item.title)).toEqual(['Fecha retirada'])
  })

  it('treats an already-removed date as removed, without a toast', async () => {
    removeSeriesDateMock.mockResolvedValueOnce({ kind: 'not-found' })
    const wrapper = mountInstances(dateList)

    await byLabel(wrapper, `Quitar la fecha ${d('2099-12-15')}`).trigger('click')
    await flushPromises()

    expect(wrapper.emitted('update:series')).toHaveLength(1)
    expect(toastState.items).toHaveLength(0)
  })

  it('keeps the date and shows an error when removing fails', async () => {
    removeSeriesDateMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountInstances(dateList)

    await byLabel(wrapper, `Quitar la fecha ${d('2099-12-15')}`).trigger('click')
    await flushPromises()

    expect(wrapper.emitted('update:series')).toBeUndefined()
    expect(wrapper.text()).toContain('No pudimos guardar el cambio')
  })

  it('adds an explicit date', async () => {
    addSeriesDateMock.mockResolvedValueOnce({ kind: 'success' })
    const wrapper = mountInstances(dateList)

    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-12-20')
    await flushPromises()
    await button(wrapper, 'Agregar fecha').trigger('click')
    await flushPromises()

    expect(addSeriesDateMock).toHaveBeenCalledWith('b-1', 'ser-d', '2099-12-20')
    expect(wrapper.emitted('update:series')?.[0]![0]).toMatchObject({
      dates: [
        { blockDate: '2099-12-16' },
        { blockDate: '2099-12-15' },
        { blockDate: '2099-12-20' },
      ],
    })
    expect(toastState.items.map((item) => item.title)).toEqual(['Fecha agregada'])
  })

  it.each([
    ['duplicate', 'Esa fecha ya está en la serie.'],
    ['validation-error', 'La fecha debe quedar dentro de la vigencia de la serie.'],
  ])('explains a %s answer when adding a date', async (kind, message) => {
    addSeriesDateMock.mockResolvedValueOnce({ kind })
    const wrapper = mountInstances(dateList)

    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-12-20')
    await flushPromises()
    await button(wrapper, 'Agregar fecha').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain(message)
    expect(wrapper.emitted('update:series')).toBeUndefined()
  })

  it('adds an exception with its reason', async () => {
    addSeriesExceptionMock.mockResolvedValueOnce({ kind: 'success' })
    const wrapper = mountInstances(weekly)

    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-01-12')
    wrapper.getComponent(BaseInput).vm.$emit('update:modelValue', ' Libró ')
    await flushPromises()
    await button(wrapper, 'Agregar excepción').trigger('click')
    await flushPromises()

    expect(addSeriesExceptionMock).toHaveBeenCalledWith('b-1', 'ser-w', '2099-01-12', 'Libró')
    expect(wrapper.emitted('update:series')?.[0]![0]).toMatchObject({
      exceptions: [{ excludedDate: '2099-01-12', reason: 'Libró' }],
    })
    expect(toastState.items.map((item) => item.title)).toEqual(['Excepción agregada'])
  })

  it('sends a null reason when none was written and reports a duplicate exception', async () => {
    addSeriesExceptionMock.mockResolvedValueOnce({ kind: 'duplicate' })
    const wrapper = mountInstances(weekly)

    pickers(wrapper)[0]!.vm.$emit('update:modelValue', '2099-01-12')
    await flushPromises()
    await button(wrapper, 'Agregar excepción').trigger('click')
    await flushPromises()

    expect(addSeriesExceptionMock).toHaveBeenCalledWith('b-1', 'ser-w', '2099-01-12', null)
    expect(wrapper.text()).toContain('Esa instancia ya tiene una excepción.')
    expect(wrapper.emitted('update:series')).toBeUndefined()
  })

  it('restores an excepted instance', async () => {
    removeSeriesExceptionMock.mockResolvedValueOnce({ kind: 'success' })
    const wrapper = mountInstances(dateList)

    await byLabel(wrapper, `Restaurar la instancia del ${d('2099-12-20')}`).trigger('click')
    await flushPromises()

    expect(removeSeriesExceptionMock).toHaveBeenCalledWith('b-1', 'ser-d', '2099-12-20')
    expect(wrapper.emitted('update:series')?.[0]).toEqual([{ ...dateList, exceptions: [] }])
    expect(toastState.items.map((item) => item.title)).toEqual(['Instancia restaurada'])
  })
})
