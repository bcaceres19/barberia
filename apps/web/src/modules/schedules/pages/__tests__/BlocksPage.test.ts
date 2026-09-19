/**
 * Pruebas de BlocksPage centradas en el feedback de cada cambio (DEC-095):
 * una confirmación emergente por alta o retiro, y un aviso de error con
 * "Reintentar" cuando un retiro falla, porque no hay formulario donde
 * mostrarlo. Los datos de la API se sustituyen por dobles controlables.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { BaseInput } from '@/shared/ui'
import { toastState } from '@/shared/model/toastStore'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchBarbershopTimezoneMock = vi.hoisted(() => vi.fn())
const fetchTimeBlocksMock = vi.hoisted(() => vi.fn())
const fetchTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const createTimeBlockMock = vi.hoisted(() => vi.fn())
const createTimeBlockSeriesMock = vi.hoisted(() => vi.fn())
const deleteTimeBlockMock = vi.hoisted(() => vi.fn())
const deleteTimeBlockSeriesMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/schedulesApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchBarbershopTimezone: fetchBarbershopTimezoneMock,
}))
vi.mock('../../api/timeBlocksApi', () => ({
  fetchTimeBlocks: fetchTimeBlocksMock,
  fetchTimeBlockSeries: fetchTimeBlockSeriesMock,
  createTimeBlock: createTimeBlockMock,
  createTimeBlockSeries: createTimeBlockSeriesMock,
  deleteTimeBlock: deleteTimeBlockMock,
  deleteTimeBlockSeries: deleteTimeBlockSeriesMock,
}))

const { default: BlocksPage } = await import('../BlocksPage.vue')

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
  reason: null,
  deletedAt: null,
  deletedBy: null,
  createdAt: '2026-09-01T12:00:00Z',
  updatedAt: '2026-09-01T12:00:00Z',
}

async function mountReady() {
  fetchBarberSummariesMock.mockResolvedValueOnce({
    kind: 'success',
    items: [{ id: 'b-1', fullName: 'Carlos Ramírez' }],
  })
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
  const wrapper = mount(BlocksPage, { global: { stubs: { teleport: true, RouterLink: true } } })
  await flushPromises()
  return wrapper
}

// Los dos "Retirar" de la pantalla: el primero es el bloqueo puntual y el
// segundo la serie semanal.
function retireButtons(wrapper: VueWrapper) {
  return wrapper.findAll('button').filter((button) => button.text() === 'Retirar')
}

describe('BlocksPage · avisos emergentes (DEC-095)', () => {
  beforeEach(() => {
    for (const mock of [
      fetchBarberSummariesMock,
      fetchBarbershopTimezoneMock,
      fetchTimeBlocksMock,
      fetchTimeBlockSeriesMock,
      createTimeBlockMock,
      createTimeBlockSeriesMock,
      deleteTimeBlockMock,
      deleteTimeBlockSeriesMock,
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
    // Se escribe por el v-model de cada BaseInput: fecha y hora de inicio y de fin.
    const inputs = form.findAllComponents(BaseInput)
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
    expect(wrapper.text()).toContain('Este barbero no tiene bloqueos puntuales vigentes.')
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
})
