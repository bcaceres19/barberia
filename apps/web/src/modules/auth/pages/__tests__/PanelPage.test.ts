import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PanelPage from '../PanelPage.vue'

describe('PanelPage', () => {
  it('renders a heading without claiming an agenda that does not exist yet', () => {
    const wrapper = mount(PanelPage)
    expect(wrapper.get('h1').text()).toBe('Panel del barbero')
    expect(wrapper.text().toLowerCase()).not.toContain('cita')
  })
})
