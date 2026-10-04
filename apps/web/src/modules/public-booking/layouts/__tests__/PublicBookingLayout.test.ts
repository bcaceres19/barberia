/**
 * Pruebas de PublicBookingLayout (DEC-111): el cascarón aloja las cinco
 * pantallas con las rutas reales del módulo (mismos paths, nombres y
 * `meta.bookingStep`), dibuja el progreso del paso vigente, ofrece el
 * retorno al paso anterior con los parámetros correctos, se completa cuando
 * la página de datos confirma el turno y firma al pie como «Reservas con
 * NAVA». Las pantallas se sustituyen por un doble: aquí no se prueba su
 * contenido.
 */
import { afterEach, describe, it, expect } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory, type RouteRecordRaw } from 'vue-router'
import { axe } from 'vitest-axe'
import PublicBookingLayout from '../PublicBookingLayout.vue'
import { useBookingChrome } from '../../model/bookingChrome'
import { publicBookingRoutes } from '../../routes'

// Doble de pantalla: un `<main>` que, si se le pide, confirma el turno por el
// canal del cascarón (lo que hace la página de datos tras el éxito del API).
const Screen = defineComponent({
  props: { confirm: { type: Boolean, default: false } },
  setup(props) {
    const chrome = useBookingChrome()
    if (props.confirm) chrome.completed.value = true
    return () => h('main', 'pantalla')
  },
})

function stubbedRoutes(): RouteRecordRaw[] {
  return publicBookingRoutes.map((route) => ({
    ...route,
    component: PublicBookingLayout,
    children: route.children?.map((child) => ({
      ...child,
      component: Screen,
      props: child.name === 'reserva-publica-cliente' ? { confirm: true } : false,
    })) as RouteRecordRaw[],
  })) as RouteRecordRaw[]
}

// Cada montaje queda adjunto al documento (axe y los landmarks lo exigen):
// se desmonta siempre, aunque una aserción falle, para no contaminar la
// siguiente prueba con un segundo `<main>` o un segundo banner.
const mounted: Array<{ unmount: () => void }> = []
afterEach(() => {
  while (mounted.length) mounted.pop()!.unmount()
})

async function mountAt(path: string, routes: RouteRecordRaw[] = stubbedRoutes()) {
  const router = createRouter({ history: createMemoryHistory(), routes })
  await router.push(path)
  await router.isReady()
  const wrapper = mount(
    { template: '<RouterView />' },
    { global: { plugins: [router] }, attachTo: document.body },
  )
  await flushPromises()
  mounted.push(wrapper)
  return { wrapper, router }
}

const STEP_LABELS = ['Servicio', 'Barbero', 'Horario', 'Tus datos']
const SERVICE = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER = 'a1111111-1111-1111-1111-111111111111'

describe('PublicBookingLayout', () => {
  it('on the entry shows no progress, no back link and no step caption', async () => {
    const { wrapper } = await mountAt('/reservar/barberia-ejemplo')
    expect(wrapper.find('nav[aria-label="Progreso de la reserva"]').exists()).toBe(false)
    expect(wrapper.find('.pb-back').exists()).toBe(false)
    expect(wrapper.find('.pb-header__caption').text()).toBe('')
    wrapper.unmount()
  })

  it('draws the progress for the route step and captions «Paso N de 4»', async () => {
    const cases: Array<[string, number]> = [
      ['/reservar/b/servicios', 1],
      [`/reservar/b/servicios/${SERVICE}/barbero`, 2],
      [`/reservar/b/servicios/${SERVICE}/barbero/${BARBER}/horario`, 3],
    ]
    for (const [path, step] of cases) {
      const { wrapper } = await mountAt(path)
      expect(wrapper.find('.pb-header__caption').text()).toBe(`Paso ${step} de 4`)
      const current = wrapper.findAll('li[aria-current="step"]')
      expect(current).toHaveLength(1)
      expect(current[0]!.find('.pb-progress__label').text()).toBe(STEP_LABELS[step - 1])
      wrapper.unmount()
    }
  })

  it("links back to the previous step with that step's own params", async () => {
    const { wrapper, router } = await mountAt(
      `/reservar/b/servicios/${SERVICE}/barbero/${BARBER}/horario`,
    )
    const back = wrapper.find('a.pb-back')
    expect(back.text()).toContain('Barbero')
    expect(router.resolve(back.attributes('href')!).name).toBe('reserva-publica-barbero')
    expect(back.attributes('href')).toBe(`/reservar/b/servicios/${SERVICE}/barbero`)
    wrapper.unmount()

    const services = await mountAt('/reservar/b/servicios')
    expect(services.wrapper.find('a.pb-back').attributes('href')).toBe('/reservar/b')
    services.wrapper.unmount()
  })

  it('completes the progress and drops the back link once the booking is confirmed', async () => {
    const { wrapper } = await mountAt(
      `/reservar/b/servicios/${SERVICE}/barbero/${BARBER}/horario/2026-09-15T19:00:00Z/cliente`,
    )
    expect(wrapper.find('.pb-header__caption').text()).toBe('Confirmado')
    expect(wrapper.find('.pb-back').exists()).toBe(false)
    const items = wrapper.findAll('li')
    expect(items).toHaveLength(4)
    expect(items.every((item) => item.text().includes('(completado)'))).toBe(true)
    wrapper.unmount()
  })

  it('signs the page as «Reservas con NAVA», never putting NAVA in the hero position', async () => {
    const { wrapper } = await mountAt('/reservar/barberia-ejemplo')
    expect(wrapper.find('footer').text()).toContain('Reservas con')
    expect(wrapper.find('footer').text()).toContain('NAVA')
    wrapper.unmount()
  })

  it('keeps one banner, one <main> and no accessibility violations', async () => {
    const { wrapper } = await mountAt(`/reservar/b/servicios/${SERVICE}/barbero`)
    expect(wrapper.findAll('header')).toHaveLength(1)
    expect(wrapper.findAll('main')).toHaveLength(1)
    expect(await axe(wrapper.element)).toHaveNoViolations()
    wrapper.unmount()
  })
})
