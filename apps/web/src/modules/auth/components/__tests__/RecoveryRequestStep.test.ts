/**
 * Pruebas de RecoveryRequestStep (HU-011, paso 1/3, DEC-092): elección de
 * canal (WhatsApp o correo), mensaje solo del canal elegido, validación de
 * forma del valor, envío, avance no enumerable y errores de transporte reales.
 * `requestRecovery` se mockea (mismo patrón que `LoginForm.test.ts`); el
 * recorrido de red real vive en `e2e/recuperacion.spec.ts`.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import RecoveryRequestStep from '../RecoveryRequestStep.vue'

const requestRecoveryMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/recoveryApi', () => ({ requestRecovery: requestRecoveryMock }))

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

type Wrapper = ReturnType<typeof mount>

function channelButton(wrapper: Wrapper, label: 'WhatsApp' | 'Correo') {
  const button = wrapper.findAll('button.recovery-channel').find((b) => b.text() === label)
  if (!button) throw new Error(`Botón de canal no encontrado: ${label}`)
  return button
}

async function chooseEmailAndSubmit(wrapper: Wrapper, email = 'barbero@ejemplo.test') {
  await channelButton(wrapper, 'Correo').trigger('click')
  await wrapper.get('input[name="email"]').setValue(email)
  await wrapper.get('form').trigger('submit')
}

describe('RecoveryRequestStep', () => {
  beforeEach(() => requestRecoveryMock.mockReset())

  describe('choosing the channel (CA-011-09, CA-011-10)', () => {
    it('shows two channel buttons and neither field nor submit until one is chosen', () => {
      const wrapper = mount(RecoveryRequestStep)

      const buttons = wrapper.findAll('button.recovery-channel')
      expect(buttons.map((b) => b.text())).toEqual(['WhatsApp', 'Correo'])
      expect(buttons.map((b) => b.attributes('aria-pressed'))).toEqual(['false', 'false'])
      expect(wrapper.find('input').exists()).toBe(false)
      expect(wrapper.find('button[type="submit"]').exists()).toBe(false)
      expect(wrapper.get('[aria-live="polite"]').text()).toContain(
        'Elige por dónde quieres recibir tu código',
      )
    })

    it('shows only WhatsApp in the message and asks for the number when WhatsApp is chosen', async () => {
      const wrapper = mount(RecoveryRequestStep)

      await channelButton(wrapper, 'WhatsApp').trigger('click')

      const message = wrapper.get('[aria-live="polite"]').text()
      expect(message).toContain('número de WhatsApp')
      expect(message).toContain('por WhatsApp')
      expect(message.toLowerCase()).not.toContain('correo')
      const input = wrapper.get('input[name="whatsapp"]')
      expect(input.attributes('type')).toBe('tel')
      expect(input.attributes('autocomplete')).toBe('tel')
      expect(wrapper.find('input[name="email"]').exists()).toBe(false)
      expect(channelButton(wrapper, 'WhatsApp').attributes('aria-pressed')).toBe('true')
      expect(channelButton(wrapper, 'Correo').attributes('aria-pressed')).toBe('false')
      expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
    })

    it('shows only the email in the message and asks for it when Correo is chosen', async () => {
      const wrapper = mount(RecoveryRequestStep)

      await channelButton(wrapper, 'Correo').trigger('click')

      const message = wrapper.get('[aria-live="polite"]').text()
      expect(message).toContain('correo de tu cuenta')
      expect(message).toContain('por correo')
      expect(message).not.toContain('WhatsApp')
      const input = wrapper.get('input[name="email"]')
      expect(input.attributes('type')).toBe('email')
      expect(wrapper.find('input[name="whatsapp"]').exists()).toBe(false)
      expect(channelButton(wrapper, 'Correo').attributes('aria-pressed')).toBe('true')
    })

    it('replaces message and field when switching channel and keeps what was typed in each', async () => {
      const wrapper = mount(RecoveryRequestStep)

      await channelButton(wrapper, 'WhatsApp').trigger('click')
      await wrapper.get('input[name="whatsapp"]').setValue('+57 300 123 4567')
      await channelButton(wrapper, 'Correo').trigger('click')
      await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
      await channelButton(wrapper, 'WhatsApp').trigger('click')

      expect((wrapper.get('input[name="whatsapp"]').element as HTMLInputElement).value).toBe(
        '+57 300 123 4567',
      )
      await channelButton(wrapper, 'Correo').trigger('click')
      expect((wrapper.get('input[name="email"]').element as HTMLInputElement).value).toBe(
        'barbero@ejemplo.test',
      )
    })

    it('clears the field error of the previous channel when switching', async () => {
      const wrapper = mount(RecoveryRequestStep)

      await channelButton(wrapper, 'WhatsApp').trigger('click')
      await wrapper.get('form').trigger('submit')
      expect(wrapper.text()).toContain('Escribe tu número de WhatsApp')

      await channelButton(wrapper, 'Correo').trigger('click')

      expect(wrapper.text()).not.toContain('Escribe tu número de WhatsApp')
      expect(wrapper.text()).not.toContain('Escribe tu correo')
    })

    it('keeps the focus on the pressed channel button (the message is announced, focus does not move)', async () => {
      const wrapper = mount(RecoveryRequestStep, { attachTo: document.body })
      const button = channelButton(wrapper, 'WhatsApp')
      ;(button.element as HTMLButtonElement).focus()

      await button.trigger('click')

      expect(document.activeElement).toBe(button.element)
      wrapper.unmount()
    })
  })

  describe('validation', () => {
    it('shows a field error and does not submit when the email is empty', async () => {
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'Correo').trigger('click')

      await wrapper.get('form').trigger('submit')

      expect(requestRecoveryMock).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain('Escribe tu correo')
    })

    it('shows a field error and does not submit when the phone is empty', async () => {
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'WhatsApp').trigger('click')

      await wrapper.get('form').trigger('submit')

      expect(requestRecoveryMock).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain('Escribe tu número de WhatsApp')
    })

    it('asks for the international prefix when the number is not in E.164', async () => {
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'WhatsApp').trigger('click')
      await wrapper.get('input[name="whatsapp"]').setValue('3001234567')

      await wrapper.get('form').trigger('submit')

      expect(requestRecoveryMock).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain('indicativo de tu país')
    })
  })

  describe('sending', () => {
    it('advances with the trimmed email on accepted (CA-011-01, non-enumerable)', async () => {
      requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mount(RecoveryRequestStep)

      await chooseEmailAndSubmit(wrapper, '  barbero@ejemplo.test  ')
      await flushPromises()

      const target = { channel: 'email', value: 'barbero@ejemplo.test' }
      expect(requestRecoveryMock).toHaveBeenCalledWith(target)
      expect(wrapper.emitted('advance')).toEqual([[{ target }]])
    })

    it('advances with the phone without presentation separators on accepted', async () => {
      requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'WhatsApp').trigger('click')
      await wrapper.get('input[name="whatsapp"]').setValue(' +57 (300) 123-4567 ')

      await wrapper.get('form').trigger('submit')
      await flushPromises()

      const target = { channel: 'whatsapp', value: '+573001234567' }
      expect(requestRecoveryMock).toHaveBeenCalledWith(target)
      expect(wrapper.emitted('advance')).toEqual([[{ target }]])
    })

    it('shows a real transport error and does not advance on network-error', async () => {
      requestRecoveryMock.mockResolvedValueOnce({ kind: 'network-error' })
      const wrapper = mount(RecoveryRequestStep)

      await chooseEmailAndSubmit(wrapper)
      await flushPromises()

      expect(wrapper.emitted('advance')).toBeUndefined()
      expect(wrapper.text()).toContain('No pudimos conectar')
    })

    it('shows a safe unexpected-error message and does not advance', async () => {
      requestRecoveryMock.mockResolvedValueOnce({ kind: 'unexpected-error', requestId: 'req-1' })
      const wrapper = mount(RecoveryRequestStep)

      await chooseEmailAndSubmit(wrapper)
      await flushPromises()

      expect(wrapper.emitted('advance')).toBeUndefined()
      expect(wrapper.text()).toContain('Ocurrió un error inesperado')
    })

    it('sends exactly one request on double submit while one is in flight', async () => {
      let resolve: (value: { kind: 'accepted' }) => void = () => {}
      requestRecoveryMock.mockImplementationOnce(
        () =>
          new Promise((res) => {
            resolve = res
          }),
      )
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'Correo').trigger('click')
      await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')

      await wrapper.get('form').trigger('submit')
      await wrapper.get('form').trigger('submit')

      expect(requestRecoveryMock).toHaveBeenCalledTimes(1)
      resolve({ kind: 'accepted' })
      await flushPromises()
    })
  })

  describe('accessibility', () => {
    it('has no obvious accessibility violations before choosing a channel', async () => {
      const wrapper = mount(RecoveryRequestStep)
      const results = await axe(wrapper.element, axeOptions)
      expect(results.violations).toEqual([])
    })

    it('has no obvious accessibility violations with WhatsApp chosen', async () => {
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'WhatsApp').trigger('click')
      const results = await axe(wrapper.element, axeOptions)
      expect(results.violations).toEqual([])
    })

    it('has no obvious accessibility violations with Correo chosen', async () => {
      const wrapper = mount(RecoveryRequestStep)
      await channelButton(wrapper, 'Correo').trigger('click')
      const results = await axe(wrapper.element, axeOptions)
      expect(results.violations).toEqual([])
    })
  })
})
