/**
 * LocationMap (DEC-130, HU-102): iframe con título accesible, dirección en texto
 * y enlace para abrir Google Maps. En los tests `import.meta.env.DEV` es true,
 * así que se usa el incrustado sin clave.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import LocationMap from '../LocationMap.vue'

// jsdom no resuelve el contenido del iframe: axe revisa el resto (título, enlace).
const axeOptions = {
  iframes: false,
  rules: { region: { enabled: false }, 'color-contrast': { enabled: false } },
}
const props = { latitude: 4.598056, longitude: -74.076111, address: 'Calle 10 # 5-20, Bogotá' }

describe('LocationMap', () => {
  it('muestra el mapa centrado en las coordenadas y la dirección en texto (CA-102-04)', () => {
    const wrapper = mount(LocationMap, { props })
    const frame = wrapper.get('iframe')
    expect(frame.attributes('src')).toContain('4.598056')
    expect(frame.attributes('src')).toContain('-74.076111')
    expect(wrapper.get('.location-map__address').text()).toBe('Calle 10 # 5-20, Bogotá')
  })

  it('el iframe tiene un título que nombra la dirección y carga perezosa', () => {
    const frame = mount(LocationMap, { props }).get('iframe')
    expect(frame.attributes('title')).toBe('Mapa de Calle 10 # 5-20, Bogotá')
    expect(frame.attributes('loading')).toBe('lazy')
  })

  it('ofrece un enlace seguro para abrir Google Maps en otra pestaña', () => {
    const link = mount(LocationMap, { props }).get('a')
    expect(link.attributes('href')).toContain('https://www.google.com/maps/search/')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toContain('noopener')
    expect(link.text()).toContain('se abre en una pestaña nueva')
  })

  it('con un punto inválido no pinta nada (nunca un mapa sin coordenadas reales)', () => {
    const wrapper = mount(LocationMap, { props: { ...props, latitude: 120 } })
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.html()).not.toContain('Calle 10')
  })

  it('no tiene violaciones de accesibilidad', async () => {
    const wrapper = mount(LocationMap, { props, attachTo: document.body })
    expect(await axe(wrapper.element as HTMLElement, axeOptions)).toHaveNoViolations()
    wrapper.unmount()
  })
})
