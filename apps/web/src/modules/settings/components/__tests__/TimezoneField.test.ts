/**
 * TimezoneField (DEC-110, RN-DIS-07): combobox accesible con buscador, reloj
 * de la zona elegida y atajo a la zona del dispositivo (nunca automática).
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import TimezoneField from '../TimezoneField.vue'

let mounted: VueWrapper | undefined

function mountField(props: Record<string, unknown> = {}) {
  mounted = mount(TimezoneField, {
    attachTo: document.body,
    props: { modelValue: 'America/Bogota', label: 'Zona horaria', ...props },
  })
  return mounted
}

const combo = (w: VueWrapper) => w.get('[role="combobox"]')
const options = (w: VueWrapper) => w.findAll('[role="option"]')

beforeEach(() => {
  // El reloj se detiene en un instante conocido: Bogotá = 07:00.
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-10-03T12:00:00Z'))
})

afterEach(() => {
  mounted?.unmount()
  mounted = undefined
  vi.useRealTimers()
})

describe('TimezoneField', () => {
  it('is a labelled combobox that shows the current zone and starts closed', () => {
    const wrapper = mountField()

    expect(wrapper.get('label').text()).toBe('Zona horaria')
    expect((combo(wrapper).element as HTMLInputElement).value).toBe('America/Bogota')
    expect(combo(wrapper).attributes('aria-expanded')).toBe('false')
    expect(combo(wrapper).attributes('aria-autocomplete')).toBe('list')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('shows the live clock of the chosen zone, not of the device', () => {
    const wrapper = mountField()

    expect(wrapper.get('.tz-field__time').text()).toBe('07:00')
    expect(wrapper.text()).toContain('Hora en Bogota')
    expect(wrapper.text()).toContain('UTC−5')
    expect(wrapper.text()).toContain('Sábado, 3 de octubre')
  })

  it('ticks forward with time', async () => {
    const wrapper = mountField()

    await vi.advanceTimersByTimeAsync(61_000)

    expect(wrapper.get('.tz-field__time').text()).toBe('07:01')
  })

  it('opens a listbox and wires the active option for assistive technology', async () => {
    const wrapper = mountField()

    await combo(wrapper).trigger('click')

    expect(combo(wrapper).attributes('aria-expanded')).toBe('true')
    const listbox = wrapper.get('[role="listbox"]')
    expect(combo(wrapper).attributes('aria-controls')).toBe(listbox.attributes('id'))
    const active = combo(wrapper).attributes('aria-activedescendant')!
    expect(wrapper.get(`#${active}`).text()).toContain('America/Bogota')
    expect(wrapper.get(`#${active}`).attributes('aria-selected')).toBe('true')
  })

  it('opens on the zone already chosen, first in the list', async () => {
    const wrapper = mountField({ modelValue: 'Pacific/Auckland' })

    await combo(wrapper).trigger('click')

    expect(options(wrapper)[0]!.text()).toContain('Pacific/Auckland')
    expect(options(wrapper)[0]!.attributes('aria-selected')).toBe('true')
    // Una sola vez: no se repite más abajo en la lista alfabética.
    expect(options(wrapper).filter((o) => o.text().includes('Pacific/Auckland'))).toHaveLength(1)
  })

  it('filters by city while typing', async () => {
    const wrapper = mountField()

    await combo(wrapper).setValue('madrid')

    const texts = options(wrapper).map((o) => o.text())
    expect(texts).toHaveLength(1)
    expect(texts[0]).toContain('Europe/Madrid')
  })

  it('filters by offset', async () => {
    const wrapper = mountField()

    await combo(wrapper).setValue('utc+5:30')

    const found = options(wrapper)
    expect(found.length).toBeGreaterThan(0)
    expect(found.every((o) => o.text().includes('UTC+5:30'))).toBe(true)
  })

  it('explains an empty search instead of showing a blank list', async () => {
    const wrapper = mountField()

    await combo(wrapper).setValue('zzzz')

    expect(options(wrapper)).toHaveLength(0)
    expect(wrapper.get('[role="listbox"]').text()).toContain('No encontramos esa zona')
  })

  it('chooses with the keyboard: arrows move, Enter picks and closes', async () => {
    const wrapper = mountField({ modelValue: 'America/Bogota' })
    await combo(wrapper).setValue('america/ar')
    expect(options(wrapper).length).toBeGreaterThan(1)

    await combo(wrapper).trigger('keydown', { key: 'ArrowDown' })
    const active = wrapper.get(`#${combo(wrapper).attributes('aria-activedescendant')!}`)
    const zone = active.find('.tz-field__zone').text()
    await combo(wrapper).trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('update:modelValue')).toEqual([[zone]])
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('chooses with the mouse', async () => {
    const wrapper = mountField()
    await combo(wrapper).setValue('madrid')

    await options(wrapper)[0]!.trigger('mousedown')

    expect(wrapper.emitted('update:modelValue')).toEqual([['Europe/Madrid']])
  })

  it('closes with Escape without choosing and without leaking the key to a dialog', async () => {
    const wrapper = mountField()
    await combo(wrapper).trigger('click')

    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true })
    const stop = vi.spyOn(event, 'stopPropagation')
    combo(wrapper).element.dispatchEvent(event)
    await wrapper.vm.$nextTick()

    expect(stop).toHaveBeenCalled()
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect((combo(wrapper).element as HTMLInputElement).value).toBe('America/Bogota')
  })

  it('goes back to the confirmed zone when the text is abandoned half typed', async () => {
    const wrapper = mountField()
    await combo(wrapper).setValue('mad')

    await combo(wrapper).trigger('blur')

    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    expect((combo(wrapper).element as HTMLInputElement).value).toBe('America/Bogota')
  })

  it('offers the device zone as a shortcut only when it differs, and never applies it by itself', async () => {
    vi.spyOn(Intl.DateTimeFormat.prototype, 'resolvedOptions').mockReturnValue({
      timeZone: 'Europe/Madrid',
    } as Intl.ResolvedDateTimeFormatOptions)
    const wrapper = mountField()

    const shortcut = wrapper.get('.tz-field__device')
    expect(shortcut.text()).toContain('Europe/Madrid')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()

    await shortcut.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['Europe/Madrid']])
    vi.restoreAllMocks()
  })

  it('hides the shortcut when the device is already in the configured zone', () => {
    vi.spyOn(Intl.DateTimeFormat.prototype, 'resolvedOptions').mockReturnValue({
      timeZone: 'America/Bogota',
    } as Intl.ResolvedDateTimeFormatOptions)
    const wrapper = mountField()

    expect(wrapper.find('.tz-field__device').exists()).toBe(false)
    vi.restoreAllMocks()
  })

  it('shows a field error linked to the combobox', () => {
    const wrapper = mountField({ error: 'Escribe la zona horaria.' })

    expect(combo(wrapper).attributes('aria-invalid')).toBe('true')
    expect(wrapper.get(`#${combo(wrapper).attributes('aria-describedby')!}`).text()).toBe(
      'Escribe la zona horaria.',
    )
  })

  it('keeps a zone the browser does not list, so a legacy value is never lost', async () => {
    const wrapper = mountField({ modelValue: 'Legacy/Zona_Vieja' })
    await combo(wrapper).trigger('click')

    expect(options(wrapper).some((o) => o.text().includes('Legacy/Zona_Vieja'))).toBe(true)
  })

  it('does not open while disabled', async () => {
    const wrapper = mountField({ disabled: true })

    await combo(wrapper).trigger('click')

    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('has no axe violations closed, open and with an error', async () => {
    vi.useRealTimers()
    const wrapper = mountField()
    expect(await axe(wrapper.element)).toHaveNoViolations()

    await combo(wrapper).trigger('click')
    expect(await axe(wrapper.element)).toHaveNoViolations()

    await wrapper.setProps({ error: 'Escribe la zona horaria.' })
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
