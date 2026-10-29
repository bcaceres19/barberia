/**
 * StarRatingInput (DEC-128, HU-100): grupo de radios nativos con leyenda, cada estrella
 * con nombre «N de 5 estrellas», v-model controlado y error asociado.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import StarRatingInput from '../StarRatingInput.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }
const base = { legend: 'Tu calificación' }

describe('StarRatingInput', () => {
  it('es un fieldset con leyenda y cinco radios nombrados', () => {
    const wrapper = mount(StarRatingInput, { props: { ...base, modelValue: null } })
    expect(wrapper.get('fieldset legend').text()).toBe('Tu calificación')
    const radios = wrapper.findAll('input[type="radio"]')
    expect(radios).toHaveLength(5)
    expect(radios.map((r) => r.attributes('aria-label'))).toEqual([
      '1 de 5 estrellas',
      '2 de 5 estrellas',
      '3 de 5 estrellas',
      '4 de 5 estrellas',
      '5 de 5 estrellas',
    ])
    // Mismo `name`: el navegador las trata como un solo grupo (flechas entre ellas).
    expect(new Set(radios.map((r) => r.attributes('name'))).size).toBe(1)
  })

  it('emite el valor elegido y marca las estrellas hasta ese valor', async () => {
    const wrapper = mount(StarRatingInput, { props: { ...base, modelValue: null } })
    await wrapper.findAll('input')[3]!.setValue(true)
    expect(wrapper.emitted('update:modelValue')).toEqual([[4]])

    await wrapper.setProps({ modelValue: 4 })
    const on = wrapper.findAll('.star-input__star--on')
    expect(on).toHaveLength(4)
    expect((wrapper.findAll('input')[3]!.element as HTMLInputElement).checked).toBe(true)
  })

  it('un error queda asociado al grupo y se anuncia', () => {
    const wrapper = mount(StarRatingInput, {
      props: { ...base, modelValue: null, error: 'Elige de 1 a 5 estrellas.' },
    })
    const fieldset = wrapper.get('fieldset')
    const error = wrapper.get('[role="alert"]')
    expect(error.text()).toBe('Elige de 1 a 5 estrellas.')
    expect(fieldset.attributes('aria-describedby')).toBe(error.attributes('id'))
    expect(fieldset.attributes('aria-invalid')).toBe('true')
  })

  it('deshabilitado no deja elegir', () => {
    const wrapper = mount(StarRatingInput, {
      props: { ...base, modelValue: 2, disabled: true },
    })
    expect(wrapper.get('fieldset').attributes('disabled')).toBeDefined()
  })

  it('no tiene violaciones de axe, con y sin error', async () => {
    const plain = mount(StarRatingInput, {
      props: { ...base, modelValue: 3 },
      attachTo: document.body,
    })
    expect(await axe(plain.element, axeOptions)).toHaveNoViolations()
    plain.unmount()
    const invalid = mount(StarRatingInput, {
      props: { ...base, modelValue: null, error: 'Elige de 1 a 5 estrellas.' },
      attachTo: document.body,
    })
    expect(await axe(invalid.element, axeOptions)).toHaveNoViolations()
    invalid.unmount()
  })
})
