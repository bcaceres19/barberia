/**
 * Pruebas de RecoveryVerifyStep (HU-011, paso 2/3): verificación, cooldown
 * de reenvío con cuenta regresiva (CA-011-04) y error uniforme para código
 * incorrecto/vencido/agotado (CA-011-03 según el contrato real de HU-008,
 * ver CT-007 en docs/00-control/contradicciones.md).
 *
 * El código ya no vive en un único `input[name="code"]` (issue #213 adopta
 * `OtpInput`, seis casillas sin `name`): las pruebas que necesitan leer o
 * escribir el código completo usan los helpers `fillOtp`/`readOtp`.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import RecoveryVerifyStep from '../RecoveryVerifyStep.vue'

const requestRecoveryMock = vi.hoisted(() => vi.fn())
const verifyRecoveryMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/recoveryApi', () => ({
  requestRecovery: requestRecoveryMock,
  verifyRecovery: verifyRecoveryMock,
}))

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountStep() {
  return mount(RecoveryVerifyStep, { props: { email: 'barbero@ejemplo.test' } })
}

// `OtpInput` distribuye el código completo cuando el primer slot recibe un
// valor con más de un carácter (mismo camino que un pegado real,
// `OtpInput.vue` `onInput`).
async function fillOtp(wrapper: VueWrapper, code: string) {
  await wrapper.findAll('input')[0].setValue(code)
}

function readOtp(wrapper: VueWrapper): string {
  return wrapper
    .findAll('input')
    .map((input) => (input.element as HTMLInputElement).value)
    .join('')
}

describe('RecoveryVerifyStep', () => {
  beforeEach(() => {
    requestRecoveryMock.mockReset()
    verifyRecoveryMock.mockReset()
    vi.useRealTimers()
  })
  afterEach(() => vi.useRealTimers())

  it('shows the generic non-committal sent message on mount', () => {
    const wrapper = mountStep()
    expect(wrapper.text()).toContain('Si tu cuenta existe')
  })

  it('renders exactly six OTP slots for the code', () => {
    const wrapper = mountStep()
    expect(wrapper.findAll('input')).toHaveLength(6)
  })

  it('verifies a correct code and emits advance with resetToken/masked destination', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    })
    const wrapper = mountStep()

    await fillOtp(wrapper, '482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(verifyRecoveryMock).toHaveBeenCalledWith('barbero@ejemplo.test', '482913')
    expect(wrapper.emitted('advance')).toEqual([
      [{ resetToken: 'token-abc', maskedPhone: '+57 *** *** 12', maskedEmail: 'b***@c***.test' }],
    ])
  })

  it('shows one uniform message under six filled error slots on invalid-code (incorrect/expired/exhausted, CA-011-03)', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({ kind: 'invalid-code' })
    const wrapper = mountStep()

    await fillOtp(wrapper, '000000')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.emitted('advance')).toBeUndefined()
    // El error vive bajo las casillas, sin una alerta global aparte
    // (mockup 06-verificacion-codigo-invalido, trabajo requerido §6).
    expect(wrapper.text()).toContain('El código no es correcto o ya venció')
    expect(wrapper.findAll('[role="alert"].base-alert')).toHaveLength(0)
    expect(readOtp(wrapper)).toBe('000000')
  })

  it('shows a real transport error after the actions group, without clearing the code', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountStep()

    await fillOtp(wrapper, '482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(readOtp(wrapper)).toBe('482913')
  })

  it('blocks submission with a field error when fewer than 6 digits are entered', async () => {
    const wrapper = mountStep()

    await fillOtp(wrapper, '12345')
    await wrapper.get('form').trigger('submit')

    expect(wrapper.text()).toContain('6 dígitos')
    expect(verifyRecoveryMock).not.toHaveBeenCalled()
  })

  it('starts the resend cooldown immediately on mount and disables Reenviar (CA-011-04)', () => {
    const wrapper = mountStep()
    const resendButton = wrapper.findAll('button').find((b) => b.text().includes('Reenviar'))
    expect(resendButton?.attributes('disabled')).toBeDefined()
    expect(resendButton?.text()).toContain('60 s')
  })

  it('counts the cooldown down and re-enables Reenviar once it reaches zero', async () => {
    vi.useFakeTimers()
    const wrapper = mountStep()

    await vi.advanceTimersByTimeAsync(60_000)

    const resendButton = wrapper.findAll('button').find((b) => b.text().includes('Reenviar'))
    expect(resendButton?.attributes('disabled')).toBeUndefined()
    expect(resendButton?.text()).toBe('Reenviar código')
  })

  it('resends and restarts the cooldown when Reenviar is pressed after it elapses', async () => {
    vi.useFakeTimers()
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountStep()
    await vi.advanceTimersByTimeAsync(60_000)

    const resendButton = () => wrapper.findAll('button').find((b) => b.text().includes('Reenviar'))
    await resendButton()?.trigger('click')
    await flushPromises()

    expect(requestRecoveryMock).toHaveBeenCalledWith('barbero@ejemplo.test')
    expect(resendButton()?.attributes('disabled')).toBeDefined()
  })

  it('has no obvious accessibility violations', async () => {
    const wrapper = mountStep()
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
