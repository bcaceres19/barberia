/**
 * Tests for AgendaDatePicker: calendario propio del selector de fecha de
 * /panel. Fechas civiles AAAA-MM-DD, semana desde el lunes, teclado APG y
 * accesibilidad del disparador y del diálogo.
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import AgendaDatePicker from '@/shared/ui/BaseDatePicker.vue'

const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountPicker(props: Record<string, unknown> = {}) {
  return mount(AgendaDatePicker, {
    attachTo: document.body,
    props: {
      modelValue: '2026-09-26',
      today: '2026-09-26',
      triggerId: 'daily-agenda-date-picker',
      ...props,
    },
    global: { stubs: { teleport: true } },
  })
}

async function openPicker(wrapper: ReturnType<typeof mountPicker>) {
  await wrapper.get('.agenda-date-picker__trigger').trigger('click')
}

describe('AgendaDatePicker', () => {
  it('muestra la fecha ISO en el disparador y no abre el calendario hasta pulsarlo', () => {
    const wrapper = mountPicker()

    expect(wrapper.get('.agenda-date-picker__value').text()).toBe('2026-09-26')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(wrapper.get('.agenda-date-picker__trigger').attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })

  it('abre en el mes de la fecha elegida con seis semanas que empiezan en lunes', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)

    expect(wrapper.get('.agenda-date-picker__title').text()).toBe('Septiembre de 2026')
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(6)
    // 1 de septiembre de 2026 es martes: la primera celda es el lunes 31 de agosto.
    expect(rows[0]!.findAll('button')[0]!.attributes('data-date')).toBe('2026-08-31')
    expect(wrapper.get('[data-date="2026-09-26"]').attributes('aria-current')).toBe('date')
    expect(wrapper.get('[data-date="2026-09-26"]').classes()).toContain(
      'agenda-date-picker__day--selected',
    )
    wrapper.unmount()
  })

  it('emite la fecha civil elegida y cierra el calendario', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)

    await wrapper.get('[data-date="2026-09-30"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([['2026-09-30']])
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('navega entre meses con los botones sin cambiar la fecha elegida', async () => {
    const wrapper = mountPicker({ modelValue: '2026-01-31' })
    await openPicker(wrapper)

    await wrapper.get('button[aria-label="Mes siguiente"]').trigger('click')
    // Febrero no tiene 31: se ajusta al último día, sin desbordar a marzo.
    expect(wrapper.get('.agenda-date-picker__title').text()).toBe('Febrero de 2026')
    expect(wrapper.find('[data-date="2026-02-28"][tabindex="0"]').exists()).toBe(true)

    await wrapper.get('button[aria-label="Mes anterior"]').trigger('click')
    expect(wrapper.get('.agenda-date-picker__title').text()).toBe('Enero de 2026')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('«Hoy» elige el día vigente de la barbería y solo aparece si se conoce', async () => {
    const wrapper = mountPicker({ modelValue: '2026-03-01', today: '2026-09-26' })
    await openPicker(wrapper)

    await wrapper.get('.agenda-date-picker__today').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['2026-09-26']])
    wrapper.unmount()

    const withoutToday = mountPicker({ today: null })
    await openPicker(withoutToday)
    expect(withoutToday.find('.agenda-date-picker__today').exists()).toBe(false)
    withoutToday.unmount()
  })

  it('el teclado mueve el día enfocado: flechas ±1/±7, Inicio/Fin de semana y RePág/AvPág ±mes', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)
    const grid = wrapper.get('tbody')
    const focusedDate = () => wrapper.get('[tabindex="0"][data-date]').attributes('data-date')

    await grid.trigger('keydown', { key: 'ArrowRight' })
    expect(focusedDate()).toBe('2026-09-27')
    await grid.trigger('keydown', { key: 'ArrowDown' })
    expect(focusedDate()).toBe('2026-10-04')
    await grid.trigger('keydown', { key: 'Home' })
    expect(focusedDate()).toBe('2026-09-28')
    await grid.trigger('keydown', { key: 'End' })
    expect(focusedDate()).toBe('2026-10-04')
    await grid.trigger('keydown', { key: 'PageUp' })
    expect(focusedDate()).toBe('2026-09-04')
    await grid.trigger('keydown', { key: 'PageDown', shiftKey: true })
    expect(focusedDate()).toBe('2027-09-04')
    wrapper.unmount()
  })

  it('Esc cierra el calendario y devuelve el foco al disparador', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)

    await wrapper.get('[role="dialog"]').trigger('keydown', { key: 'Escape' })

    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(document.activeElement).toBe(wrapper.get('.agenda-date-picker__trigger').element)
    wrapper.unmount()
  })

  it('la rueda sobre el calendario lo acerca y lo aleja dentro de los límites, sin tocar la fecha', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)
    const dialog = wrapper.get('[role="dialog"]')
    const zoomOf = () => Number((dialog.element as HTMLElement).style.getPropertyValue('--dp-zoom'))

    await dialog.trigger('wheel', { deltaY: -100 })
    expect(zoomOf()).toBeGreaterThan(1)
    for (let i = 0; i < 20; i++) await dialog.trigger('wheel', { deltaY: -100 })
    expect(zoomOf()).toBe(1.6)
    for (let i = 0; i < 40; i++) await dialog.trigger('wheel', { deltaY: 100 })
    expect(zoomOf()).toBe(0.85)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('el teclado hace zoom con + y - y lo restablece con 0', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)
    const dialog = wrapper.get('[role="dialog"]')
    const zoomOf = () => Number((dialog.element as HTMLElement).style.getPropertyValue('--dp-zoom'))

    await dialog.trigger('keydown', { key: '+' })
    expect(zoomOf()).toBeCloseTo(1.1, 3)
    await dialog.trigger('keydown', { key: '-' })
    await dialog.trigger('keydown', { key: '-' })
    expect(zoomOf()).toBeLessThan(1)
    await dialog.trigger('keydown', { key: '0' })
    expect(zoomOf()).toBe(1)
    wrapper.unmount()
  })

  it('el pellizco de dos dedos escala el calendario y el clic del mismo gesto no elige fecha', async () => {
    const wrapper = mountPicker()
    await openPicker(wrapper)
    const dialog = wrapper.get('[role="dialog"]')
    const zoomOf = () => Number((dialog.element as HTMLElement).style.getPropertyValue('--dp-zoom'))
    const touch = (pointerId: number, x: number) => ({
      pointerId,
      pointerType: 'touch',
      clientX: x,
      clientY: 100,
    })

    await dialog.trigger('pointerdown', touch(1, 100))
    await dialog.trigger('pointerdown', touch(2, 200))
    await dialog.trigger('pointermove', touch(2, 300))
    expect(zoomOf()).toBeCloseTo(1.6, 1)
    await dialog.trigger('pointermove', touch(2, 150))
    expect(zoomOf()).toBeLessThan(1)

    await dialog.trigger('pointerup', touch(2, 150))
    await dialog.trigger('pointerup', touch(1, 100))
    await wrapper.get('[data-date="2026-09-30"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('deshabilitado no abre el calendario', async () => {
    const wrapper = mountPicker({ disabled: true })

    await openPicker(wrapper)

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('el disparador se nombra con su rótulo y el valor, y no tiene violaciones axe abierto', async () => {
    const wrapper = mount(
      {
        components: { AgendaDatePicker },
        template: `<div><label id="lbl" for="daily-agenda-date-picker">Fecha</label><AgendaDatePicker model-value="2026-09-26" today="2026-09-26" trigger-id="daily-agenda-date-picker" label-id="lbl" /></div>`,
      },
      { attachTo: document.body, global: { stubs: { teleport: true } } },
    )
    const trigger = wrapper.get('#daily-agenda-date-picker')
    expect(trigger.attributes('aria-labelledby')).toBe('lbl daily-agenda-date-picker')

    await trigger.trigger('click')
    expect(await axe(wrapper.element as HTMLElement, axeOptions)).toHaveNoViolations()
    wrapper.unmount()
  })
})
