import { describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import HomeLocationField from '../HomeLocationField.vue'
import type { PlaceSelection } from '@/shared/model/placesAutocomplete'

function mountIt(modelValue: PlaceSelection | null = null, error?: string) {
  return mount(HomeLocationField, {
    attachTo: document.body,
    props: { modelValue, error, start: { latitude: 4.7, longitude: -74.07 } },
  })
}

describe('HomeLocationField', () => {
  it('offers the three ways in and says the address is optional (CA-138-01)', async () => {
    const wrapper = mountIt()
    await flushPromises()

    expect(wrapper.text()).toContain('Busca tu dirección (opcional)')
    expect(wrapper.text()).toContain('no hace falta: marca tu ubicación en el mapa')
    expect(wrapper.find('input[role="combobox"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Mueve el mapa hasta que el pin quede sobre el lugar')
    expect(wrapper.find('[aria-label="Mover el mapa al norte"]').exists()).toBe(true)
  })

  it('emits the point marked on the map', async () => {
    const wrapper = mountIt()
    await flushPromises()
    await wrapper.get('[aria-label="Mover el mapa al norte"]').trigger('click')
    await flushPromises()

    const all = wrapper.emitted('update:modelValue')!
    expect(all[0]![0]).toMatchObject({
      address: 'Ubicación marcada en el mapa',
      latitude: 4.701,
      longitude: -74.07,
    })
  })

  it('does not wipe the marked point when the person edits the search text', async () => {
    const chosen: PlaceSelection = { address: 'Punto marcado', latitude: 4.7, longitude: -74.07 }
    const wrapper = mountIt(chosen)
    await flushPromises()
    await wrapper.get('input[role="combobox"]').setValue('Carrera 7')

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('shows the validation error under the map', async () => {
    const wrapper = mountIt(null, 'Marca tu ubicación en el mapa.')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Marca tu ubicación en el mapa.')
  })

  it('has no accessibility violations', async () => {
    const wrapper = mountIt({ address: 'Calle 9', latitude: 4.6, longitude: -74.1 })
    await flushPromises()
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
