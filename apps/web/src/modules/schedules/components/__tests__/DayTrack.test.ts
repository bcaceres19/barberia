import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import DayTrack from '../DayTrack.vue'

const scale = { startMinute: 420, endMinute: 1260 }

describe('DayTrack', () => {
  it('is decorative: hidden from assistive technology', () => {
    const wrapper = mount(DayTrack, { props: { segments: [], scale } })
    expect(wrapper.attributes('aria-hidden')).toBe('true')
  })

  it('draws one bar per valid tramo, positioned as a percentage of the ruler', () => {
    const wrapper = mount(DayTrack, {
      props: {
        segments: [
          { id: 'a', startsTime: '09:00', durationMinutes: 240 },
          { id: 'b', startsTime: '', durationMinutes: 60 },
          { id: 'c', startsTime: '14:00', durationMinutes: 0 },
        ],
        scale,
      },
    })

    const bars = wrapper.findAll('.day-track__bar')
    expect(bars).toHaveLength(1)
    expect(bars[0]!.attributes('style')).toContain('--bar-from: 14.28')
    expect(bars[0]!.attributes('style')).toContain('--bar-span: 28.57')
  })

  it('highlights the active tramo and mutes context tramos', () => {
    const wrapper = mount(DayTrack, {
      props: {
        segments: [
          { id: 'a', startsTime: '09:00', durationMinutes: 120 },
          { id: 'b', startsTime: '14:00', durationMinutes: 120, muted: true },
        ],
        scale,
        activeId: 'a',
      },
    })

    const [first, second] = wrapper.findAll('.day-track__bar')
    expect(first!.classes()).toContain('day-track__bar--active')
    expect(second!.classes()).toContain('day-track__bar--muted')
  })

  it('draws the now marker only inside the ruler', () => {
    const inside = mount(DayTrack, { props: { segments: [], scale, nowMinute: 840 } })
    const outside = mount(DayTrack, { props: { segments: [], scale, nowMinute: 120 } })

    expect(inside.find('.day-track__now').attributes('style')).toContain('left: 50%')
    expect(outside.find('.day-track__now').exists()).toBe(false)
  })

  it('flags a tramo that crosses midnight', () => {
    const wrapper = mount(DayTrack, {
      props: {
        segments: [{ id: 'n', startsTime: '20:00', durationMinutes: 480 }],
        scale: { startMinute: 420, endMinute: 1440 },
      },
    })

    expect(wrapper.find('.day-track__bar').classes()).toContain('day-track__bar--cut')
  })
})
