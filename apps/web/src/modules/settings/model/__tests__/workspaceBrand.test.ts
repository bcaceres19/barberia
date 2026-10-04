/**
 * Carga de la marca al autenticarse (DEC-110): un fallo no bloquea el panel y
 * una respuesta obsoleta nunca pisa la de la sesión vigente.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { brandState, DEFAULT_BRAND, resetBrand } from '@/shared/model'

const fetchBrandMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/brandApi', () => ({ fetchBrand: fetchBrandMock, saveBrand: vi.fn() }))

const { loadWorkspaceBrand } = await import('../workspaceBrand')

const salon = {
  ...DEFAULT_BRAND,
  accent: 'ruby' as const,
  businessTerm: 'salón',
  businessTermGender: 'masculine' as const,
}

beforeEach(() => {
  window.localStorage.clear()
  resetBrand()
  fetchBrandMock.mockReset()
})

describe('loadWorkspaceBrand', () => {
  it('publishes the confirmed brand to the shared state', async () => {
    fetchBrandMock.mockResolvedValueOnce({ kind: 'success', brand: salon })

    await loadWorkspaceBrand()

    expect({ ...brandState.brand }).toEqual(salon)
  })

  it.each([{ kind: 'network-error' }, { kind: 'unexpected-error' }])(
    'keeps what was painted when the server answers %o',
    async (outcome) => {
      fetchBrandMock.mockResolvedValueOnce(outcome)

      await loadWorkspaceBrand()

      expect({ ...brandState.brand }).toEqual(DEFAULT_BRAND)
    },
  )

  it('ignores a stale answer when a newer request is already in flight', async () => {
    let resolveFirst: (value: unknown) => void = () => {}
    fetchBrandMock.mockReturnValueOnce(new Promise((resolve) => (resolveFirst = resolve)))
    fetchBrandMock.mockResolvedValueOnce({
      kind: 'success',
      brand: { ...salon, accent: 'sapphire' },
    })

    const first = loadWorkspaceBrand()
    await loadWorkspaceBrand()
    resolveFirst({ kind: 'success', brand: salon })
    await first

    expect(brandState.brand.accent).toBe('sapphire')
  })
})
