/**
 * Pintado de las preferencias en <html> (DEC-110): modo, animaciones, acento y
 * tamaño de texto, solo mientras el cascarón privado está montado.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

type Listener = (event: { matches: boolean }) => void

function mockSystem(light: boolean) {
  const listeners = new Set<Listener>()
  const media = {
    matches: light,
    addEventListener: (_: string, l: Listener) => listeners.add(l),
    removeEventListener: (_: string, l: Listener) => listeners.delete(l),
  }
  window.matchMedia = vi.fn().mockReturnValue(media) as unknown as typeof window.matchMedia
  return {
    flip(matches: boolean) {
      media.matches = matches
      listeners.forEach((l) => l({ matches }))
    },
    listeners,
  }
}

async function load() {
  vi.resetModules()
  const [theme, appearance, brand] = await Promise.all([
    import('../workspaceTheme'),
    import('../appearanceStore'),
    import('../brandStore'),
  ])
  return { theme, appearance, brand }
}

const root = document.documentElement
const prop = (name: string) => root.style.getPropertyValue(name)

beforeEach(() => {
  window.localStorage.clear()
  mockSystem(false)
  Object.defineProperty(window, 'innerWidth', { value: 1280, configurable: true })
})

afterEach(() => {
  root.removeAttribute('data-app-theme')
  root.removeAttribute('data-motion')
  root.removeAttribute('style')
})

describe('effectiveZoom', () => {
  it('never reduces the width below 320 px', async () => {
    const { theme } = await load()
    expect(theme.effectiveZoom(1.25, 1280)).toBe(1.25)
    expect(theme.effectiveZoom(1.25, 360)).toBeCloseTo(1.125, 5)
    expect(theme.effectiveZoom(1.25, 320)).toBe(1)
    expect(theme.effectiveZoom(1.25, 280)).toBe(1)
  })

  it('does not limit the smaller sizes', async () => {
    const { theme } = await load()
    expect(theme.effectiveZoom(0.9, 320)).toBe(0.9)
    expect(theme.effectiveZoom(1, 320)).toBe(1)
  })
})

describe('attachWorkspaceTheme', () => {
  it('paints the initial look: Tinta, brass untouched, no zoom', async () => {
    const { theme } = await load()
    const detach = theme.attachWorkspaceTheme()

    expect(root.dataset.appTheme).toBe('ink')
    expect(root.dataset.motion).toBeUndefined()
    // El latón sobre Tinta es el de tokens.css: no se sobrescribe.
    expect(prop('--color-brand-accent-surface')).toBe('')
    expect(prop('--color-focus')).toBe('')
    expect(prop('zoom')).toBe('')
    expect(prop('--ui-zoom')).toBe('')
    detach()
  })

  it('switches the mode and the brass with it', async () => {
    const { theme, appearance } = await load()
    const detach = theme.attachWorkspaceTheme()

    appearance.setTheme('ivory')
    await nextTick()

    expect(root.dataset.appTheme).toBe('ivory')
    expect(prop('--color-brand-accent-surface')).toBe('#765c2f')
    expect(prop('--color-brand-accent-text')).toBe('#f4f0e7')
    detach()
  })

  it('paints another accent with the value measured for the current mode', async () => {
    const { theme, appearance, brand } = await load()
    const detach = theme.attachWorkspaceTheme()

    brand.setBrand({ ...brand.brandState.brand, accent: 'emerald' })
    await nextTick()
    expect(prop('--color-brand-accent-surface')).toBe('#4cb58a')
    expect(prop('--color-brand-accent-text')).toBe('#101b2b')
    expect(prop('--color-focus')).toBe('#4cb58a')

    appearance.setTheme('ivory')
    await nextTick()
    expect(prop('--color-brand-accent-surface')).toBe('#1c6a4d')
    expect(prop('--color-brand-accent-text')).toBe('#f4f0e7')
    detach()
  })

  it('previews an unsaved accent and returns to the confirmed one', async () => {
    const { theme, brand } = await load()
    const detach = theme.attachWorkspaceTheme()

    brand.setAccentPreview('ruby')
    await nextTick()
    expect(prop('--color-brand-accent-surface')).toBe('#e8808c')

    brand.setAccentPreview(null)
    await nextTick()
    expect(prop('--color-brand-accent-surface')).toBe('')
    detach()
  })

  it('follows the system in automatic mode, live', async () => {
    const system = mockSystem(false)
    const { theme, appearance } = await load()
    appearance.setTheme('auto')
    const detach = theme.attachWorkspaceTheme()
    expect(root.dataset.appTheme).toBe('ink')

    system.flip(true)
    await nextTick()
    expect(root.dataset.appTheme).toBe('ivory')

    system.flip(false)
    await nextTick()
    expect(root.dataset.appTheme).toBe('ink')
    detach()
    expect(system.listeners.size).toBe(0)
  })

  it('reduces motion on request and restores it', async () => {
    const { theme, appearance } = await load()
    const detach = theme.attachWorkspaceTheme()

    appearance.setMotion('reduced')
    await nextTick()
    expect(root.dataset.motion).toBe('reduced')

    appearance.setMotion('full')
    await nextTick()
    expect(root.dataset.motion).toBeUndefined()
    detach()
  })

  // jsdom no implementa la propiedad no estándar `zoom` y descarta su valor: se
  // comprueba la orden que el código le da al navegador y la variable propia
  // `--ui-zoom`, que sí se conserva. El efecto real se verifica en el E2E.
  it('scales the interface with zoom and exposes the real factor', async () => {
    const setProperty = vi.spyOn(root.style, 'setProperty')
    const { theme, appearance } = await load()
    const detach = theme.attachWorkspaceTheme()

    appearance.setTextScale('large')
    await nextTick()
    expect(setProperty).toHaveBeenCalledWith('zoom', '1.125')
    expect(prop('--ui-zoom')).toBe('1.125')

    appearance.setTextScale('normal')
    await nextTick()
    expect(prop('--ui-zoom')).toBe('')
    detach()
    setProperty.mockRestore()
  })

  it('clamps the zoom when the window is narrow and follows a resize', async () => {
    Object.defineProperty(window, 'innerWidth', { value: 360, configurable: true })
    const { theme, appearance } = await load()
    appearance.setTextScale('xlarge')
    const detach = theme.attachWorkspaceTheme()
    expect(Number(prop('--ui-zoom'))).toBeCloseTo(1.125, 5)

    Object.defineProperty(window, 'innerWidth', { value: 1440, configurable: true })
    window.dispatchEvent(new Event('resize'))
    await nextTick()
    expect(prop('--ui-zoom')).toBe('1.25')
    detach()
  })

  it('is idempotent and removes everything on detach, so public pages stay untouched', async () => {
    const { theme, appearance, brand } = await load()
    appearance.setTheme('ivory')
    appearance.setMotion('reduced')
    appearance.setTextScale('large')
    brand.setBrand({ ...brand.brandState.brand, accent: 'copper' })

    const detach = theme.attachWorkspaceTheme()
    expect(theme.attachWorkspaceTheme()).toBe(detach)
    expect(root.dataset.appTheme).toBe('ivory')

    detach()

    expect(root.dataset.appTheme).toBeUndefined()
    expect(root.dataset.motion).toBeUndefined()
    expect(prop('--color-brand-accent-surface')).toBe('')
    expect(prop('--color-brand-accent-text')).toBe('')
    expect(prop('--color-focus')).toBe('')
    expect(prop('zoom')).toBe('')
    expect(prop('--ui-zoom')).toBe('')
    expect(root.getAttribute('style') ?? '').toBe('')
  })

  it('can be attached again after a detach', async () => {
    const { theme } = await load()
    theme.attachWorkspaceTheme()()
    const detach = theme.attachWorkspaceTheme()
    expect(root.dataset.appTheme).toBe('ink')
    detach()
  })
})
