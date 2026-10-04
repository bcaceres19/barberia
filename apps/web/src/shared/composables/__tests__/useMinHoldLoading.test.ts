import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { useMinHoldLoading, type MinHoldLoading } from '../useMinHoldLoading'

// onUnmounted (dentro del composable) solo se registra con una instancia de
// componente activa: se monta un host mínimo en vez de llamar al composable
// suelto.
function mountWithHold(minMs?: number) {
  let hold!: MinHoldLoading
  const wrapper = mount(
    defineComponent({
      setup() {
        hold = useMinHoldLoading(minMs)
        return () => h('div')
      },
    }),
  )
  return { wrapper, hold }
}

beforeEach(() => vi.useFakeTimers())
afterEach(() => vi.useRealTimers())

describe('useMinHoldLoading', () => {
  it('reveals immediately when hold() is called without a prior start()', () => {
    const { hold } = mountWithHold(400)
    const reveal = vi.fn()

    hold.hold(reveal)

    expect(reveal).toHaveBeenCalledOnce()
  })

  it('delays reveal until minMs elapsed since start()', () => {
    const { hold } = mountWithHold(400)
    const reveal = vi.fn()

    hold.start()
    vi.advanceTimersByTime(150)
    hold.hold(reveal)
    expect(reveal).not.toHaveBeenCalled()

    vi.advanceTimersByTime(249)
    expect(reveal).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    expect(reveal).toHaveBeenCalledOnce()
  })

  it('reveals immediately when hold() is called after minMs already elapsed', () => {
    const { hold } = mountWithHold(400)
    const reveal = vi.fn()

    hold.start()
    vi.advanceTimersByTime(500)
    hold.hold(reveal)

    expect(reveal).toHaveBeenCalledOnce()
  })

  it('a second start() restarts the minimum window', () => {
    const { hold } = mountWithHold(400)
    const reveal = vi.fn()

    hold.start()
    vi.advanceTimersByTime(300)
    hold.start()
    vi.advanceTimersByTime(300)
    hold.hold(reveal)
    expect(reveal).not.toHaveBeenCalled()

    vi.advanceTimersByTime(100)
    expect(reveal).toHaveBeenCalledOnce()
  })

  it('clears the pending timer on unmount without ever calling reveal', () => {
    const { wrapper, hold } = mountWithHold(400)
    const reveal = vi.fn()

    hold.start()
    hold.hold(reveal)
    wrapper.unmount()
    vi.advanceTimersByTime(1000)

    expect(reveal).not.toHaveBeenCalled()
  })
})
