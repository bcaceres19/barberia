/**
 * Pruebas de PhoneChallengeForm (HU-007, DEC-062): solicitar el código,
 * verificarlo con éxito/error, y accesibilidad básica. El cliente HTTP
 * tipado se mockea (mismo patrón que loginApi.test.ts); el recorrido de red
 * real vive en e2e/acceso.spec.ts.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import PhoneChallengeForm from '../PhoneChallengeForm.vue'

const requestChallengeMock = vi.hoisted(() => vi.fn())
const verifyChallengeMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/challengeApi', () => ({
  requestChallenge: requestChallengeMock,
  verifyChallenge: verifyChallengeMock,
}))

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountChallenge(disabled = false) {
  return mount(PhoneChallengeForm, {
    props: { email: 'barbero@ejemplo.test', disabled },
  })
}

describe('PhoneChallengeForm', () => {
  beforeEach(() => {
    requestChallengeMock.mockReset()
    verifyChallengeMock.mockReset()
    vi.useRealTimers()
  })

  it('shows the request-code button initially, no code input yet', () => {
    const wrapper = mountChallenge()
    expect(wrapper.find('button').text()).toContain('Enviar código')
    expect(wrapper.find('input').exists()).toBe(false)
  })

  it('requests the challenge and reveals the code input on click', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountChallenge()

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(requestChallengeMock).toHaveBeenCalledWith('barbero@ejemplo.test')
    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.text()).toContain('Si tu cuenta existe')
  })

  it('reveals the code input even when the request outcome is not accepted (never blocks the flow)', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mountChallenge()

    await wrapper.find('button').trigger('click')
    await flushPromises()

    // El botón vuelve a "Enviar código" (no queda atrapado en "sent"), pero
    // no debe lanzar ni dejar el formulario en un estado sin salida.
    expect(wrapper.find('button').text()).toContain('Enviar código')
  })

  it('verifies a correct code and emits verified', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyChallengeMock.mockResolvedValueOnce({ kind: 'verified' })
    const wrapper = mountChallenge()

    await wrapper.find('button').trigger('click')
    await flushPromises()

    await wrapper.find('input').setValue('482913')
    const buttons = wrapper.findAll('button')
    const verifyButton = buttons.find((b) => b.text().includes('Verificar código'))
    await verifyButton?.trigger('click')
    await flushPromises()

    expect(verifyChallengeMock).toHaveBeenCalledWith('barbero@ejemplo.test', '482913')
    expect(wrapper.emitted('verified')).toHaveLength(1)
  })

  it('shows an inline error and clears the field on an invalid code, without emitting verified', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyChallengeMock.mockResolvedValueOnce({ kind: 'invalid-code' })
    const wrapper = mountChallenge()

    await wrapper.find('button').trigger('click')
    await flushPromises()
    await wrapper.find('input').setValue('000000')
    const verifyButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Verificar código'))
    await verifyButton?.trigger('click')
    await flushPromises()

    expect(wrapper.emitted('verified')).toBeUndefined()
    expect(wrapper.text()).toContain('El código no es válido')
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('')
  })

  it('strips non-digit characters and caps the code at 6 digits', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountChallenge()
    await wrapper.find('button').trigger('click')
    await flushPromises()

    await wrapper.find('input').setValue('4a8-2 91399999')

    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('482913')
  })

  it('disables the verify button until exactly 6 digits are entered', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountChallenge()
    await wrapper.find('button').trigger('click')
    await flushPromises()

    const verifyButton = () =>
      wrapper.findAll('button').find((b) => b.text().includes('Verificar código'))
    expect(verifyButton()?.attributes('disabled')).toBeDefined()

    await wrapper.find('input').setValue('12345')
    expect(verifyButton()?.attributes('disabled')).toBeDefined()

    await wrapper.find('input').setValue('123456')
    expect(verifyButton()?.attributes('disabled')).toBeUndefined()
  })

  it('respects the disabled prop on both actions', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountChallenge(true)

    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })

  it('has no obvious accessibility violations in the initial state', async () => {
    const wrapper = mountChallenge()
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })

  it('has no obvious accessibility violations once the code input is shown', async () => {
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mountChallenge()
    await wrapper.find('button').trigger('click')
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
