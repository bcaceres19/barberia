/**
 * Pruebas de GoogleCalendarPage (issue #325, DEC-099/DEC-101/DEC-122): cada estado
 * (no disponible, falta vincular el barbero, no conectado, conectado, sincronizando,
 * requiere reconexión, error de sincronización), las acciones (conectar, sincronizar,
 * desconectar con confirmación) seguras ante el doble clic, el recordatorio validado
 * de 0 a 40320 y el texto de «cómo funciona». googleCalendarApi se sustituye por un
 * doble; el recorrido real vive en el E2E.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { resetToasts, toastState } from '@/shared/model/toastStore'

const fetchMock = vi.hoisted(() => vi.fn())
const startConnectMock = vi.hoisted(() => vi.fn())
const syncMock = vi.hoisted(() => vi.fn())
const disconnectMock = vi.hoisted(() => vi.fn())
const reminderMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/googleCalendarApi', () => ({
  fetchConnection: fetchMock,
  startConnect: startConnectMock,
  syncNow: syncMock,
  disconnect: disconnectMock,
  updateReminder: reminderMock,
}))

const { default: GoogleCalendarPage } = await import('../GoogleCalendarPage.vue')

const base = {
  enabled: true,
  barberLinked: true,
  status: 'not_connected',
  accountEmail: null,
  reminderMinutes: null,
  connectedAt: null,
  lastSyncedAt: null,
  pendingSyncJobs: 0,
  failedSyncJobs: 0,
}
const connected = {
  ...base,
  status: 'connected',
  accountEmail: 'barbero@ejemplo.test',
  connectedAt: '2026-10-10T15:00:00Z',
  lastSyncedAt: '2026-10-10T16:30:00Z',
}

const axeOptions = { rules: { 'color-contrast': { enabled: false } } }

async function mountWith(connection: object): Promise<VueWrapper> {
  fetchMock.mockResolvedValue({ kind: 'success', connection })
  const wrapper = mount(GoogleCalendarPage, {
    global: {
      stubs: { teleport: true, transition: true, RouterLink: { template: '<a><slot /></a>' } },
    },
  })
  await flushPromises()
  return wrapper
}

function button(wrapper: VueWrapper, text: string) {
  const found = wrapper.findAll('button').find((b) => b.text() === text)
  if (!found) throw new Error(`button ${JSON.stringify(text)} not found`)
  return found
}

// BaseDialog deja sus diálogos SIEMPRE en el DOM (v-show) y el stub de Teleport los
// monta dentro del wrapper: se busca el que esté visualmente abierto.
function dialogButton(wrapper: VueWrapper, text: string): HTMLButtonElement {
  const root = wrapper.element as Element
  const buttons = Array.from(
    root.querySelectorAll('.base-dialog--open button'),
  ) as HTMLButtonElement[]
  const found = buttons.find((b) => b.textContent?.trim() === text)
  if (!found) throw new Error(`dialog button ${JSON.stringify(text)} not found`)
  return found as HTMLButtonElement
}

describe('GoogleCalendarPage', () => {
  beforeEach(() => {
    resetToasts()
    for (const mock of [fetchMock, startConnectMock, syncMock, disconnectMock, reminderMock]) {
      mock.mockReset()
    }
  })

  it('shows a loading state and then the content', async () => {
    let resolve: (value: unknown) => void = () => {}
    fetchMock.mockReturnValue(new Promise((r) => (resolve = r)))
    const wrapper = mount(GoogleCalendarPage, {
      global: { stubs: { teleport: true, transition: true } },
    })
    expect(wrapper.text()).toContain('Cargando')
    resolve({ kind: 'success', connection: base })
    await flushPromises()
    expect(wrapper.text()).not.toContain('Cargando Google Calendar')
    expect(wrapper.text()).toContain('No conectado')
  })

  it('offers a retry when the status cannot be loaded', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mount(GoogleCalendarPage, {
      global: { stubs: { teleport: true, transition: true } },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos cargar tu Google Calendar')

    fetchMock.mockResolvedValueOnce({ kind: 'success', connection: base })
    await button(wrapper, 'Reintentar').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Conectar Google Calendar')
  })

  it('says the integration is unavailable and offers no actions', async () => {
    const wrapper = await mountWith({ ...base, enabled: false })
    expect(wrapper.text()).toContain('No disponible')
    expect(wrapper.text()).toContain('no está disponible en este entorno')
    expect(wrapper.findAll('.gcal-panel button')).toHaveLength(0)
    expect(wrapper.text()).not.toContain('Cómo funciona')
  })

  it('asks to link the barber first and points to the team screen', async () => {
    const wrapper = await mountWith({ ...base, barberLinked: false })
    expect(wrapper.text()).toContain('Este soy yo')
    expect(wrapper.text()).not.toContain('Conectar Google Calendar')
  })

  it('walks through the three steps and marks where the barber is', async () => {
    const linking = await mountWith({ ...base, barberLinked: false })
    const first = linking.findAll('.gcal-step')
    expect(first).toHaveLength(3)
    expect(first[0].attributes('aria-current')).toBe('step')
    expect(first[1].attributes('aria-current')).toBeUndefined()

    const connecting = await mountWith(base)
    const second = connecting.findAll('.gcal-step')
    expect(second[0].classes()).toContain('gcal-step--done')
    expect(second[1].attributes('aria-current')).toBe('step')
  })

  it('hides the steps once Google Calendar is connected', async () => {
    const wrapper = await mountWith(connected)
    expect(wrapper.find('.gcal-steps').exists()).toBe(false)
  })

  it('shows the how-it-works cards in plain, short language', async () => {
    const wrapper = await mountWith(connected)
    const cards = wrapper.findAll('.gcal-how__item')
    expect(cards).toHaveLength(4)
    expect(cards[0].text()).toContain('NAVA manda')
  })

  it('shows «No conectado» with the connect action, and the how-it-works copy always', async () => {
    const wrapper = await mountWith(base)
    expect(wrapper.text()).toContain('No conectado')
    expect(button(wrapper, 'Conectar Google Calendar').exists()).toBe(true)
    const how = wrapper.text()
    expect(how).toContain('NAVA manda')
    expect(how).toContain('no modifica tus turnos')
    expect(how).toContain('NAVA lo vuelve a crear')
    expect(how).toContain('cancélalo desde NAVA')
    expect(how).toContain('Google les envía la invitación')
  })

  it('shows the connected state with account, last sync and the three actions', async () => {
    const wrapper = await mountWith(connected)
    expect(wrapper.text()).toContain('Conectado')
    expect(wrapper.text()).toContain('barbero@ejemplo.test')
    expect(wrapper.text()).toContain('Última sincronización')
    for (const label of ['Sincronizar ahora', 'Desconectar']) {
      expect(button(wrapper, label).exists()).toBe(true)
    }
    expect(wrapper.text()).toContain('Recordatorio')
  })

  it('says it has not synced yet when there is no last sync', async () => {
    const wrapper = await mountWith({ ...connected, lastSyncedAt: null })
    expect(wrapper.text()).toContain('Aún no hay eventos sincronizados')
  })

  it('shows «Sincronizando» with the queued change count while changes are pending', async () => {
    const wrapper = await mountWith({ ...connected, pendingSyncJobs: 3 })
    expect(wrapper.text()).toContain('Sincronizando')
    expect(wrapper.text()).toContain('3 cambios')
  })

  it('shows the recoverable sync error and keeps «Sincronizar ahora» available', async () => {
    const wrapper = await mountWith({ ...connected, failedSyncJobs: 2 })
    expect(wrapper.text()).toContain('Error de sincronización')
    expect(wrapper.text()).toContain('Tus turnos están a salvo en NAVA')
    expect(button(wrapper, 'Sincronizar ahora').exists()).toBe(true)
  })

  it('asks to reconnect when Google revoked the permission', async () => {
    const wrapper = await mountWith({ ...base, status: 'reauth_required' })
    expect(wrapper.text()).toContain('Requiere reconexión')
    expect(button(wrapper, 'Volver a conectar').exists()).toBe(true)
    expect(wrapper.text()).toContain('Tus turnos no se pierden')
    expect(wrapper.text()).not.toContain('Recordatorio')
  })

  // --- Acciones -----------------------------------------------------------------

  it('connects by sending the browser to Google, and ignores a second click', async () => {
    const assign = vi.fn()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, assign },
    })
    startConnectMock.mockResolvedValue({
      kind: 'success',
      authorizationUrl: 'https://accounts.example.test/auth',
    })
    const wrapper = await mountWith(base)

    await button(wrapper, 'Conectar Google Calendar').trigger('click')
    await button(wrapper, 'Abriendo Google…').trigger('click')
    await flushPromises()

    expect(startConnectMock).toHaveBeenCalledTimes(1)
    expect(assign).toHaveBeenCalledTimes(1)
    expect(assign).toHaveBeenCalledWith('https://accounts.example.test/auth')
    expect(button(wrapper, 'Abriendo Google…').attributes('disabled')).toBeDefined()
  })

  it('explains a failed connect without leaving the screen', async () => {
    startConnectMock.mockResolvedValue({ kind: 'network-error' })
    const wrapper = await mountWith(base)
    await button(wrapper, 'Conectar Google Calendar').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos abrir Google')
    expect(button(wrapper, 'Conectar Google Calendar').attributes('disabled')).toBeUndefined()
  })

  it('syncs now once even with repeated clicks and confirms with a toast', async () => {
    let finish: (value: unknown) => void = () => {}
    syncMock.mockReturnValue(new Promise((r) => (finish = r)))
    const wrapper = await mountWith(connected)

    await button(wrapper, 'Sincronizar ahora').trigger('click')
    await button(wrapper, 'Sincronizando…').trigger('click')
    expect(syncMock).toHaveBeenCalledTimes(1)
    expect(button(wrapper, 'Desconectar').attributes('disabled')).toBeDefined()

    finish({ kind: 'success', pendingSyncJobs: 2, failedSyncJobs: 0 })
    await flushPromises()
    expect(
      toastState.items.some((t) => t.title.includes('Sincronizando con Google Calendar')),
    ).toBe(true)
    expect(fetchMock).toHaveBeenCalledTimes(2) // recarga el estado
  })

  it('keeps the screen usable when syncing fails', async () => {
    syncMock.mockResolvedValue({ kind: 'unexpected-error' })
    const wrapper = await mountWith(connected)
    await button(wrapper, 'Sincronizar ahora').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('No pudimos sincronizar')
    expect(button(wrapper, 'Sincronizar ahora').attributes('disabled')).toBeUndefined()
  })

  it('asks for confirmation before disconnecting and then disconnects once', async () => {
    disconnectMock.mockResolvedValue({ kind: 'success' })
    const wrapper = await mountWith(connected)

    await button(wrapper, 'Desconectar').trigger('click')
    await flushPromises()
    expect(disconnectMock).not.toHaveBeenCalled()
    const open = wrapper.element.querySelector('.base-dialog--open')!
    expect(open.textContent).toContain('¿Desconectar Google Calendar?')
    expect(open.textContent).toContain('se quedan allí')

    fetchMock.mockResolvedValue({
      kind: 'success',
      connection: { ...base, status: 'disconnected' },
    })
    dialogButton(wrapper, 'Desconectar').click()
    await flushPromises()

    expect(disconnectMock).toHaveBeenCalledTimes(1)
    expect(toastState.items.some((t) => t.title === 'Google Calendar desconectado')).toBe(true)
    expect(wrapper.text()).toContain('No conectado')
  })

  it('can cancel the disconnect confirmation without side effects', async () => {
    const wrapper = await mountWith(connected)
    await button(wrapper, 'Desconectar').trigger('click')
    await flushPromises()
    dialogButton(wrapper, 'Cancelar').click()
    await flushPromises()
    expect(disconnectMock).not.toHaveBeenCalled()
  })

  // --- Recordatorio -----------------------------------------------------------------

  it('starts on the Google defaults when there is no reminder and saves null', async () => {
    reminderMock.mockResolvedValue({ kind: 'success', connection: connected })
    const wrapper = await mountWith(connected)
    const defaults = wrapper.get('input[name="reminderMode"][value="default"]')
    expect((defaults.element as HTMLInputElement).checked).toBe(true)
    expect(wrapper.find('input[name="reminderMinutes"]').exists()).toBe(false)

    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(reminderMock).toHaveBeenCalledTimes(1)
    expect(reminderMock).toHaveBeenCalledWith(null)
  })

  it('switches to custom minutes and lets a shortcut fill the field', async () => {
    reminderMock.mockResolvedValue({
      kind: 'success',
      connection: { ...connected, reminderMinutes: 60 },
    })
    const wrapper = await mountWith(connected)
    await wrapper.get('input[name="reminderMode"][value="custom"]').setValue(true)
    expect(wrapper.find('input[name="reminderMinutes"]').exists()).toBe(true)

    const hour = wrapper.findAll('.gcal-chip').find((c) => c.text() === '1 hora')
    expect(hour?.attributes('aria-pressed')).toBe('false')
    await hour?.trigger('click')
    expect(hour?.attributes('aria-pressed')).toBe('true')
    expect((wrapper.get('input[name="reminderMinutes"]').element as HTMLInputElement).value).toBe(
      '60',
    )

    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(reminderMock).toHaveBeenCalledWith(60)
  })

  it('loads the saved minutes, validates the range and saves a valid value', async () => {
    reminderMock.mockResolvedValue({
      kind: 'success',
      connection: { ...connected, reminderMinutes: 45 },
    })
    const wrapper = await mountWith({ ...connected, reminderMinutes: 30 })
    const input = wrapper.get('input[name="reminderMinutes"]')
    expect((input.element as HTMLInputElement).value).toBe('30')

    await input.setValue('99999')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(reminderMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('de 0 a 40320')

    await input.setValue('45')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(reminderMock).toHaveBeenCalledTimes(1)
    expect(reminderMock).toHaveBeenCalledWith(45)
    expect(toastState.items.some((t) => t.title === 'Recordatorio guardado')).toBe(true)
  })

  it('shows a server validation error under the field and keeps what was typed', async () => {
    reminderMock.mockResolvedValue({ kind: 'validation-error' })
    const wrapper = await mountWith({ ...connected, reminderMinutes: 30 })
    await wrapper.get('input[name="reminderMinutes"]').setValue('10')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('de 0 a 40320')
    expect((wrapper.get('input[name="reminderMinutes"]').element as HTMLInputElement).value).toBe(
      '10',
    )
  })

  it('disables the form while the reminder is saving', async () => {
    let finish: (value: unknown) => void = () => {}
    reminderMock.mockReturnValue(new Promise((r) => (finish = r)))
    const wrapper = await mountWith(connected)
    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')
    expect(reminderMock).toHaveBeenCalledTimes(1)
    expect(button(wrapper, 'Guardando…').attributes('disabled')).toBeDefined()
    finish({ kind: 'success', connection: connected })
    await flushPromises()
  })

  // --- Accesibilidad ----------------------------------------------------------------

  it.each([
    ['no conectado', base],
    ['conectado', connected],
    ['sincronizando', { ...connected, pendingSyncJobs: 2 }],
    ['error', { ...connected, failedSyncJobs: 1 }],
    ['reconexión', { ...base, status: 'reauth_required' }],
    ['falta vincular', { ...base, barberLinked: false }],
    ['no disponible', { ...base, enabled: false }],
  ])('has no axe violations in the «%s» state', async (_name, connection) => {
    const wrapper = await mountWith(connection)
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })
})
