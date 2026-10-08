/**
 * PublicLinkPanel (DEC-117): ver y copiar el enlace público de reservas.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchPublicLinkMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/publicLinkApi', () => ({ fetchPublicLink: fetchPublicLinkMock }))

const { default: PublicLinkPanel } = await import('../PublicLinkPanel.vue')

const SLUG = 'corte-fino-k7x2m9'
const URL_ = `${window.location.origin}/reservar/${SLUG}`

function mountPanel() {
  return mount(PublicLinkPanel, { attachTo: document.body })
}

describe('PublicLinkPanel', () => {
  beforeEach(() => {
    fetchPublicLinkMock.mockReset()
    fetchPublicLinkMock.mockResolvedValue({ kind: 'success', slug: SLUG })
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('announces the wait and then shows the full public URL in a read-only field', async () => {
    const wrapper = mountPanel()
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(wrapper.find('[role="textbox"]').exists()).toBe(false)

    await flushPromises()

    const field = wrapper.get('[role="textbox"]')
    expect(field.text()).toBe(URL_)
    expect(field.attributes('aria-readonly')).toBe('true')
    const label = wrapper.get(`#${field.attributes('aria-labelledby')}`)
    expect(label.text()).toBe('Tu enlace de reservas')
  })

  it('opens the link in a new tab without leaking the opener', async () => {
    const wrapper = mountPanel()
    await flushPromises()

    const open = wrapper.get('a')
    expect(open.attributes('href')).toBe(URL_)
    expect(open.attributes('target')).toBe('_blank')
    expect(open.attributes('rel')).toBe('noopener noreferrer')
    expect(open.text()).toContain('se abre en una pestaña nueva')
  })

  it('copies the URL and says so in a live region that was already on screen', async () => {
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] })
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })
    const wrapper = mountPanel()
    await flushPromises()
    const feedback = wrapper.get('.public-link__feedback')
    expect(feedback.attributes('aria-live')).toBe('polite')
    expect(feedback.text()).toBe('')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(writeText).toHaveBeenCalledWith(URL_)
    expect(feedback.text()).toBe('Enlace copiado.')

    vi.advanceTimersByTime(3100)
    await flushPromises()
    expect(feedback.text()).toBe('')
  })

  it('selects the link and explains how to copy by hand when the clipboard is unavailable', async () => {
    vi.stubGlobal('navigator', {
      clipboard: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    })
    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    const field = wrapper.get('[role="textbox"]').element
    expect(document.activeElement).toBe(field)
    expect(window.getSelection()?.toString().trim()).toBe(URL_)
    expect(wrapper.get('.public-link__feedback').text()).toContain('Ctrl+C')
  })

  it.each([
    ['network-error', 'No pudimos conectar'],
    ['unexpected-error', 'No pudimos cargar tu enlace'],
  ])('shows a recoverable error for %s and retries', async (kind, title) => {
    fetchPublicLinkMock.mockResolvedValueOnce({ kind })
    const wrapper = mountPanel()
    await flushPromises()

    expect(wrapper.text()).toContain(title)
    expect(wrapper.find('[role="textbox"]').exists()).toBe(false)

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(fetchPublicLinkMock).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[role="textbox"]').text()).toBe(URL_)
  })

  it('has no accessibility violations once loaded', async () => {
    const wrapper = mountPanel()
    await flushPromises()

    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
