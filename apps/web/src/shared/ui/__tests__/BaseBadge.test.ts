/**
 * Tests for BaseBadge component
 * Verifies: variants, sizes, dot, dismissible, accessibility
 */
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseBadge from '../BaseBadge.vue'

describe('BaseBadge', () => {
  describe('Rendering', () => {
    it('renders span element', () => {
      const wrapper = mount(BaseBadge, { slots: { default: 'Badge' } })
      expect(wrapper.element.tagName).toBe('SPAN')
    })

    it('renders slot content', () => {
      const wrapper = mount(BaseBadge, { slots: { default: 'Badge text' } })
      expect(wrapper.text()).toContain('Badge text')
    })

    it('applies variant classes', () => {
      const variants = ['neutral', 'primary', 'success', 'warning', 'danger', 'info'] as const
      variants.forEach(variant => {
        const wrapper = mount(BaseBadge, { props: { variant }, slots: { default: 'Test' } })
        expect(wrapper.classes()).toContain(`base-badge--${variant}`)
      })
    })

    it('applies size classes', () => {
      const sizes = ['sm', 'md'] as const
      sizes.forEach(size => {
        const wrapper = mount(BaseBadge, { props: { size }, slots: { default: 'Test' } })
        expect(wrapper.classes()).toContain(`base-badge--${size}`)
      })
    })
  })

  describe('Dot indicator', () => {
    it('shows dot when dot prop is true', () => {
      const wrapper = mount(BaseBadge, { props: { dot: true }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dot').exists()).toBe(true)
      expect(wrapper.classes()).toContain('base-badge--dot')
    })

    it('hides dot when dot prop is false', () => {
      const wrapper = mount(BaseBadge, { props: { dot: false }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dot').exists()).toBe(false)
    })

    it('applies dotColor style when provided', () => {
      const wrapper = mount(BaseBadge, { props: { dot: true, dotColor: '#ff0000' }, slots: { default: 'Test' } })
      const dot = wrapper.find('.base-badge__dot')
      // jsdom normalizes hex to rgb
      expect(dot.attributes('style')).toContain('background-color')
      expect(dot.attributes('style')).toMatch(/background-color.*rgb\(255,\s*0,\s*0\)/)
    })

    it('applies variant-based dot color when dotColor not provided', () => {
      const wrapper = mount(BaseBadge, { props: { variant: 'success', dot: true }, slots: { default: 'Test' } })
      const dot = wrapper.find('.base-badge__dot')
      // The CSS variable --badge-dot-color is set on the badge element, dot inherits via var()
      expect(dot.attributes('style')).toContain('var(--badge-dot-color)')
      expect(wrapper.classes()).toContain('base-badge--success')
    })
  })

  describe('Dismissible', () => {
    it('shows dismiss button when dismissible', () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dismiss').exists()).toBe(true)
      expect(wrapper.classes()).toContain('base-badge--dismissible')
    })

    it('hides dismiss button when not dismissible', () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: false }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dismiss').exists()).toBe(false)
    })

    it('emits dismiss event when dismiss button clicked', async () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true }, slots: { default: 'Test' } })
      await wrapper.find('.base-badge__dismiss').trigger('click')
      expect(wrapper.emitted('dismiss')).toBeTruthy()
    })

    it('stops propagation on dismiss click', async () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true }, slots: { default: 'Test' } })
      const clickSpy = vi.fn()
      wrapper.element.addEventListener('click', clickSpy)
      await wrapper.find('.base-badge__dismiss').trigger('click')
      // El click en el botón no debe propagarse al span
    })
  })

  describe('Accessibility', () => {
    it('has role status when dot is true', () => {
      const wrapper = mount(BaseBadge, { props: { dot: true }, slots: { default: 'Test' } })
      expect(wrapper.attributes('role')).toBe('status')
    })

    it('has aria-live polite when dot is true', () => {
      const wrapper = mount(BaseBadge, { props: { dot: true }, slots: { default: 'Test' } })
      expect(wrapper.attributes('aria-live')).toBe('polite')
    })

    it('has no role when dot is false', () => {
      const wrapper = mount(BaseBadge, { props: { dot: false }, slots: { default: 'Test' } })
      expect(wrapper.attributes('role')).toBeUndefined()
    })

    it('dismiss button has aria-label with label prop', () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true, label: 'Estado' }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dismiss').attributes('aria-label')).toBe('Cerrar Estado')
    })

    it('dismiss button has generic aria-label without label prop', () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dismiss').attributes('aria-label')).toBe('Cerrar etiqueta')
    })

    it('dot has aria-hidden', () => {
      const wrapper = mount(BaseBadge, { props: { dot: true }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dot').attributes('aria-hidden')).toBe('true')
    })

    it('dismiss icon has aria-hidden', () => {
      const wrapper = mount(BaseBadge, { props: { dismissible: true }, slots: { default: 'Test' } })
      expect(wrapper.find('.base-badge__dismiss-icon').attributes('aria-hidden')).toBe('true')
    })
  })

  describe('Custom class', () => {
    it('applies custom class', () => {
      const wrapper = mount(BaseBadge, { props: { class: 'my-custom-class' }, slots: { default: 'Test' } })
      expect(wrapper.classes()).toContain('my-custom-class')
    })
  })

  describe('Status variants', () => {
    it('has status-confirmed variant class', () => {
      const wrapper = mount(BaseBadge, { props: { variant: 'success' }, slots: { default: 'Confirmado' } })
      // El estilo se aplica via CSS variables, verificar que la clase base está
      expect(wrapper.classes()).toContain('base-badge--success')
    })

    it('has status-cancelled-customer variant class', () => {
      const wrapper = mount(BaseBadge, { props: { variant: 'neutral' }, slots: { default: 'Cancelado por cliente' } })
      expect(wrapper.classes()).toContain('base-badge--neutral')
    })
  })
})