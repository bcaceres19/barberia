/**
 * Pruebas de RecoveryRequestStep (HU-011, paso 1/3): validación de forma,
 * envío, avance no enumerable y errores de transporte reales. `requestRecovery`
 * se mockea (mismo patrón que `LoginForm.test.ts`); el recorrido de red real
 * vive en `e2e/recuperacion.spec.ts`.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import RecoveryRequestStep from '../RecoveryRequestStep.vue'

const requestRecoveryMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/recoveryApi', () => ({ requestRecovery: requestRecoveryMock }))

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

async function fillAndSubmit(wrapper: ReturnType<typeof mount>, email = 'barbero@ejemplo.test') {
  await wrapper.get('input[name="email"]').setValue(email)
  await wrapper.get('form').trigger('submit')
}

describe('RecoveryRequestStep', () => {
  beforeEach(() => requestRecoveryMock.mockReset())

  it('shows a field error and does not submit when the email is empty', async () => {
    const wrapper = mount(RecoveryRequestStep)

    await wrapper.get('form').trigger('submit')

    expect(requestRecoveryMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe tu correo')
  })

  it('advances with the trimmed email on accepted (CA-011-01, non-enumerable)', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    const wrapper = mount(RecoveryRequestStep)

    await fillAndSubmit(wrapper, '  barbero@ejemplo.test  ')
    await flushPromises()

    expect(requestRecoveryMock).toHaveBeenCalledWith('barbero@ejemplo.test')
    expect(wrapper.emitted('advance')).toEqual([[{ email: 'barbero@ejemplo.test' }]])
  })

  it('shows a real transport error and does not advance on network-error', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mount(RecoveryRequestStep)

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.emitted('advance')).toBeUndefined()
    expect(wrapper.text()).toContain('No pudimos conectar')
  })

  it('shows a safe unexpected-error message and does not advance', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'unexpected-error', requestId: 'req-1' })
    const wrapper = mount(RecoveryRequestStep)

    await fillAndSubmit(wrapper)
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
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')

    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')

    expect(requestRecoveryMock).toHaveBeenCalledTimes(1)
    resolve({ kind: 'accepted' })
    await flushPromises()
  })

  it('has no obvious accessibility violations', async () => {
    const wrapper = mount(RecoveryRequestStep)
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
