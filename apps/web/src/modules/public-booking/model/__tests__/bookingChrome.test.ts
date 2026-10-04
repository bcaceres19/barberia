/**
 * Pruebas de useBookingChrome (DEC-111): canal acotado entre una pantalla y
 * el cascarón. Sin cascarón (pruebas de componente en aislamiento) es
 * inerte, y al desmontar la pantalla el cascarón vuelve a «no confirmado».
 */
import { describe, it, expect } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { bookingChromeKey, useBookingChrome, type BookingChrome } from '../bookingChrome'

function makeProbe(onChrome: (chrome: BookingChrome) => void) {
  return defineComponent({
    setup() {
      onChrome(useBookingChrome())
      return () => h('div')
    },
  })
}

describe('useBookingChrome', () => {
  it('is inert without a shell: writing to it throws nothing and leaks nowhere', () => {
    let chrome!: BookingChrome
    const wrapper = mount(makeProbe((value) => (chrome = value)))
    expect(chrome.completed.value).toBe(false)
    chrome.completed.value = true
    expect(chrome.completed.value).toBe(true)
    wrapper.unmount()
  })

  it('shares the shell flag and resets it when the screen unmounts', () => {
    const completed = ref(false)
    let chrome!: BookingChrome
    const wrapper = mount(
      makeProbe((value) => (chrome = value)),
      { global: { provide: { [bookingChromeKey as symbol]: { completed } } } },
    )

    chrome.completed.value = true
    expect(completed.value).toBe(true)

    wrapper.unmount()
    expect(completed.value).toBe(false)
  })
})
