/**
 * LoginOtpPanel (DEC-139): elegir por dónde llega el código de acceso.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchMock = vi.hoisted(() => vi.fn())
const saveMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/loginOtpApi', () => ({
  fetchLoginOtpChannel: fetchMock,
  saveLoginOtpChannel: saveMock,
}))

const { default: LoginOtpPanel } = await import('../LoginOtpPanel.vue')

function mountPanel() {
  return mount(LoginOtpPanel, { attachTo: document.body })
}

const radios = (w: ReturnType<typeof mountPanel>) =>
  w.findAll<HTMLInputElement>('input[type="radio"]')

describe('LoginOtpPanel', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    saveMock.mockReset()
    fetchMock.mockResolvedValue({ kind: 'success', channel: 'email' })
    document.body.innerHTML = ''
  })

  it('announces the wait and then shows the three channels with the current one checked', async () => {
    fetchMock.mockResolvedValue({ kind: 'success', channel: 'both' })
    const wrapper = mountPanel()
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)

    await flushPromises()

    const group = wrapper.get('[role="radiogroup"]')
    expect(group.attributes('aria-label')).toBe('Canal del código de acceso')
    expect(radios(wrapper).map((r) => r.element.value)).toEqual(['email', 'whatsapp', 'both'])
    expect(radios(wrapper).map((r) => r.element.checked)).toEqual([false, false, true])
    expect(wrapper.text()).toContain('Correo y WhatsApp')
    expect(wrapper.text()).toContain('cada persona elige su propio canal')
  })

  it('saves the new channel on choosing it and says so (the page never asks for a destination)', async () => {
    saveMock.mockResolvedValue({ kind: 'success', channel: 'whatsapp' })
    const wrapper = mountPanel()
    await flushPromises()

    await radios(wrapper)[1]!.setValue(true)
    await flushPromises()

    expect(saveMock).toHaveBeenCalledTimes(1)
    expect(saveMock).toHaveBeenCalledWith('whatsapp')
    expect(radios(wrapper)[1]!.element.checked).toBe(true)
    expect(wrapper.text()).toContain('Guardado.')
    expect(
      wrapper.find('input[type="email"], input[type="tel"], input[type="text"]').exists(),
    ).toBe(false)
  })

  it('keeps the previous option and explains when saving fails', async () => {
    saveMock.mockResolvedValue({ kind: 'network-error' })
    const wrapper = mountPanel()
    await flushPromises()

    await radios(wrapper)[2]!.setValue(true)
    await flushPromises()

    expect(radios(wrapper).map((r) => r.element.checked)).toEqual([true, false, false])
    expect(wrapper.text()).toContain('se mantiene la opción anterior')
  })

  it('blocks the group while saving and ignores a second change', async () => {
    let finish: (v: unknown) => void = () => {}
    saveMock.mockReturnValue(new Promise((r) => (finish = r)))
    const wrapper = mountPanel()
    await flushPromises()

    await radios(wrapper)[1]!.setValue(true)
    expect(wrapper.text()).toContain('Guardando…')
    expect(radios(wrapper).every((r) => r.element.disabled)).toBe(true)

    finish({ kind: 'success', channel: 'whatsapp' })
    await flushPromises()
    expect(saveMock).toHaveBeenCalledTimes(1)
    expect(radios(wrapper).every((r) => !r.element.disabled)).toBe(true)
  })

  it.each([
    ['network-error', 'No pudimos conectar'],
    ['unexpected-error', 'No pudimos cargar esta opción'],
  ])('offers a retry when loading ends in %s', async (kind, title) => {
    fetchMock.mockResolvedValueOnce({ kind })
    const wrapper = mountPanel()
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toContain(title)
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)

    fetchMock.mockResolvedValueOnce({ kind: 'success', channel: 'whatsapp' })
    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(radios(wrapper)[1]!.element.checked).toBe(true)
  })

  it('has no accessibility violations', async () => {
    const wrapper = mountPanel()
    await flushPromises()
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
