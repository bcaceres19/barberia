/**
 * Tests for BaseButton component
 * Verifies: variants, sizes, states, accessibility, events
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseButton from '../BaseButton.vue'

describe('BaseButton', () => {
  describe('Rendering', () => {
    it('renders as button element', () => {
      const wrapper = mount(BaseButton, { slots: { default: 'Click me' } })
      expect(wrapper.element.tagName).toBe('BUTTON')
    })

    it('renders slot content', () => {
      const wrapper = mount(BaseButton, { slots: { default: 'Click me' } })
      expect(wrapper.text()).toContain('Click me')
    })

    it('applies variant classes', () => {
      const variants = ['primary', 'secondary', 'soft', 'ghost', 'danger'] as const
      variants.forEach(variant => {
        const wrapper = mount(BaseButton, { props: { variant }, slots: { default: 'Test' } })
        expect(wrapper.classes()).toContain(`base-button--${variant}`)
      })
    })

    it('applies size classes', () => {
      const sizes = ['sm', 'md', 'lg'] as const
      sizes.forEach(size => {
        const wrapper = mount(BaseButton, { props: { size }, slots: { default: 'Test' } })
        expect(wrapper.classes()).toContain(`base-button--${size}`)
      })
    })
  })

  describe('States', () => {
    it('applies disabled class and attributes when disabled', () => {
      const wrapper = mount(BaseButton, { props: { disabled: true }, slots: { default: 'Test' } })
      expect(wrapper.classes()).toContain('base-button--disabled')
      expect(wrapper.attributes('disabled')).toBeDefined()
      expect(wrapper.attributes('aria-disabled')).toBe('true')
    })

    it('applies loading class and attributes when loading', () => {
      const wrapper = mount(BaseButton, { props: { loading: true }, slots: { default: 'Test' } })
      expect(wrapper.classes()).toContain('base-button--loading')
      expect(wrapper.attributes('disabled')).toBeDefined()
      expect(wrapper.attributes('aria-disabled')).toBe('true')
      expect(wrapper.attributes('aria-busy')).toBe('true')
      expect(wrapper.find('.base-button__spinner').exists()).toBe(true)
    })

    it('applies pressed class when pressed', () => {
      const wrapper = mount(BaseButton, { props: { pressed: true }, slots: { default: 'Test' } })
      expect(wrapper.classes()).toContain('base-button--pressed')
      expect(wrapper.attributes('aria-pressed')).toBe('true')
    })

    it('sets type attribute', () => {
      const types = ['button', 'submit', 'reset'] as const
      types.forEach(type => {
        const wrapper = mount(BaseButton, { props: { type }, slots: { default: 'Test' } })
        expect(wrapper.attributes('type')).toBe(type)
      })
    })
  })

  describe('Events', () => {
    it('emits click event when clicked', async () => {
      const wrapper = mount(BaseButton, { slots: { default: 'Test' } })
      await wrapper.trigger('click')
      expect(wrapper.emitted('click')).toBeTruthy()
    })

    it('does not emit click when disabled', async () => {
      const wrapper = mount(BaseButton, { props: { disabled: true }, slots: { default: 'Test' } })
      await wrapper.trigger('click')
      expect(wrapper.emitted('click')).toBeFalsy()
    })

    it('does not emit click when loading', async () => {
      const wrapper = mount(BaseButton, { props: { loading: true }, slots: { default: 'Test' } })
      await wrapper.trigger('click')
      expect(wrapper.emitted('click')).toBeFalsy()
    })

    it('passes native event to click handler', async () => {
      const wrapper = mount(BaseButton, { slots: { default: 'Test' } })
      await wrapper.trigger('click')
      const clickEvent = wrapper.emitted('click')?.[0]?.[0]
      expect(clickEvent).toBeInstanceOf(MouseEvent)
    })
  })

  describe('Accessibility', () => {
    it('has native button role', () => {
      const wrapper = mount(BaseButton, { slots: { default: 'Test' } })
      // button element has implicit role="button"
      expect(wrapper.element.tagName).toBe('BUTTON')
    })

    it('has aria-disabled when disabled', () => {
      const wrapper = mount(BaseButton, { props: { disabled: true }, slots: { default: 'Test' } })
      expect(wrapper.attributes('aria-disabled')).toBe('true')
    })

    it('has aria-pressed when pressed', () => {
      const wrapper = mount(BaseButton, { props: { pressed: true }, slots: { default: 'Test' } })
      expect(wrapper.attributes('aria-pressed')).toBe('true')
    })

    it('has aria-busy when loading', () => {
      const wrapper = mount(BaseButton, { props: { loading: true }, slots: { default: 'Test' } })
      expect(wrapper.attributes('aria-busy')).toBe('true')
    })

    it('spinner has aria-hidden', () => {
      const wrapper = mount(BaseButton, { props: { loading: true }, slots: { default: 'Test' } })
      const spinner = wrapper.find('.base-button__spinner')
      expect(spinner.attributes('aria-hidden')).toBe('true')
    })
  })

  describe('Custom class', () => {
    it('applies custom class', () => {
      const wrapper = mount(BaseButton, { props: { class: 'my-custom-class' }, slots: { default: 'Test' } })
      expect(wrapper.classes()).toContain('my-custom-class')
    })
  })
})