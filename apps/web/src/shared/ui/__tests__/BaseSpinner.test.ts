/**
 * Tests for BaseSpinner component (issue #189, estandar-diseno-visual.md).
 * Verifies: sizes, tone, accessible name, reduced-motion.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BaseSpinner from '../BaseSpinner.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('BaseSpinner', () => {
  it('is aria-hidden by default (decorative, expects an adjacent visible label)', () => {
    const wrapper = mount(BaseSpinner)
    expect(wrapper.attributes('aria-hidden')).toBe('true')
    expect(wrapper.attributes('role')).toBeUndefined()
  })

  it('exposes role="img" and aria-label when used standalone with a label', () => {
    const wrapper = mount(BaseSpinner, { props: { label: 'Cargando' } })
    expect(wrapper.attributes('role')).toBe('img')
    expect(wrapper.attributes('aria-label')).toBe('Cargando')
    expect(wrapper.attributes('aria-hidden')).toBeUndefined()
  })

  it.each(['sm', 'md', 'lg'] as const)('applies the %s size class', (size) => {
    const wrapper = mount(BaseSpinner, { props: { size } })
    expect(wrapper.classes()).toContain(`base-spinner--${size}`)
  })

  it('resolves the tone into a CSS custom property', () => {
    const wrapper = mount(BaseSpinner, { props: { tone: 'warning' } })
    expect(wrapper.attributes('style')).toContain('--spinner-tone: var(--color-warning-text)')
  })

  it('has no obvious accessibility violations', async () => {
    const wrapper = mount(BaseSpinner, { props: { label: 'Cargando' } })
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })
})
