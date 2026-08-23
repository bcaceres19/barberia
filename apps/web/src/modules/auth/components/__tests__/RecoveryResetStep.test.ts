/**
 * Pruebas de RecoveryResetStep (HU-011, paso 3/3): política explicada antes
 * de escribir (CA-011-06), confirmación del destino enmascarado
 * (CA-011-02), éxito, token vencido/consumido y violación de política.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import RecoveryResetStep from '../RecoveryResetStep.vue'

const resetRecoveryPasswordMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/recoveryApi', () => ({ resetRecoveryPassword: resetRecoveryPasswordMock }))

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountStep() {
  return mount(RecoveryResetStep, {
    props: {
      email: 'barbero@ejemplo.test',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    },
  })
}

async function fillAndSubmit(
  wrapper: ReturnType<typeof mount>,
  password = 'contraseña-nueva-valida',
  confirm = password,
) {
  await wrapper.get('input[name="newPassword"]').setValue(password)
  await wrapper.get('input[name="confirmPassword"]').setValue(confirm)
  await wrapper.get('form').trigger('submit')
}

describe('RecoveryResetStep', () => {
  beforeEach(() => resetRecoveryPasswordMock.mockReset())

  it('confirms the masked destination and explains the password policy before typing (CA-011-02, CA-011-06)', () => {
    const wrapper = mountStep()
    expect(wrapper.text()).toContain('+57 *** *** 12')
    expect(wrapper.text()).toContain('b***@c***.test')
    expect(wrapper.text()).toContain('entre 10 y 128 caracteres')
  })

  it('rejects a password equal to the email without calling the API', async () => {
    const wrapper = mountStep()

    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'barbero@ejemplo.test')

    expect(resetRecoveryPasswordMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('igual a tu correo')
  })

  it('rejects mismatched confirmation without calling the API', async () => {
    const wrapper = mountStep()

    await fillAndSubmit(wrapper, 'contraseña-nueva-valida', 'otra-contraseña-distinta')

    expect(resetRecoveryPasswordMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('no coinciden')
  })

  it('emits done on success (CA-011-05)', async () => {
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'success' })
    const wrapper = mountStep()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(resetRecoveryPasswordMock).toHaveBeenCalledWith(
      'barbero@ejemplo.test',
      'token-abc',
      'contraseña-nueva-valida',
    )
    expect(wrapper.emitted('done')).toHaveLength(1)
  })

  it('offers a way to restart on invalid-token, hiding the password fields', async () => {
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'invalid-token' })
    const wrapper = mountStep()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.find('input[name="newPassword"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('venció')

    const restartButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Solicitar de nuevo'))
    await restartButton?.trigger('click')
    expect(wrapper.emitted('restart')).toHaveLength(1)
  })

  it('shows a policy-violation message without advancing', async () => {
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'policy-violation' })
    const wrapper = mountStep()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.emitted('done')).toBeUndefined()
    expect(wrapper.text()).toContain('no cumple la política')
  })

  it('has no obvious accessibility violations', async () => {
    const wrapper = mountStep()
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
