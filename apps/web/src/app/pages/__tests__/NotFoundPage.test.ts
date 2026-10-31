// Issue #425: una ruta desconocida muestra «No encontramos esta página» en vez de una pantalla en blanco.
import { describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { router } from '../../router'
import NotFoundPage from '../NotFoundPage.vue'

describe('ruta no encontrada (#425)', () => {
  it.each(['/nada-total', '/panel/nada', '/panel/servicios/7f1c0f0e-0000-4000-8000-000000000000'])(
    '%s resuelve a la pantalla de no encontrada',
    (path) => {
      expect(router.resolve(path).name).toBe('no-encontrada')
    },
  )

  it('no desvía las rutas reales ni cambia la sesión que exigen', () => {
    expect(router.resolve('/panel/servicios').name).not.toBe('no-encontrada')
    expect(router.resolve('/acceso').name).toBe('acceso')
    const privateMatch = router.resolve('/panel/servicios').matched
    expect(privateMatch.some((record) => record.beforeEnter)).toBe(true)
    expect(router.resolve('/panel/nada').matched.some((record) => record.beforeEnter)).toBe(false)
  })

  it('nombra la pantalla y ofrece volver al inicio', async () => {
    await router.push('/nada-total')
    await router.isReady()
    const wrapper = mount(NotFoundPage, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Página no encontrada')
    expect(wrapper.get('h2').text()).toBe('No encontramos esta página')
    const action = wrapper.get('a')
    expect(action.text()).toBe('Ir al inicio')
    expect(action.attributes('href')).toBe('/')
  })
})
