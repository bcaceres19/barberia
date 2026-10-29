import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import PhoneField from '../PhoneField.vue'

function mountField(props: Partial<{ country: string; national: string; error: string }> = {}) {
  return mount(PhoneField, { props: { country: 'CO', national: '', required: true, ...props } })
}

describe('PhoneField', () => {
  it('muestra la bandera y el prefijo del país elegido, fuera de la casilla del número', () => {
    const wrapper = mountField({ country: 'MX', national: '5512345678' })

    expect(wrapper.get('.phone-field__flag').text()).toBe('🇲🇽')
    expect(wrapper.get('.phone-field__dial').text()).toBe('+52')
    expect((wrapper.get('input[type="tel"]').element as HTMLInputElement).value).toBe('5512345678')
  })

  it('emite el país al cambiar el selector', async () => {
    const wrapper = mountField()
    await wrapper.get('select').setValue('EC')

    expect(wrapper.emitted('update:country')).toEqual([['EC']])
  })

  it('deja solo dígitos en la casilla del número', async () => {
    const wrapper = mountField()
    await wrapper.get('input[type="tel"]').setValue('300 12a3-4567')

    expect(wrapper.emitted('update:national')).toEqual([['3001234567']])
  })

  it('al pegar un número con prefijo elige el país y deja el número nacional', async () => {
    const wrapper = mountField()
    await wrapper.get('input[type="tel"]').setValue('+52 55 1234 5678')

    expect(wrapper.emitted('update:country')).toEqual([['MX']])
    expect(wrapper.emitted('update:national')).toEqual([['5512345678']])
  })

  it('anuncia el error y lo asocia a la casilla', () => {
    const wrapper = mountField({ error: 'El teléfono es obligatorio.' })
    const input = wrapper.get('input[type="tel"]')

    expect(input.attributes('aria-invalid')).toBe('true')
    const errorId = input.attributes('aria-describedby')!
    expect(wrapper.get(`#${errorId}`).text()).toBe('El teléfono es obligatorio.')
  })

  it('no tiene violaciones de accesibilidad', async () => {
    // Dentro de un landmark, como lo aloja la página real.
    const wrapper = mount({
      render: () =>
        h('main', [
          h(PhoneField, {
            country: 'CO',
            national: '',
            required: true,
            error: 'El teléfono es obligatorio.',
          }),
        ]),
    })
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
