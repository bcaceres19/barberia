/**
 * Pruebas de integración de RecoveryPage (HU-011): indicador de paso
 * (CA-011-01), recorrido completo de los tres pasos, foco en el
 * encabezado en cada transición (CA-011-07) y confirmación final con
 * salida al acceso (CA-011-05). Las funciones del API se sustituyen por
 * dobles; el recorrido real contra el API real vive en
 * `e2e/recuperacion.spec.ts`.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'

const requestRecoveryMock = vi.hoisted(() => vi.fn())
const verifyRecoveryMock = vi.hoisted(() => vi.fn())
const resetRecoveryPasswordMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/recoveryApi', () => ({
  requestRecovery: requestRecoveryMock,
  verifyRecovery: verifyRecoveryMock,
  resetRecoveryPassword: resetRecoveryPasswordMock,
}))

const { default: RecoveryPage } = await import('../RecoveryPage.vue')

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/acceso', name: 'acceso', component: { template: '<div>acceso</div>' } },
      { path: '/recuperar-acceso', name: 'recuperar-acceso', component: RecoveryPage },
    ],
  })
}

async function fillOtp(wrapper: ReturnType<typeof mount>, code: string) {
  // `OtpInput` conserva un valor lógico de seis dígitos, aunque lo renderiza
  // en seis casillas accesibles. Escribir/pegar en la primera distribuye el
  // valor completo, igual que en el recorrido real.
  await wrapper.get('.otp-input__slot').setValue(code)
}

// El paso 1 no muestra campo hasta elegir el canal (DEC-092, CA-011-09).
async function chooseChannel(wrapper: ReturnType<typeof mount>, label: 'WhatsApp' | 'Correo') {
  const button = wrapper.findAll('button.recovery-channel').find((b) => b.text() === label)
  await button?.trigger('click')
}

async function mountPage(options: { attachToBody?: boolean } = {}) {
  const router = buildRouter()
  await router.push({ name: 'recuperar-acceso' })
  await router.isReady()
  const wrapper = mount(RecoveryPage, {
    attachTo: options.attachToBody ? document.body : undefined,
    global: {
      plugins: [router],
      stubs: { RouterLink: { props: ['to'], template: '<a><slot /></a>' } },
    },
  })
  return { wrapper, router }
}

describe('RecoveryPage', () => {
  beforeEach(() => {
    requestRecoveryMock.mockReset()
    verifyRecoveryMock.mockReset()
    resetRecoveryPasswordMock.mockReset()
  })

  it('shows "Paso 1 de 3" on the first step (CA-011-01)', async () => {
    const { wrapper } = await mountPage()
    expect(wrapper.text()).toContain('Paso 1 de 3')
  })

  it('completes the full journey: request, verify and reset, ending in the confirmation (CA-011-05)', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyRecoveryMock.mockResolvedValueOnce({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    })
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'success' })
    const { wrapper, router } = await mountPage()

    await chooseChannel(wrapper, 'Correo')
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Paso 2 de 3')

    await fillOtp(wrapper, '482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Paso 3 de 3')
    expect(wrapper.text()).toContain('+57 *** *** 12')

    await wrapper.get('input[name="newPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('input[name="confirmPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('cerramos todas tus sesiones activas')
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.name).toBe('acceso')
  })

  it('completes the journey by WhatsApp, reusing the same channel and phone in steps 2 and 3 (DEC-092, DEC-093)', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyRecoveryMock.mockResolvedValueOnce({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 67',
      maskedEmail: 'b***@c***.test',
    })
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'success' })
    const { wrapper } = await mountPage()
    const target = { channel: 'whatsapp', value: '+573001234567' }

    await chooseChannel(wrapper, 'WhatsApp')
    await wrapper.get('input[name="whatsapp"]').setValue('+57 300 123 4567')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(requestRecoveryMock).toHaveBeenCalledWith(target)
    expect(wrapper.text()).toContain('Paso 2 de 3')
    expect(wrapper.text()).toContain('por WhatsApp')

    await fillOtp(wrapper, '482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(verifyRecoveryMock).toHaveBeenCalledWith(target, '482913')
    expect(wrapper.text()).toContain('Paso 3 de 3')

    await wrapper.get('input[name="newPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('input[name="confirmPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(resetRecoveryPasswordMock).toHaveBeenCalledWith(
      target,
      'token-abc',
      'contraseña-nueva-valida',
    )
    expect(wrapper.text()).toContain('cerramos todas tus sesiones activas')
  })

  it('goes back from step 2 to step 1 to choose another channel, with nothing chosen yet (DP-SEG-16)', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    const { wrapper } = await mountPage()

    await chooseChannel(wrapper, 'Correo')
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Paso 2 de 3')

    const change = wrapper.findAll('button').find((b) => b.text().includes('Elegir otro canal'))
    await change?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Paso 1 de 3')
    expect(wrapper.findAll('button.recovery-channel')).toHaveLength(2)
    expect(wrapper.find('input').exists()).toBe(false)
  })

  it('moves focus to the step heading on each transition (CA-011-07)', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    const { wrapper } = await mountPage({ attachToBody: true })

    await chooseChannel(wrapper, 'Correo')
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(document.activeElement).toBe(wrapper.get('h1').element)
    wrapper.unmount()
  })

  it('restarts to step 1 when the reset step reports an invalid token', async () => {
    requestRecoveryMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyRecoveryMock.mockResolvedValueOnce({
      kind: 'verified',
      resetToken: 'token-abc',
      maskedPhone: '+57 *** *** 12',
      maskedEmail: 'b***@c***.test',
    })
    resetRecoveryPasswordMock.mockResolvedValueOnce({ kind: 'invalid-token' })
    const { wrapper } = await mountPage()

    await chooseChannel(wrapper, 'Correo')
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await fillOtp(wrapper, '482913')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await wrapper.get('input[name="newPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('input[name="confirmPassword"]').setValue('contraseña-nueva-valida')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const restartButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Solicitar de nuevo'))
    await restartButton?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Paso 1 de 3')
  })

  it('always offers a way back to /acceso before the final step', async () => {
    const { wrapper } = await mountPage()
    expect(wrapper.text()).toContain('Volver al acceso')
  })

  it('has no obvious accessibility violations on the first step', async () => {
    const { wrapper } = await mountPage()
    const results = await axe(wrapper.element, axeOptions)
    expect(results.violations).toEqual([])
  })
})
