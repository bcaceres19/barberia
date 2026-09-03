/**
 * Tests for RecordRow component (estandar-diseno-visual.md §6.6)
 * Verifies: tag, leading/main/trailing regions, axe.
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import RecordRow from '../RecordRow.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

describe('RecordRow', () => {
  it('renders as li by default (list semantics)', () => {
    const wrapper = mount(RecordRow, { slots: { default: 'Barbero QA' } })
    expect(wrapper.element.tagName).toBe('LI')
  })

  it('renders as the requested tag', () => {
    const wrapper = mount(RecordRow, { props: { tag: 'div' }, slots: { default: 'Barbero QA' } })
    expect(wrapper.element.tagName).toBe('DIV')
  })

  it('renders default slot content in the main region', () => {
    const wrapper = mount(RecordRow, { slots: { default: 'Barbero QA' } })
    expect(wrapper.get('.record-row__main').text()).toBe('Barbero QA')
  })

  it('renders leading slot only when provided', () => {
    const withLeading = mount(RecordRow, {
      slots: { default: 'Barbero QA', leading: '<span>BQ</span>' },
    })
    expect(withLeading.find('.record-row__leading').exists()).toBe(true)

    const withoutLeading = mount(RecordRow, { slots: { default: 'Barbero QA' } })
    expect(withoutLeading.find('.record-row__leading').exists()).toBe(false)
  })

  it('renders trailing slot only when provided', () => {
    const withTrailing = mount(RecordRow, {
      slots: { default: 'Corte QA', trailing: '<button>Editar</button>' },
    })
    expect(withTrailing.find('.record-row__trailing').exists()).toBe(true)
    expect(withTrailing.text()).toContain('Editar')

    const withoutTrailing = mount(RecordRow, { slots: { default: 'Corte QA' } })
    expect(withoutTrailing.find('.record-row__trailing').exists()).toBe(false)
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones dentro de una lista con las tres regiones', async () => {
      const wrapper = mount(
        {
          components: { RecordRow },
          template: `
            <ul aria-label="Barberos de la barbería">
              <RecordRow>
                <template #leading><span aria-hidden="true">BQ</span></template>
                Barbero QA
                <template #trailing><button>Editar</button></template>
              </RecordRow>
            </ul>
          `,
        },
        {},
      )
      expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
    })
  })
})
