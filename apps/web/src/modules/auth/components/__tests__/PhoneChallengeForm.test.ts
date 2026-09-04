/**
 * Pruebas de PhoneChallengeForm (HU-007, DEC-062; sección del reto reglada
 * del issue #213): solicitar el código, verificarlo con éxito/error, y
 * accesibilidad básica. El cliente HTTP tipado se mockea (mismo patrón que
 * loginApi.test.ts); el recorrido de red real vive en e2e/acceso.spec.ts.
 *
 * El código ya no vive en un único `<input>` (issue #213 adopta `OtpInput`,
 * seis casillas): las pruebas que necesitan leer o escribir el código
 * completo usan los helpers `fillOtp`/`readOtp` de este archivo en vez de
 * `wrapper.find('input').setValue(...)`.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
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

describe('PhoneChallengeForm', () => {
  beforeEach(() => {
    requestChallengeMock.mockReset()
    verifyChallengeMock.mockReset()
    vi.useRealTimers()
  })

  describe('Fase previa (no representada en los mockups, composición actual)', () => {
    it('shows the request-code button initially, no code input yet', () => {
      const wrapper = mountChallenge()
      expect(wrapper.text()).toContain('Verifica tu teléfono')
      expect(wrapper.find('button').text()).toContain('Enviar código')
      expect(wrapper.find('input').exists()).toBe(false)
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

    it('respects the disabled prop on both actions', () => {
      const wrapper = mountChallenge(true)
      expect(wrapper.find('button').attributes('disabled')).toBeDefined()
    })

    it('has no obvious accessibility violations in the initial state', async () => {
      const wrapper = mountChallenge()
      const results = await axe(wrapper.element, axeOptions)
      expect(results.violations).toEqual([])
    })
  })

  describe('Reto con código enviado (mockups 07-12, sección reglada)', () => {
    it('drops the standalone "Verifica tu teléfono" alert and shows the channel section instead', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mountChallenge()

      await wrapper.find('button').trigger('click')
      await flushPromises()

      expect(wrapper.find('input').exists()).toBe(true)
      expect(wrapper.text()).not.toContain('Verifica tu teléfono')
      expect(wrapper.text()).toContain('Si tu cuenta existe')
      // Canal real vigente (DEC-081: el backend solo resuelve WhatsApp hoy).
      expect(wrapper.text()).toContain('WhatsApp')
    })

    it('renders exactly six OTP slots for the code', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mountChallenge()
      await wrapper.find('button').trigger('click')
      await flushPromises()

      expect(wrapper.findAll('input')).toHaveLength(6)
    })

    it('never shows two primary actions at once: credentials stay owned by LoginPage, this section keeps a single primary button', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mountChallenge()
      await wrapper.find('button').trigger('click')
      await flushPromises()

      const primaryButtons = wrapper.findAll('.base-button--primary')
      expect(primaryButtons).toHaveLength(1)
      expect(primaryButtons[0].text()).toContain('Verificar código')
    })

    it('verifies a correct code and emits verified', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      verifyChallengeMock.mockResolvedValueOnce({ kind: 'verified' })
      const wrapper = mountChallenge()

      await wrapper.find('button').trigger('click')
      await flushPromises()

      await fillOtp(wrapper, '482913')
      const verifyButton = wrapper
        .findAll('button')
        .find((b) => b.text().includes('Verificar código'))
      await verifyButton?.trigger('click')
      await flushPromises()

      expect(verifyChallengeMock).toHaveBeenCalledWith('barbero@ejemplo.test', '482913')
      expect(wrapper.emitted('verified')).toHaveLength(1)
    })

    it('shows an inline error and keeps all six slots filled on an invalid code, without emitting verified', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      verifyChallengeMock.mockResolvedValueOnce({ kind: 'invalid-code' })
      const wrapper = mountChallenge()

      await wrapper.find('button').trigger('click')
      await flushPromises()
      await fillOtp(wrapper, '000000')
      const verifyButton = wrapper
        .findAll('button')
        .find((b) => b.text().includes('Verificar código'))
      await verifyButton?.trigger('click')
      await flushPromises()

      expect(wrapper.emitted('verified')).toBeUndefined()
      expect(wrapper.text()).toContain('El código no es válido')
      expect(readOtp(wrapper)).toBe('000000')
    })

    it('disables the verify button until exactly 6 digits are entered', async () => {
      requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mountChallenge()
      await wrapper.find('button').trigger('click')
      await flushPromises()

      const verifyButton = () =>
        wrapper.findAll('button').find((b) => b.text().includes('Verificar código'))
      expect(verifyButton()?.attributes('disabled')).toBeDefined()

      await fillOtp(wrapper, '12345')
      expect(verifyButton()?.attributes('disabled')).toBeDefined()

      await fillOtp(wrapper, '123456')
      expect(verifyButton()?.attributes('disabled')).toBeUndefined()
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
})
