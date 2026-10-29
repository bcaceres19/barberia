/**
 * MyProfileCard (DEC-115, DEC-127): la tarjeta del perfil `solo` muestra la
 * descripción pública del barbero cuando existe y nada cuando no.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import MyProfileCard from '../MyProfileCard.vue'

function mountCard(description: string | null) {
  return mount(MyProfileCard, {
    props: {
      fullName: 'Carlos Ramírez',
      description,
      photoUrl: null,
      hasPhoto: false,
      sinceLabel: '23 ago 2026',
      roleLabel: 'Barbero',
    },
  })
}

describe('MyProfileCard', () => {
  it('shows the public description under the name', () => {
    const wrapper = mountCard('Fade y barba.')
    expect(wrapper.get('.my-profile-card__description').text()).toBe('Fade y barba.')
  })

  it('shows no description block without one', () => {
    expect(mountCard(null).find('.my-profile-card__description').exists()).toBe(false)
  })

  it('has no accessibility violations with a description', async () => {
    const wrapper = mountCard('Fade y barba.')
    const results = await axe(wrapper.element, {
      rules: { 'color-contrast': { enabled: false } },
    })
    expect(results).toHaveNoViolations()
  })
})
