import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import { installChunkLoadRecovery, isChunkLoadError } from '../chunkLoadRecovery'

const chunkError = () =>
  new TypeError('Failed to fetch dynamically imported module: http://localhost/Page.vue')

describe('isChunkLoadError', () => {
  it.each([
    new TypeError('Failed to fetch dynamically imported module: x'),
    new TypeError('error loading dynamically imported module: x'),
    new TypeError('Importing a module script failed.'),
    new Error('Unable to preload CSS for /assets/x.css'),
    Object.assign(new Error('x'), { name: 'ChunkLoadError' }),
  ])('recognises %s', (error) => {
    expect(isChunkLoadError(error)).toBe(true)
  })

  it('ignores any other navigation error', () => {
    expect(isChunkLoadError(new Error('boom'))).toBe(false)
    expect(isChunkLoadError('Failed to fetch dynamically imported module')).toBe(false)
  })
})

// Con un fallo de descarga `push` rechaza además de avisar; la pantalla real lo absorbe.
const go = (router: Router, path: string) => router.push(path).catch(() => undefined)

describe('installChunkLoadRecovery (CA-090-05, PENTRY-F-01)', () => {
  let router: Router
  let online: boolean
  let failLoad: boolean
  const navigate = vi.fn()

  beforeEach(async () => {
    resetToasts()
    navigate.mockReset()
    online = true
    failLoad = false
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: { template: '<i />' } },
        {
          path: '/servicios',
          component: () =>
            failLoad ? Promise.reject(chunkError()) : Promise.resolve({ template: '<i />' }),
        },
        { path: '/otra', component: { template: '<i />' } },
      ],
    })
    installChunkLoadRecovery(router, { navigate, isOnline: () => online })
    await router.push('/')
  })

  it('warns with a retry action when the screen cannot be downloaded, and stays where it was', async () => {
    failLoad = true
    online = false
    await go(router, '/servicios')

    expect(router.currentRoute.value.fullPath).toBe('/')
    expect(toastState.items).toHaveLength(1)
    expect(toastState.items[0]).toMatchObject({
      variant: 'danger',
      title: 'No pudimos abrir esta pantalla',
      detail: 'Sin conexión. Revisa tu red e inténtalo de nuevo.',
    })
    expect(toastState.items[0]!.action?.label).toBe('Reintentar')
  })

  it('«Reintentar» with the network back reloads the address: the browser remembers the failed module', async () => {
    failLoad = true
    online = false
    await go(router, '/servicios')

    online = true
    toastState.items[0]!.action!.run()

    expect(navigate).toHaveBeenCalledWith('/servicios')
  })

  it('«Reintentar» still offline warns again instead of leaving the page', async () => {
    failLoad = true
    online = false
    await go(router, '/servicios')
    const retry = toastState.items[0]!.action!
    resetToasts()

    retry.run()

    expect(navigate).not.toHaveBeenCalled()
    expect(toastState.items.map((t) => t.title)).toEqual(['No pudimos abrir esta pantalla'])
  })

  it('pressing the same destination again counts as retry', async () => {
    failLoad = true
    await go(router, '/servicios')
    expect(navigate).not.toHaveBeenCalled()

    await go(router, '/servicios')

    expect(navigate).toHaveBeenCalledWith('/servicios')
    expect(router.currentRoute.value.fullPath).toBe('/')
  })

  it('offline, pressing again warns again and does not leave the page', async () => {
    failLoad = true
    online = false
    await go(router, '/servicios')
    await go(router, '/servicios')

    expect(navigate).not.toHaveBeenCalled()
    expect(toastState.items.length).toBeGreaterThanOrEqual(1)
  })

  it('forgets the failure when the person goes somewhere else', async () => {
    failLoad = true
    await go(router, '/servicios')
    await go(router, '/otra')
    failLoad = false

    await go(router, '/servicios')

    expect(navigate).not.toHaveBeenCalled()
    expect(router.currentRoute.value.fullPath).toBe('/servicios')
  })

  it('does not react to a navigation that works', async () => {
    await go(router, '/servicios')

    expect(toastState.items).toHaveLength(0)
    expect(navigate).not.toHaveBeenCalled()
  })
})
