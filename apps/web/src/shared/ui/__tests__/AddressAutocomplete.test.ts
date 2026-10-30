import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import AddressAutocomplete from '../AddressAutocomplete.vue'
import type { PlaceSelection, PlacesClient } from '../../model/placesAutocomplete'

const selection: PlaceSelection = {
  address: 'Carrera 7 # 71-21, Bogotá',
  latitude: 4.6486,
  longitude: -74.0628,
}

function fakeClient(overrides: Partial<PlacesClient> = {}): PlacesClient {
  return {
    suggest: vi.fn().mockResolvedValue([
      { placeId: 'p1', text: 'Carrera 7 # 71-21, Bogotá' },
      { placeId: 'p2', text: 'Carrera 7 # 72-10, Bogotá' },
    ]),
    resolve: vi.fn().mockResolvedValue(selection),
    ...overrides,
  }
}

function mountIt(client: PlacesClient | null, modelValue: PlaceSelection | null = null) {
  return mount(AddressAutocomplete, {
    attachTo: document.body,
    props: { modelValue, label: 'Tu dirección', client, debounceMs: 0 },
  })
}

async function type(wrapper: ReturnType<typeof mountIt>, value: string) {
  const input = wrapper.get('input')
  await input.setValue(value)
  await new Promise((r) => setTimeout(r, 5))
  await flushPromises()
}

describe('AddressAutocomplete', () => {
  it('is an accessible combobox with a status region', async () => {
    const wrapper = mountIt(fakeClient())
    await type(wrapper, 'Carrera 7')

    const input = wrapper.get('input')
    expect(input.attributes('role')).toBe('combobox')
    expect(input.attributes('aria-expanded')).toBe('true')
    expect(wrapper.findAll('[role="option"]')).toHaveLength(2)
    expect(wrapper.get('[role="status"]').text()).toContain('2 sugerencias')
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })

  it('a typed text is not an address: nothing is emitted until a suggestion is chosen', async () => {
    const wrapper = mountIt(fakeClient())
    await type(wrapper, 'Carrera 7')

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('choosing a suggestion resolves it and emits the selection with coordinates', async () => {
    const client = fakeClient()
    const wrapper = mountIt(client)
    await type(wrapper, 'Carrera 7')

    await wrapper.findAll('[role="option"]')[0]!.trigger('click')
    await flushPromises()

    expect(client.resolve).toHaveBeenCalledWith('p1')
    expect(wrapper.emitted('update:modelValue')).toEqual([[selection]])
  })

  it('works from the keyboard: arrows, Enter and Escape', async () => {
    const wrapper = mountIt(fakeClient())
    await type(wrapper, 'Carrera 7')
    const input = wrapper.get('input')

    await input.trigger('keydown', { key: 'ArrowDown' })
    expect(input.attributes('aria-activedescendant')).toBe(
      wrapper.findAll('[role="option"]')[0]!.attributes('id'),
    )
    await input.trigger('keydown', { key: 'ArrowDown' })
    expect(wrapper.findAll('[role="option"]')[1]!.attributes('aria-selected')).toBe('true')

    await input.trigger('keydown', { key: 'Escape' })
    expect(input.attributes('aria-expanded')).toBe('false')

    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(wrapper.emitted('update:modelValue')).toEqual([[selection]])
  })

  it('editing the text drops the previous selection', async () => {
    const wrapper = mountIt(fakeClient(), selection)

    await wrapper.get('input').setValue('Otra calle')

    expect(wrapper.emitted('update:modelValue')).toEqual([[null]])
  })

  it('shows the map of the selected point', () => {
    const wrapper = mountIt(fakeClient(), selection)

    expect(wrapper.find('.address-autocomplete__map').exists()).toBe(true)
    expect(wrapper.get('input').element.value).toBe(selection.address)
  })

  it('says so when the search fails, and when nothing matches', async () => {
    const failing = mountIt(fakeClient({ suggest: vi.fn().mockRejectedValue(new Error('x')) }))
    await type(failing, 'Carrera 7')
    expect(failing.get('[role="status"]').text()).toContain('No pudimos buscar')

    const empty = mountIt(fakeClient({ suggest: vi.fn().mockResolvedValue([]) }))
    await type(empty, 'zzzz')
    expect(empty.get('[role="status"]').text()).toContain('No encontramos esa dirección')
  })

  it('does not search before the minimum length', async () => {
    const client = fakeClient()
    const wrapper = mountIt(client)

    await type(wrapper, 'ca')

    expect(client.suggest).not.toHaveBeenCalled()
  })

  it('shows the error with role alert and marks the field invalid', async () => {
    const wrapper = mount(AddressAutocomplete, {
      attachTo: document.body,
      props: {
        modelValue: null,
        label: 'Tu dirección',
        client: fakeClient(),
        error: 'Elige una dirección de la lista.',
      },
    })

    expect(wrapper.get('[role="alert"]').text()).toContain('Elige una dirección')
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('true')
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })

  describe('when a map picker owns the location (DEC-138)', () => {
    it('keeps the chosen point while typing if clearOnInput is off', async () => {
      const wrapper = mount(AddressAutocomplete, {
        attachTo: document.body,
        props: {
          modelValue: selection,
          label: 'Tu dirección',
          client: fakeClient(),
          debounceMs: 0,
          clearOnInput: false,
        },
      })
      await type(wrapper, 'Calle 80')

      expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    })

    it('still clears the point while typing by default', async () => {
      const wrapper = mountIt(fakeClient(), selection)
      await type(wrapper, 'Calle 80')

      expect(wrapper.emitted('update:modelValue')![0]).toEqual([null])
    })

    it('hides its own map when showMap is off', () => {
      const wrapper = mount(AddressAutocomplete, {
        props: {
          modelValue: selection,
          label: 'Tu dirección',
          client: fakeClient(),
          showMap: false,
        },
      })
      expect(wrapper.findComponent({ name: 'LocationMap' }).exists()).toBe(false)
    })
  })
})
