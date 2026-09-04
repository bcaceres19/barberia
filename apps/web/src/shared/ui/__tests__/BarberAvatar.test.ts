/**
 * Tests for BarberAvatar component (issue #189, atlas panel-agenda-eventos
 * evento 12). Verifies: monogram algorithm (1-word and 3-word names, middle
 * word ignored), sizes, photo branch, decorative aria-hidden.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BarberAvatar from '../BarberAvatar.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('BarberAvatar', () => {
  it('derives a single-letter monogram from a one-word name', () => {
    const wrapper = mount(BarberAvatar, { props: { fullName: 'Cher' } })
    expect(wrapper.get('.barber-avatar__monogram').text()).toBe('C')
  })

  it('derives a two-letter monogram from first+last word, ignoring middle words', () => {
    const wrapper = mount(BarberAvatar, { props: { fullName: 'Ana María Gómez' } })
    expect(wrapper.get('.barber-avatar__monogram').text()).toBe('AG')
  })

  it('derives a two-letter monogram from a two-word name', () => {
    const wrapper = mount(BarberAvatar, { props: { fullName: 'Julián Rodríguez' } })
    expect(wrapper.get('.barber-avatar__monogram').text()).toBe('JR')
  })

  it.each(['closed', 'option'] as const)('applies the %s size class', (size) => {
    const wrapper = mount(BarberAvatar, { props: { fullName: 'Julián Rodríguez', size } })
    expect(wrapper.classes()).toContain(`barber-avatar--${size}`)
  })

  it('renders a decorative photo and suppresses the monogram when photoUrl is set', () => {
    const wrapper = mount(BarberAvatar, {
      props: { fullName: 'Andrés Beltrán', photoUrl: 'https://example.test/andres.jpg' },
    })
    const img = wrapper.get('.barber-avatar__photo')
    expect(img.attributes('src')).toBe('https://example.test/andres.jpg')
    expect(img.attributes('alt')).toBe('')
    expect(wrapper.find('.barber-avatar__monogram').exists()).toBe(false)
  })

  it('is always decorative: the adjacent visible name carries the accessible name', () => {
    const wrapper = mount(BarberAvatar, { props: { fullName: 'Julián Rodríguez' } })
    expect(wrapper.attributes('aria-hidden')).toBe('true')
  })

  it('has no obvious accessibility violations next to a visible name', async () => {
    const wrapper = mount(
      {
        components: { BarberAvatar },
        template: `<span><BarberAvatar full-name="Julián Rodríguez" /> Julián Rodríguez</span>`,
      },
      {},
    )
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })
})
