/**
 * «Lo ofrezco» (DEC-115): los servicios del único barbero de una barbería
 * individual. La API se sustituye por un doble; el recorrido real contra el
 * servidor vive en el E2E del perfil de barbero individual.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'

const fetchBarberSummariesMock = vi.hoisted(() => vi.fn())
const fetchAssignmentsMock = vi.hoisted(() => vi.fn())
const assignServiceMock = vi.hoisted(() => vi.fn())
const unassignServiceMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/barberServicesApi', () => ({
  fetchBarberSummaries: fetchBarberSummariesMock,
  fetchAssignments: fetchAssignmentsMock,
  assignService: assignServiceMock,
  unassignService: unassignServiceMock,
}))

const { useBarberOffering } = await import('../barberOffering')

const oneBarber = { kind: 'success', items: [{ id: 'b-1', fullName: 'Mateo Rojas' }] }
const assignment = (serviceId: string) => ({ barberId: 'b-1', serviceId, createdAt: '2026-10-04' })
const page = (serviceIds: string[], nextCursor: string | null = null) => ({
  kind: 'success',
  page: { items: serviceIds.map(assignment), nextCursor },
})

async function ready(offered: string[] = ['s-1']) {
  fetchBarberSummariesMock.mockResolvedValueOnce(oneBarber)
  fetchAssignmentsMock.mockResolvedValueOnce(page(offered))
  const offering = useBarberOffering()
  await offering.load()
  return offering
}

describe('useBarberOffering', () => {
  beforeEach(() => {
    fetchBarberSummariesMock.mockReset()
    fetchAssignmentsMock.mockReset()
    assignServiceMock.mockReset()
    unassignServiceMock.mockReset()
  })

  describe('load', () => {
    it('reads what the only barber offers, across every page of the cursor', async () => {
      fetchBarberSummariesMock.mockResolvedValueOnce(oneBarber)
      fetchAssignmentsMock
        .mockResolvedValueOnce(page(['s-1', 's-2'], 'next'))
        .mockResolvedValueOnce(page(['s-3']))
      const offering = useBarberOffering()

      await offering.load()

      expect(offering.status.value).toBe('ready')
      expect(['s-1', 's-2', 's-3'].map(offering.isOffered)).toEqual([true, true, true])
      expect(offering.isOffered('s-4')).toBe(false)
      expect(fetchAssignmentsMock).toHaveBeenNthCalledWith(1, 'b-1', undefined)
      expect(fetchAssignmentsMock).toHaveBeenNthCalledWith(2, 'b-1', 'next')
    })

    it.each([
      ['no barber', []],
      [
        'several barbers',
        [
          { id: 'b-1', fullName: 'A' },
          { id: 'b-2', fullName: 'B' },
        ],
      ],
    ])('is unavailable with %s: there is nobody to assign to without asking', async (_n, items) => {
      fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'success', items })
      const offering = useBarberOffering()

      await offering.load()

      expect(offering.status.value).toBe('unavailable')
      expect(fetchAssignmentsMock).not.toHaveBeenCalled()
      expect(await offering.setOffered('s-1', true)).toBe('error')
      expect(assignServiceMock).not.toHaveBeenCalled()
    })

    it('reports an error when the barbers or the assignments cannot be read', async () => {
      fetchBarberSummariesMock.mockResolvedValueOnce({ kind: 'network-error' })
      const a = useBarberOffering()
      await a.load()
      expect(a.status.value).toBe('error')

      fetchBarberSummariesMock.mockResolvedValueOnce(oneBarber)
      fetchAssignmentsMock.mockResolvedValueOnce({ kind: 'unexpected-error' })
      const b = useBarberOffering()
      await b.load()
      expect(b.status.value).toBe('error')
    })

    it('stops following a cursor that never ends', async () => {
      fetchBarberSummariesMock.mockResolvedValueOnce(oneBarber)
      fetchAssignmentsMock.mockResolvedValue(page(['s-1'], 'again'))
      const offering = useBarberOffering()

      await offering.load()

      expect(fetchAssignmentsMock).toHaveBeenCalledTimes(20)
      expect(offering.status.value).toBe('ready')
    })
  })

  describe('setOffered', () => {
    it('offers a service only once the server confirms it', async () => {
      const offering = await ready([])
      let resolve: (value: unknown) => void = () => {}
      assignServiceMock.mockReturnValueOnce(new Promise((r) => (resolve = r)))

      const pending = offering.setOffered('s-9', true)
      expect(offering.isOffered('s-9')).toBe(false)
      expect(offering.isBusy('s-9')).toBe(true)

      resolve({ kind: 'success', assignment: assignment('s-9') })
      expect(await pending).toBe('changed')
      expect(offering.isOffered('s-9')).toBe(true)
      expect(offering.isBusy('s-9')).toBe(false)
      expect(assignServiceMock).toHaveBeenCalledWith('b-1', 's-9')
    })

    it('stops offering a service once the server confirms it', async () => {
      const offering = await ready(['s-1'])
      unassignServiceMock.mockResolvedValueOnce({ kind: 'success' })

      expect(await offering.setOffered('s-1', false)).toBe('changed')
      expect(offering.isOffered('s-1')).toBe(false)
      expect(unassignServiceMock).toHaveBeenCalledWith('b-1', 's-1')
    })

    it('treats a service that was already gone as stopped offering', async () => {
      const offering = await ready(['s-1'])
      unassignServiceMock.mockResolvedValueOnce({ kind: 'not-found' })

      expect(await offering.setOffered('s-1', false)).toBe('changed')
      expect(offering.isOffered('s-1')).toBe(false)
    })

    it('keeps what was saved when the server refuses or the network fails', async () => {
      const offering = await ready(['s-1'])

      unassignServiceMock.mockResolvedValueOnce({ kind: 'last-active-conflict' })
      expect(await offering.setOffered('s-1', false)).toBe('error')
      expect(offering.isOffered('s-1')).toBe(true)

      unassignServiceMock.mockResolvedValueOnce({ kind: 'network-error' })
      expect(await offering.setOffered('s-1', false)).toBe('network-error')
      expect(offering.isOffered('s-1')).toBe(true)

      assignServiceMock.mockResolvedValueOnce({ kind: 'not-found' })
      expect(await offering.setOffered('s-2', true)).toBe('error')
      expect(offering.isOffered('s-2')).toBe(false)
      expect(offering.isBusy('s-1')).toBe(false)
    })

    it('does nothing when the service is already in the requested state', async () => {
      const offering = await ready(['s-1'])

      expect(await offering.setOffered('s-1', true)).toBe('unchanged')
      expect(await offering.setOffered('s-2', false)).toBe('unchanged')
      expect(assignServiceMock).not.toHaveBeenCalled()
      expect(unassignServiceMock).not.toHaveBeenCalled()
    })

    it('does not send a second request while one is in flight for the same service', async () => {
      const offering = await ready([])
      let resolve: (value: unknown) => void = () => {}
      assignServiceMock.mockReturnValueOnce(new Promise((r) => (resolve = r)))

      const first = offering.setOffered('s-1', true)
      expect(await offering.setOffered('s-1', true)).toBe('unchanged')
      resolve({ kind: 'success', assignment: assignment('s-1') })
      await first

      expect(assignServiceMock).toHaveBeenCalledTimes(1)
    })
  })
})
