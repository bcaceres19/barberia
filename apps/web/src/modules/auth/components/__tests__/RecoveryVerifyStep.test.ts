/**
 * Pruebas de RecoveryVerifyStep (HU-011, paso 2/3): verificación, cooldown
 * de reenvío con cuenta regresiva (CA-011-04) y error uniforme para código
 * incorrecto/vencido/agotado (CA-011-03 según el contrato real de HU-008,
 * ver CT-007 en docs/00-control/contradicciones.md).
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
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

  it('strips non-digit characters and caps the code at 6 digits', async () => {
    const wrapper = mountStep()
    await wrapper.get('input[name="code"]').setValue('4a8-2 91399999')
    expect((wrapper.get('input[name="code"]').element as HTMLInputElement).value).toBe('482913')
  })

  it('verifies a correct code and emits advance with resetToken/masked destination', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    })
    const wrapper = mountStep()

    await wrapper.get('input[name="code"]').setValue('482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(verifyRecoveryMock).toHaveBeenCalledWith('barbero@ejemplo.test', '482913')
    expect(wrapper.emitted('advance')).toEqual([
      [{ resetToken: 'token-abc', maskedPhone: '+57 *** *** 12', maskedEmail: 'b***@c***.test' }],
    ])
  })

  it('shows one uniform message and clears the field on invalid-code (incorrect/expired/exhausted, CA-011-03)', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({ kind: 'invalid-code' })
    const wrapper = mountStep()

    await wrapper.get('input[name="code"]').setValue('000000')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.emitted('advance')).toBeUndefined()
    expect(wrapper.text()).toContain('El código no es válido')
    expect((wrapper.get('input[name="code"]').element as HTMLInputElement).value).toBe('')
  })

  it('shows a real transport error on network-error, without clearing the code', async () => {
    verifyRecoveryMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountStep()

    await wrapper.get('input[name="code"]').setValue('482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect((wrapper.get('input[name="code"]').element as HTMLInputElement).value).toBe('482913')
  })

  it('blocks submission with a field error when fewer than 6 digits are entered', async () => {
    const wrapper = mountStep()

    await wrapper.get('input[name="code"]').setValue('12345')
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
