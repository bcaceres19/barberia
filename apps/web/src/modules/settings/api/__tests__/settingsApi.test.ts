/**
 * Pruebas del cliente tipado (`settingsApi` sobre `shared/api/httpClient`):
 * mapeo por `status` (nunca por `detail`), mismo criterio que
 * `auth/api/__tests__/loginApi.test.ts`. El recorrido de red real vive en
 * un E2E propio de HU-020.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const getMock = vi.hoisted(() => vi.fn())
const patchMock = vi.hoisted(() => vi.fn())

vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: getMock, PATCH: patchMock },
}))

const { fetchBarbershopSettings, saveBarbershopSettings } = await import('../settingsApi')

function ok(body: unknown) {
  return { data: body, error: undefined, response: new Response(null, { status: 200 }) }
}

function problem(status: number, body: Record<string, unknown> = {}) {
  return { data: undefined, error: body, response: new Response(null, { status }) }
}

const settingsBody = {
  name: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@ejemplo.test',
  contactPhone: '+573001234567',
}

describe('settingsApi.fetchBarbershopSettings', () => {
  beforeEach(() => getMock.mockReset())

  it('maps a 200 success body to a success outcome', async () => {
    getMock.mockResolvedValueOnce(ok(settingsBody))

    const outcome = await fetchBarbershopSettings()

    expect(outcome).toEqual({ kind: 'success', settings: settingsBody })
  })

  it('maps null contact fields through unchanged (no contact configured)', async () => {
    getMock.mockResolvedValueOnce(ok({ ...settingsBody, contactEmail: null, contactPhone: null }))

    const outcome = await fetchBarbershopSettings()

    expect(outcome).toEqual({
      kind: 'success',
      settings: { ...settingsBody, contactEmail: null, contactPhone: null },
    })
  })

  it('maps any non-2xx status to unexpected-error (401 is handled globally by installSessionHandling)', async () => {
    getMock.mockResolvedValueOnce(problem(401))

    const outcome = await fetchBarbershopSettings()

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    getMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await fetchBarbershopSettings()

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})

describe('settingsApi.saveBarbershopSettings', () => {
  beforeEach(() => patchMock.mockReset())

  const formValues = {
    name: 'Barbería Ejemplo',
    timezone: 'America/Bogota',
    contactEmail: 'contacto@ejemplo.test',
    contactPhone: '+573001234567',
  }

  it('sends exactly the four contract fields as the body', async () => {
    patchMock.mockResolvedValueOnce(ok(settingsBody))

    await saveBarbershopSettings(formValues)

    expect(patchMock).toHaveBeenCalledTimes(1)
    const [path, options] = patchMock.mock.calls[0] as [string, { body: unknown }]
    expect(path).toBe('/private/settings/barbershop')
    expect(options.body).toEqual(formValues)
  })

  it('maps a 200 success body to a success outcome', async () => {
    patchMock.mockResolvedValueOnce(ok(settingsBody))

    const outcome = await saveBarbershopSettings(formValues)

    expect(outcome).toEqual({ kind: 'success', settings: settingsBody })
  })

  it('maps 422 (field validation, e.g. unrecognized timezone) to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(422, { code: 'validation-error' }))

    const outcome = await saveBarbershopSettings({ ...formValues, timezone: 'COT' })

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps 400 (unknown field / malformed JSON) to validation-error', async () => {
    patchMock.mockResolvedValueOnce(problem(400, { code: 'invalid-request' }))

    const outcome = await saveBarbershopSettings(formValues)

    expect(outcome).toEqual({ kind: 'validation-error' })
  })

  it('maps a 500 to unexpected-error', async () => {
    patchMock.mockResolvedValueOnce(problem(500, { code: 'internal-error' }))

    const outcome = await saveBarbershopSettings(formValues)

    expect(outcome).toEqual({ kind: 'unexpected-error' })
  })

  it('maps a rejected request (no HTTP response) to network-error', async () => {
    patchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'))

    const outcome = await saveBarbershopSettings(formValues)

    expect(outcome).toEqual({ kind: 'network-error' })
  })
})
