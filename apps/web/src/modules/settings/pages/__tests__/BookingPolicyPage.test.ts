/**
 * Pruebas de BookingPolicyPage (HU-093): carga, edición, guardado exitoso
 * (renueva versionToken), conflicto de versión con recarga, error de
 * validación y error de red conservando lo escrito, doble envío bloqueado,
 * validación local de rangos/coherencia. bookingPolicyApi se sustituye por
 * un doble de prueba; el recorrido real contra el API vive en el E2E de
 * HU-093.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchMock = vi.hoisted(() => vi.fn())
const saveMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/bookingPolicyApi', () => ({
  fetchBookingPolicy: fetchMock,
  saveBookingPolicy: saveMock,
}))

const { default: BookingPolicyPage } = await import('../BookingPolicyPage.vue')

const loadedPolicy = {
  minAdvanceMinutes: 60,
  maxAdvanceDays: 3,
  slotGridMinutes: 15,
  cancellationDeadlineMinutes: 20,
  lateCancellationClientAllowed: true,
  lateCancellationReasonRequired: true,
  versionToken: 'tok-1',
}

// Los dos interruptores de «Cancelación fuera de plazo», en el orden de la
// pantalla: primero si puede cancelar, después si debe indicar un motivo.
type Mounted = Awaited<ReturnType<typeof mountReady>>
const lateCancellationSwitch = (wrapper: Mounted) => wrapper.findAll('button[role="switch"]')[0]!
const reasonRequiredSwitch = (wrapper: Mounted) => wrapper.findAll('button[role="switch"]')[1]!

async function mountReady() {
  fetchMock.mockResolvedValueOnce({ kind: 'success', policy: loadedPolicy })
  const wrapper = mount(BookingPolicyPage)
  await flushPromises()
  return wrapper
}

describe('BookingPolicyPage', () => {
  beforeEach(() => {
    resetToasts()
    fetchMock.mockReset()
    saveMock.mockReset()
  })

  it('shows a non-blank loading state, then the loaded values', async () => {
    let resolveFetch: (value: unknown) => void = () => {}
    fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
    const wrapper = mount(BookingPolicyPage)

    expect(wrapper.text()).toContain('Cargando')

    resolveFetch({ kind: 'success', policy: loadedPolicy })
    await flushPromises()

    expect(
      (wrapper.find('input[name="minAdvanceMinutes"]').element as HTMLInputElement).value,
    ).toBe('60')
    expect((wrapper.find('input[name="maxAdvanceDays"]').element as HTMLInputElement).value).toBe(
      '3',
    )
    expect(
      (wrapper.get('input[type="radio"][value="15"]').element as HTMLInputElement).checked,
    ).toBe(true)
    expect(
      (wrapper.find('input[name="cancellationDeadlineMinutes"]').element as HTMLInputElement).value,
    ).toBe('20')
    expect(lateCancellationSwitch(wrapper).attributes('aria-checked')).toBe('true')
    expect(reasonRequiredSwitch(wrapper).attributes('aria-checked')).toBe('true')
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mount(BookingPolicyPage)
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar la configuración')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({ kind: 'success', policy: loadedPolicy })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('input[name="minAdvanceMinutes"]').exists()).toBe(true)
  })

  it('blocks submission and shows a field error for an out-of-range value (client-side validation)', async () => {
    const wrapper = await mountReady()

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('9999')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe un número entero entre 0 y 1440 minutos.')
  })

  it('blocks submission for an incoherent cancellation policy (client-side validation)', async () => {
    const wrapper = await mountReady()

    await lateCancellationSwitch(wrapper).trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(
      'No puedes exigir motivo si el cliente no puede cancelar tarde.',
    )
  })

  it('blocks submission when minAdvanceMinutes reaches the full window', async () => {
    const wrapper = await mountReady()

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('1440')
    await wrapper.get('input[name="maxAdvanceDays"]').setValue('1')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(
      'La anticipación mínima no puede alcanzar ni superar la ventana máxima.',
    )
  })

  it('on success, sends the current versionToken as If-Match and renews it from the response', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({
      kind: 'success',
      policy: { ...loadedPolicy, minAdvanceMinutes: 90, versionToken: 'tok-2' },
    })

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('90')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).toHaveBeenCalledTimes(1)
    expect(saveMock.mock.calls[0]![1]).toBe('tok-1')
    expect(wrapper.text()).toContain('Guardado')

    // Un segundo guardado exitoso debe usar el token RENOVADO, no el
    // original: confirma que la página lo actualizó tras la respuesta.
    saveMock.mockResolvedValueOnce({
      kind: 'success',
      policy: { ...loadedPolicy, minAdvanceMinutes: 100, versionToken: 'tok-3' },
    })
    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('100')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(saveMock.mock.calls[1]![1]).toBe('tok-2')
    // Este recorrido guarda dos veces: cada confirmación es su propio aviso.
    expect(toastState.items.map((item) => item.title)).toEqual([
      'Política de reservas guardada',
      'Política de reservas guardada',
    ])
  })

  it('on a version conflict, shows a distinct message with a reload action and keeps the typed values', async () => {
    const wrapper = await mountReady()
    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('90')
    saveMock.mockResolvedValueOnce({ kind: 'version-conflict' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('La configuración cambió')
    expect(
      (wrapper.find('input[name="minAdvanceMinutes"]').element as HTMLInputElement).value,
    ).toBe('90')

    // La acción "Recargar" reemplaza el formulario con la representación
    // vigente del servidor (CA-093-02: nunca reintentar con el mismo token).
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      policy: { ...loadedPolicy, minAdvanceMinutes: 45, versionToken: 'tok-servidor' },
    })
    await wrapper.get('button:not([type="submit"])').trigger('click')
    await flushPromises()

    expect(
      (wrapper.find('input[name="minAdvanceMinutes"]').element as HTMLInputElement).value,
    ).toBe('45')
  })

  it('on a recoverable network error, keeps the typed values', async () => {
    const wrapper = await mountReady()
    await wrapper.get('input[name="cancellationDeadlineMinutes"]').setValue('30')
    saveMock.mockResolvedValueOnce({ kind: 'network-error' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(
      (wrapper.find('input[name="cancellationDeadlineMinutes"]').element as HTMLInputElement).value,
    ).toBe('30')
  })

  it('on a server validation-error, keeps the typed values', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({ kind: 'validation-error' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos guardar los cambios')
  })

  it('blocks a second submit while the first is still in flight (no double PUT)', async () => {
    const wrapper = await mountReady()
    let resolveSave: (value: unknown) => void = () => {}
    saveMock.mockReturnValueOnce(new Promise((resolve) => (resolveSave = resolve)))

    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).toHaveBeenCalledTimes(1)
    resolveSave({ kind: 'success', policy: loadedPolicy })
    await flushPromises()
  })

  it('shows the unsaved-changes bar only while something differs from the saved values', async () => {
    const wrapper = await mountReady()
    expect(wrapper.text()).not.toContain('Cambios sin guardar')

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('90')
    expect(wrapper.text()).toContain('Cambios sin guardar')
    expect(wrapper.text()).toContain('Ventana de reserva')

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('60')
    expect(wrapper.text()).not.toContain('Cambios sin guardar')
  })

  it('Descartar restores the last saved values and clears local errors', async () => {
    const wrapper = await mountReady()
    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('9999')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Escribe un número entero entre 0 y 1440 minutos.')

    const discard = wrapper.findAll('button').find((button) => button.text() === 'Descartar')!
    await discard.trigger('click')
    await flushPromises()

    expect(
      (wrapper.find('input[name="minAdvanceMinutes"]').element as HTMLInputElement).value,
    ).toBe('60')
    expect(wrapper.text()).not.toContain('Escribe un número entero')
    expect(wrapper.text()).not.toContain('Cambios sin guardar')
  })

  it('stops claiming «Guardado» as soon as new unsaved changes appear', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({
      kind: 'success',
      policy: { ...loadedPolicy, minAdvanceMinutes: 90, versionToken: 'tok-2' },
    })
    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('90')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Guardado')

    await wrapper.get('input[name="minAdvanceMinutes"]').setValue('120')
    expect(wrapper.text()).not.toContain('Los cambios se guardaron correctamente.')
  })

  it('changes the slot grid with the radio group and mirrors it in the live preview', async () => {
    const wrapper = await mountReady()
    expect(wrapper.get('.policy-preview').text()).toContain('una cada 15 min')

    await wrapper.get('input[type="radio"][value="30"]').setValue(true)

    expect(wrapper.get('.policy-preview').text()).toContain('una cada 30 min')
    expect(wrapper.get('.policy-preview').text()).toContain('10:30')
    saveMock.mockResolvedValueOnce({
      kind: 'success',
      policy: { ...loadedPolicy, slotGridMinutes: 30, versionToken: 'tok-2' },
    })
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(saveMock.mock.calls[0]![0]).toMatchObject({ slotGridMinutes: 30 })
  })

  it('describes the cancellation rule in the preview', async () => {
    const wrapper = await mountReady()
    expect(wrapper.get('.policy-preview').text()).toContain('puede cancelar, indicando un motivo')

    await reasonRequiredSwitch(wrapper).trigger('click')
    expect(wrapper.get('.policy-preview').text()).toContain('puede cancelar sin motivo')

    await wrapper.get('input[name="cancellationDeadlineMinutes"]').setValue('0')
    expect(wrapper.get('.policy-preview').text()).toContain('no puede cancelar por su cuenta')
  })

  it('has no axe violations once loaded', async () => {
    const wrapper = await mountReady()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  // heading-order: BaseAlert.vue fija el título de una alerta como `<h4>`
  // (mismo criterio ya documentado en SettingsPage.test.ts): no es un
  // defecto de esta pantalla.
  const axeOptionsWithAlert = { rules: { 'heading-order': { enabled: false } } }

  it('has no axe violations with the version-conflict alert visible', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({ kind: 'version-conflict' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptionsWithAlert)
    expect(results).toHaveNoViolations()
  })

  it('has no axe violations with the saved confirmation visible', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({ kind: 'success', policy: loadedPolicy })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptionsWithAlert)
    expect(results).toHaveNoViolations()
  })
})
