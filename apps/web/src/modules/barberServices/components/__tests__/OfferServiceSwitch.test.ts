/**
 * Interruptor «Lo ofrezco» (DEC-115): estado anunciado, ocupado mientras viaja la
 * petición, avisos por resultado y accesibilidad. El estado real vive en
 * `useBarberOffering` (probado aparte); aquí se usa un doble mínimo.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { axe } from 'vitest-axe'
import { resetToasts, toastState } from '@/shared/model/toastStore'
import type {
  BarberOffering,
  OfferingChangeOutcome,
  OfferingStatus,
} from '../../model/barberOffering'
import OfferServiceSwitch from '../OfferServiceSwitch.vue'

function fakeOffering(
  options: {
    status?: OfferingStatus
    offered?: string[]
    busy?: string[]
    outcome?: OfferingChangeOutcome
  } = {},
) {
  const setOffered = vi.fn(async () => options.outcome ?? 'changed')
  const offering = {
    status: ref<OfferingStatus>(options.status ?? 'ready'),
    isOffered: (id: string) => (options.offered ?? []).includes(id),
    isBusy: (id: string) => (options.busy ?? []).includes(id),
    load: vi.fn(),
    setOffered,
  } as unknown as BarberOffering
  return { offering, setOffered }
}

function mountSwitch(offering: BarberOffering, compact = false) {
  return mount(OfferServiceSwitch, {
    props: { offering, serviceId: 's-1', serviceName: 'Corte clásico', compact },
  })
}

describe('OfferServiceSwitch', () => {
  beforeEach(() => resetToasts())

  it('is a switch named after its service and announces whether it is offered', () => {
    const on = mountSwitch(fakeOffering({ offered: ['s-1'] }).offering)
    const sw = on.get('[role="switch"]')
    expect(sw.attributes('aria-label')).toBe('Lo ofrezco: Corte clásico')
    expect(sw.attributes('aria-checked')).toBe('true')
    expect(on.text()).toContain('Lo ofrezco')

    const off = mountSwitch(fakeOffering().offering)
    expect(off.get('[role="switch"]').attributes('aria-checked')).toBe('false')
    expect(off.text()).toContain('No lo ofrezco')
  })

  it('shows only the track in compact mode, keeping the accessible name', () => {
    const wrapper = mountSwitch(fakeOffering({ offered: ['s-1'] }).offering, true)

    expect(wrapper.find('.offer-switch__text').exists()).toBe(false)
    expect(wrapper.get('[role="switch"]').attributes('aria-label')).toBe(
      'Lo ofrezco: Corte clásico',
    )
  })

  it.each<OfferingStatus>(['idle', 'loading', 'unavailable', 'error'])(
    'renders nothing while the offering is %s',
    (status) => {
      expect(mountSwitch(fakeOffering({ status }).offering).find('button').exists()).toBe(false)
    },
  )

  it('asks to offer a service that is not offered and confirms it with a notice', async () => {
    const { offering, setOffered } = fakeOffering()
    const wrapper = mountSwitch(offering)

    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()

    expect(setOffered).toHaveBeenCalledWith('s-1', true)
    expect(toastState.items.map((t) => [t.variant, t.title])).toEqual([
      ['success', 'Ahora lo ofreces'],
    ])
    expect(toastState.items[0]!.detail).toContain('Corte clásico')
  })

  it('asks to stop offering a service that is offered', async () => {
    const { offering, setOffered } = fakeOffering({ offered: ['s-1'] })
    const wrapper = mountSwitch(offering)

    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()

    expect(setOffered).toHaveBeenCalledWith('s-1', false)
    expect(toastState.items.map((t) => t.title)).toEqual(['Ya no lo ofreces'])
  })

  it.each<[OfferingChangeOutcome, string]>([
    ['error', 'sigue como estaba'],
    ['network-error', 'Revisa tu conexión'],
  ])('warns without claiming a change when the result is %s', async (outcome, detail) => {
    const { offering } = fakeOffering({ outcome })
    const wrapper = mountSwitch(offering)

    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()

    expect(toastState.items.map((t) => [t.variant, t.title])).toEqual([
      ['danger', 'No pudimos guardar el cambio'],
    ])
    expect(toastState.items[0]!.detail).toContain(detail)
  })

  it('stays silent when nothing changed', async () => {
    const { offering } = fakeOffering({ outcome: 'unchanged' })
    const wrapper = mountSwitch(offering)

    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()

    expect(toastState.items).toEqual([])
  })

  it('is busy and disabled while the request is in flight', () => {
    const wrapper = mountSwitch(fakeOffering({ busy: ['s-1'] }).offering)
    const sw = wrapper.get('[role="switch"]')

    expect(sw.attributes('disabled')).toBeDefined()
    expect(sw.attributes('aria-busy')).toBe('true')
  })

  it('has no axe violations', async () => {
    const wrapper = mountSwitch(fakeOffering({ offered: ['s-1'] }).offering, true)

    expect(
      await axe(wrapper.element.outerHTML, {
        rules: { 'color-contrast': { enabled: false }, region: { enabled: false } },
      }),
    ).toHaveNoViolations()
  })
})
