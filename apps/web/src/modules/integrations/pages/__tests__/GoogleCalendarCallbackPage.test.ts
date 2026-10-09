/**
 * Pruebas de GoogleCalendarCallbackPage (issue #325, DEC-102): reenvía una sola vez
 * `state` y `code` al API, limpia la URL de inmediato y distingue conectado,
 * rechazo, fallo y autorización inválida. Nunca guarda el código.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { resetToasts, toastState } from '@/shared/model/toastStore'

const completeMock = vi.hoisted(() => vi.fn())
const replaceMock = vi.hoisted(() => vi.fn())
const routeMock = vi.hoisted(() => ({ query: {} as Record<string, unknown> }))

vi.mock('../../api/googleCalendarApi', () => ({ completeCallback: completeMock }))
vi.mock('vue-router', () => ({
  useRoute: () => routeMock,
  useRouter: () => ({ replace: replaceMock }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const { default: GoogleCalendarCallbackPage } = await import('../GoogleCalendarCallbackPage.vue')

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

async function mountWith(query: Record<string, unknown>) {
  routeMock.query = query
  const wrapper = mount(GoogleCalendarCallbackPage, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
  await flushPromises()
  return wrapper
}

describe('GoogleCalendarCallbackPage', () => {
  beforeEach(() => {
    resetToasts()
    completeMock.mockReset()
    replaceMock.mockReset()
    replaceMock.mockResolvedValue(undefined)
  })

  it('forwards state and code once, clears the URL first, and goes back on success', async () => {
    completeMock.mockResolvedValue({ kind: 'success', result: 'connected' })
    await mountWith({ state: 'estado-1', code: 'codigo-1' })

    expect(completeMock).toHaveBeenCalledTimes(1)
    expect(completeMock).toHaveBeenCalledWith({
      state: 'estado-1',
      code: 'codigo-1',
      error: '',
    })
    // La primera navegación vacía el query: el código no se queda en la URL.
    expect(replaceMock.mock.calls[0]![0]).toEqual({
      name: 'integraciones-google-calendar-retorno',
      query: {},
    })
    expect(replaceMock).toHaveBeenLastCalledWith({ name: 'integraciones-google-calendar' })
    expect(toastState.items.some((t) => t.title === 'Google Calendar conectado')).toBe(true)
  })

  it('forwards a denial from Google and explains it', async () => {
    completeMock.mockResolvedValue({ kind: 'success', result: 'denied' })
    const wrapper = await mountWith({ state: 's', error: 'access_denied' })
    expect(completeMock).toHaveBeenCalledWith({ state: 's', code: '', error: 'access_denied' })
    expect(wrapper.text()).toContain('No diste el permiso')
    expect(replaceMock).not.toHaveBeenCalledWith({ name: 'integraciones-google-calendar' })
  })

  it('reports Google failing to complete the connection', async () => {
    completeMock.mockResolvedValue({ kind: 'success', result: 'failed' })
    const wrapper = await mountWith({ state: 's', code: 'c' })
    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(wrapper.text()).toContain('tus turnos no se tocaron')
  })

  it('treats an invalid or expired authorization as such, with a way back', async () => {
    completeMock.mockResolvedValue({ kind: 'invalid' })
    const wrapper = await mountWith({ state: 'viejo', code: 'c' })
    expect(wrapper.text()).toContain('La autorización venció')
    expect(wrapper.text()).toContain('Volver a Google Calendar')
  })

  it('does not call the API without a state', async () => {
    const wrapper = await mountWith({ code: 'solo-codigo' })
    expect(completeMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('La autorización venció')
  })

  it('shows a network problem as a generic connection error', async () => {
    completeMock.mockResolvedValue({ kind: 'network-error' })
    const wrapper = await mountWith({ state: 's', code: 'c' })
    expect(wrapper.text()).toContain('Revisa tu conexión')
  })

  it('takes the first value when a query parameter is repeated', async () => {
    completeMock.mockResolvedValue({ kind: 'success', result: 'connected' })
    await mountWith({ state: ['a', 'b'], code: ['c1', 'c2'] })
    expect(completeMock).toHaveBeenCalledWith({ state: 'a', code: 'c1', error: '' })
  })

  it('has no axe violations while working or on an error', async () => {
    completeMock.mockReturnValue(new Promise(() => {}))
    const working = await mountWith({ state: 's', code: 'c' })
    expect(await axe(working.element, axeOptions)).toHaveNoViolations()

    completeMock.mockResolvedValue({ kind: 'invalid' })
    const failed = await mountWith({ state: 's', code: 'c' })
    expect(await axe(failed.element, axeOptions)).toHaveNoViolations()
  })
})
